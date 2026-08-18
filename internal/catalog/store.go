package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

var ErrPlaceNotFound = errors.New("place not found")

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) SearchPlaces(ctx context.Context, query string, limit int) ([]PlaceSummary, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []PlaceSummary{}, nil
	}
	limit = boundedLimit(limit, 10, 25)
	pattern := "%" + escapeLike(query) + "%"
	prefix := escapeLike(query) + "%"

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, slug, name, kind, country_code, latitude, longitude,
		       COALESCE(timezone, ''), population
		FROM places
		WHERE name LIKE ? ESCAPE '\' COLLATE NOCASE
		ORDER BY
			CASE
				WHEN name = ? COLLATE NOCASE THEN 0
				WHEN name LIKE ? ESCAPE '\' COLLATE NOCASE THEN 1
				ELSE 2
			END,
			population DESC NULLS LAST,
			name COLLATE NOCASE
		LIMIT ?
	`, pattern, query, prefix, limit)
	if err != nil {
		return nil, fmt.Errorf("search places: %w", err)
	}
	defer rows.Close()

	return scanPlaces(rows, limit)
}

func (s *Store) MapPlaces(ctx context.Context, limit int) ([]PlaceSummary, error) {
	limit = boundedLimit(limit, 150, 500)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, slug, name, kind, country_code, latitude, longitude,
		       COALESCE(timezone, ''), population
		FROM places
		WHERE kind = 'destination'
			AND latitude IS NOT NULL
			AND longitude IS NOT NULL
		ORDER BY population DESC NULLS LAST, name COLLATE NOCASE
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list map places: %w", err)
	}
	defer rows.Close()

	return scanPlaces(rows, limit)
}

func (s *Store) DestinationsByCountry(ctx context.Context, countryCode string, limit int) ([]PlaceSummary, error) {
	limit = boundedLimit(limit, 24, 50)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, slug, name, kind, country_code, latitude, longitude,
		       COALESCE(timezone, ''), population
		FROM places
		WHERE country_code = ? AND kind = 'destination'
		ORDER BY name COLLATE NOCASE
		LIMIT ?
	`, countryCode, limit)
	if err != nil {
		return nil, fmt.Errorf("list country destinations: %w", err)
	}
	defer rows.Close()

	return scanPlaces(rows, limit)
}

func (s *Store) CountryByCode(ctx context.Context, countryCode string) (PlaceSummary, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, slug, name, kind, country_code, latitude, longitude,
		       COALESCE(timezone, ''), population
		FROM places
		WHERE country_code = ? AND kind = 'country'
	`, countryCode)
	country, err := scanPlace(row)
	if errors.Is(err, sql.ErrNoRows) {
		return PlaceSummary{}, ErrPlaceNotFound
	}
	if err != nil {
		return PlaceSummary{}, fmt.Errorf("get country by code: %w", err)
	}
	return country, nil
}

func (s *Store) PlaceBySlug(ctx context.Context, slug string) (PlaceSummary, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, slug, name, kind, country_code, latitude, longitude,
		       COALESCE(timezone, ''), population
		FROM places
		WHERE slug = ?
	`, slug)
	place, err := scanPlace(row)
	if errors.Is(err, sql.ErrNoRows) {
		return PlaceSummary{}, ErrPlaceNotFound
	}
	if err != nil {
		return PlaceSummary{}, fmt.Errorf("get place by slug: %w", err)
	}
	if place.Kind == PlaceKindCountry {
		place.Bounds, err = s.countryBounds(ctx, place.CountryCode)
		if err != nil {
			return PlaceSummary{}, err
		}
	}
	return place, nil
}

