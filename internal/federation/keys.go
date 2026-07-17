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

func workoutObjectURL(authorNickname, workoutID string) string {
	return fmt.Sprintf("%s/workouts/%s", actorURL(authorNickname), workoutID)
}

var keyMu sync.Mutex

func LoadOrCreateActorKey(blobs blob.Store, nickname string) (publicKeyPEM string, keyID string, err error) {
	keyMu.Lock()
	defer keyMu.Unlock()

	ctx := context.Background()
	keyPath := keys.UserActorKey(nickname)

	data, readErr := blob.ReadAll(ctx, blobs, keyPath)
	if readErr == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			key, parseErr := x509.ParsePKCS1PrivateKey(block.Bytes)
			if parseErr == nil {
				pubDER, marshalErr := x509.MarshalPKIXPublicKey(&key.PublicKey)
				if marshalErr != nil {
					return "", "", marshalErr
				}
				pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
				return string(pubPEM), actorURL(nickname) + "#main-key", nil
			}
		}
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}
	privDER := x509.MarshalPKCS1PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER})
	if _, err := blob.PutBytes(ctx, blobs, keyPath, privPEM, blob.PutOptions{}); err != nil {
		return "", "", err
	}

	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return "", "", err
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return string(pubPEM), actorURL(nickname) + "#main-key", nil
}

func ActorKeyID(nickname string) string {
	return fmt.Sprintf("%s#main-key", actorURL(nickname))
}
