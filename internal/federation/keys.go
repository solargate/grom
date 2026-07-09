package federation

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/solargate/travka/internal/config"
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

var keyMu sync.Mutex

func actorKeyPath(nickname string) string {
	return filepath.Join(config.Cfg.Data.ResolvedDir, nickname, "federation", "actor_key.pem")
}

func LoadOrCreateActorKey(nickname string) (publicKeyPEM string, keyID string, err error) {
	keyMu.Lock()
	defer keyMu.Unlock()

	path := actorKeyPath(nickname)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", "", err
	}

	data, readErr := os.ReadFile(path)
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
	if err := os.WriteFile(path, privPEM, 0600); err != nil {
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
