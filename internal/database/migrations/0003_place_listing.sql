ALTER TABLE places
ADD COLUMN is_listed INTEGER NOT NULL DEFAULT 1 CHECK (is_listed IN (0, 1));

DROP INDEX places_country_kind_idx;

CREATE INDEX places_country_listing_population_idx
    ON places(country_code, kind, is_listed, population DESC);
