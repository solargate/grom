package tlsutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GenOptions struct {
	IP     string
	Domain string
	OutDir string
}

func GenerateCerts(opts GenOptions) error {
	if opts.OutDir == "" {
		opts.OutDir = "tls"
	}
	host := strings.TrimSpace(opts.Domain)
	sanIP := strings.TrimSpace(opts.IP)
	if host == "" && sanIP == "" {
		return fmt.Errorf("either domain or ip is required")
	}
	if host == "" {
		host = sanIP
	}

	if err := os.MkdirAll(opts.OutDir, 0700); err != nil {
		return err
	}

	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Grom Test CA"},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		return err
	}
	if err := writeKeyPair(filepath.Join(opts.OutDir, "ca.crt"), caDER, caKey, false); err != nil {
		return err
	}

	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	serverTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: host},
		NotBefore:    time.Now().UTC(),
		NotAfter:     time.Now().UTC().Add(825 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	if sanIP != "" {
		serverTmpl.IPAddresses = []net.IP{net.ParseIP(sanIP)}
	}
	if opts.Domain != "" {
		serverTmpl.DNSNames = []string{opts.Domain}
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		return err
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTmpl, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		return err
	}
	if err := writePEM(filepath.Join(opts.OutDir, "server.crt"), "CERTIFICATE", serverDER); err != nil {
		return err
	}
	if err := writePEM(filepath.Join(opts.OutDir, "server.key"), "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverKey)); err != nil {
		return err
	}
	return writePEM(filepath.Join(opts.OutDir, "ca.crt"), "CERTIFICATE", caDER)
}

func writeKeyPair(path string, certDER []byte, key *rsa.PrivateKey, includeKey bool) error {
	if err := writePEM(path, "CERTIFICATE", certDER); err != nil {
		return err
	}
	if includeKey {
		return writePEM(strings.TrimSuffix(path, ".crt")+".key", "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
	}
	return writePEM(filepath.Join(filepath.Dir(path), "ca.key"), "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(key))
}

func writePEM(path, blockType string, der []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: blockType, Bytes: der})
}
