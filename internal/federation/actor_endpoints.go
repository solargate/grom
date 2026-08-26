package federation

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/federation/httpsig"
	"github.com/solargate/grom/internal/storage/blob"
)

// RemoteActorEndpoints are inbox URLs extracted from an actor document.
type RemoteActorEndpoints struct {
	Inbox       string
	SharedInbox string
}

// ExtractActorEndpoints reads inbox / endpoints.sharedInbox from an actor map.
func ExtractActorEndpoints(actor map[string]any) RemoteActorEndpoints {
	var out RemoteActorEndpoints
	if actor == nil {
		return out
	}
	out.Inbox, _ = actor["inbox"].(string)
	if ep, ok := actor["endpoints"].(map[string]any); ok {
		out.SharedInbox, _ = ep["sharedInbox"].(string)
	}
	return out
}

// PreferDeliveryInbox returns sharedInbox if set, otherwise inbox.
func PreferDeliveryInbox(ep RemoteActorEndpoints) string {
	if strings.TrimSpace(ep.SharedInbox) != "" {
		return strings.TrimSpace(ep.SharedInbox)
	}
	return strings.TrimSpace(ep.Inbox)
}

// DeduplicateDeliveryInboxes prefers SharedInbox per follower, falls back to Inbox, unique URLs.
func DeduplicateDeliveryInboxes(followers []InboundFollower) []string {
	seen := make(map[string]struct{}, len(followers))
	out := make([]string, 0, len(followers))
	for i := range followers {
		inbox := strings.TrimSpace(followers[i].SharedInbox)
		if inbox == "" {
			inbox = strings.TrimSpace(followers[i].Inbox)
		}
		if inbox == "" {
			continue
		}
		if _, ok := seen[inbox]; ok {
			continue
		}
		seen[inbox] = struct{}{}
		out = append(out, inbox)
	}
	return out
}

// DeduplicateInboxURLs returns unique non-empty inbox URLs in first-seen order.
func DeduplicateInboxURLs(inboxes []string) []string {
	seen := make(map[string]struct{}, len(inboxes))
	out := make([]string, 0, len(inboxes))
	for _, inbox := range inboxes {
		inbox = strings.TrimSpace(inbox)
		if inbox == "" {
			continue
		}
		if _, ok := seen[inbox]; ok {
			continue
		}
		seen[inbox] = struct{}{}
		out = append(out, inbox)
	}
	return out
}

// CachedPublicKey is a remote actor signing key.
type CachedPublicKey struct {
	KeyID     string
	Owner     string
	PublicKey any // *rsa.PublicKey
	FetchedAt time.Time
}

// KeyResolver resolves ActivityPub keyId URLs to public keys.
type KeyResolver interface {
	Resolve(keyID string) (pub any, owner string, err error)
}

// HTTPKeyResolver fetches and caches remote public keys.
type HTTPKeyResolver struct {
	client *http.Client
	blobs  blob.Store
	mu     sync.Mutex
	cache  map[string]CachedPublicKey
	ttl    time.Duration
}

func NewHTTPKeyResolver(client *http.Client, blobs blob.Store) *HTTPKeyResolver {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPKeyResolver{
		client: client,
		blobs:  blobs,
		cache:  make(map[string]CachedPublicKey),
		ttl:    time.Hour,
	}
}

func (r *HTTPKeyResolver) SetClient(client *http.Client) {
	if r == nil || client == nil {
		return
	}
	r.mu.Lock()
	r.client = client
	r.mu.Unlock()
}

func (r *HTTPKeyResolver) Resolve(keyID string) (pub any, owner string, err error) {
	return r.resolve(keyID, false)
}

func (r *HTTPKeyResolver) ResolveFresh(keyID string) (pub any, owner string, err error) {
	return r.resolve(keyID, true)
}

func (r *HTTPKeyResolver) resolve(keyID string, bypassCache bool) (pub any, owner string, err error) {
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return nil, "", fmt.Errorf("empty keyId")
	}
	if !bypassCache {
		r.mu.Lock()
		if c, ok := r.cache[keyID]; ok && time.Since(c.FetchedAt) < r.ttl {
			r.mu.Unlock()
			return c.PublicKey, c.Owner, nil
		}
		r.mu.Unlock()
	}

	pub, owner, status, err := r.fetchKey(keyID)
	if err != nil {
		return nil, "", err
	}
	if status == http.StatusGone {
		return nil, owner, errKeyGone
	}
	r.mu.Lock()
	r.cache[keyID] = CachedPublicKey{KeyID: keyID, Owner: owner, PublicKey: pub, FetchedAt: time.Now().UTC()}
	r.mu.Unlock()
	return pub, owner, nil
}

var errKeyGone = fmt.Errorf("remote key gone")

func (r *HTTPKeyResolver) fetchKey(keyID string) (pub any, owner string, status int, err error) {
	fetchURL := strings.SplitN(keyID, "#", 2)[0]
	req, err := http.NewRequest(http.MethodGet, fetchURL, nil)
	if err != nil {
		return nil, "", 0, err
	}
	req.Header.Set("Accept", "application/activity+json")
	if err := r.signGETAsInstance(req); err != nil {
		slog.Debug("federation key fetch unsigned", "key_id", keyID, "err", err)
	}
	r.mu.Lock()
	client := r.client
	r.mu.Unlock()
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, "", resp.StatusCode, err
	}
	if resp.StatusCode == http.StatusGone {
		return nil, "", resp.StatusCode, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, "", resp.StatusCode, fmt.Errorf("key fetch %s: %s", fetchURL, resp.Status)
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, "", resp.StatusCode, err
	}
	return publicKeyFromDocument(doc, keyID)
}

