// Package geonames refreshes sourced facts from a GeoNames dump file.
//
// The refresh is update-only by design: it matches dump rows against
// place_sources.external_id and updates provider facts (name, coordinates,
// population, timezone, retrieved_at) on places Atlas already knows. It
// never inserts or deletes places and never touches Atlas-owned fields:
// id, slug, place_type, parent_id, country_code, is_destination and
// curated_rank stay exactly as they are.
package geonames

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Record is one row of the standard 19-column GeoNames dump format, reduced
// to the fields Atlas stores.
type Record struct {
	ID          string
	Name        string
	Latitude    float64
	Longitude   float64
	HasCoords   bool
	Population  int64
	CountryCode string
	Timezone    string
}

// ParseLine parses one tab-separated dump line. Comment lines and rows
// without a geonameid or name return ok=false.
func ParseLine(line string) (Record, bool) {
	if strings.HasPrefix(line, "#") {
		return Record{}, false
	}
	fields := strings.Split(line, "\t")
	if len(fields) < 19 || fields[0] == "" || fields[1] == "" {
		return Record{}, false
	}
	record := Record{
		ID:          fields[0],
		Name:        fields[1],
		CountryCode: fields[8],
		Timezone:    fields[17],
	}
	latitude, latErr := strconv.ParseFloat(fields[4], 64)
	longitude, lonErr := strconv.ParseFloat(fields[5], 64)
	if latErr == nil && lonErr == nil {
		record.Latitude = latitude
		record.Longitude = longitude
		record.HasCoords = true
	}
	if population, err := strconv.ParseInt(fields[14], 10, 64); err == nil {
		record.Population = population
	}
	return record, true
}

// Stats reports what a refresh run saw and did.
type Stats struct {
	Scanned int
	Matched int
}

type knownPlace struct {
	placeID   string
	placeType string
	hasCoords bool
}

// Refresh streams a dump and updates every matched place in one
// transaction. With dryRun the transaction is rolled back at the end, so
// the run only reports what it would do.
func Refresh(ctx context.Context, db *sql.DB, dump io.Reader, now time.Time, dryRun bool) (Stats, error) {
	var stats Stats

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return stats, fmt.Errorf("begin refresh: %w", err)
	}
	defer tx.Rollback()

	known, err := knownPlaces(ctx, tx)
	if err != nil {
		return stats, err
	}

	scanner := bufio.NewScanner(dump)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		record, ok := ParseLine(scanner.Text())
		if !ok {
			continue
		}
		stats.Scanned++
		place, ok := known[record.ID]
		if !ok {
			continue
		}
		if err := applyRecord(ctx, tx, place, record, now); err != nil {
			return stats, err
		}
		stats.Matched++
	}
	if err := scanner.Err(); err != nil {
		return stats, fmt.Errorf("read dump: %w", err)
	}

	if dryRun {
		return stats, tx.Rollback()
	}
	if err := tx.Commit(); err != nil {
		return stats, fmt.Errorf("commit refresh: %w", err)
	}
	return stats, nil
}

func knownPlaces(ctx context.Context, tx *sql.Tx) (map[string]knownPlace, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT ps.external_id, p.id, p.place_type,
		       p.latitude IS NOT NULL AND p.longitude IS NOT NULL
		FROM place_sources ps
		JOIN places p ON p.id = ps.place_id
		WHERE ps.source_id = 'geonames'
	`)
	if err != nil {
		return nil, fmt.Errorf("list geonames places: %w", err)
	}
	defer rows.Close()

	known := make(map[string]knownPlace)
	for rows.Next() {
		var externalID string
		var place knownPlace
		if err := rows.Scan(&externalID, &place.placeID, &place.placeType, &place.hasCoords); err != nil {
			return nil, fmt.Errorf("scan geonames place: %w", err)
		}
		known[externalID] = place
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read geonames places: %w", err)
	}
	return known, nil
}

func applyRecord(ctx context.Context, tx *sql.Tx, place knownPlace, record Record, now time.Time) error {
	// Coordinates only refresh where Atlas already stores them: countries
	// intentionally have none, their map view derives bounds instead.
	updateCoords := record.HasCoords && place.hasCoords
	// GeoNames uses zero for unknown population; never replace a known
	// value with unknown.
	updatePopulation := record.Population > 0

	if _, err := tx.ExecContext(ctx, `
		UPDATE places SET
			name = ?,
			latitude = CASE WHEN ? THEN ? ELSE latitude END,
			longitude = CASE WHEN ? THEN ? ELSE longitude END,
			population = CASE WHEN ? THEN ? ELSE population END,
			updated_at = ?
		WHERE id = ?
	`,
		record.Name,
		updateCoords, record.Latitude,
		updateCoords, record.Longitude,
		updatePopulation, record.Population,
		now.UTC().Format(time.RFC3339Nano),
		place.placeID,
	); err != nil {
		return fmt.Errorf("update place %s: %w", place.placeID, err)
	}

	// Country time zones come from the IANA source and can be plural;
	// only non-countries take the single dump zone.
	if place.placeType != "country" && record.Timezone != "" {
		if _, err := tx.ExecContext(ctx, `DELETE FROM place_timezones WHERE place_id = ?`, place.placeID); err != nil {
			return fmt.Errorf("clear timezones for %s: %w", place.placeID, err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO place_timezones (place_id, timezone_id) VALUES (?, ?)
		`, place.placeID, record.Timezone); err != nil {
			return fmt.Errorf("set timezone for %s: %w", place.placeID, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE place_sources SET retrieved_at = ?
		WHERE place_id = ? AND source_id = 'geonames'
	`, now.UTC().Format(time.RFC3339Nano), place.placeID); err != nil {
		return fmt.Errorf("update retrieval date for %s: %w", place.placeID, err)
	}
	return nil
}
