package server

import (
	"log"
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
		Cache:      autocert.DirCache(ac.CacheDir),
		HostPolicy: autocert.HostWhitelist(ac.Domains...),
		Email:      ac.Email,
	}

	log.Printf("autocert: domains=%v cache=%s", ac.Domains, ac.CacheDir)

	go func() {
		handler := m.HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}))
		if err := http.ListenAndServe(":80", handler); err != nil {
			log.Printf("autocert HTTP listener: %v", err)
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
