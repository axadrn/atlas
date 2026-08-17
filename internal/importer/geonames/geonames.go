package geonames

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"atlas/internal/catalog"
)

const (
	DefaultCountryInfoURL = "https://download.geonames.org/export/dump/countryInfo.txt"
	DefaultCitiesURL      = "https://download.geonames.org/export/dump/cities5000.zip"

	geonamesSourceID = "src_geonames"
	maxCountryBytes  = 5 << 20
	maxCitiesBytes   = 100 << 20
)

type Options struct {
	CountryInfoURL string
	CitiesURL      string
	HTTPClient     *http.Client
	Now            func() time.Time
}

type Stats struct {
	Countries int
	Cities    int
	Skipped   int
}

type download struct {
	body        []byte
	originURL   string
	version     string
	retrievedAt time.Time
	checksum    string
	snapshotID  string
}

type countryRecord struct {
	code       string
	name       string
	continent  string
	geonameID  string
	population *int64
}

var continents = map[string]string{
	"AF": "Africa",
	"AN": "Antarctica",
	"AS": "Asia",
	"EU": "Europe",
	"NA": "North America",
	"OC": "Oceania",
	"SA": "South America",
}

func Import(ctx context.Context, db *sql.DB, options Options) (Stats, error) {
	options = options.withDefaults()

	countryDownload, err := fetch(ctx, options.HTTPClient, options.CountryInfoURL, maxCountryBytes, options.Now)
	if err != nil {
		return Stats{}, fmt.Errorf("download GeoNames countries: %w", err)
	}
	citiesDownload, err := fetch(ctx, options.HTTPClient, options.CitiesURL, maxCitiesBytes, options.Now)
	if err != nil {
		return Stats{}, fmt.Errorf("download GeoNames cities: %w", err)
	}
	countries, err := parseCountries(countryDownload.body)
	if err != nil {
		return Stats{}, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Stats{}, fmt.Errorf("begin GeoNames import: %w", err)
	}
	defer tx.Rollback()
	store := catalog.NewTxStore(tx)

	if err := addSource(ctx, store, countryDownload, citiesDownload); err != nil {
		return Stats{}, err
	}
	countryIDs, err := importCountries(ctx, store, countries, countryDownload.snapshotID)
	if err != nil {
		return Stats{}, err
	}

	stats := Stats{Countries: len(countries)}
	stats.Cities, stats.Skipped, err = importCities(
		ctx,
		store,
		citiesDownload.body,
		citiesDownload.snapshotID,
		countryIDs,
	)
	if err != nil {
		return Stats{}, err
	}
	if err := tx.Commit(); err != nil {
		return Stats{}, fmt.Errorf("commit GeoNames import: %w", err)
	}
	return stats, nil
}

func (o Options) withDefaults() Options {
	if o.CountryInfoURL == "" {
		o.CountryInfoURL = DefaultCountryInfoURL
	}
	if o.CitiesURL == "" {
		o.CitiesURL = DefaultCitiesURL
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: 90 * time.Second}
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	return o
}

