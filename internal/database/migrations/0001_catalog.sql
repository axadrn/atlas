CREATE TABLE sources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    publisher TEXT NOT NULL,
    homepage_url TEXT NOT NULL,
    license_name TEXT NOT NULL,
    license_url TEXT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
) STRICT;

CREATE TABLE source_snapshots (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL REFERENCES sources(id) ON DELETE RESTRICT,
    version TEXT,
    retrieved_at TEXT NOT NULL,
    checksum_sha256 TEXT NOT NULL,
    origin_url TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (source_id, checksum_sha256)
) STRICT;

CREATE TABLE import_runs (
    id TEXT PRIMARY KEY,
    source_snapshot_id TEXT NOT NULL REFERENCES source_snapshots(id) ON DELETE RESTRICT,
    status TEXT NOT NULL CHECK (status IN ('running', 'completed', 'failed')),
    started_at TEXT NOT NULL,
    finished_at TEXT,
    records_seen INTEGER NOT NULL DEFAULT 0 CHECK (records_seen >= 0),
    records_changed INTEGER NOT NULL DEFAULT 0 CHECK (records_changed >= 0),
    error_message TEXT,
    CHECK (
        (status = 'running' AND finished_at IS NULL) OR
        (status IN ('completed', 'failed') AND finished_at IS NOT NULL)
    )
) STRICT;

CREATE TABLE places (
    id TEXT PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    kind TEXT NOT NULL CHECK (kind IN (
        'world', 'continent', 'country', 'territory', 'region', 'city', 'metro'
    )),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN (
        'active', 'historic', 'disputed'
    )),
    parent_id TEXT REFERENCES places(id) ON DELETE RESTRICT,
    country_code TEXT,
    latitude REAL,
    longitude REAL,
    timezone TEXT,
    population INTEGER CHECK (population IS NULL OR population >= 0),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (country_code IS NULL OR (
        length(country_code) = 2 AND country_code = upper(country_code)
    )),
    CHECK (
        (latitude IS NULL AND longitude IS NULL) OR
        (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
    ),
    CHECK (parent_id IS NULL OR parent_id <> id)
) STRICT;

CREATE INDEX places_parent_id_idx ON places(parent_id);
CREATE INDEX places_country_kind_idx ON places(country_code, kind);
CREATE INDEX places_population_idx ON places(population DESC);

CREATE TABLE place_names (
    place_id TEXT NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    language_tag TEXT NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL DEFAULT 'common' CHECK (kind IN (
        'common', 'official', 'short', 'historic', 'alternate'
    )),
    is_preferred INTEGER NOT NULL DEFAULT 0 CHECK (is_preferred IN (0, 1)),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    PRIMARY KEY (place_id, language_tag, name, kind)
) STRICT;

CREATE UNIQUE INDEX place_names_preferred_idx
    ON place_names(place_id, language_tag)
    WHERE is_preferred = 1;

CREATE INDEX place_names_lookup_idx ON place_names(name COLLATE NOCASE);

CREATE TABLE external_references (
    provider TEXT NOT NULL,
    external_id TEXT NOT NULL,
    place_id TEXT NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    source_snapshot_id TEXT REFERENCES source_snapshots(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    PRIMARY KEY (provider, external_id),
    UNIQUE (place_id, provider)
) STRICT;

CREATE INDEX external_references_place_id_idx ON external_references(place_id);
