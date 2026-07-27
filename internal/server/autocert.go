package server

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/solargate/grom/internal/config"
	"golang.org/x/crypto/acme/autocert"
)

func runAutocertTLS(router *gin.Engine) error {
	ac := config.Cfg.Server.TLS.Autocert
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(ac.ResolvedCacheDir),
		HostPolicy: autocert.HostWhitelist(ac.Domains...),
		Email:      ac.Email,
	}

	slog.Info("autocert enabled", "domains", ac.Domains, "cache_dir", ac.ResolvedCacheDir)

	go func() {
		handler := m.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}))
		if err := http.ListenAndServe(":80", handler); err != nil {
			slog.Error("autocert HTTP listener failed", "err", err)
		}
	}()

	tlsPort := config.Cfg.Server.TLS.Port
	srv := &http.Server{
		Addr:      ":" + strconv.Itoa(tlsPort),
		Handler:   router,
		TLSConfig: m.TLSConfig(),
	}
	return srv.ListenAndServeTLS("", "")
}
