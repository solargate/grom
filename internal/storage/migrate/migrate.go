package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/data"
	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/federation"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage"
	storebbolt "github.com/solargate/grom/internal/storage/bbolt"
	"github.com/solargate/grom/internal/storage/file"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
	"gopkg.in/yaml.v3"
)

type Options struct {
	From   config.StorageDriver
	To     config.StorageDriver
	Config config.StorageConfig
	// FederationDomain is used to reconstruct legacy like activity ids for
	// local workouts (https://{domain}/users/{nick}/workouts/{id}). Empty
	// skips that reconstruction path.
	FederationDomain string
	DryRun           bool
	Verify           bool
	Force            bool
}

type Result struct {
	Users                int
	Profiles             int
	Equipment            int
	Follows              int
	Workouts             int
	FedFollowers         int
	FedAuthors           int
	FedInboxWorkouts     int
	LocalLikes           int
	FedLikes             int
	LikeActivities       int
	LocalComments        int
	FedComments          int
	CommentActivities    int
	LocalSpeedCharts     int
	LocalHeartRateCharts int
	FedSpeedCharts       int
	FedHeartRateCharts   int
	PersonalAccessTokens int
}

// Run copies metadata from one storage driver to another. Track/media/avatar
// blob files under storage.location are shared and not copied. Speed and
// heart-rate charts are converted between file JSON blobs and bbolt binary
// buckets. Workout likes and comments (local, federated cache, and outbound
// activity ids) and personal access tokens are copied so they remain readable
// after switching drivers. Password-reset tokens are not copied.
func Run(opts Options) (*Result, error) {
	if opts.From == opts.To {
		return nil, fmt.Errorf("from and to drivers must differ")
	}
	if opts.From != config.StorageDriverFile && opts.From != config.StorageDriverBBolt {
		return nil, fmt.Errorf("unsupported from driver %q", opts.From)
	}
	if opts.To != config.StorageDriverFile && opts.To != config.StorageDriverBBolt {
		return nil, fmt.Errorf("unsupported to driver %q", opts.To)
	}
	cfg := opts.Config
	if cfg.ResolvedLocation == "" {
		return nil, fmt.Errorf("storage location is not resolved")
	}

	dbPath := cfg.ResolvedBBoltPath
	if dbPath == "" {
		dbPath = filepath.Join(cfg.ResolvedLocation, "grom.db")
	}

	if opts.DryRun {
		src, err := openDriver(opts.From, cfg)
		if err != nil {
			return nil, err
		}
		defer src.Close()
		return countAll(src, cfg.ResolvedLocation, opts.FederationDomain)
	}

	if opts.To == config.StorageDriverBBolt {
		if !opts.Force {
			if info, err := os.Stat(dbPath); err == nil && info.Size() > 0 {
				return nil, fmt.Errorf("bbolt database %q already exists (use --force to overwrite)", dbPath)
			}
		} else {
			_ = os.Remove(dbPath)
		}
	}

	src, err := openDriver(opts.From, cfg)
	if err != nil {
		return nil, fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	dst, err := openDriver(opts.To, cfg)
	if err != nil {
		return nil, fmt.Errorf("open target: %w", err)
	}

	result, err := copyAll(src, dst, cfg.ResolvedLocation, opts.FederationDomain)
	closeErr := dst.Close()
	if err != nil {
		return result, err
	}
	if closeErr != nil {
		return result, closeErr
	}

	if opts.Verify {
		srcCount, err := countAll(src, cfg.ResolvedLocation, opts.FederationDomain)
		if err != nil {
			return result, fmt.Errorf("verify source count: %w", err)
		}
		dst2, err := openDriver(opts.To, cfg)
		if err != nil {
			return result, fmt.Errorf("reopen target for verify: %w", err)
		}
		defer dst2.Close()
		dstCount, err := countAll(dst2, cfg.ResolvedLocation, opts.FederationDomain)
		if err != nil {
			return result, fmt.Errorf("verify target count: %w", err)
		}
		if *srcCount != *dstCount {
			return result, fmt.Errorf("verify failed: source=%+v target=%+v", *srcCount, *dstCount)
		}
	}

	return result, nil
}

func openDriver(driver config.StorageDriver, cfg config.StorageConfig) (storage.Backend, error) {
	c := cfg
	c.Driver = driver
	return storage.Open(c)
}

func copyAll(src, dst storage.Backend, location, federationDomain string) (*Result, error) {
	result := &Result{}

	usersList, err := src.Users().ListAll()
	if err != nil {
		return result, fmt.Errorf("list users: %w", err)
	}
	for _, u := range usersList {
		if err := importUser(dst, u); err != nil {
			return result, fmt.Errorf("import user %s: %w", u.Nickname, err)
		}
		result.Users++

		profile, err := src.Users().GetProfile(u.ID)
		if err != nil {
			return result, fmt.Errorf("get profile %s: %w", u.Nickname, err)
		}
		if profile != nil && profileNonEmpty(profile) {
			if err := dst.Users().PutProfile(u.ID, *profile); err != nil {
				return result, fmt.Errorf("import profile %s: %w", u.Nickname, err)
			}
			result.Profiles++
		}

		eq, err := src.Equipment().List(u.Nickname)
		if err != nil {
			return result, fmt.Errorf("list equipment %s: %w", u.Nickname, err)
		}
		for _, item := range eq {
			if err := importEquipment(dst, u.Nickname, item); err != nil {
				return result, fmt.Errorf("import equipment: %w", err)
			}
			result.Equipment++
		}

		ws, err := src.Workouts().List(u.Nickname)
		if err != nil {
			return result, fmt.Errorf("list workouts %s: %w", u.Nickname, err)
		}
		for i := range ws {
			w := ws[i]
			if err := importWorkout(dst, u.Nickname, &w); err != nil {
				return result, fmt.Errorf("import workout %s: %w", w.ID, err)
			}
			speed, hr, err := copyLocalCharts(src, dst, u.Nickname, &w)
			if err != nil {
				return result, fmt.Errorf("copy charts for workout %s: %w", w.ID, err)
			}
			result.LocalSpeedCharts += speed
			result.LocalHeartRateCharts += hr
			if err := copyLocalLikes(src, dst, u.Nickname, &w); err != nil {
				return result, fmt.Errorf("copy likes for workout %s: %w", w.ID, err)
			}
			likes, err := src.Likes().GetLocal(u.Nickname, w.ID)
			if err != nil {
				return result, fmt.Errorf("count likes for workout %s: %w", w.ID, err)
			}
			if likes != nil && likes.Likes > 0 {
				result.LocalLikes++
			}
			if err := copyLocalComments(src, dst, u.Nickname, &w); err != nil {
				return result, fmt.Errorf("copy comments for workout %s: %w", w.ID, err)
			}
			comments, err := src.Comments().GetLocal(u.Nickname, w.ID)
			if err != nil {
				return result, fmt.Errorf("count comments for workout %s: %w", w.ID, err)
			}
			if comments != nil && comments.CommentsNum > 0 {
				result.LocalComments++
			}
			result.Workouts++
		}

		followers, err := src.Federation().Followers().List(u.Nickname)
		if err != nil {
			return result, fmt.Errorf("list followers %s: %w", u.Nickname, err)
		}
		if len(followers) > 0 {
			if err := importFollowers(dst, u.Nickname, followers); err != nil {
				return result, fmt.Errorf("import followers: %w", err)
			}
			result.FedFollowers += len(followers)
		}
	}

	follows, err := listFollows(src)
	if err != nil {
		return result, err
	}
	for _, f := range follows {
		if err := importFollow(dst, f); err != nil {
			return result, fmt.Errorf("import follow: %w", err)
		}
		result.Follows++
	}

	authors, inbox, err := loadFederationInbox(src, location)
	if err != nil {
		return result, err
	}
	for viewer, byOwner := range authors {
		for ownerKey, meta := range byOwner {
			if err := importAuthor(dst, viewer, ownerKey, meta); err != nil {
				return result, fmt.Errorf("import author: %w", err)
			}
			result.FedAuthors++
		}
	}
	for viewer, byOwner := range inbox {
		for ownerKey, list := range byOwner {
			for i := range list {
				w := list[i]
				if err := importInboxWorkout(dst, viewer, ownerKey, &w); err != nil {
					return result, fmt.Errorf("import inbox workout: %w", err)
				}
				speed, hr, err := copyFederatedCharts(src, dst, viewer, ownerKey, &w)
				if err != nil {
					return result, fmt.Errorf("copy federated charts for workout %s: %w", w.ID, err)
				}
				result.FedSpeedCharts += speed
				result.FedHeartRateCharts += hr
				result.FedInboxWorkouts++
			}
		}
	}

	fedLikes, err := copyFederatedLikes(src, dst)
	if err != nil {
		return result, fmt.Errorf("copy federated likes: %w", err)
	}
	result.FedLikes = fedLikes

	likeActivities, err := copyLikeActivities(src, dst, authors, inbox, federationDomain)
	if err != nil {
		return result, fmt.Errorf("copy like activities: %w", err)
	}
	result.LikeActivities = likeActivities

	fedComments, err := copyFederatedComments(src, dst)
	if err != nil {
		return result, fmt.Errorf("copy federated comments: %w", err)
	}
	result.FedComments = fedComments

	commentActivities, err := copyCommentActivities(src, dst)
	if err != nil {
		return result, fmt.Errorf("copy comment activities: %w", err)
	}
	result.CommentActivities = commentActivities

	pats, err := copyPersonalAccessTokens(src, dst)
	if err != nil {
		return result, fmt.Errorf("copy personal access tokens: %w", err)
	}
	result.PersonalAccessTokens = pats

	return result, nil
}

func countAll(backend storage.Backend, location, federationDomain string) (*Result, error) {
	result := &Result{}
	usersList, err := backend.Users().ListAll()
	if err != nil {
		return nil, err
	}
	result.Users = len(usersList)
	for _, u := range usersList {
		profile, err := backend.Users().GetProfile(u.ID)
		if err != nil {
			return nil, err
		}
		if profile != nil && profileNonEmpty(profile) {
			result.Profiles++
		}
		eq, err := backend.Equipment().List(u.Nickname)
		if err != nil {
			return nil, err
		}
		result.Equipment += len(eq)
		ws, err := backend.Workouts().List(u.Nickname)
		if err != nil {
			return nil, err
		}
		result.Workouts += len(ws)
		followers, err := backend.Federation().Followers().List(u.Nickname)
		if err != nil {
			return nil, err
		}
		result.FedFollowers += len(followers)
	}
	follows, err := listFollows(backend)
	if err != nil {
		return nil, err
	}
	result.Follows = len(follows)
	authors, inbox, err := loadFederationInbox(backend, location)
	if err != nil {
		return nil, err
	}
	for _, byOwner := range authors {
		result.FedAuthors += len(byOwner)
	}
	for _, byOwner := range inbox {
		for _, list := range byOwner {
			result.FedInboxWorkouts += len(list)
		}
	}
	localLikes, err := countLocalLikes(backend)
	if err != nil {
		return nil, err
	}
	result.LocalLikes = localLikes
	fedLikes, err := countFederatedLikes(backend)
	if err != nil {
		return nil, err
	}
	result.FedLikes = fedLikes
	likeActivities, err := countLikeActivities(backend, authors, inbox, federationDomain)
	if err != nil {
		return nil, err
	}
	result.LikeActivities = likeActivities
	localComments, err := countLocalComments(backend)
	if err != nil {
		return nil, err
	}
	result.LocalComments = localComments
	fedComments, err := countFederatedComments(backend)
	if err != nil {
		return nil, err
	}
	result.FedComments = fedComments
	commentActivities, err := countCommentActivities(backend)
	if err != nil {
		return nil, err
	}
	result.CommentActivities = commentActivities
	if err := countCharts(backend, location, result); err != nil {
		return nil, err
	}
	pats, err := countPersonalAccessTokens(backend)
	if err != nil {
		return nil, err
	}
	result.PersonalAccessTokens = pats
	return result, nil
}

func profileNonEmpty(profile *users.Profile) bool {
	return profile != nil && (profile.LastSportType != "" || len(profile.LastEquipmentBySport) > 0 || len(profile.UsedSportTypes) > 0)
}

func listFollows(backend storage.Backend) ([]social.Follow, error) {
	switch b := backend.(type) {
	case *file.Backend:
		return b.Social().(*file.SocialStore).ListAll()
	case *storebbolt.Backend:
		return b.Social().(*storebbolt.SocialStore).ListAll()
	default:
		return nil, fmt.Errorf("unsupported backend type %T", backend)
	}
}

func importUser(dst storage.Backend, u users.User) error {
	switch b := dst.(type) {
	case *file.Backend:
		return b.Users().(*file.UsersStore).Import(u)
	case *storebbolt.Backend:
		return b.Users().(*storebbolt.UsersStore).PutExisting(u)
	default:
		return fmt.Errorf("unsupported backend type %T", dst)
	}
}

func importEquipment(dst storage.Backend, nickname string, item equipment.Equipment) error {
	switch b := dst.(type) {
	case *file.Backend:
		return b.Equipment().(*file.EquipmentStore).Import(nickname, item)
	case *storebbolt.Backend:
		return b.Equipment().(*storebbolt.EquipmentStore).PutExisting(nickname, item)
	default:
		return fmt.Errorf("unsupported backend type %T", dst)
	}
}

func importWorkout(dst storage.Backend, nickname string, w *workouts.Workout) error {
	switch b := dst.(type) {
	case *file.Backend:
		return b.WorkoutsRepo().Import(nickname, w)
	case *storebbolt.Backend:
		return b.WorkoutsRepo().PutExisting(nickname, w)
	default:
		return fmt.Errorf("unsupported backend type %T", dst)
	}
}

func importFollow(dst storage.Backend, f social.Follow) error {
	switch b := dst.(type) {
	case *file.Backend:
		return b.Social().(*file.SocialStore).Import(f)
	case *storebbolt.Backend:
		return b.Social().(*storebbolt.SocialStore).PutExisting(f)
	default:
		return fmt.Errorf("unsupported backend type %T", dst)
	}
}

func importFollowers(dst storage.Backend, nickname string, followers []federation.InboundFollower) error {
	switch b := dst.(type) {
	case *file.Backend:
		return b.Federation().Followers().(*file.FederationFollowersStore).Import(nickname, followers)
	case *storebbolt.Backend:
		store := b.Federation().Followers().(*storebbolt.FederationFollowersStore)
		for _, f := range followers {
			if err := store.Add(nickname, f); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported backend type %T", dst)
	}
}

func importAuthor(dst storage.Backend, viewer, ownerKey string, meta federation.AuthorMeta) error {
	switch b := dst.(type) {
	case *file.Backend:
		dir := filepath.Join(data.UserDir(b.Location(), viewer), "federation", "inbox", "workouts", ownerKey)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
		raw, err := yaml.Marshal(meta)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "author.yaml"), raw, 0600)
	case *storebbolt.Backend:
		return b.Federation().Inbox().(*storebbolt.InboxStore).PutExistingAuthor(viewer, ownerKey, meta)
	default:
		return fmt.Errorf("unsupported backend type %T", dst)
	}
}

