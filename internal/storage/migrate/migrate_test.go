package migrate_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/auth/pat"
	"github.com/solargate/grom/internal/auth/reset"
	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/equipment"
	"github.com/solargate/grom/internal/social"
	"github.com/solargate/grom/internal/storage"
	"github.com/solargate/grom/internal/storage/keys"
	"github.com/solargate/grom/internal/storage/migrate"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

func TestMigrateFileToBBoltAndBack(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}

	fileBackend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}

	user, err := fileBackend.Users().Create("alice", "Alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	bike := &equipment.Equipment{Type: equipment.TypeBike, Name: "Road", BikeType: "road"}
	if _, err := fileBackend.Equipment().Create("alice", bike); err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.Workouts().Create("alice", &workouts.Workout{
		Name: "Run", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.Social().Create(social.Follow{
		FollowerID:     user.ID,
		TargetHandle:   "bob@local",
		TargetNickname: "bob",
		TargetIsLocal:  true,
		Status:         social.StatusActive,
		CreatedAt:      time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	_ = fileBackend.Close()

	result, err := migrate.Run(migrate.Options{
		From:   config.StorageDriverFile,
		To:     config.StorageDriverBBolt,
		Config: cfg,
		Verify: true,
	})
	if err != nil {
		t.Fatalf("file→bbolt: %v", err)
	}
	if result.Users != 1 || result.Workouts != 1 || result.Equipment != 1 || result.Follows != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}

	bboltCfg := cfg
	bboltCfg.Driver = config.StorageDriverBBolt
	boltBackend, err := storage.Open(bboltCfg)
	if err != nil {
		t.Fatal(err)
	}
	got, err := boltBackend.Users().FindByNickname("alice")
	if err != nil {
		t.Fatal(err)
	}
	if got.PasswordHash == "" {
		t.Fatal("password hash missing after migrate")
	}
	ws, err := boltBackend.Workouts().List("alice")
	if err != nil || len(ws) != 1 {
		t.Fatalf("workouts=%v err=%v", ws, err)
	}
	_ = boltBackend.Close()

	// Round-trip back to file into the same location (upsert).
	result2, err := migrate.Run(migrate.Options{
		From:   config.StorageDriverBBolt,
		To:     config.StorageDriverFile,
		Config: cfg,
		Verify: true,
	})
	if err != nil {
		t.Fatalf("bbolt→file: %v", err)
	}
	if result2.Users != 1 || result2.Workouts != 1 {
		t.Fatalf("round-trip result: %+v", result2)
	}
}

func TestMigratePreservesLocalCharts(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}

	fileBackend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.Users().Create("alice", "Alice", "alice@example.com", "password123"); err != nil {
		t.Fatal(err)
	}

	gpx, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "tracks", "1-sample.gpx"))
	if err != nil {
		t.Fatal(err)
	}
	fit, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "tracks", "1-ride.fit"))
	if err != nil {
		t.Fatal(err)
	}

	gpxWorkout, err := fileBackend.Workouts().CreateWithTrack("alice", &workouts.Workout{
		Name: "GPX run", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	}, &workouts.TrackInput{Filename: "1-sample.gpx", Data: gpx})
	if err != nil {
		t.Fatalf("CreateWithTrack GPX: %v", err)
	}
	fitWorkout, err := fileBackend.Workouts().CreateWithTrack("alice", &workouts.Workout{
		Name: "FIT ride", SportType: "Ride",
		StartDate: time.Date(2026, 7, 9, 10, 0, 0, 0, time.UTC),
	}, &workouts.TrackInput{Filename: "1-ride.fit", Data: fit})
	if err != nil {
		t.Fatalf("CreateWithTrack FIT: %v", err)
	}

	_, wantSpeed, err := fileBackend.Workouts().GetSpeedChart("alice", gpxWorkout.ID)
	if err != nil || len(wantSpeed) < 1 {
		t.Fatalf("file speed chart: len=%d err=%v", len(wantSpeed), err)
	}
	_, wantHR, err := fileBackend.Workouts().GetHeartRateChart("alice", fitWorkout.ID)
	if err != nil || len(wantHR) < 1 {
		t.Fatalf("file HR chart: len=%d err=%v", len(wantHR), err)
	}
	_ = fileBackend.Close()

	if _, err := migrate.Run(migrate.Options{
		From: config.StorageDriverFile, To: config.StorageDriverBBolt, Config: cfg, Force: true, Verify: true,
	}); err != nil {
		t.Fatalf("file→bbolt: %v", err)
	}

	bboltCfg := cfg
	bboltCfg.Driver = config.StorageDriverBBolt
	boltBackend, err := storage.Open(bboltCfg)
	if err != nil {
		t.Fatal(err)
	}
	_, gotSpeed, err := boltBackend.Workouts().GetSpeedChart("alice", gpxWorkout.ID)
	if err != nil {
		t.Fatalf("bbolt GetSpeedChart: %v", err)
	}
	if len(gotSpeed) != len(wantSpeed) {
		t.Fatalf("bbolt speed len=%d want %d", len(gotSpeed), len(wantSpeed))
	}
	_, gotHR, err := boltBackend.Workouts().GetHeartRateChart("alice", fitWorkout.ID)
	if err != nil {
		t.Fatalf("bbolt GetHeartRateChart: %v", err)
	}
	if len(gotHR) != len(wantHR) {
		t.Fatalf("bbolt HR len=%d want %d", len(gotHR), len(wantHR))
	}
	_ = boltBackend.Close()

	// Remove JSON chart blobs so bbolt→file must rewrite them from binary buckets.
	gpxDir := filepath.Join(dir, "users", "alice", "workouts")
	entries, err := os.ReadDir(gpxDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		_ = os.Remove(filepath.Join(gpxDir, e.Name(), "speed-chart.json"))
		_ = os.Remove(filepath.Join(gpxDir, e.Name(), "heartrate-chart.json"))
	}

	if _, err := migrate.Run(migrate.Options{
		From: config.StorageDriverBBolt, To: config.StorageDriverFile, Config: cfg, Verify: true,
	}); err != nil {
		t.Fatalf("bbolt→file: %v", err)
	}

	fileCfg := cfg
	fileCfg.Driver = config.StorageDriverFile
	fileBackend2, err := storage.Open(fileCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer fileBackend2.Close()
	_, roundSpeed, err := fileBackend2.Workouts().GetSpeedChart("alice", gpxWorkout.ID)
	if err != nil || len(roundSpeed) != len(wantSpeed) {
		t.Fatalf("round-trip speed len=%d want %d err=%v", len(roundSpeed), len(wantSpeed), err)
	}
	_, roundHR, err := fileBackend2.Workouts().GetHeartRateChart("alice", fitWorkout.ID)
	if err != nil || len(roundHR) != len(wantHR) {
		t.Fatalf("round-trip HR len=%d want %d err=%v", len(roundHR), len(wantHR), err)
	}
}

