package geonames_test

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"atlas/internal/catalog"
	"atlas/internal/database"
	"atlas/internal/importer/geonames"
)

func TestImportIsAtomicAndIdempotent(t *testing.T) {
	t.Parallel()
	countryBody := strings.Join([]string{
		countryLine("DE", "Germany", "EU", "2921044", "83536115"),
		countryLine("JP", "Japan", "AS", "1861060", "124214766"),
	}, "\n")
	cityBody := citiesArchive(t, strings.Join([]string{
		cityLine("2950159", "Berlin", "Berlin", "52.52437", "13.41053", "DE", "3426354", "Europe/Berlin"),
		cityLine("1850147", "Tokyo", "Tokyo", "35.6895", "139.69171", "JP", "8336599", "Asia/Tokyo"),
		cityLine("2911285", "Wandsbek", "Wandsbek", "53.58334", "10.08305", "DE", "411422", "Europe/Berlin", "PPLX"),
	}, "\n"))

	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	options := geonames.Options{
		CountryInfoURL: "https://example.test/countryInfo.txt",
		CitiesURL:      "https://example.test/cities5000.zip",
		HTTPClient:     dataClient([]byte(countryBody), cityBody),
		Now: func() time.Time {
			return time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)
		},
	}

	for run := 1; run <= 2; run++ {
		stats, err := geonames.Import(ctx, db, options)
		if err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		if stats.Countries != 2 || stats.Cities != 2 || stats.Skipped != 1 {
			t.Fatalf("run %d: unexpected stats %#v", run, stats)
		}
	}

	assertCount(t, db, "places", 12)
	assertCount(t, db, "external_references", 4)
	assertCount(t, db, "source_snapshots", 2)

	var name, countryCode, timezone string
	var latitude, longitude float64
	err = db.QueryRowContext(ctx, `
		SELECT pn.name, p.country_code, p.timezone, p.latitude, p.longitude
		FROM external_references er
		JOIN places p ON p.id = er.place_id
		JOIN place_names pn ON pn.place_id = p.id AND pn.is_preferred = 1
		WHERE er.provider = 'geonames' AND er.external_id = '2950159'
	`).Scan(&name, &countryCode, &timezone, &latitude, &longitude)
	if err != nil {
		t.Fatal(err)
	}
	if name != "Berlin" || countryCode != "DE" || timezone != "Europe/Berlin" ||
		latitude != 52.52437 || longitude != 13.41053 {
		t.Fatalf("unexpected imported city: %q %q %q %f %f", name, countryCode, timezone, latitude, longitude)
	}
}

func TestImportRollsBackMalformedData(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	_, err = geonames.Import(ctx, db, geonames.Options{
		CountryInfoURL: "https://example.test/countryInfo.txt",
		CitiesURL:      "https://example.test/cities5000.zip",
		HTTPClient: dataClient(
			[]byte(countryLine("DE", "Germany", "EU", "2921044", "83536115")),
			citiesArchive(t, "not-enough-fields"),
		),
	})
	if err == nil {
		t.Fatal("expected malformed import to fail")
	}
	assertCount(t, db, "places", 0)
	assertCount(t, db, "sources", 0)
}

func TestImportUnlistsPlacesMissingFromTheCurrentCatalog(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "atlas.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	countryBody := []byte(countryLine("DE", "Germany", "EU", "2921044", "83536115"))
	now := func() time.Time { return time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC) }
	firstCities := citiesArchive(t, strings.Join([]string{
		cityLine("2950159", "Berlin", "Berlin", "52.52437", "13.41053", "DE", "3426354", "Europe/Berlin"),
		cityLine("2911285", "Wandsbek", "Wandsbek", "53.58334", "10.08305", "DE", "411422", "Europe/Berlin"),
	}, "\n"))
	if _, err := geonames.Import(ctx, db, geonames.Options{
		CountryInfoURL: "https://example.test/countryInfo.txt",
		CitiesURL:      "https://example.test/cities5000.zip",
		HTTPClient:     dataClient(countryBody, firstCities),
		Now:            now,
	}); err != nil {
		t.Fatal(err)
	}

	secondCities := citiesArchive(t, strings.Join([]string{
		cityLine("2950159", "Berlin", "Berlin", "52.52437", "13.41053", "DE", "3426354", "Europe/Berlin"),
		cityLine("2911285", "Wandsbek", "Wandsbek", "53.58334", "10.08305", "DE", "411422", "Europe/Berlin", "PPLX"),
	}, "\n"))
	if _, err := geonames.Import(ctx, db, geonames.Options{
		CountryInfoURL: "https://example.test/countryInfo.txt",
		CitiesURL:      "https://example.test/cities5000.zip",
		HTTPClient:     dataClient(countryBody, secondCities),
		Now:            now,
	}); err != nil {
		t.Fatal(err)
	}

	var listed int
	if err := db.QueryRowContext(ctx, `
		SELECT is_listed
		FROM places p
		JOIN external_references er ON er.place_id = p.id
		WHERE er.provider = 'geonames' AND er.external_id = '2911285'
	`).Scan(&listed); err != nil {
		t.Fatal(err)
	}
	if listed != 0 {
		t.Fatalf("expected a reclassified city section to be unlisted, got %d", listed)
	}
	results, err := catalog.NewStore(db).SearchPlaces(ctx, "Wandsbek", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected the unlisted city section to stay out of discovery, got %#v", results)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func dataClient(countryBody []byte, cityBody []byte) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body []byte
		switch request.URL.Path {
		case "/countryInfo.txt":
			body = countryBody
		case "/cities5000.zip":
			body = cityBody
		default:
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Status:     "404 Not Found",
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("not found")),
				Request:    request,
			}, nil
		}
		header := make(http.Header)
		header.Set("Last-Modified", "Mon, 17 Aug 2026 03:49:00 GMT")
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     header,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    request,
		}, nil
	})}
}

func countryLine(code string, name string, continent string, geonameID string, population string) string {
	fields := []string{
		code, "XXX", "000", code, name, "Capital", "1", population,
		continent, ".xx", "XXX", "Currency", "1", "", "", "en", geonameID, "", "",
	}
	return strings.Join(fields, "\t")
}

func cityLine(
	id string,
	name string,
	asciiName string,
	latitude string,
	longitude string,
	countryCode string,
	population string,
	timezone string,
	featureCodes ...string,
) string {
	featureCode := "PPLA"
	if len(featureCodes) > 0 {
		featureCode = featureCodes[0]
	}
	fields := []string{
		id, name, asciiName, "", latitude, longitude, "P", featureCode, countryCode,
		"", "", "", "", "", population, "", "", timezone, "2026-08-17",
	}
	return strings.Join(fields, "\t")
}

func citiesArchive(t *testing.T, contents string) []byte {
	t.Helper()
	var body bytes.Buffer
	archive := zip.NewWriter(&body)
	file, err := archive.Create("cities5000.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte(contents)); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return body.Bytes()
}

func assertCount(t *testing.T, db interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, table string, want int) {
	t.Helper()
	var got int
	if err := db.QueryRowContext(context.Background(), "SELECT count(*) FROM "+table).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("expected %d %s records, got %d", want, table, got)
	}
}
