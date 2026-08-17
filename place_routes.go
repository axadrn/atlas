package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/a-h/templ"

	"atlas/components/placesearch"
	"atlas/internal/catalog"
	"atlas/pages"
)

type placeResults struct {
	Results []catalog.PlaceSummary `json:"results"`
}

func setupPlaceRoutes(mux *http.ServeMux, store *catalog.Store) {
	mux.HandleFunc("GET /api/places", func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if utf8.RuneCountInString(query) > 100 {
			writeAPIError(w, http.StatusBadRequest, "query must be 100 characters or fewer")
			return
		}
		results, err := store.SearchPlaces(r.Context(), query, 10)
		if err != nil {
			log.Printf("place search: %v", err)
			writeAPIError(w, http.StatusInternalServerError, "place search failed")
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=60")
		writeJSON(w, http.StatusOK, placeResults{Results: results})
	})

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
		templ.Handler(placesearch.Results(results)).ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /api/map/cities", func(w http.ResponseWriter, r *http.Request) {
		limit := 150
		if value := r.URL.Query().Get("limit"); value != "" {
			parsed, err := strconv.Atoi(value)
			if err != nil {
				writeAPIError(w, http.StatusBadRequest, "limit must be a number")
				return
			}
			limit = parsed
		}
		results, err := store.MapCities(r.Context(), limit)
		if err != nil {
			log.Printf("map cities: %v", err)
			writeAPIError(w, http.StatusInternalServerError, "map data failed")
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=900")
		writeJSON(w, http.StatusOK, placeResults{Results: results})
	})

	mux.HandleFunc("GET /places/{slug}", func(w http.ResponseWriter, r *http.Request) {
		place, err := store.PlaceBySlug(r.Context(), r.PathValue("slug"))
		if errors.Is(err, catalog.ErrPlaceNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Printf("place page: %v", err)
			http.Error(w, "Could not load this place.", http.StatusInternalServerError)
			return
		}
		templ.Handler(pages.Place(place)).ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