func importInboxWorkout(dst storage.Backend, viewer, ownerKey string, w *workouts.Workout) error {
	switch b := dst.(type) {
	case *file.Backend:
		dir := filepath.Join(data.UserDir(b.Location(), viewer), "federation", "inbox", "workouts", ownerKey)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
		raw, err := yaml.Marshal(w)
		if err != nil {
			return err
		}
		tmp := filepath.Join(dir, w.ID+".yaml.tmp")
		path := filepath.Join(dir, w.ID+".yaml")
		if err := os.WriteFile(tmp, raw, 0600); err != nil {
			return err
		}
		return os.Rename(tmp, path)
	case *storebbolt.Backend:
		return b.Federation().Inbox().(*storebbolt.InboxStore).PutExistingWorkout(viewer, ownerKey, w)
	default:
		return fmt.Errorf("unsupported backend type %T", dst)
	}
}

func loadFederationInbox(backend storage.Backend, location string) (
	authors map[string]map[string]federation.AuthorMeta,
	inbox map[string]map[string][]workouts.Workout,
	err error,
) {
	switch b := backend.(type) {
	case *file.Backend:
		return loadFileFederationInbox(b.Location())
	case *storebbolt.Backend:
		inboxStore := b.Federation().Inbox().(*storebbolt.InboxStore)
		authors, err = inboxStore.ListAllAuthors()
		if err != nil {
			return nil, nil, err
		}
		inbox, err = inboxStore.ListAllWorkouts()
		return authors, inbox, err
	default:
		_ = location
		return nil, nil, fmt.Errorf("unsupported backend type %T", backend)
	}
}

