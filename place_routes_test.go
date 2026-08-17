package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"atlas/internal/catalog"
	"atlas/internal/database"
)

func TestPlaceRoutes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store := catalog.NewStore(db)

	population := int64(3_426_354)
	place := catalog.Place{
		ID:          "plc_berlin_routes",
		Slug:        "berlin-2950159",
		Kind:        catalog.PlaceKindCity,
		Status:      catalog.PlaceStatusActive,
		CountryCode: "DE",
		Coordinates: &catalog.Coordinates{Latitude: 52.52437, Longitude: 13.41053},
		Timezone:    "Europe/Berlin",
		Population:  &population,
	}
	if err := store.UpsertPlace(ctx, place); err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertPlaceName(ctx, catalog.PlaceName{
		PlaceID:     place.ID,
		LanguageTag: "und",
		Name:        "Berlin",
		Kind:        "common",
		Preferred:   true,
	}); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	setupPlaceRoutes(mux, store)

	t.Run("search", func(t *testing.T) {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/places?q=ber", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
		}
		var body placeResults
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body.Results) != 1 || body.Results[0].Name != "Berlin" {
			t.Fatalf("unexpected body: %#v", body)
		}
	})

	t.Run("map", func(t *testing.T) {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/map/cities?limit=20", nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"latitude":52.52437`) {
			t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
		}
	})

	t.Run("search fragment", func(t *testing.T) {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/fragments/place-search?q=ber", nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "data-tui-combobox-item") {
			t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
		}
	})

	t.Run("place page", func(t *testing.T) {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/places/berlin-2950159", nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "OpenStreetMap contributors") {
			t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
		}
	})

	t.Run("missing place", func(t *testing.T) {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/places/missing", nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", response.Code)
		}
	})
}
