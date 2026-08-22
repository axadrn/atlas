# Data

Atlas is database-first. Open source does not require every live row to be a
file in Git.

## Storage boundary

Git contains code, migrations, methodology, policies, documentation and small
test fixtures. SQLite contains the live catalog, community contributions,
review state and private application data.

The application database is never a public download. A future API or public
dataset gets its own schema, privacy review, license and rate limits. Accounts,
sessions, private presence, raw risk signals and private toolbox inputs are
never public data products.

SSR handlers call the Go store directly. They do not call an internal HTTP API.
Browser features receive only the narrow data they need: the globe gets its
visible markers embedded in the page and search receives a small HTML fragment.
This is data minimization, not a promise that facts shown on public pages cannot
be collected.

Future first-party mobile and external APIs live under `/api/v1`. They are
authenticated by default. First-party apps use user sessions or short-lived
tokens; external customers use revocable, scoped API keys. A mobile or browser
bundle never contains a shared secret.

The initial system is one Go process and one SQLite database:

```text
external facts  -> validate and attribute -> SQLite
community input -> validate and review    -> SQLite
SQLite          -> application policy     -> pages and selected APIs
```

PostgreSQL becomes an option only after measured write contention or multiple
application writers justify it.

## Place catalog

The catalog starts with four tables:

- `places` stores the stable Atlas ID and current display fields;
- `place_timezones` stores zero or more IANA time zone identifiers per place;
- `sources` stores provider and license information once;
- `place_sources` links a place to an external record and describes what that
  provider contributed.

`places.id` is the only identity referenced by ratings, observations, saved
places and future community records. Provider IDs can change without moving
community data.

Keys follow one rule set, decided August 2026:

- Entity tables get a generated ID with a table-unique prefix and a random
  payload, Stripe style. Registry: `plc_` places. Reserved for later: `usr_`
  users, `obs_` observations, `rpt_` reports. A prefix is never reused.
- Small curated lookup tables use their name as the key, like
  `sources.id = 'geonames'`. Code reads them by name.
- Pure link tables use the pair as composite primary key, like
  `place_sources (place_id, source_id)`. No surrogate ID. If a link ever
  becomes an entity that other tables reference, it gets a prefixed ID in
  that migration.

Random payloads stay random on purpose: time-sortable IDs (UUIDv7, ULID)
only pay off at high insert volume and leak creation time, which matters
once user IDs exist.

Place geography and product selection are separate:

- `place_type` describes what a row is: `country`, `region`, `locality` or
  `neighborhood`;
- `parent_id` creates the useful hierarchy, such as country to city to
  neighborhood;
- `is_destination` is the vetting flag: imported rows merely exist, vetted
  rows are the ones Atlas promotes on maps and lists;
- `curated_rank` orders the promoted destinations editorially until
  community-derived scores replace it. One ranking signal at a time, never
  several in parallel.

The hierarchy is intentionally sparse. Atlas stores useful destinations and
neighborhoods, not every administrative division in the world. Missing levels
are valid, so a city can link directly to its country. A neighborhood inherits
its time zone from the nearest ancestor unless it has an explicit override.

The hierarchy stays boring on purpose: no closure table, no materialized
path. Ancestor walks are a handful of primary key lookups, country-wide
queries use `country_code`, and a recursive CTE covers anything deeper if
it ever comes up.

Provider facts have exactly one correction path: fix upstream and
re-import, or ship a small curation migration. There is no field-level
override table. Community knowledge lives in its own tables with its own
provenance and never edits provider columns, so the two worlds cannot
corrupt each other.

Names and slugs are display values, not identity. Localization, aliases,
boundaries and merge support are added when the product needs them. They are
not speculative launch tables.

## Provenance

`place_sources` is intentionally narrow. It covers place identity and current
geographic display facts such as name, country code, coordinates, population
and timezone. It does not become a generic table for every fact about a place.

Each link can store:

- provider;
- external record ID and URL;
- a short human-readable contribution description;
- retrieval time.

The source license lives in `sources`. Checksums, archived snapshots and
field-level lineage are added only if a real update or compliance workflow
requires them.

Every place page presents its provenance through one Sources drawer. It states
what each provider contributed and that Atlas may select, normalize and combine
source data. Map attribution remains visible with the map because map licenses
and provider terms can require it there.

## Separate data domains

Different facts have different meaning and cannot share one generic record.

### Prices

A future price report needs a place, category, amount, currency, observation
date and context. Public output is an aggregate with sample size, period and
methodology. Money is never stored as a floating-point approximation.

### Weather and exchange rates

Machine-measured data comes from a provider through a bounded cache. It is not
a permanent property of `places`, and page requests degrade gracefully when a
provider is unavailable.

### Crime and public statistics

Statistics need a geography, definition, period and source. Different legal
definitions and reporting rates must not be presented as directly comparable.

### Experiences and ratings

Ratings are subjective community responses. Atlas shows distributions and
segmented views instead of pretending that one universal city score is truth.
Accounts have bounded influence and cannot gain extra votes through payment.

### Claims and articles

Visa, tax, health and safety information is versioned content with an effective
period and evidence. Popularity alone cannot publish a high-risk claim.

### Images

Image attribution belongs to the image record because creator and license vary
per file. A wrong or illegally reused image is worse than no image.

These domains may reuse the central `sources` registry, and their source details
can be collected in the same UI drawer. Their actual records remain separate.

## Initial external sources

| Source | Initial use |
| --- | --- |
| GeoNames | Initial place names, country codes, coordinates, approximate population, time zones and external IDs |
| IANA Time Zone Database | Country time zone identifiers |
| OpenStreetMap | Map data, neighborhood identity and coordinates where suitable records exist |
| Wikivoyage | Traveler-facing neighborhood identity where geographic catalogs do not model it cleanly |
| OpenFreeMap | Initial vector tile and map style delivery |

GeoNames data is available under CC BY 4.0 and without a guarantee of accuracy,
timeliness or completeness. Atlas stores GeoNames facts instead of fetching
them during page requests. The initial selection process is temporary and not a
production synchronization pipeline.

Official references:

- [GeoNames data dumps](https://download.geonames.org/export/dump/)
- [IANA Time Zone Database](https://www.iana.org/time-zones)
- [OpenStreetMap attribution](https://www.openstreetmap.org/copyright)
- [Wikivoyage](https://www.wikivoyage.org/)
- [OpenFreeMap](https://openfreemap.org/)

## Freshness

Stored values keep serving when a provider is unavailable. A relevant date is
shown instead of implying that every value is live. Scheduled synchronization
is added only when real maintenance shows it is necessary.

Population needs special care. Country population, city proper, urban area and
metro population are different measurements. Atlas must label the definition
before combining or comparing them.

## Contributions

People contribute through structured product forms so Atlas can validate units,
context and privacy. Observations are append-only. Corrections supersede prior
records, and public aggregates expose sample size, period, methodology and
uncertainty.

GitHub remains the place for methodology, schemas, reproducible bugs and code.
It is not the live community database.

## Offline storage

The PWA service worker caches only public static assets and a generic offline
page. It does not cache navigated pages, API responses, search results, account
data or private toolbox inputs. Features that need intentional encrypted or
offline data storage must receive a separate privacy and security design.