func loadFileFederationInbox(location string) (
	map[string]map[string]federation.AuthorMeta,
	map[string]map[string][]workouts.Workout,
	error,
) {
	authors := make(map[string]map[string]federation.AuthorMeta)
	inbox := make(map[string]map[string][]workouts.Workout)

	usersRoot := filepath.Join(location, data.UsersSubdir)
	userEntries, err := os.ReadDir(usersRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return authors, inbox, nil
		}
		return nil, nil, err
	}

	for _, userEntry := range userEntries {
		if !userEntry.IsDir() {
			continue
		}
		viewer := userEntry.Name()
		root := filepath.Join(usersRoot, viewer, "federation", "inbox", "workouts")
		ownerEntries, err := os.ReadDir(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, err
		}
		for _, ownerEntry := range ownerEntries {
			if !ownerEntry.IsDir() {
				continue
			}
			ownerKey := ownerEntry.Name()
			ownerDir := filepath.Join(root, ownerKey)

			metaPath := filepath.Join(ownerDir, "author.yaml")
			if raw, err := os.ReadFile(metaPath); err == nil {
				var meta federation.AuthorMeta
				if err := yaml.Unmarshal(raw, &meta); err != nil {
					return nil, nil, fmt.Errorf("parse author meta %s: %w", metaPath, err)
				}
				if authors[viewer] == nil {
					authors[viewer] = make(map[string]federation.AuthorMeta)
				}
				authors[viewer][ownerKey] = meta
			} else if !os.IsNotExist(err) {
				return nil, nil, err
			}

			files, err := os.ReadDir(ownerDir)
			if err != nil {
				return nil, nil, err
			}
			for _, f := range files {
				name := f.Name()
				if f.IsDir() || !strings.HasSuffix(name, ".yaml") || name == "author.yaml" {
					continue
				}
				raw, err := os.ReadFile(filepath.Join(ownerDir, name))
				if err != nil {
					return nil, nil, err
				}
				var w workouts.Workout
				if err := yaml.Unmarshal(raw, &w); err != nil {
					return nil, nil, fmt.Errorf("parse inbox workout %s: %w", name, err)
				}
				if inbox[viewer] == nil {
					inbox[viewer] = make(map[string][]workouts.Workout)
				}
				inbox[viewer][ownerKey] = append(inbox[viewer][ownerKey], w)
			}
		}
	}
	return authors, inbox, nil
}