func (s *Store) SourcesForPlace(ctx context.Context, placeID string) ([]SourceAttribution, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.name, s.homepage_url, s.license_name,
		       COALESCE(s.license_url, ''), COALESCE(ps.external_id, ''),
		       COALESCE(ps.record_url, ''), ps.contribution, ps.retrieved_at
		FROM place_sources ps
		JOIN sources s ON s.id = ps.source_id
		WHERE ps.place_id = ?
		ORDER BY s.name COLLATE NOCASE
	`, placeID)
	if err != nil {
		return nil, fmt.Errorf("list place sources: %w", err)
	}
	defer rows.Close()

	sources := make([]SourceAttribution, 0, 2)
	for rows.Next() {
		var source SourceAttribution
		if err := rows.Scan(
			&source.Name,
			&source.HomepageURL,
			&source.LicenseName,
			&source.LicenseURL,
			&source.ExternalID,
			&source.RecordURL,
			&source.Contribution,
			&source.RetrievedAt,
		); err != nil {
			return nil, fmt.Errorf("scan place source: %w", err)
		}
		sources = append(sources, source)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read place sources: %w", err)
	}
	return sources, nil
}

func (s *Store) countryBounds(ctx context.Context, countryCode string) (*Bounds, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT longitude, latitude
		FROM places
		WHERE country_code = ?
			AND kind = 'destination'
			AND longitude IS NOT NULL
			AND latitude IS NOT NULL
	`, countryCode)
	if err != nil {
		return nil, fmt.Errorf("get country map bounds: %w", err)
	}
	defer rows.Close()

	longitudes := make([]float64, 0, 32)
	south := math.Inf(1)
	north := math.Inf(-1)
	for rows.Next() {
		var longitude, latitude float64
		if err := rows.Scan(&longitude, &latitude); err != nil {
			return nil, fmt.Errorf("scan country map bounds: %w", err)
		}
		longitudes = append(longitudes, longitude)
		south = min(south, latitude)
		north = max(north, latitude)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read country map bounds: %w", err)
	}
	if len(longitudes) == 0 {
		return nil, nil
	}

	west, east := minimalLongitudeRange(longitudes)
	if south == north {
		south = max(-90, south-0.5)
		north = min(90, north+0.5)
	}
	return &Bounds{West: west, South: south, East: east, North: north}, nil
}

func minimalLongitudeRange(longitudes []float64) (float64, float64) {
	normalized := make([]float64, len(longitudes))
	for index, longitude := range longitudes {
		normalized[index] = math.Mod(longitude+360, 360)
	}
	sort.Float64s(normalized)

	largestGap := -1.0
	largestGapIndex := 0
	for index, longitude := range normalized {
		next := normalized[(index+1)%len(normalized)]
		if index == len(normalized)-1 {
			next += 360
		}
		if gap := next - longitude; gap > largestGap {
			largestGap = gap
			largestGapIndex = index
		}
	}

	west := normalized[(largestGapIndex+1)%len(normalized)]
	if west >= 180 {
		west -= 360
	}
	east := west + 360 - largestGap
	if east == west {
		west -= 0.5
		east += 0.5
	}
	return west, east
}

type rowScanner interface {
	Scan(...any) error
}

func scanPlaces(rows *sql.Rows, capacity int) ([]PlaceSummary, error) {
	places := make([]PlaceSummary, 0, capacity)
	for rows.Next() {
		place, err := scanPlace(rows)
		if err != nil {
			return nil, fmt.Errorf("scan place: %w", err)
		}
		places = append(places, place)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read places: %w", err)
	}
	return places, nil
}

func scanPlace(row rowScanner) (PlaceSummary, error) {
	var place PlaceSummary
	var latitude, longitude sql.NullFloat64
	var population sql.NullInt64
	if err := row.Scan(
		&place.ID,
		&place.Slug,
		&place.Name,
		&place.Kind,
		&place.CountryCode,
		&latitude,
		&longitude,
		&place.Timezone,
		&population,
	); err != nil {
		return PlaceSummary{}, err
	}
	if latitude.Valid && longitude.Valid {
		place.Coordinates = &Coordinates{Latitude: latitude.Float64, Longitude: longitude.Float64}
	}
	if population.Valid {
		place.Population = &population.Int64
	}
	return place, nil
}

func boundedLimit(limit, defaultLimit, maximum int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maximum {
		return maximum
	}
	return limit
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "%", "\\%")
	return strings.ReplaceAll(value, "_", "\\_")
}
