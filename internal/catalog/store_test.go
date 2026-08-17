package catalog_test

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"atlas/internal/catalog"
	"atlas/internal/database"
)

func TestStoreUpsertsAndReadsPlace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store := catalog.NewStore(db)

	population := int64(3_897_145)
	berlin := catalog.Place{
		ID:          "plc_berlin",
		Slug:        "berlin-germany",
		Kind:        catalog.PlaceKindCity,
		Status:      catalog.PlaceStatusActive,
		CountryCode: "DE",
		Coordinates: &catalog.Coordinates{Latitude: 52.52, Longitude: 13.405},
		Timezone:    "Europe/Berlin",
		Population:  &population,
	}
	if err := store.UpsertPlace(ctx, berlin); err != nil {
		t.Fatal(err)
	}

	berlin.Population = nil
	if err := store.UpsertPlace(ctx, berlin); err != nil {
		t.Fatal(err)
	}

	got, err := store.PlaceByID(ctx, berlin.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Slug != berlin.Slug || got.Kind != berlin.Kind || got.CountryCode != "DE" {
		t.Fatalf("unexpected place: %#v", got)
	}
	if got.Population != nil {
		t.Fatalf("expected population to be cleared, got %d", *got.Population)
	}
	if got.Coordinates == nil || got.Coordinates.Latitude != 52.52 || got.Coordinates.Longitude != 13.405 {
		t.Fatalf("unexpected coordinates: %#v", got.Coordinates)
	}
}

func TestStoreValidatesPlace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store := catalog.NewStore(db)

	err = store.UpsertPlace(ctx, catalog.Place{
		ID:          "plc_invalid",
		Slug:        "Invalid Slug",
		Kind:        catalog.PlaceKindCity,
		Status:      catalog.PlaceStatusActive,
		CountryCode: "de",
	})
	if err == nil {
		t.Fatal("expected invalid place to fail")
	}
}

func TestStoreReturnsNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store := catalog.NewStore(db)

	_, err = store.PlaceByID(ctx, "plc_missing")
	if !errors.Is(err, catalog.ErrPlaceNotFound) {
		t.Fatalf("expected ErrPlaceNotFound, got %v", err)
	}
}

