package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrPlaceNotFound = errors.New("place not found")

type queryExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
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
