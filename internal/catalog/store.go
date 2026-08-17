package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrPlaceNotFound = errors.New("place not found")

type queryExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type Store struct {
	db queryExecutor
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func NewTxStore(tx *sql.Tx) *Store {
	return &Store{db: tx}
}

func (s *Store) UpsertPlace(ctx context.Context, place Place) error {
	if err := place.Validate(); err != nil {
		return fmt.Errorf("validate place: %w", err)
	}

	var latitude, longitude any
	if place.Coordinates != nil {
		latitude = place.Coordinates.Latitude
		longitude = place.Coordinates.Longitude
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO places (
			id, slug, kind, status, parent_id, country_code,
			latitude, longitude, timezone, population
		) VALUES (?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, NULLIF(?, ''), ?)
		ON CONFLICT(id) DO UPDATE SET
			slug = excluded.slug,
			kind = excluded.kind,
			status = excluded.status,
			parent_id = excluded.parent_id,
			country_code = excluded.country_code,
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			timezone = excluded.timezone,
			population = excluded.population,
			updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
	`,
		place.ID,
		place.Slug,
		place.Kind,
		place.Status,
		place.ParentID,
		place.CountryCode,
		latitude,
		longitude,
		place.Timezone,
		place.Population,
	)
	if err != nil {
		return fmt.Errorf("upsert place: %w", err)
	}
	return nil
}

func (s *Store) PlaceByID(ctx context.Context, id string) (Place, error) {
	var place Place
	var parentID, countryCode, timezone sql.NullString
	var latitude, longitude sql.NullFloat64
	var population sql.NullInt64

	err := s.db.QueryRowContext(ctx, `
		SELECT id, slug, kind, status, parent_id, country_code,
		       latitude, longitude, timezone, population
		FROM places
		WHERE id = ?
	`, id).Scan(
		&place.ID,
		&place.Slug,
		&place.Kind,
		&place.Status,
		&parentID,
		&countryCode,
		&latitude,
		&longitude,
		&timezone,
		&population,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Place{}, ErrPlaceNotFound
	}
	if err != nil {
		return Place{}, fmt.Errorf("get place: %w", err)
	}

	place.ParentID = parentID.String
	place.CountryCode = countryCode.String
	place.Timezone = timezone.String
	if latitude.Valid && longitude.Valid {
		place.Coordinates = &Coordinates{
			Latitude:  latitude.Float64,
			Longitude: longitude.Float64,
		}
	}
	if population.Valid {
		place.Population = &population.Int64
	}
	return place, nil
}

type PlaceSummary struct {
	ID          string       `json:"id"`
	Slug        string       `json:"slug"`
	Name        string       `json:"name"`
	Kind        PlaceKind    `json:"kind"`
	CountryCode string       `json:"country_code,omitempty"`
	Coordinates *Coordinates `json:"coordinates,omitempty"`
	Timezone    string       `json:"timezone,omitempty"`
	Population  *int64       `json:"population,omitempty"`
}

type SourceAttribution struct {
	Name        string
	Publisher   string
	HomepageURL string
	LicenseName string
	LicenseURL  string
	OriginURL   string
	Version     string
	RetrievedAt string
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
		SELECT
			p.id,
			p.slug,
			(
				SELECT preferred.name
				FROM place_names preferred
				WHERE preferred.place_id = p.id AND preferred.is_preferred = 1
				ORDER BY CASE preferred.language_tag WHEN 'en' THEN 0 WHEN 'und' THEN 1 ELSE 2 END
				LIMIT 1
			) AS name,
			p.kind,
			COALESCE(p.country_code, ''),
			p.latitude,
			p.longitude,
			COALESCE(p.timezone, ''),
			p.population
		FROM places p
		WHERE p.status = 'active'
			AND p.kind IN ('country', 'territory', 'city', 'metro')
			AND EXISTS (
				SELECT 1
				FROM place_names matched
				WHERE matched.place_id = p.id
					AND matched.name LIKE ? ESCAPE '\' COLLATE NOCASE
			)
		ORDER BY
			CASE
				WHEN name = ? COLLATE NOCASE THEN 0
				WHEN name LIKE ? ESCAPE '\' COLLATE NOCASE THEN 1
				ELSE 2
			END,
			p.population DESC NULLS LAST,
			name COLLATE NOCASE
		LIMIT ?
	`, pattern, query, prefix, limit)
	if err != nil {
		return nil, fmt.Errorf("search places: %w", err)
	}
	defer rows.Close()

	places := make([]PlaceSummary, 0, limit)
	for rows.Next() {
		place, err := scanPlaceSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("scan place search result: %w", err)
		}
		places = append(places, place)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read place search results: %w", err)
	}
	return places, nil
}

