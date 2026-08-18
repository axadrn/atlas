package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"atlas/pages"
)

func TestSecurityHeadersAndScriptNonces(t *testing.T) {
	handler := withSecurityHeaders(templ.Handler(pages.Home(nil)))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	policy := response.Header().Get("Content-Security-Policy")
	nonceMatch := regexp.MustCompile(`script-src 'nonce-([^']+)'`).FindStringSubmatch(policy)
	if len(nonceMatch) != 2 || len(nonceMatch[1]) < 40 {
		t.Fatalf("CSP does not contain a strong nonce: %q", policy)
	}
	nonce := nonceMatch[1]
	if !strings.Contains(policy, "'strict-dynamic'") || !strings.Contains(policy, "script-src-attr 'none'") {
		t.Fatalf("CSP is missing strict script controls: %q", policy)
	}
	if !strings.Contains(policy, "connect-src 'self' https://tiles.openfreemap.org") ||
		!strings.Contains(policy, "frame-src 'none'") ||
		!strings.Contains(policy, "worker-src 'self'") {
		t.Fatalf("CSP is missing map isolation controls: %q", policy)
	}

	scriptTags := regexp.MustCompile(`<script\b[^>]*>`).FindAllString(response.Body.String(), -1)
	if len(scriptTags) == 0 {
		t.Fatal("expected rendered script tags")
	}
	for _, tag := range scriptTags {
		if !strings.Contains(tag, `nonce="`+nonce+`"`) {
			t.Fatalf("script tag does not use the response nonce: %s", tag)
		}
	}
	if !strings.Contains(response.Body.String(), `rel="manifest" href="/manifest.webmanifest"`) ||
		!strings.Contains(response.Body.String(), `src="/assets/js/pwa.js"`) {
		t.Fatal("page is missing PWA metadata or registration")
	}

	expectedHeaders := map[string]string{
		"Cross-Origin-Opener-Policy":   "same-origin",
		"Cross-Origin-Resource-Policy": "same-origin",
		"Referrer-Policy":              "strict-origin-when-cross-origin",
		"X-Content-Type-Options":       "nosniff",
		"X-Frame-Options":              "DENY",
	}
	for name, want := range expectedHeaders {
		if got := response.Header().Get(name); got != want {
			t.Fatalf("expected %s %q, got %q", name, want, got)
		}
	}
}

func TestStaticJavaScriptContentType(t *testing.T) {
	mux := http.NewServeMux()
	setupAssetsRoutes(mux)
	for _, path := range []string{
		"/assets/js/globe.js",
		"/assets/js/place-map.js",
		"/assets/js/pwa.js",
		"/assets/vendor/maplibre-gl-csp-5.24.0.js",
		"/assets/vendor/maplibre-gl-csp-worker-5.24.0.js",
	} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))

			if response.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
			}
			if got := response.Header().Get("Content-Type"); got != "application/javascript; charset=utf-8" {
				t.Fatalf("unexpected JavaScript content type: %q", got)
			}
		})
	}
}

func TestPWAPublicFiles(t *testing.T) {
	mux := http.NewServeMux()
	setupAssetsRoutes(mux)

	t.Run("manifest", func(t *testing.T) {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/manifest.webmanifest", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
		}
		if got := response.Header().Get("Content-Type"); got != "application/manifest+json; charset=utf-8" {
			t.Fatalf("unexpected manifest content type: %q", got)
		}
		var manifest struct {
			Name     string `json:"name"`
			StartURL string `json:"start_url"`
			Display  string `json:"display"`
			Icons    []struct {
				Source string `json:"src"`
			} `json:"icons"`
		}
		if err := json.NewDecoder(response.Body).Decode(&manifest); err != nil {
			t.Fatal(err)
		}
		if manifest.Name != "Atlas" || manifest.StartURL != "/" || manifest.Display != "standalone" || len(manifest.Icons) < 3 {
			t.Fatalf("unexpected manifest: %#v", manifest)
		}
	})

	t.Run("service worker", func(t *testing.T) {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/service-worker.js", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
		}
		if got := response.Header().Get("Content-Type"); got != "application/javascript; charset=utf-8" {
			t.Fatalf("unexpected service worker content type: %q", got)
		}
		if response.Header().Get("Cache-Control") != "no-cache" || response.Header().Get("Service-Worker-Allowed") != "/" {
			t.Fatalf("unexpected service worker headers: %#v", response.Header())
		}
		body := response.Body.String()
		if !strings.Contains(body, `request.mode === "navigate"`) ||
			!strings.Contains(body, "OFFLINE_ASSETS.indexOf(url.pathname)") ||
			strings.Contains(body, "atlas-static-v") {
			t.Fatal("service worker must stay network-first with an offline-only cache fallback")
		}
	})
}
