package main

import (
	"context"
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

	if _, err := db.ExecContext(ctx, `
		INSERT INTO places (
			id, slug, name, place_type, parent_id, country_code, is_destination,
			latitude, longitude, population
		) VALUES (
			'plc_germany_routes', 'germany', 'Germany', 'country', NULL, 'DE', 0,
			NULL, NULL, 84000000
		);
		INSERT INTO places (
			id, slug, name, place_type, parent_id, country_code, is_destination,
			latitude, longitude, population
		) VALUES (
			'plc_berlin_routes', 'berlin-2950159', 'Berlin', 'locality',
			'plc_germany_routes', 'DE', 1, 52.52437, 13.41053, 3426354
		);
		INSERT INTO place_timezones (place_id, timezone_id)
		VALUES
			('plc_germany_routes', 'Europe/Berlin'),
			('plc_berlin_routes', 'Europe/Berlin');
		INSERT INTO places (
			id, slug, name, place_type, parent_id, country_code, is_destination,
			latitude, longitude, population
		) VALUES (
			'plc_mitte_routes', 'mitte-berlin', 'Mitte', 'neighborhood',
			'plc_berlin_routes', 'DE', 0, 52.52, 13.405, NULL
		);
		INSERT INTO sources (id, name, homepage_url, license_name, license_url)
		VALUES ('geonames', 'GeoNames', 'https://www.geonames.org/', 'CC BY 4.0', 'https://creativecommons.org/licenses/by/4.0/');
		INSERT INTO place_sources (
			place_id, source_id, external_id, record_url, contribution, retrieved_at
		) VALUES (
			'plc_berlin_routes', 'geonames', '2950159', 'https://www.geonames.org/2950159',
			'Name, country code, coordinates, population and timezone', '2026-08-17T12:00:00Z'
		);
	`); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	setupPlaceRoutes(mux, store, func() int { return 0 })

	t.Run("unversioned JSON APIs stay private", func(t *testing.T) {
		for _, path := range []string{"/api/places?q=ber", "/api/map/places?limit=20"} {
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusNotFound {
				t.Fatalf("expected %s to return 404, got %d", path, response.Code)
			}
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
		if response.Code != http.StatusOK ||
			!strings.Contains(response.Body.String(), "OpenStreetMap contributors") ||
			!strings.Contains(response.Body.String(), "data-place-map") ||
			!strings.Contains(response.Body.String(), "data-map-canvas") ||
			!strings.Contains(response.Body.String(), "/assets/vendor/maplibre-gl-csp-5.24.0.js") ||
			!strings.Contains(response.Body.String(), "/assets/js/place-map.js") ||
			!strings.Contains(response.Body.String(), `aria-controls="place-sources"`) ||
			!strings.Contains(response.Body.String(), `data-slot="sheet-content"`) ||
			!strings.Contains(response.Body.String(), `role="dialog"`) ||
			!strings.Contains(response.Body.String(), `href="/places/germany"`) ||
			!strings.Contains(response.Body.String(), `href="/places/mitte-berlin"`) ||
			!strings.Contains(response.Body.String(), "Neighborhoods") ||
			!strings.Contains(response.Body.String(), "Germany") ||
			!strings.Contains(response.Body.String(), "Contribution") ||
			!strings.Contains(response.Body.String(), "Source record 2950159") ||
			!strings.Contains(response.Body.String(), "GeoNames") ||
			!strings.Contains(response.Body.String(), "CC BY 4.0") {
			t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
		}
	})

	t.Run("neighborhood page", func(t *testing.T) {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/places/mitte-berlin", nil))
		if response.Code != http.StatusOK ||
			!strings.Contains(response.Body.String(), `href="/places/germany"`) ||
			!strings.Contains(response.Body.String(), `href="/places/berlin-2950159"`) ||
			!strings.Contains(response.Body.String(), "Europe/Berlin") ||
			!strings.Contains(response.Body.String(), "Neighborhood") {
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