func TestMigratePreservesLikes(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}

	fileBackend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.Users().Create("alice", "Alice", "alice@example.com", "password123"); err != nil {
		t.Fatal(err)
	}
	localWorkout, err := fileBackend.Workouts().Create("alice", &workouts.Workout{
		Name: "Run", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	localLikes := &workouts.WorkoutLikes{Users: []workouts.WorkoutLikeUser{
		{Handle: "bob@localhost", Nickname: "bob", Name: "Bob", IsLocal: true},
		{Handle: "carol@remote.test", Nickname: "carol", Name: "Carol", IsLocal: false},
	}}
	if err := fileBackend.Likes().PutLocal("alice", localWorkout.ID, localLikes); err != nil {
		t.Fatal(err)
	}

	ownerHandle := "bob@remote.test"
	fedWorkoutID := "38472901"
	fedLikes := &workouts.WorkoutLikes{Users: []workouts.WorkoutLikeUser{
		{Handle: "alice@localhost", Nickname: "alice", Name: "Alice", IsLocal: true},
	}}
	if err := fileBackend.Likes().PutFederated("alice", ownerHandle, fedWorkoutID, fedLikes); err != nil {
		t.Fatal(err)
	}

	objectID := "https://remote.test/users/bob/workouts/" + fedWorkoutID
	activityID := "https://localhost/users/alice/activities/like-1"
	if err := fileBackend.Likes().PutLikeActivityID("alice", objectID, activityID); err != nil {
		t.Fatal(err)
	}

	// Federated inbox metadata so activity ids can be reconstructed on migrate.
	ownerKey := "bob_remote.test"
	inboxDir := filepath.Join(dir, "users", "alice", "federation", "inbox", "workouts", ownerKey)
	if err := os.MkdirAll(inboxDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inboxDir, "author.yaml"), []byte("nickname: bob\nhandle: bob@remote.test\nname: Bob\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inboxDir, fedWorkoutID+".yaml"), []byte("id: \"38472901\"\nname: Remote Ride\nsport_type: Ride\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_ = fileBackend.Close()

	result, err := migrate.Run(migrate.Options{
		From: config.StorageDriverFile, To: config.StorageDriverBBolt, Config: cfg, Verify: true, Force: true,
	})
	if err != nil {
		t.Fatalf("file→bbolt: %v", err)
	}
	if result.LocalLikes != 1 || result.FedLikes != 1 || result.LikeActivities != 1 {
		t.Fatalf("likes result: %+v", result)
	}

	bboltCfg := cfg
	bboltCfg.Driver = config.StorageDriverBBolt
	boltBackend, err := storage.Open(bboltCfg)
	if err != nil {
		t.Fatal(err)
	}
	gotLocal, err := boltBackend.Likes().GetLocal("alice", localWorkout.ID)
	if err != nil || gotLocal.Likes != 2 {
		t.Fatalf("bbolt local likes: %#v err=%v", gotLocal, err)
	}
	gotFed, err := boltBackend.Likes().GetFederated("alice", ownerHandle, fedWorkoutID)
	if err != nil || gotFed.Likes != 1 {
		t.Fatalf("bbolt fed likes: %#v err=%v", gotFed, err)
	}
	gotActivity, err := boltBackend.Likes().GetLikeActivityID("alice", objectID)
	if err != nil || gotActivity != activityID {
		t.Fatalf("bbolt activity id: %q err=%v", gotActivity, err)
	}
	_ = boltBackend.Close()

	// Clear file likes so bbolt→file must rewrite them.
	_ = os.Remove(filepath.Join(dir, "users", "alice", "workouts",
		keys.WorkoutDirName(localWorkout.StartDate, localWorkout.ID), "likes.yaml"))
	_ = os.RemoveAll(filepath.Join(inboxDir, "likes"))
	_ = os.RemoveAll(filepath.Join(dir, "users", "alice", "federation", "outbox", "likes"))

	result2, err := migrate.Run(migrate.Options{
		From: config.StorageDriverBBolt, To: config.StorageDriverFile, Config: cfg, Verify: true,
	})
	if err != nil {
		t.Fatalf("bbolt→file: %v", err)
	}
	if result2.LocalLikes != 1 || result2.FedLikes != 1 || result2.LikeActivities != 1 {
		t.Fatalf("round-trip likes result: %+v", result2)
	}

	fileCfg := cfg
	fileCfg.Driver = config.StorageDriverFile
	fileBackend2, err := storage.Open(fileCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer fileBackend2.Close()
	gotLocal, err = fileBackend2.Likes().GetLocal("alice", localWorkout.ID)
	if err != nil || gotLocal.Likes != 2 {
		t.Fatalf("round-trip local likes: %#v err=%v", gotLocal, err)
	}
	gotFed, err = fileBackend2.Likes().GetFederated("alice", ownerHandle, fedWorkoutID)
	if err != nil || gotFed.Likes != 1 {
		t.Fatalf("round-trip fed likes: %#v err=%v", gotFed, err)
	}
	gotActivity, err = fileBackend2.Likes().GetLikeActivityID("alice", objectID)
	if err != nil || gotActivity != activityID {
		t.Fatalf("round-trip activity id: %q err=%v", gotActivity, err)
	}
}

func TestMigratePreservesComments(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}

	fileBackend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.Users().Create("alice", "Alice", "alice@example.com", "password123"); err != nil {
		t.Fatal(err)
	}
	localWorkout, err := fileBackend.Workouts().Create("alice", &workouts.Workout{
		Name: "Run", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	localComments := &workouts.WorkoutComments{Comments: []workouts.WorkoutComment{
		{
			ID: "c1",
			User: workouts.WorkoutLikeUser{
				Handle: "bob@localhost", Nickname: "bob", Name: "Bob", IsLocal: true,
			},
			Datetime: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC),
			Text:     "Nice!",
			NoteID:   "https://localhost/users/bob/notes/c1",
		},
	}}
	if err := fileBackend.Comments().PutLocal("alice", localWorkout.ID, localComments); err != nil {
		t.Fatal(err)
	}

	ownerHandle := "bob@remote.test"
	fedWorkoutID := "38472901"
	fedComments := &workouts.WorkoutComments{Comments: []workouts.WorkoutComment{
		{
			ID: "c2",
			User: workouts.WorkoutLikeUser{
				Handle: "alice@localhost", Nickname: "alice", Name: "Alice", IsLocal: true,
			},
			Datetime: time.Date(2026, 8, 6, 13, 0, 0, 0, time.UTC),
			Text:     "hi",
			NoteID:   "https://localhost/users/alice/notes/c2",
		},
	}}
	if err := fileBackend.Comments().PutFederated("alice", ownerHandle, fedWorkoutID, fedComments); err != nil {
		t.Fatal(err)
	}

	noteID := "https://localhost/users/alice/notes/c2"
	activityID := "https://localhost/users/alice/activities/comment-1"
	if err := fileBackend.Comments().PutCommentActivityID("alice", noteID, activityID); err != nil {
		t.Fatal(err)
	}

	ownerKey := "bob_remote.test"
	inboxDir := filepath.Join(dir, "users", "alice", "federation", "inbox", "workouts", ownerKey)
	if err := os.MkdirAll(inboxDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inboxDir, "author.yaml"), []byte("nickname: bob\nhandle: bob@remote.test\nname: Bob\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inboxDir, fedWorkoutID+".yaml"), []byte("id: \"38472901\"\nname: Remote Ride\nsport_type: Ride\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_ = fileBackend.Close()

	result, err := migrate.Run(migrate.Options{
		From: config.StorageDriverFile, To: config.StorageDriverBBolt, Config: cfg, Verify: true, Force: true,
	})
	if err != nil {
		t.Fatalf("file→bbolt: %v", err)
	}
	if result.LocalComments != 1 || result.FedComments != 1 || result.CommentActivities != 1 {
		t.Fatalf("comments result: %+v", result)
	}

	bboltCfg := cfg
	bboltCfg.Driver = config.StorageDriverBBolt
	boltBackend, err := storage.Open(bboltCfg)
	if err != nil {
		t.Fatal(err)
	}
	gotLocal, err := boltBackend.Comments().GetLocal("alice", localWorkout.ID)
	if err != nil || gotLocal.CommentsNum != 1 || gotLocal.Comments[0].Text != "Nice!" {
		t.Fatalf("bbolt local comments: %#v err=%v", gotLocal, err)
	}
	gotFed, err := boltBackend.Comments().GetFederated("alice", ownerHandle, fedWorkoutID)
	if err != nil || gotFed.CommentsNum != 1 {
		t.Fatalf("bbolt fed comments: %#v err=%v", gotFed, err)
	}
	gotActivity, err := boltBackend.Comments().GetCommentActivityID("alice", noteID)
	if err != nil || gotActivity != activityID {
		t.Fatalf("bbolt activity id: %q err=%v", gotActivity, err)
	}
	_ = boltBackend.Close()

	_ = os.Remove(filepath.Join(dir, "users", "alice", "workouts",
		keys.WorkoutDirName(localWorkout.StartDate, localWorkout.ID), "comments.yaml"))
	_ = os.RemoveAll(filepath.Join(inboxDir, "comments"))
	_ = os.RemoveAll(filepath.Join(dir, "users", "alice", "federation", "outbox", "comments"))

	result2, err := migrate.Run(migrate.Options{
		From: config.StorageDriverBBolt, To: config.StorageDriverFile, Config: cfg, Verify: true,
	})
	if err != nil {
		t.Fatalf("bbolt→file: %v", err)
	}
	if result2.LocalComments != 1 || result2.FedComments != 1 || result2.CommentActivities != 1 {
		t.Fatalf("round-trip comments result: %+v", result2)
	}

	fileCfg := cfg
	fileCfg.Driver = config.StorageDriverFile
	fileBackend2, err := storage.Open(fileCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer fileBackend2.Close()
	gotLocal, err = fileBackend2.Comments().GetLocal("alice", localWorkout.ID)
	if err != nil || gotLocal.CommentsNum != 1 || gotLocal.Comments[0].Text != "Nice!" {
		t.Fatalf("round-trip local comments: %#v err=%v", gotLocal, err)
	}
	gotFed, err = fileBackend2.Comments().GetFederated("alice", ownerHandle, fedWorkoutID)
	if err != nil || gotFed.CommentsNum != 1 {
		t.Fatalf("round-trip fed comments: %#v err=%v", gotFed, err)
	}
	gotActivity, err = fileBackend2.Comments().GetCommentActivityID("alice", noteID)
	if err != nil || gotActivity != activityID {
		t.Fatalf("round-trip activity id: %q err=%v", gotActivity, err)
	}
}

func TestMigrateLegacyLikeActivityID(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}

	fileBackend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.Users().Create("alice", "Alice", "alice@example.com", "password123"); err != nil {
		t.Fatal(err)
	}

	fedWorkoutID := "38472901"
	objectID := "https://remote.test/users/bob/workouts/" + fedWorkoutID
	activityID := "https://localhost/users/alice/activities/legacy-1"
	ownerKey := "bob_remote.test"
	inboxDir := filepath.Join(dir, "users", "alice", "federation", "inbox", "workouts", ownerKey)
	if err := os.MkdirAll(inboxDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inboxDir, "author.yaml"), []byte("nickname: bob\nhandle: bob@remote.test\nname: Bob\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inboxDir, fedWorkoutID+".yaml"), []byte("id: \"38472901\"\nname: Remote Ride\nsport_type: Ride\n"), 0600); err != nil {
		t.Fatal(err)
	}

	// Legacy plain-text activity file (pre-YAML format).
	sumPath := filepath.Join(dir, "users", "alice", "federation", "outbox", "likes")
	if err := os.MkdirAll(sumPath, 0700); err != nil {
		t.Fatal(err)
	}
	// Filename must match sha256(objectID) as used by the file likes store.
	if err := fileBackend.Likes().PutLikeActivityID("alice", objectID, activityID); err != nil {
		t.Fatal(err)
	}
	// Overwrite with legacy plain text after Put wrote YAML.
	entries, err := os.ReadDir(sumPath)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 activity file: %v len=%d", err, len(entries))
	}
	if err := os.WriteFile(filepath.Join(sumPath, entries[0].Name()), []byte(activityID), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := fileBackend.Likes().GetLikeActivityID("alice", objectID)
	if err != nil || got != activityID {
		t.Fatalf("legacy read: %q err=%v", got, err)
	}
	_ = fileBackend.Close()

	if _, err := migrate.Run(migrate.Options{
		From: config.StorageDriverFile, To: config.StorageDriverBBolt, Config: cfg, Verify: true, Force: true,
	}); err != nil {
		t.Fatalf("file→bbolt: %v", err)
	}

	bboltCfg := cfg
	bboltCfg.Driver = config.StorageDriverBBolt
	boltBackend, err := storage.Open(bboltCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer boltBackend.Close()
	got, err = boltBackend.Likes().GetLikeActivityID("alice", objectID)
	if err != nil || got != activityID {
		t.Fatalf("migrated legacy activity id: %q err=%v", got, err)
	}
}

func TestMigrateDryRun(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}
	backend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := backend.Users().Create("alice", "Alice", "alice@example.com", "password123"); err != nil {
		t.Fatal(err)
	}
	_ = backend.Close()

	result, err := migrate.Run(migrate.Options{
		From:   config.StorageDriverFile,
		To:     config.StorageDriverBBolt,
		Config: cfg,
		DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Users != 1 {
		t.Fatalf("dry-run users=%d", result.Users)
	}
	if _, err := os.Stat(filepath.Join(dir, "grom.db")); err == nil {
		t.Fatal("dry-run should not create bbolt db")
	}
}

func TestMigratePreservesUserProfile(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}

	fileBackend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	user, err := fileBackend.Users().Create("alice", "Alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := fileBackend.Users().PutProfile(user.ID, users.Profile{
		LastSportType: "Ride",
		LastEquipmentBySport: map[string][]string{
			"Run":  {"eq-1"},
			"Ride": {"bike-1", "bike-2"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	_ = fileBackend.Close()

	result, err := migrate.Run(migrate.Options{
		From: config.StorageDriverFile, To: config.StorageDriverBBolt, Config: cfg, Verify: true,
	})
	if err != nil {
		t.Fatalf("file→bbolt: %v", err)
	}
	if result.Profiles != 1 {
		t.Fatalf("profiles result: %+v", result)
	}

	bboltCfg := cfg
	bboltCfg.Driver = config.StorageDriverBBolt
	boltBackend, err := storage.Open(bboltCfg)
	if err != nil {
		t.Fatal(err)
	}
	gotUser, err := boltBackend.Users().FindByNickname("alice")
	if err != nil {
		t.Fatal(err)
	}
	profile, err := boltBackend.Users().GetProfile(gotUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	if profile.LastSportType != "Ride" {
		t.Fatalf("sport = %q", profile.LastSportType)
	}
	if got := profile.LastEquipmentBySport["Ride"]; len(got) != 2 || got[0] != "bike-1" {
		t.Fatalf("equipment: %#v", profile.LastEquipmentBySport)
	}
	_ = boltBackend.Close()

	if err := os.RemoveAll(filepath.Join(dir, "users", "alice", "profile.yaml")); err != nil && !os.IsNotExist(err) {
		// profile may live next to user; clear via empty put after reopen is enough below
		_ = err
	}
	// Force bbolt→file to rewrite profile by wiping file profile if present.
	fileBackend2, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	alice2, err := fileBackend2.Users().FindByNickname("alice")
	if err != nil {
		t.Fatal(err)
	}
	_ = fileBackend2.Users().PutProfile(alice2.ID, users.Profile{})
	_ = fileBackend2.Close()

	if _, err := migrate.Run(migrate.Options{
		From: config.StorageDriverBBolt, To: config.StorageDriverFile, Config: cfg, Verify: true,
	}); err != nil {
		t.Fatalf("bbolt→file: %v", err)
	}

	fileBackend3, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer fileBackend3.Close()
	alice3, err := fileBackend3.Users().FindByNickname("alice")
	if err != nil {
		t.Fatal(err)
	}
	roundTrip, err := fileBackend3.Users().GetProfile(alice3.ID)
	if err != nil {
		t.Fatal(err)
	}
	if roundTrip.LastSportType != "Ride" {
		t.Fatalf("round-trip sport = %q", roundTrip.LastSportType)
	}
	if got := roundTrip.LastEquipmentBySport["Run"]; len(got) != 1 || got[0] != "eq-1" {
		t.Fatalf("round-trip equipment: %#v", roundTrip.LastEquipmentBySport)
	}
}

func TestMigrateResetTokensNotCopied(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}

	fileBackend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	user, err := fileBackend.Users().Create("alice", "Alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	if err := fileBackend.ResetTokens().ReplaceForUser(user.ID, reset.TokenRecord{
		TokenHash: "hash-before-migrate",
		UserID:    user.ID,
		ExpiresAt: now.Add(time.Hour),
		CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.ResetTokens().GetByHash("hash-before-migrate"); err != nil {
		t.Fatalf("source token: %v", err)
	}
	_ = fileBackend.Close()

	if _, err := migrate.Run(migrate.Options{
		From:   config.StorageDriverFile,
		To:     config.StorageDriverBBolt,
		Config: cfg,
		Verify: true,
	}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	bboltCfg := cfg
	bboltCfg.Driver = config.StorageDriverBBolt
	boltBackend, err := storage.Open(bboltCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer boltBackend.Close()

	if _, err := boltBackend.ResetTokens().GetByHash("hash-before-migrate"); !errors.Is(err, reset.ErrInvalidToken) {
		t.Fatalf("expected reset token absent on bbolt, got %v", err)
	}
}

func TestMigratePreservesPersonalAccessTokens(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}

	fileBackend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	user, err := fileBackend.Users().Create("alice", "Alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)
	rec := pat.TokenRecord{
		ID:          "pat-1",
		TokenHash:   "hash-pat-1",
		TokenPrefix: "grom_pat_xx",
		UserID:      user.ID,
		Name:        "CI",
		Scopes:      []string{pat.ScopeWorkoutsRead, pat.ScopeEquipmentRead},
		CreatedAt:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		ExpiresAt:   &expires,
	}
	if err := fileBackend.PAT().Create(rec); err != nil {
		t.Fatal(err)
	}
	_ = fileBackend.Close()

	result, err := migrate.Run(migrate.Options{
		From: config.StorageDriverFile, To: config.StorageDriverBBolt, Config: cfg, Verify: true, Force: true,
	})
	if err != nil {
		t.Fatalf("file→bbolt: %v", err)
	}
	if result.PersonalAccessTokens != 1 {
		t.Fatalf("pats result: %+v", result)
	}

	bboltCfg := cfg
	bboltCfg.Driver = config.StorageDriverBBolt
	boltBackend, err := storage.Open(bboltCfg)
	if err != nil {
		t.Fatal(err)
	}
	got, err := boltBackend.PAT().GetByHash("hash-pat-1")
	if err != nil {
		t.Fatalf("bbolt pat: %v", err)
	}
	if got.ID != "pat-1" || got.Name != "CI" || got.UserID != user.ID || len(got.Scopes) != 2 {
		t.Fatalf("unexpected pat: %#v", got)
	}
	_ = boltBackend.Close()

	result2, err := migrate.Run(migrate.Options{
		From: config.StorageDriverBBolt, To: config.StorageDriverFile, Config: cfg, Verify: true,
	})
	if err != nil {
		t.Fatalf("bbolt→file: %v", err)
	}
	if result2.PersonalAccessTokens != 1 {
		t.Fatalf("round-trip pats: %+v", result2)
	}

	fileCfg := cfg
	fileCfg.Driver = config.StorageDriverFile
	fileBackend2, err := storage.Open(fileCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer fileBackend2.Close()
	got, err = fileBackend2.PAT().GetByHash("hash-pat-1")
	if err != nil || got.Name != "CI" {
		t.Fatalf("round-trip pat: %#v err=%v", got, err)
	}
}

func TestMigrateLegacyLocalLikeActivity(t *testing.T) {
	dir := t.TempDir()
	cfg := config.StorageConfig{
		Driver:            config.StorageDriverFile,
		ResolvedLocation:  dir,
		ResolvedBBoltPath: filepath.Join(dir, "grom.db"),
	}

	fileBackend, err := storage.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.Users().Create("alice", "Alice", "alice@example.com", "password123"); err != nil {
		t.Fatal(err)
	}
	if _, err := fileBackend.Users().Create("bob", "Bob", "bob@example.com", "password123"); err != nil {
		t.Fatal(err)
	}
	w, err := fileBackend.Workouts().Create("bob", &workouts.Workout{
		Name: "Run", SportType: "Run",
		StartDate: time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	domain := "grom.test"
	objectID := "https://" + domain + "/users/bob/workouts/" + w.ID
	activityID := "https://" + domain + "/users/alice/activities/legacy-local-1"
	if err := fileBackend.Likes().PutLikeActivityID("alice", objectID, activityID); err != nil {
		t.Fatal(err)
	}
	sumPath := filepath.Join(dir, "users", "alice", "federation", "outbox", "likes")
	entries, err := os.ReadDir(sumPath)
	if err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 activity file: %v len=%d", err, len(entries))
	}
	if err := os.WriteFile(filepath.Join(sumPath, entries[0].Name()), []byte(activityID), 0600); err != nil {
		t.Fatal(err)
	}
	_ = fileBackend.Close()

	result, err := migrate.Run(migrate.Options{
		From:             config.StorageDriverFile,
		To:               config.StorageDriverBBolt,
		Config:           cfg,
		FederationDomain: domain,
		Verify:           true,
		Force:            true,
	})
	if err != nil {
		t.Fatalf("file→bbolt: %v", err)
	}
	if result.LikeActivities != 1 {
		t.Fatalf("like activities: %+v", result)
	}

	bboltCfg := cfg
	bboltCfg.Driver = config.StorageDriverBBolt
	boltBackend, err := storage.Open(bboltCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer boltBackend.Close()
	got, err := boltBackend.Likes().GetLikeActivityID("alice", objectID)
	if err != nil || got != activityID {
		t.Fatalf("migrated local legacy activity id: %q err=%v", got, err)
	}
}

