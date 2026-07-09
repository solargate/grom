package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/solargate/travka/internal/config"
)

func Run(router *gin.Engine) error {
	if config.Cfg.Server.TLS.Enabled {
		return runTLS(router)
	}
	return router.Run(":" + strconv.Itoa(config.Cfg.Server.Port))
}

func runTLS(router *gin.Engine) error {
	certFile := config.Cfg.Server.TLS.CertFile
	keyFile := config.Cfg.Server.TLS.KeyFile
	if certFile == "" || keyFile == "" {
		return fmt.Errorf("server.tls.cert_file and server.tls.key_file are required when TLS is enabled")
	}

	tlsPort := config.Cfg.Server.TLS.Port
	addr := ":" + strconv.Itoa(tlsPort)

	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	if config.Cfg.Server.Port > 0 && config.Cfg.Server.Port != tlsPort {
		go func() {
			_ = router.Run(":" + strconv.Itoa(config.Cfg.Server.Port))
		}()
	}

	return srv.ListenAndServeTLS(certFile, keyFile)
}

func FederationHTTPClient() (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if config.Cfg.Federation.TLSInsecureSkipVerify {
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = true
	} else if config.Cfg.Federation.CACertFile != "" {
		data, err := os.ReadFile(config.Cfg.Federation.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("read federation ca cert: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("invalid federation ca cert")
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.RootCAs = pool
	}

	return &http.Client{Transport: transport}, nil
}
