package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/a-h/templ"
)

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, err := contentSecurityNonce()
		if err != nil {
			log.Printf("generate CSP nonce: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		policy := []string{
			"default-src 'none'",
			"base-uri 'none'",
			"connect-src 'self' https://tiles.openfreemap.org",
			"font-src 'self' https://fonts.gstatic.com",
			"form-action 'self'",
			"frame-ancestors 'none'",
			"frame-src 'none'",
			"img-src 'self' data: blob:",
			"manifest-src 'self'",
			"media-src 'self'",
			"object-src 'none'",
			fmt.Sprintf("script-src 'nonce-%s' 'strict-dynamic' 'self'", nonce),
			"script-src-attr 'none'",
			"style-src 'self' https://fonts.googleapis.com",
			"style-src-attr 'unsafe-inline'",
			"worker-src 'self'",
		}
		if os.Getenv("GO_ENV") == "production" {
			policy = append(policy, "upgrade-insecure-requests")
			w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}

		w.Header().Set("Content-Security-Policy", strings.Join(policy, "; "))
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		w.Header().Set("Permissions-Policy", "camera=(), geolocation=(self), microphone=(), payment=(), usb=(), browsing-topics=()")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")

		ctx := templ.WithNonce(r.Context(), nonce)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func contentSecurityNonce() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