func TestStoreSearchesAndListsPlaces(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store := catalog.NewStore(db)

	population := int64(3_426_354)
	berlin := catalog.Place{
		ID:          "plc_berlin_search",
		Slug:        "berlin-2950159",
		Kind:        catalog.PlaceKindCity,
		Status:      catalog.PlaceStatusActive,
		CountryCode: "DE",
		Coordinates: &catalog.Coordinates{Latitude: 52.52437, Longitude: 13.41053},
		Timezone:    "Europe/Berlin",
		Population:  &population,
	}
	if err := store.UpsertPlace(ctx, berlin); err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertPlaceName(ctx, catalog.PlaceName{
		PlaceID:     berlin.ID,
		LanguageTag: "und",
		Name:        "Berlin",
		Kind:        "common",
		Preferred:   true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertPlaceName(ctx, catalog.PlaceName{
		PlaceID:     berlin.ID,
		LanguageTag: "de",
		Name:        "Bundeshauptstadt Berlin",
		Kind:        "alternate",
	}); err != nil {
		t.Fatal(err)
	}

	results, err := store.SearchPlaces(ctx, "Bundeshauptstadt", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != "Berlin" || results[0].CountryCode != "DE" {
		t.Fatalf("unexpected search results: %#v", results)
	}

	mapCities, err := store.MapCities(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(mapCities) != 1 || mapCities[0].Coordinates == nil {
		t.Fatalf("unexpected map cities: %#v", mapCities)
	}

	got, err := store.PlaceBySlug(ctx, berlin.Slug)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != berlin.ID || got.Timezone != "Europe/Berlin" {
		t.Fatalf("unexpected place: %#v", got)
	}

	_, err = store.PlaceBySlug(ctx, "missing")
	if !errors.Is(err, catalog.ErrPlaceNotFound) {
		t.Fatalf("expected ErrPlaceNotFound, got %v", err)
	}
}

func TestStoreReplacesProviderReferenceWithoutLosingItOnConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store := catalog.NewStore(db)

	for _, place := range []catalog.Place{
		{ID: "plc_first", Slug: "first", Kind: catalog.PlaceKindCity, Status: catalog.PlaceStatusActive},
		{ID: "plc_second", Slug: "second", Kind: catalog.PlaceKindCity, Status: catalog.PlaceStatusActive},
	} {
		if err := store.UpsertPlace(ctx, place); err != nil {
			t.Fatal(err)
		}
	}
	for _, ref := range []catalog.ExternalReference{
		{Provider: "example", ExternalID: "old", PlaceID: "plc_first"},
		{Provider: "example", ExternalID: "taken", PlaceID: "plc_second"},
	} {
		if err := store.UpsertExternalReference(ctx, ref); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.UpsertExternalReference(ctx, catalog.ExternalReference{
		Provider: "example", ExternalID: "new", PlaceID: "plc_first",
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertExternalReference(ctx, catalog.ExternalReference{
		Provider: "example", ExternalID: "taken", PlaceID: "plc_first",
	}); err == nil {
		t.Fatal("expected a provider ID collision")
	}

	var externalID string
	if err := db.QueryRowContext(ctx, `
		SELECT external_id FROM external_references
		WHERE place_id = 'plc_first' AND provider = 'example'
	`).Scan(&externalID); err != nil {
		t.Fatal(err)
	}
	if externalID != "new" {
		t.Fatalf("expected the last valid provider ID, got %q", externalID)
	}
}

func TestStoreDerivesCountryMapDataFromCities(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store := catalog.NewStore(db)

	fiji := catalog.Place{
		ID:          "plc_fiji_map",
		Slug:        "fiji",
		Kind:        catalog.PlaceKindCountry,
		Status:      catalog.PlaceStatusActive,
		CountryCode: "FJ",
	}
	if err := store.UpsertPlace(ctx, fiji); err != nil {
		t.Fatal(err)
	}
	if err := store.UpsertPlaceName(ctx, catalog.PlaceName{
		PlaceID: fiji.ID, LanguageTag: "en", Name: "Fiji", Kind: "common", Preferred: true,
	}); err != nil {
		t.Fatal(err)
	}
	for index, city := range []catalog.Place{
		{
			ID: "plc_suva", Slug: "suva", Kind: catalog.PlaceKindCity, Status: catalog.PlaceStatusActive,
			CountryCode: "FJ", Coordinates: &catalog.Coordinates{Latitude: -18.1416, Longitude: 178.4419}, Timezone: "Pacific/Fiji",
		},
		{
			ID: "plc_levuka", Slug: "levuka", Kind: catalog.PlaceKindCity, Status: catalog.PlaceStatusActive,
			CountryCode: "FJ", Coordinates: &catalog.Coordinates{Latitude: -17.6833, Longitude: -179.3000}, Timezone: "Pacific/Fiji",
		},
	} {
		if err := store.UpsertPlace(ctx, city); err != nil {
			t.Fatal(err)
		}
		if err := store.UpsertPlaceName(ctx, catalog.PlaceName{
			PlaceID: city.ID, LanguageTag: "en", Name: fmt.Sprintf("City %d", index), Kind: "common", Preferred: true,
		}); err != nil {
			t.Fatal(err)
		}
	}

	got, err := store.PlaceBySlug(ctx, fiji.Slug)
	if err != nil {
		t.Fatal(err)
	}
	if got.Bounds == nil || got.Bounds.East-got.Bounds.West > 5 {
		t.Fatalf("expected date-line-safe city bounds, got %#v", got.Bounds)
	}
	if len(got.Timezones) != 1 || got.Timezones[0] != "Pacific/Fiji" {
		t.Fatalf("unexpected timezones: %#v", got.Timezones)
	}
}
