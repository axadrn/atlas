CREATE TABLE place_redirects (
    old_place_id TEXT PRIMARY KEY REFERENCES places(id) ON DELETE RESTRICT,
    place_id TEXT NOT NULL REFERENCES places(id) ON DELETE RESTRICT,
    reason TEXT NOT NULL CHECK (length(trim(reason)) > 0),
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (old_place_id <> place_id)
) STRICT;

CREATE INDEX place_redirects_place_id_idx ON place_redirects(place_id);
