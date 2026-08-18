-- Curated destination ranking, independent of population: lower is better,
-- NULL means unranked and sorts after every ranked place. This is editorial
-- curation until community observation data can compute real scores.
ALTER TABLE places ADD COLUMN curated_rank INTEGER
    CHECK (curated_rank IS NULL OR curated_rank >= 1);