func (r *HTTPKeyResolver) signGETAsInstance(req *http.Request) error {
	if r.blobs == nil {
		return fmt.Errorf("no blob store")
	}
	ak, err := LoadOrCreateInstanceActorKey(r.blobs)
	if err != nil {
		return err
	}
	return httpsig.SignGET(req, ak.Private, ak.KeyID)
}

func publicKeyFromDocument(doc map[string]any, keyID string) (pub any, owner string, status int, err error) {
	// Actor with embedded publicKey, or a Key object.
	if pk := findPublicKeyObject(doc, keyID); pk != nil {
		owner, _ = pk["owner"].(string)
		if owner == "" {
			owner, _ = pk["controller"].(string)
		}
		pemStr, _ := pk["publicKeyPem"].(string)
		parsed, err := parsePublicKeyPEM(pemStr)
		if err != nil {
			return nil, "", 0, err
		}
		id, _ := pk["id"].(string)
		if id != "" && id != keyID && !strings.HasPrefix(keyID, id) {
			// tolerate fragment-only mismatches when fetch URL matched
		}
		if owner == "" {
			if t, _ := doc["type"].(string); t != "Key" {
				if id, _ := doc["id"].(string); id != "" {
					owner = id
				}
			}
		}
		return parsed, owner, http.StatusOK, nil
	}
	return nil, "", 0, fmt.Errorf("publicKey not found for %s", keyID)
}

func findPublicKeyObject(doc map[string]any, keyID string) map[string]any {
	if isKeyObject(doc) {
		if id, _ := doc["id"].(string); id == "" || id == keyID || strings.HasPrefix(keyID, strings.SplitN(id, "#", 2)[0]) {
			return doc
		}
	}
	switch v := doc["publicKey"].(type) {
	case map[string]any:
		return v
	case []any:
		for _, item := range v {
			m, _ := item.(map[string]any)
			if m == nil {
				continue
			}
			id, _ := m["id"].(string)
			if id == keyID || id == "" {
				return m
			}
		}
		if len(v) == 1 {
			m, _ := v[0].(map[string]any)
			return m
		}
	}
	return nil
}

func isKeyObject(doc map[string]any) bool {
	t, _ := doc["type"].(string)
	if t == "Key" {
		return true
	}
	_, hasPEM := doc["publicKeyPem"]
	return hasPEM && doc["publicKey"] == nil
}

func parsePublicKeyPEM(pemStr string) (any, error) {
	pemStr = strings.TrimSpace(pemStr)
	if pemStr == "" {
		return nil, fmt.Errorf("empty publicKeyPem")
	}
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("invalid publicKeyPem")
	}
	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		return pub, nil
	}
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("unsupported public key encoding")
}

// LocalActorNickname extracts {nick} from https://{domain}/users/{nick}.
func LocalActorNickname(actorURI string) (string, bool) {
	actorURI = strings.TrimSpace(actorURI)
	prefix := fmt.Sprintf("https://%s/users/", publicDomain())
	if !strings.HasPrefix(actorURI, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(actorURI, prefix)
	rest = strings.SplitN(rest, "/", 2)[0]
	rest = strings.SplitN(rest, "#", 2)[0]
	if rest == "" || strings.Contains(rest, "/") {
		return "", false
	}
	return rest, true
}

func collectAddressedURIs(activity map[string]any) []string {
	var out []string
	for _, key := range []string{"to", "cc", "bto", "bcc", "audience"} {
		out = append(out, stringOrSlice(activity[key])...)
	}
	return out
}

func stringOrSlice(v any) []string {
	switch t := v.(type) {
	case string:
		if t != "" {
			return []string{t}
		}
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return t
	}
	return nil
}

func objectIDString(activity map[string]any) string {
	switch o := activity["object"].(type) {
	case string:
		return o
	case map[string]any:
		id, _ := o["id"].(string)
		return id
	}
	return ""
}

func ownerMatchesKey(actorURI, keyOwner, keyID string) bool {
	actorURI = strings.TrimRight(strings.TrimSpace(actorURI), "/")
	keyOwner = strings.TrimRight(strings.TrimSpace(keyOwner), "/")
	if actorURI == "" {
		return false
	}
	if keyOwner != "" && actorURI == keyOwner {
		return true
	}
	// keyId https://host/users/a#main-key → owner https://host/users/a
	base := strings.SplitN(keyID, "#", 2)[0]
	base = strings.TrimRight(base, "/")
	return actorURI == base
}

func sameHost(a, b string) bool {
	ua, errA := url.Parse(a)
	ub, errB := url.Parse(b)
	if errA != nil || errB != nil {
		return false
	}
	return strings.EqualFold(ua.Host, ub.Host)
}

// AuthorizedFetchRequired reports whether AF is on for this process.
func AuthorizedFetchRequired() bool {
	return config.Cfg.Federation.Enabled && config.Cfg.AuthorizedFetchEnabled()
}
