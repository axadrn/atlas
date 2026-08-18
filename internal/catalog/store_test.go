package catalog_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"atlas/internal/catalog"
	"atlas/internal/database"
)

func TestStoreReadsCatalog(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	insertPlace(t, db, "plc_de", "germany", "Germany", "country", "DE", nil, nil, nil, 84_000_000)
	insertPlace(t, db, "plc_berlin", "berlin-2950159", "Berlin", "destination", "DE", 52.52437, 13.41053, "Europe/Berlin", 3_426_354)
	store := catalog.NewStore(db)

	results, err := store.SearchPlaces(ctx, "ber", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != "Berlin" {
		t.Fatalf("unexpected search results: %#v", results)
	}

	mapPlaces, err := store.MapPlaces(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(mapPlaces) != 1 || mapPlaces[0].Coordinates == nil {
		t.Fatalf("unexpected map places: %#v", mapPlaces)
	}

	destinations, err := store.DestinationsByCountry(ctx, "DE", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(destinations) != 1 || destinations[0].Name != "Berlin" {
		t.Fatalf("unexpected destinations: %#v", destinations)
	}

	country, err := store.CountryByCode(ctx, "DE")
	if err != nil {
		t.Fatal(err)
	}
	if country.Name != "Germany" || country.Slug != "germany" {
		t.Fatalf("unexpected country: %#v", country)
	}

	place, err := store.PlaceBySlug(ctx, "berlin-2950159")
	if err != nil {
		t.Fatal(err)
	}
	if place.ID != "plc_berlin" || place.Timezone != "Europe/Berlin" {
		t.Fatalf("unexpected place: %#v", place)
	}

	_, err = store.PlaceBySlug(ctx, "missing")
	if !errors.Is(err, catalog.ErrPlaceNotFound) {
		t.Fatalf("expected ErrPlaceNotFound, got %v", err)
	}

	_, err = store.CountryByCode(ctx, "XX")
	if !errors.Is(err, catalog.ErrPlaceNotFound) {
		t.Fatalf("expected ErrPlaceNotFound, got %v", err)
	}
}

func TestStoreReadsPlaceSources(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	insertPlace(t, db, "plc_berlin", "berlin-2950159", "Berlin", "destination", "DE", 52.52437, 13.41053, "Europe/Berlin", 3_426_354)
	if _, err := db.ExecContext(ctx, `
		INSERT INTO sources (id, name, homepage_url, license_name, license_url)
		VALUES ('geonames', 'GeoNames', 'https://www.geonames.org/', 'CC BY 4.0', 'https://creativecommons.org/licenses/by/4.0/')
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO place_sources (
			place_id, source_id, external_id, record_url, contribution, retrieved_at
		) VALUES (
			'plc_berlin', 'geonames', '2950159', 'https://www.geonames.org/2950159',
			'Name, country code, coordinates, population and timezone', '2026-08-17T12:00:00Z'
		)
	`); err != nil {
		t.Fatal(err)
	}

	sources, err := catalog.NewStore(db).SourcesForPlace(ctx, "plc_berlin")
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 || sources[0].ExternalID != "2950159" || sources[0].Contribution == "" {
		t.Fatalf("unexpected sources: %#v", sources)
	}
}

func TestStoreDerivesDateLineSafeCountryBounds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	insertPlace(t, db, "plc_fj", "fiji", "Fiji", "country", "FJ", nil, nil, nil, 930_000)
	insertPlace(t, db, "plc_suva", "suva", "Suva", "destination", "FJ", -18.1416, 178.4419, "Pacific/Fiji", 93_970)
	insertPlace(t, db, "plc_levuka", "levuka", "Levuka", "destination", "FJ", -17.6833, -179.3, "Pacific/Fiji", 1_130)

	place, err := catalog.NewStore(db).PlaceBySlug(ctx, "fiji")
	if err != nil {
		t.Fatal(err)
	}
	if place.Bounds == nil || place.Bounds.East-place.Bounds.West > 5 {
		t.Fatalf("expected date-line-safe bounds, got %#v", place.Bounds)
	}
}

func insertPlace(t *testing.T, db interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, id, slug, name, kind, countryCode string, latitude, longitude, timezone any, population int64) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
		INSERT INTO places (
			id, slug, name, kind, country_code, latitude, longitude, timezone, population
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, slug, name, kind, countryCode, latitude, longitude, timezone, population); err != nil {
		t.Fatal(err)
	}
}