func (s *Store) MapCities(ctx context.Context, limit int) ([]PlaceSummary, error) {
	limit = boundedLimit(limit, 150, 500)
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			p.id,
			p.slug,
			(
				SELECT preferred.name
				FROM place_names preferred
				WHERE preferred.place_id = p.id AND preferred.is_preferred = 1
				ORDER BY CASE preferred.language_tag WHEN 'en' THEN 0 WHEN 'und' THEN 1 ELSE 2 END
				LIMIT 1
			) AS name,
			p.kind,
			COALESCE(p.country_code, ''),
			p.latitude,
			p.longitude,
			COALESCE(p.timezone, ''),
			p.population
		FROM places p
		WHERE p.status = 'active'
			AND p.kind IN ('city', 'metro')
			AND p.latitude IS NOT NULL
			AND p.longitude IS NOT NULL
		ORDER BY p.population DESC NULLS LAST, name COLLATE NOCASE
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list map cities: %w", err)
	}
	defer rows.Close()

	places := make([]PlaceSummary, 0, limit)
	for rows.Next() {
		place, err := scanPlaceSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("scan map city: %w", err)
		}
		places = append(places, place)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read map cities: %w", err)
	}
	return places, nil
}

func (s *Store) PlaceBySlug(ctx context.Context, slug string) (PlaceSummary, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT
			p.id,
			p.slug,
			(
				SELECT preferred.name
				FROM place_names preferred
				WHERE preferred.place_id = p.id AND preferred.is_preferred = 1
				ORDER BY CASE preferred.language_tag WHEN 'en' THEN 0 WHEN 'und' THEN 1 ELSE 2 END
				LIMIT 1
			) AS name,
			p.kind,
			COALESCE(p.country_code, ''),
			p.latitude,
			p.longitude,
			COALESCE(p.timezone, ''),
			p.population
		FROM places p
		WHERE p.slug = ? AND p.status = 'active'
	`, slug)
	place, err := scanPlaceSummary(row)
	if errors.Is(err, sql.ErrNoRows) {
		return PlaceSummary{}, ErrPlaceNotFound
	}
	if err != nil {
		return PlaceSummary{}, fmt.Errorf("get place by slug: %w", err)
	}
	return place, nil
}

func (s *Store) SourcesForPlace(ctx context.Context, placeID string) ([]SourceAttribution, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT
			s.name,
			s.publisher,
			s.homepage_url,
			s.license_name,
			COALESCE(s.license_url, ''),
			ss.origin_url,
			COALESCE(ss.version, ''),
			ss.retrieved_at
		FROM external_references er
		JOIN source_snapshots ss ON ss.id = er.source_snapshot_id
		JOIN sources s ON s.id = ss.source_id
		WHERE er.place_id = ?
		ORDER BY s.name COLLATE NOCASE, ss.retrieved_at DESC
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
			&source.Publisher,
			&source.HomepageURL,
			&source.LicenseName,
			&source.LicenseURL,
			&source.OriginURL,
			&source.Version,
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

type rowScanner interface {
	Scan(...any) error
}

func scanPlaceSummary(row rowScanner) (PlaceSummary, error) {
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

func boundedLimit(limit int, defaultLimit int, maximum int) int {
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

type PlaceName struct {
	PlaceID     string
	LanguageTag string
	Name        string
	Kind        string
	Preferred   bool
}

func (s *Store) UpsertPlaceName(ctx context.Context, name PlaceName) error {
	if name.PlaceID == "" || name.LanguageTag == "" || name.Name == "" || name.Kind == "" {
		return errors.New("place name fields are required")
	}
	if name.Preferred {
		if _, err := s.db.ExecContext(ctx, `
			UPDATE place_names
			SET is_preferred = 0
			WHERE place_id = ? AND language_tag = ? AND is_preferred = 1
		`, name.PlaceID, name.LanguageTag); err != nil {
			return fmt.Errorf("clear preferred place name: %w", err)
		}
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO place_names (place_id, language_tag, name, kind, is_preferred)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(place_id, language_tag, name, kind) DO UPDATE SET
			is_preferred = excluded.is_preferred
	`, name.PlaceID, name.LanguageTag, name.Name, name.Kind, name.Preferred)
	if err != nil {
		return fmt.Errorf("upsert place name: %w", err)
	}
	return nil
}

