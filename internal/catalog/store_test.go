package catalog_test

import (
	"context"
	"errors"
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
