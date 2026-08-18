-- Collapse place types to the four levels Atlas actually uses: country,
-- region, locality and neighborhood. The city and town labels duplicated
-- locality, and island duplicated region. The CHECK constraint keeps
-- accepting the old values because narrowing it would require a full table
-- rebuild; the application only reads and writes the four remaining types.
UPDATE places SET place_type = 'locality' WHERE place_type IN ('city', 'town');
UPDATE places SET place_type = 'region' WHERE place_type = 'island';