type Source struct {
	ID          string
	Name        string
	Publisher   string
	HomepageURL string
	LicenseName string
	LicenseURL  string
}

func (s *Store) UpsertSource(ctx context.Context, source Source) error {
	if source.ID == "" || source.Name == "" || source.Publisher == "" ||
		source.HomepageURL == "" || source.LicenseName == "" {
		return errors.New("source fields are required")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sources (id, name, publisher, homepage_url, license_name, license_url)
		VALUES (?, ?, ?, ?, ?, NULLIF(?, ''))
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			publisher = excluded.publisher,
			homepage_url = excluded.homepage_url,
			license_name = excluded.license_name,
			license_url = excluded.license_url,
			updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
	`, source.ID, source.Name, source.Publisher, source.HomepageURL, source.LicenseName, source.LicenseURL)
	if err != nil {
		return fmt.Errorf("upsert source: %w", err)
	}
	return nil
}

type SourceSnapshot struct {
	ID             string
	SourceID       string
	Version        string
	RetrievedAt    string
	ChecksumSHA256 string
	OriginURL      string
}

func (s *Store) AddSourceSnapshot(ctx context.Context, snapshot SourceSnapshot) error {
	if snapshot.ID == "" || snapshot.SourceID == "" || snapshot.RetrievedAt == "" ||
		snapshot.ChecksumSHA256 == "" || snapshot.OriginURL == "" {
		return errors.New("source snapshot fields are required")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO source_snapshots (
			id, source_id, version, retrieved_at, checksum_sha256, origin_url
		) VALUES (?, ?, NULLIF(?, ''), ?, ?, ?)
		ON CONFLICT(id) DO NOTHING
	`, snapshot.ID, snapshot.SourceID, snapshot.Version, snapshot.RetrievedAt,
		snapshot.ChecksumSHA256, snapshot.OriginURL)
	if err != nil {
		return fmt.Errorf("add source snapshot: %w", err)
	}
	return nil
}

type ExternalReference struct {
	Provider         string
	ExternalID       string
	PlaceID          string
	SourceSnapshotID string
}

func (s *Store) UpsertExternalReference(ctx context.Context, ref ExternalReference) error {
	if ref.Provider == "" || ref.ExternalID == "" || ref.PlaceID == "" {
		return errors.New("external reference fields are required")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO external_references (
			provider, external_id, place_id, source_snapshot_id
		) VALUES (?, ?, ?, NULLIF(?, ''))
		ON CONFLICT(provider, external_id) DO UPDATE SET
			place_id = excluded.place_id,
			source_snapshot_id = excluded.source_snapshot_id
	`, ref.Provider, ref.ExternalID, ref.PlaceID, ref.SourceSnapshotID)
	if err != nil {
		return fmt.Errorf("upsert external reference: %w", err)
	}
	return nil
}
