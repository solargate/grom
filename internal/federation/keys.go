package federation

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/storage/blob"
	"github.com/solargate/grom/internal/storage/keys"
)

func publicDomain() string {
	if config.Cfg.Federation.Domain != "" {
		return config.Cfg.Federation.Domain
	}
	return "localhost"
}

func actorURL(nickname string) string {
	return fmt.Sprintf("https://%s/users/%s", publicDomain(), nickname)
}

func instanceActorURL() string {
	return fmt.Sprintf("https://%s/actor", publicDomain())
}

func workoutObjectURL(authorNickname, workoutID string) string {
	return fmt.Sprintf("%s/workouts/%s", actorURL(authorNickname), workoutID)
}

var keyMu sync.Mutex

// ActorKey holds a loaded RSA keypair and its ActivityPub key id.
type ActorKey struct {
	Private *rsa.PrivateKey
	KeyID   string
	PubPEM  string
}

func LoadOrCreateActorKey(blobs blob.Store, nickname string) (publicKeyPEM string, keyID string, err error) {
	ak, err := LoadOrCreateUserActorKey(blobs, nickname)
	if err != nil {
		return "", "", err
	}
	return ak.PubPEM, ak.KeyID, nil
}

func LoadOrCreateUserActorKey(blobs blob.Store, nickname string) (*ActorKey, error) {
	return loadOrCreateKey(blobs, keys.UserActorKey(nickname), actorURL(nickname)+"#main-key")
}

func LoadOrCreateInstanceActorKey(blobs blob.Store) (*ActorKey, error) {
	return loadOrCreateKey(blobs, keys.InstanceActorKey(), instanceActorURL()+"#main-key")
}

func ActorKeyID(nickname string) string {
	return fmt.Sprintf("%s#main-key", actorURL(nickname))
}

func InstanceActorKeyID() string {
	return instanceActorURL() + "#main-key"
}

func loadOrCreateKey(blobs blob.Store, keyPath, keyID string) (*ActorKey, error) {
	keyMu.Lock()
	defer keyMu.Unlock()

	ctx := context.Background()
	data, readErr := blob.ReadAll(ctx, blobs, keyPath)
	if readErr == nil {
		if ak, err := parseActorKeyPEM(data, keyID); err == nil {
			return ak, nil
		}
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	privDER := x509.MarshalPKCS1PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER})
	if _, err := blob.PutBytes(ctx, blobs, keyPath, privPEM, blob.PutOptions{}); err != nil {
		return nil, err
	}
	return actorKeyFromPrivate(key, keyID)
}

func parseActorKeyPEM(data []byte, keyID string) (*ActorKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid actor key PEM")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return actorKeyFromPrivate(key, keyID)
}

func actorKeyFromPrivate(key *rsa.PrivateKey, keyID string) (*ActorKey, error) {
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return &ActorKey{
		Private: key,
		KeyID:   keyID,
		PubPEM:  string(pubPEM),
	}, nil
}
