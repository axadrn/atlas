package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"slices"
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
	limit = boundedLimit(limit, 10, 25)
	// An empty query is the search popup's initial state: show the whole
	// curated top list instead of nothing, regardless of the search limit.
	if query == "" {
		limit = 50
		rows, err := s.db.QueryContext(ctx, `
			SELECT id, slug, name, place_type, parent_id, country_code,
			       is_destination, latitude, longitude, population, curated_rank
			FROM places
			WHERE curated_rank IS NOT NULL
			ORDER BY curated_rank
			LIMIT ?
		`, limit)
		if err != nil {
			return nil, fmt.Errorf("list top places: %w", err)
		}
		defer rows.Close()
		return scanPlaces(rows, limit)
	}
	pattern := "%" + escapeLike(query) + "%"
	prefix := escapeLike(query) + "%"

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, slug, name, place_type, parent_id, country_code,
		       is_destination, latitude, longitude, population, curated_rank
		FROM places
		WHERE name LIKE ? ESCAPE '\' COLLATE NOCASE
		ORDER BY
			CASE
				WHEN name = ? COLLATE NOCASE THEN 0
				WHEN name LIKE ? ESCAPE '\' COLLATE NOCASE THEN 1
				ELSE 2
			END,
			curated_rank IS NULL, curated_rank,
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

// MapPlaces returns the ranked destinations for the landing globe: only the
// curated hotspots, nothing else.
func (s *Store) MapPlaces(ctx context.Context, limit int) ([]PlaceSummary, error) {
	limit = boundedLimit(limit, 50, 500)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, slug, name, place_type, parent_id, country_code,
		       is_destination, latitude, longitude, population, curated_rank
		FROM places
		WHERE is_destination = 1
			AND curated_rank IS NOT NULL
			AND latitude IS NOT NULL
			AND longitude IS NOT NULL
		ORDER BY curated_rank
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list map places: %w", err)
	}
	defer rows.Close()

	return scanPlaces(rows, limit)
}

func (s *Store) ChildrenByParent(ctx context.Context, parentID string, limit int) ([]PlaceSummary, error) {
	limit = boundedLimit(limit, 24, 100)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, slug, name, place_type, parent_id, country_code,
		       is_destination, latitude, longitude, population, curated_rank
		FROM places
		WHERE parent_id = ?
		ORDER BY curated_rank IS NULL, curated_rank,
			population DESC NULLS LAST, name COLLATE NOCASE
		LIMIT ?
	`, parentID, limit)
	if err != nil {
		return nil, fmt.Errorf("list child places: %w", err)
	}
	defer rows.Close()

	return scanPlaces(rows, limit)
}

func (s *Store) placeByID(ctx context.Context, id string) (PlaceSummary, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, slug, name, place_type, parent_id, country_code,
		       is_destination, latitude, longitude, population, curated_rank
		FROM places
		WHERE id = ?
	`, id)
	place, err := scanPlace(row)
	if errors.Is(err, sql.ErrNoRows) {
		return PlaceSummary{}, ErrPlaceNotFound
	}
	if err != nil {
		return PlaceSummary{}, fmt.Errorf("get place by id: %w", err)
	}
	return place, nil
}

func (s *Store) PlaceBySlug(ctx context.Context, slug string) (PlaceSummary, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, slug, name, place_type, parent_id, country_code,
		       is_destination, latitude, longitude, population, curated_rank
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
	if place.Type == PlaceTypeCountry {
		place.Bounds, err = s.countryBounds(ctx, place.CountryCode)
		if err != nil {
			return PlaceSummary{}, err
		}
	}
	place.Timezones, err = s.timezonesForPlace(ctx, place)
	if err != nil {
		return PlaceSummary{}, err
	}
	return place, nil
}

func (s *Store) AncestorsForPlace(ctx context.Context, place PlaceSummary) ([]PlaceSummary, error) {
	ancestors := make([]PlaceSummary, 0, 3)
	parentID := place.ParentID
	for parentID != "" {
		if len(ancestors) == 8 {
			return nil, errors.New("place hierarchy exceeds eight levels")
		}
		parent, err := s.placeByID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("get place ancestor: %w", err)
		}
		ancestors = append(ancestors, parent)
		parentID = parent.ParentID
	}
	slices.Reverse(ancestors)
	return ancestors, nil
}

func (s *Store) timezonesForPlace(ctx context.Context, place PlaceSummary) ([]string, error) {
	for range 8 {
		rows, err := s.db.QueryContext(ctx, `
			SELECT timezone_id
			FROM place_timezones
			WHERE place_id = ?
			ORDER BY timezone_id
		`, place.ID)
		if err != nil {
			return nil, fmt.Errorf("list place timezones: %w", err)
		}
		timezones := make([]string, 0, 2)
		for rows.Next() {
			var timezone string
			if err := rows.Scan(&timezone); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan place timezone: %w", err)
			}
			timezones = append(timezones, timezone)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("read place timezones: %w", err)
		}
		if err := rows.Close(); err != nil {
			return nil, fmt.Errorf("close place timezones: %w", err)
		}
		if len(timezones) > 0 {
			return timezones, nil
		}
		if place.ParentID == "" {
			return []string{}, nil
		}
		place, err = s.placeByID(ctx, place.ParentID)
		if err != nil {
			return nil, fmt.Errorf("get timezone parent: %w", err)
		}
	}
	return nil, errors.New("place timezone hierarchy exceeds eight levels")
}

// Sources lists every provider for the credits page.
func (s *Store) Sources(ctx context.Context) ([]SourceAttribution, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT name, homepage_url, license_name, COALESCE(license_url, '')
		FROM sources
		ORDER BY name COLLATE NOCASE
	`)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()

	sources := make([]SourceAttribution, 0, 4)
	for rows.Next() {
		var source SourceAttribution
		if err := rows.Scan(&source.Name, &source.HomepageURL, &source.LicenseName, &source.LicenseURL); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, source)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read sources: %w", err)
	}
	return sources, nil
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
			AND is_destination = 1
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
	slices.Sort(normalized)

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
	var parentID sql.NullString
	var destination int
	var latitude, longitude sql.NullFloat64
	var population, curatedRank sql.NullInt64
	if err := row.Scan(
		&place.ID,
		&place.Slug,
		&place.Name,
		&place.Type,
		&parentID,
		&place.CountryCode,
		&destination,
		&latitude,
		&longitude,
		&population,
		&curatedRank,
	); err != nil {
		return PlaceSummary{}, err
	}
	if parentID.Valid {
		place.ParentID = parentID.String
	}
	place.Destination = destination == 1
	if latitude.Valid && longitude.Valid {
		place.Coordinates = &Coordinates{Latitude: latitude.Float64, Longitude: longitude.Float64}
	}
	if population.Valid {
		place.Population = &population.Int64
	}
	if curatedRank.Valid {
		place.CuratedRank = &curatedRank.Int64
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