func fetch(
	ctx context.Context,
	client *http.Client,
	originURL string,
	maxBytes int64,
	now func() time.Time,
) (download, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, originURL, nil)
	if err != nil {
		return download{}, err
	}
	req.Header.Set("User-Agent", "atlas-geonames-importer/1.0 (+https://github.com/axadrn/atlas)")
	resp, err := client.Do(req)
	if err != nil {
		return download{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return download{}, fmt.Errorf("%s returned %s", originURL, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return download{}, err
	}
	if int64(len(body)) > maxBytes {
		return download{}, fmt.Errorf("%s exceeds the %d byte limit", originURL, maxBytes)
	}
	sum := sha256.Sum256(body)
	checksum := hex.EncodeToString(sum[:])
	version := resp.Header.Get("Last-Modified")
	if version == "" {
		version = resp.Header.Get("ETag")
	}
	return download{
		body:        body,
		originURL:   originURL,
		version:     version,
		retrievedAt: now().UTC(),
		checksum:    checksum,
		snapshotID:  stableID("snp", "geonames", checksum),
	}, nil
}

func parseCountries(body []byte) ([]countryRecord, error) {
	var countries []countryRecord
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 17 {
			return nil, fmt.Errorf("parse GeoNames countries line %d: expected at least 17 fields, got %d", lineNumber, len(fields))
		}
		population, err := optionalPositiveInt64(fields[7])
		if err != nil {
			return nil, fmt.Errorf("parse GeoNames countries line %d population: %w", lineNumber, err)
		}
		if fields[0] == "" || fields[4] == "" || fields[8] == "" || fields[16] == "" {
			return nil, fmt.Errorf("parse GeoNames countries line %d: required field is empty", lineNumber)
		}
		countries = append(countries, countryRecord{
			code:       fields[0],
			name:       fields[4],
			continent:  fields[8],
			geonameID:  fields[16],
			population: population,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read GeoNames countries: %w", err)
	}
	if len(countries) == 0 {
		return nil, errors.New("GeoNames country data is empty")
	}
	return countries, nil
}

func addSource(ctx context.Context, store *catalog.Store, downloads ...download) error {
	if err := store.UpsertSource(ctx, catalog.Source{
		ID:          geonamesSourceID,
		Name:        "GeoNames",
		Publisher:   "GeoNames",
		HomepageURL: "https://www.geonames.org/",
		LicenseName: "CC BY 4.0",
		LicenseURL:  "https://creativecommons.org/licenses/by/4.0/",
	}); err != nil {
		return err
	}
	for _, item := range downloads {
		if err := store.AddSourceSnapshot(ctx, catalog.SourceSnapshot{
			ID:             item.snapshotID,
			SourceID:       geonamesSourceID,
			Version:        item.version,
			RetrievedAt:    item.retrievedAt.Format(time.RFC3339Nano),
			ChecksumSHA256: item.checksum,
			OriginURL:      item.originURL,
		}); err != nil {
			return err
		}
	}
	return nil
}

func importCountries(
	ctx context.Context,
	store *catalog.Store,
	countries []countryRecord,
	snapshotID string,
) (map[string]string, error) {
	worldID := stableID("plc", "atlas", "world")
	if err := upsertNamedPlace(ctx, store, catalog.Place{
		ID:     worldID,
		Slug:   "world",
		Kind:   catalog.PlaceKindWorld,
		Status: catalog.PlaceStatusActive,
	}, "en", "World", ""); err != nil {
		return nil, err
	}

	continentIDs := make(map[string]string, len(continents))
	for code, name := range continents {
		id := stableID("plc", "geonames-continent", code)
		continentIDs[code] = id
		if err := upsertNamedPlace(ctx, store, catalog.Place{
			ID:       id,
			Slug:     slugify(name),
			Kind:     catalog.PlaceKindContinent,
			Status:   catalog.PlaceStatusActive,
			ParentID: worldID,
		}, "en", name, ""); err != nil {
			return nil, err
		}
	}

	countryIDs := make(map[string]string, len(countries))
	for _, country := range countries {
		continentID, ok := continentIDs[country.continent]
		if !ok {
			return nil, fmt.Errorf("country %s has unknown continent %q", country.code, country.continent)
		}
		id := stableID("plc", "geonames", country.geonameID)
		countryIDs[country.code] = id
		if err := upsertNamedPlace(ctx, store, catalog.Place{
			ID:          id,
			Slug:        strings.ToLower(country.code),
			Kind:        catalog.PlaceKindCountry,
			Status:      catalog.PlaceStatusActive,
			ParentID:    continentID,
			CountryCode: country.code,
			Population:  country.population,
		}, "en", country.name, ""); err != nil {
			return nil, err
		}
		if err := store.UpsertExternalReference(ctx, catalog.ExternalReference{
			Provider:         "geonames",
			ExternalID:       country.geonameID,
			PlaceID:          id,
			SourceSnapshotID: snapshotID,
		}); err != nil {
			return nil, err
		}
	}
	return countryIDs, nil
}

func importCities(
	ctx context.Context,
	store *catalog.Store,
	body []byte,
	snapshotID string,
	countryIDs map[string]string,
) (imported int, skipped int, err error) {
	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return 0, 0, fmt.Errorf("open GeoNames cities archive: %w", err)
	}
	var cityFile *zip.File
	for _, file := range archive.File {
		if file.Name == "cities5000.txt" {
			cityFile = file
			break
		}
	}
	if cityFile == nil {
		return 0, 0, errors.New("GeoNames cities archive does not contain cities5000.txt")
	}
	reader, err := cityFile.Open()
	if err != nil {
		return 0, 0, fmt.Errorf("open GeoNames cities file: %w", err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) < 19 {
			return 0, 0, fmt.Errorf("parse GeoNames cities line %d: expected at least 19 fields, got %d", lineNumber, len(fields))
		}
		if fields[6] != "P" {
			skipped++
			continue
		}
		countryID, ok := countryIDs[fields[8]]
		if !ok {
			skipped++
			continue
		}
		latitude, err := strconv.ParseFloat(fields[4], 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse GeoNames cities line %d latitude: %w", lineNumber, err)
		}
		longitude, err := strconv.ParseFloat(fields[5], 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse GeoNames cities line %d longitude: %w", lineNumber, err)
		}
		population, err := optionalPositiveInt64(fields[14])
		if err != nil {
			return 0, 0, fmt.Errorf("parse GeoNames cities line %d population: %w", lineNumber, err)
		}
		geonameID := fields[0]
		name := fields[1]
		asciiName := fields[2]
		if geonameID == "" || name == "" {
			return 0, 0, fmt.Errorf("parse GeoNames cities line %d: required field is empty", lineNumber)
		}
		slugBase := slugify(asciiName)
		if slugBase == "" {
			slugBase = "city"
		}
		id := stableID("plc", "geonames", geonameID)
		if err := upsertNamedPlace(ctx, store, catalog.Place{
			ID:          id,
			Slug:        slugBase + "-" + geonameID,
			Kind:        catalog.PlaceKindCity,
			Status:      catalog.PlaceStatusActive,
			ParentID:    countryID,
			CountryCode: fields[8],
			Coordinates: &catalog.Coordinates{Latitude: latitude, Longitude: longitude},
			Timezone:    fields[17],
			Population:  population,
		}, "und", name, asciiName); err != nil {
			return 0, 0, fmt.Errorf("import GeoNames city %s: %w", geonameID, err)
		}
		if err := store.UpsertExternalReference(ctx, catalog.ExternalReference{
			Provider:         "geonames",
			ExternalID:       geonameID,
			PlaceID:          id,
			SourceSnapshotID: snapshotID,
		}); err != nil {
			return 0, 0, err
		}
		imported++
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("read GeoNames cities: %w", err)
	}
	return imported, skipped, nil
}

func upsertNamedPlace(
	ctx context.Context,
	store *catalog.Store,
	place catalog.Place,
	languageTag string,
	name string,
	alternateName string,
) error {
	if err := store.UpsertPlace(ctx, place); err != nil {
		return err
	}
	if err := store.UpsertPlaceName(ctx, catalog.PlaceName{
		PlaceID:     place.ID,
		LanguageTag: languageTag,
		Name:        name,
		Kind:        "common",
		Preferred:   true,
	}); err != nil {
		return err
	}
	if alternateName != "" && alternateName != name {
		if err := store.UpsertPlaceName(ctx, catalog.PlaceName{
			PlaceID:     place.ID,
			LanguageTag: "und-Latn",
			Name:        alternateName,
			Kind:        "alternate",
		}); err != nil {
			return err
		}
	}
	return nil
}

func optionalPositiveInt64(value string) (*int64, error) {
	if value == "" || value == "0" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	if parsed < 0 {
		return nil, errors.New("value cannot be negative")
	}
	return &parsed, nil
}

func stableID(prefix string, namespace string, value string) string {
	sum := sha256.Sum256([]byte(namespace + ":" + value))
	return prefix + "_" + hex.EncodeToString(sum[:10])
}

func slugify(value string) string {
	value = strings.ToLower(value)
	var builder strings.Builder
	separator := false
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			if separator && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(char)
			separator = false
			continue
		}
		separator = true
	}
	return builder.String()
}
