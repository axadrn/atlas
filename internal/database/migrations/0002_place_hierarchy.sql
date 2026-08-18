CREATE TEMP TABLE place_sources_backup AS
SELECT place_id, source_id, external_id, record_url, contribution, retrieved_at
FROM place_sources;

CREATE TEMP TABLE place_timezones_backup AS
SELECT id AS place_id, timezone AS timezone_id
FROM places
WHERE timezone IS NOT NULL AND timezone <> '';

DROP TABLE place_sources;
ALTER TABLE places RENAME TO places_legacy;

CREATE TABLE places (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    place_type TEXT NOT NULL CHECK (
        place_type IN ('country', 'region', 'locality', 'city', 'town', 'island', 'neighborhood')
    ),
    parent_id TEXT REFERENCES places(id) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED,
    country_code TEXT NOT NULL,
    is_destination INTEGER NOT NULL DEFAULT 0 CHECK (is_destination IN (0, 1)),
    latitude REAL,
    longitude REAL,
    population INTEGER CHECK (population IS NULL OR population >= 0),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (length(country_code) = 2 AND country_code = upper(country_code)),
    CHECK (parent_id IS NULL OR parent_id <> id),
    CHECK (
        (latitude IS NULL AND longitude IS NULL) OR
        (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
    )
) STRICT;

INSERT INTO places (
    id, slug, name, place_type, parent_id, country_code, is_destination,
    latitude, longitude, population, created_at, updated_at
)
SELECT
    legacy.id,
    legacy.slug,
    legacy.name,
    CASE legacy.kind WHEN 'country' THEN 'country' ELSE 'locality' END,
    CASE legacy.kind
        WHEN 'country' THEN NULL
        ELSE (
            SELECT country.id
            FROM places_legacy country
            WHERE country.kind = 'country'
              AND country.country_code = legacy.country_code
            LIMIT 1
        )
    END,
    legacy.country_code,
    CASE legacy.kind WHEN 'destination' THEN 1 ELSE 0 END,
    legacy.latitude,
    legacy.longitude,
    legacy.population,
    legacy.created_at,
    legacy.updated_at
FROM places_legacy legacy;

DROP TABLE places_legacy;

CREATE INDEX places_parent_name_idx
    ON places(parent_id, name COLLATE NOCASE);

CREATE INDEX places_country_type_name_idx
    ON places(country_code, place_type, name COLLATE NOCASE);

CREATE INDEX places_destination_population_idx
    ON places(is_destination, population DESC);

CREATE TABLE place_timezones (
    place_id TEXT NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    timezone_id TEXT NOT NULL,
    PRIMARY KEY (place_id, timezone_id)
) STRICT;

INSERT INTO place_timezones (place_id, timezone_id)
SELECT place_id, timezone_id
FROM place_timezones_backup;

DROP TABLE place_timezones_backup;

CREATE TABLE place_sources (
    place_id TEXT NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE RESTRICT,
    external_id TEXT,
    record_url TEXT,
    contribution TEXT NOT NULL,
    retrieved_at TEXT NOT NULL,
    PRIMARY KEY (place_id, source_id),
    UNIQUE (source_id, external_id)
) STRICT;

INSERT INTO place_sources (
    place_id, source_id, external_id, record_url, contribution, retrieved_at
)
SELECT place_id, source_id, external_id, record_url, contribution, retrieved_at
FROM place_sources_backup;

DROP TABLE place_sources_backup;

CREATE INDEX place_sources_source_id_idx ON place_sources(source_id);
