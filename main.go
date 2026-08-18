package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ"

	"atlas/components"
	"atlas/internal/catalog"
	"atlas/internal/database"
	"atlas/pages"
)

func main() {
	db, err := database.Open(context.Background(), databasePath())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()

	mux := http.NewServeMux()
	setupAssetsRoutes(mux)
	setupPlaceRoutes(mux, catalog.NewStore(db))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(pages.Home(githubStars())).ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if noindex() {
			fmt.Fprint(w, "User-agent: *\nDisallow: /\n")
			return
		}
		fmt.Fprint(w, "User-agent: *\nAllow: /\n")
	})

	ln, err := listen()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Server is running on http://localhost:%d\n", ln.Addr().(*net.TCPAddr).Port)
	server := &http.Server{
		Handler:           withSecurityHeaders(withNoindexHeader(mux)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func databasePath() string {
	if path := os.Getenv("DATABASE_PATH"); path != "" {
		return path
	}
	return "atlas.db"
}

// githubStars returns the repo star count, cached for 15 minutes, so the
// GitHub API sees at most one request per deploy per window no matter how
// much traffic comes in. Errors keep the last known value.
var starsCache struct {
	sync.Mutex
	count   int
	fetched time.Time
}

func githubStars() int {
	starsCache.Lock()
	defer starsCache.Unlock()
	if time.Since(starsCache.fetched) < 15*time.Minute {
		return starsCache.count
	}
	starsCache.fetched = time.Now()
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/axadrn/atlas")
	if err != nil {
		return starsCache.count
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return starsCache.count
	}
	var data struct {
		Stars int `json:"stargazers_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return starsCache.count
	}
	starsCache.count = data.Stars
	return starsCache.count
}

// noindex keeps the site out of search engines while there is no real name
// and no real domain. NOINDEX=0 in the deploy env goes live, until then
// every response carries the header and robots.txt blocks everything.
func noindex() bool {
	v := os.Getenv("NOINDEX")
	return v != "0" && v != "false"
}

func withNoindexHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if noindex() {
			w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		}
		next.ServeHTTP(w, r)
	})
}

// listen binds the port. An explicit PORT binds exactly (the templ proxy in
// the Taskfile points at it, so silently moving would break hot reload);
// without PORT it walks from 8090 to the next free port like node dev
// servers do.
func listen() (net.Listener, error) {
	if env := os.Getenv("PORT"); env != "" {
		port, err := strconv.Atoi(env)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT %q", env)
		}
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return nil, fmt.Errorf("port %d is busy - stop the other app or run with another port, e.g. 'task dev PORT=%d'", port, port+1)
		}
		return ln, nil
	}
	for port := 8090; port < 8100; port++ {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err == nil {
			return ln, nil
		}
	}
	return nil, fmt.Errorf("no free port found from 8090 upwards")
}

func setupAssetsRoutes(mux *http.ServeMux) {
	assetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, ".js"):
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		case strings.HasSuffix(r.URL.Path, ".css"):
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		}
		if os.Getenv("GO_ENV") != "production" {
			w.Header().Set("Cache-Control", "no-store")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=31536000")
		}
		http.FileServer(http.Dir("./assets")).ServeHTTP(w, r)
	})
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assetHandler))

	// The component JS bundle (hashed name plus the shadcn-templ.js alias).
	mux.Handle("GET /components/{bundle}", components.ScriptsHandler())
}
