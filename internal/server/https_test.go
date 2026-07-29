package server_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/solargate/grom/internal/config"
	"github.com/solargate/grom/internal/server"
)

func TestFederationHTTPClientDefault(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg = config.Config{}
	config.Cfg.Federation.TLSInsecureSkipVerify = false

	client, err := server.FederationHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	if client == nil || client.Transport == nil {
		t.Fatal("expected client with transport")
	}
}

func TestFederationHTTPClientInsecure(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	config.Cfg = config.Config{}
	config.Cfg.Federation.TLSInsecureSkipVerify = true

	client, err := server.FederationHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	tr, ok := client.Transport.(*http.Transport)
	if !ok || tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("expected insecure TLS config, got %#v", client.Transport)
	}
}

func TestFederationHTTPClientCustomCA(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })

	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.crt")
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0600); err != nil {
		t.Fatal(err)
	}

	config.Cfg = config.Config{}
	config.Cfg.Federation.ResolvedCACertFile = caPath
	client, err := server.FederationHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	tr, ok := client.Transport.(*http.Transport)
	if !ok || tr.TLSClientConfig == nil || tr.TLSClientConfig.RootCAs == nil {
		t.Fatalf("expected custom root CAs, got %#v", client.Transport)
	}
}

func TestFederationHTTPClientInvalidCA(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })
	dir := t.TempDir()
	caPath := filepath.Join(dir, "bad.crt")
	if err := os.WriteFile(caPath, []byte("not-a-cert"), 0600); err != nil {
		t.Fatal(err)
	}
	config.Cfg = config.Config{}
	config.Cfg.Federation.ResolvedCACertFile = caPath
	if _, err := server.FederationHTTPClient(); err == nil {
		t.Fatal("expected invalid CA error")
	}
}
