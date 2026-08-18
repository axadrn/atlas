package main

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/a-h/templ"

	"atlas/internal/catalog"
	"atlas/pages"
)

func setupPlaceRoutes(mux *http.ServeMux, store *catalog.Store) {
	mux.HandleFunc("GET /fragments/place-search", func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if utf8.RuneCountInString(query) > 100 {
			http.Error(w, "Query must be 100 characters or fewer.", http.StatusBadRequest)
			return
		}
		results, err := store.SearchPlaces(r.Context(), query, 10)
		if err != nil {
			log.Printf("place search fragment: %v", err)
			http.Error(w, "Place search failed.", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=60")
		templ.Handler(pages.PlaceSearchResults(results)).ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /places/{slug}", func(w http.ResponseWriter, r *http.Request) {
		place, err := store.PlaceBySlug(r.Context(), r.PathValue("slug"))
		if errors.Is(err, catalog.ErrPlaceNotFound) {
			renderNotFound(w, r)
			return
		}
		if err != nil {
			log.Printf("place page: %v", err)
			http.Error(w, "Could not load this place.", http.StatusInternalServerError)
			return
		}
		sources, err := store.SourcesForPlace(r.Context(), place.ID)
		if err != nil {
			log.Printf("place sources: %v", err)
			http.Error(w, "Could not load this place.", http.StatusInternalServerError)
			return
		}
		ancestors, err := store.AncestorsForPlace(r.Context(), place)
		if err != nil {
			log.Printf("place ancestors: %v", err)
			http.Error(w, "Could not load this place.", http.StatusInternalServerError)
			return
		}
		children, err := store.ChildrenByParent(r.Context(), place.ID, 50)
		if err != nil {
			log.Printf("place children: %v", err)
			http.Error(w, "Could not load this place.", http.StatusInternalServerError)
			return
		}
		templ.Handler(pages.Place(place, ancestors, sources, children)).ServeHTTP(w, r)
	})
}
