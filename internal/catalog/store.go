package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrPlaceNotFound = errors.New("place not found")

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
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
