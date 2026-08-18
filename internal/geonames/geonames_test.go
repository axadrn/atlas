package geonames_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"atlas/internal/database"
	"atlas/internal/geonames"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	statements := []string{
		`INSERT INTO places (id, slug, name, place_type, country_code, is_destination, latitude, longitude, population, curated_rank)
		 VALUES ('plc_cnx', 'chiang-mai-1153671', 'Chiang Mai', 'locality', 'TH', 1, 18.79038, 98.98468, 127000, 1)`,
		`INSERT INTO places (id, slug, name, place_type, country_code, is_destination, population)
		 VALUES ('plc_th', 'th', 'Thailand', 'country', 'TH', 0, 69000000)`,
		`INSERT INTO sources (id, name, homepage_url, license_name) VALUES ('geonames', 'GeoNames', 'https://www.geonames.org/', 'CC BY 4.0')`,
		`INSERT INTO place_sources (place_id, source_id, external_id, contribution, retrieved_at)
		 VALUES ('plc_cnx', 'geonames', '1153671', 'Name, country code, coordinates, population and timezone', '2026-01-01T00:00:00Z')`,
		`INSERT INTO place_sources (place_id, source_id, external_id, contribution, retrieved_at)
		 VALUES ('plc_th', 'geonames', '1605651', 'Name, country code and population', '2026-01-01T00:00:00Z')`,
		`INSERT INTO place_timezones (place_id, timezone_id) VALUES ('plc_cnx', 'Asia/Bangkok')`,
		`INSERT INTO place_timezones (place_id, timezone_id) VALUES ('plc_th', 'Asia/Bangkok')`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("fixture: %v", err)
		}
	}
	return db
}

// dumpLine builds a 19-column GeoNames row with only the fields Atlas reads.
func dumpLine(id, name, lat, lon, country, population, timezone string) string {
	fields := make([]string, 19)
	fields[0] = id
	fields[1] = name
	fields[4] = lat
	fields[5] = lon
	fields[8] = country
	fields[14] = population
	fields[17] = timezone
	return strings.Join(fields, "\t")
}

func TestRefreshUpdatesProviderFacts(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	dump := dumpLine("1153671", "Chiang Mai", "18.80000", "98.99000", "TH", "131091", "Asia/Bangkok")
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)

	stats, err := geonames.Refresh(context.Background(), db, strings.NewReader(dump), now, false)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Scanned != 1 || stats.Matched != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	var population int64
	var latitude float64
	var retrievedAt string
	row := db.QueryRow(`
		SELECT p.population, p.latitude, ps.retrieved_at
		FROM places p JOIN place_sources ps ON ps.place_id = p.id
		WHERE p.id = 'plc_cnx'`)
	if err := row.Scan(&population, &latitude, &retrievedAt); err != nil {
		t.Fatal(err)
	}
	if population != 131091 || latitude != 18.8 {
		t.Fatalf("provider facts not refreshed: population=%d latitude=%v", population, latitude)
	}
	if !strings.HasPrefix(retrievedAt, "2026-08-18T12:00:00") {
		t.Fatalf("retrieved_at not refreshed: %s", retrievedAt)
	}
}

func TestRefreshNeverTouchesCuratedFields(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	// The dump claims a different name and would love to move the place.
	dump := dumpLine("1153671", "Chiang Mai City", "19.1", "99.1", "MM", "1", "Asia/Yangon")

	if _, err := geonames.Refresh(context.Background(), db, strings.NewReader(dump), time.Now(), false); err != nil {
		t.Fatal(err)
	}

	var id, slug, placeType, countryCode string
	var destination, rank int64
	row := db.QueryRow(`
		SELECT id, slug, place_type, country_code, is_destination, curated_rank
		FROM places WHERE slug = 'chiang-mai-1153671'`)
	if err := row.Scan(&id, &slug, &placeType, &countryCode, &destination, &rank); err != nil {
		t.Fatal(err)
	}
	if id != "plc_cnx" || placeType != "locality" || countryCode != "TH" || destination != 1 || rank != 1 {
		t.Fatalf("curated fields changed: id=%s type=%s country=%s destination=%d rank=%d", id, placeType, countryCode, destination, rank)
	}
}

func TestRefreshSkipsUnknownAndNeverInserts(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	dump := dumpLine("9999999", "Nowhere", "1", "1", "XX", "5", "UTC")

	stats, err := geonames.Refresh(context.Background(), db, strings.NewReader(dump), time.Now(), false)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Matched != 0 {
		t.Fatalf("unknown row must not match, got %+v", stats)
	}
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM places`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("refresh must never insert places, have %d", count)
	}
}

func TestRefreshKeepsCountryShapeAndZones(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	dump := dumpLine("1605651", "Thailand", "15.0", "101.0", "TH", "71000000", "Asia/Bangkok")

	if _, err := geonames.Refresh(context.Background(), db, strings.NewReader(dump), time.Now(), false); err != nil {
		t.Fatal(err)
	}

	var population int64
	var latitude sql.NullFloat64
	if err := db.QueryRow(`SELECT population, latitude FROM places WHERE id = 'plc_th'`).Scan(&population, &latitude); err != nil {
		t.Fatal(err)
	}
	if population != 71000000 {
		t.Fatalf("country population should refresh, got %d", population)
	}
	if latitude.Valid {
		t.Fatal("countries must not gain coordinates")
	}
	var zones int
	if err := db.QueryRow(`SELECT count(*) FROM place_timezones WHERE place_id = 'plc_th'`).Scan(&zones); err != nil {
		t.Fatal(err)
	}
	if zones != 1 {
		t.Fatalf("country zones belong to the IANA source and must stay, have %d", zones)
	}
}

func TestRefreshZeroPopulationKeepsKnownValue(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	dump := dumpLine("1153671", "Chiang Mai", "18.8", "98.99", "TH", "0", "Asia/Bangkok")

	if _, err := geonames.Refresh(context.Background(), db, strings.NewReader(dump), time.Now(), false); err != nil {
		t.Fatal(err)
	}
	var population int64
	if err := db.QueryRow(`SELECT population FROM places WHERE id = 'plc_cnx'`).Scan(&population); err != nil {
		t.Fatal(err)
	}
	if population != 127000 {
		t.Fatalf("unknown dump population must not erase a known value, got %d", population)
	}
}

func TestRefreshDryRunWritesNothing(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	dump := dumpLine("1153671", "Chiang Mai", "18.8", "98.99", "TH", "131091", "Asia/Bangkok")

	stats, err := geonames.Refresh(context.Background(), db, strings.NewReader(dump), time.Now(), true)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Matched != 1 {
		t.Fatalf("dry run should still report matches, got %+v", stats)
	}
	var population int64
	if err := db.QueryRow(`SELECT population FROM places WHERE id = 'plc_cnx'`).Scan(&population); err != nil {
		t.Fatal(err)
	}
	if population != 127000 {
		t.Fatalf("dry run must not write, population is %d", population)
	}
}
