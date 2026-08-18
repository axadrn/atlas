CREATE TABLE sources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    homepage_url TEXT NOT NULL,
    license_name TEXT NOT NULL,
    license_url TEXT
) STRICT;

CREATE TABLE places (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('country', 'destination')),
    country_code TEXT NOT NULL,
    latitude REAL,
    longitude REAL,
    timezone TEXT,
    population INTEGER CHECK (population IS NULL OR population >= 0),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (length(country_code) = 2 AND country_code = upper(country_code)),
    CHECK (
        (latitude IS NULL AND longitude IS NULL) OR
        (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
    )
) STRICT;

CREATE INDEX places_country_kind_name_idx
    ON places(country_code, kind, name COLLATE NOCASE);

CREATE INDEX places_population_idx ON places(population DESC);

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

CREATE INDEX place_sources_source_id_idx ON place_sources(source_id);
