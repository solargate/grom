package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/config"
)

func Run(router *gin.Engine) error {
	switch config.ResolveTLSMode(&config.Cfg) {
	case config.TLSModeStatic:
		return runStaticTLS(router)
	case config.TLSModeAutocert:
		return runAutocertTLS(router)
	default:
		port := config.Cfg.Server.Port
		if port <= 0 {
			port = 8080
		}
		return router.Run(":" + strconv.Itoa(port))
	}
}

func runStaticTLS(router *gin.Engine) error {
	certFile := config.Cfg.Server.TLS.CertFile
	keyFile := config.Cfg.Server.TLS.KeyFile

	tlsPort := config.Cfg.Server.TLS.Port
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(tlsPort),
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
