# Data

Atlas is database-first. Open source does not require every live row to be a
YAML file in Git. Git is excellent for code, definitions and reviewable text.
It is poor at high-volume observations, private records, moderation workflows
and frequently changing aggregates.

## The split

### Git contains

- application code;
- database migrations;
- data type and field definitions;
- scoring and aggregation methodology;
- moderation and governance rules;
- source adapters and importers;
- license and attribution requirements;
- documentation;
- small deterministic fixtures for tests.

Markdown is used for human explanations. YAML or JSON may be used for small
configuration files when it is the simplest representation. Neither is the
primary data store.

### SQLite contains

- places and their names;
- external identifiers;
- accounts and profile settings;
- optional presence;
- observations and experiences;
- claims, evidence and attestations;
- articles and discussions when they are introduced;
- contribution and review state;
- reputation and XP events;
- moderation cases;
- risk signals;
- aggregates and publication state.

### Public access is selective

The primary application database is never published, downloadable or directly
queryable by the public. Atlas does not promise a complete database dump or a
raw export of community contributions.

Public pages expose the information required by the product. A selected
aggregate, source registry or Atlas-produced reference dataset may later be
offered through an API or a dedicated download after a separate privacy,
license and business review. Each public data product has an explicit schema
and license. Openness is decided per data product, not inherited from the
internal database.

Accounts, sessions, private presence, travel history, raw risk signals,
moderation records and private toolbox inputs are never public data products.

## Database topology

The initial system remains one Go process and one primary SQLite database. A
replaceable derived database may be used for large external datasets such as
OpenStreetMap extracts.

This keeps deployment simple while preserving clear boundaries:

```text
external sources -> import and normalize -> primary SQLite
community input  -> validate and review  -> primary SQLite
primary SQLite   -> application policy   -> public pages and selected APIs
private toolbox  -> private storage      -> only the individual user
```

SQLite uses write-ahead logging, short transactions, foreign keys and bounded
connection pools. External HTTP calls never happen inside a write transaction.
PostgreSQL becomes an option only after measured write contention, multiple
writers across application nodes or operational requirements justify it.

## Core records

The exact SQL schema will evolve, but these concepts must remain distinct.

### Geography

- `places`: stable identity, kind, parent and coarse geometry;
- `place_names`: localized and historical names;
- `external_references`: identifiers from GeoNames, OpenStreetMap and other
  providers;
- `place_relations`: containment, metro membership and disputed relationships.

Slugs and display names are not identifiers. Boundaries and disputed places
retain their source and worldview instead of pretending there is one neutral
map.

### Sources and evidence

- `sources`: publisher, URL and license;
- `source_snapshots`: retrieval time, checksum and archived representation;
- `evidence`: the part of a source used to support or dispute a claim.

Every imported or community-supplied claim can be traced to its origin.

### Observations

An observation records:

- contributor or import source;
- place and optional coarse area;
- category and definition version;
- value, currency or unit;
- observation time;
- relevant context;
- eligibility and review state.

Observations are appended. Corrections supersede a previous record rather than
rewriting history. Public aggregates use eligible observations and expose
sample size, period, methodology and uncertainty.

Prices use explicit currencies and periods. Measurements use fixed units.
Money is not stored as a floating-point approximation.

### Experiences and ratings

An account has at most one active response per place, dimension and relevant
period. Updating a response supersedes the prior response instead of creating
extra votes.

Atlas shows distributions and segmented views. A visitor, long-term resident
and local may experience the same place differently. Combining them into one
universal score would hide useful information and invite patriotic voting.

### Claims

A claim is a versioned statement with a scope and effective period. It moves
through an explicit lifecycle:

```text
submitted -> pending -> published -> contested -> superseded
                     \-> rejected
```

Support and dispute actions are attestations with reasons and optional
evidence. Popularity alone cannot publish a high-risk claim. Visa, tax, health
and safety claims require stronger evidence and review policies than a coffee
price.

### Audit and reputation

- contribution events preserve who did what and when;
- moderation events preserve decisions and reasons;
- reputation events explain every increase or decrease;
- aggregate versions record the exact policy and input set used.

Current state is queryable without replaying the entire history. Important
actions also have append-only audit records. Atlas uses an audit ledger, not a
full event-sourcing architecture.

## External data

Atlas treats external data as imported and replaceable, never as an invisible
dependency.

### Source roles

No single upstream is the Atlas geography database. Each source has a narrow
job and retains its own provenance and license.

| Source | Atlas use | Initial refresh target |
| --- | --- | --- |
| GeoNames `cities5000` | Place seed, names, coordinates, time zones and approximate population | Weekly snapshot |
| Natural Earth | Country boundaries, country centers and low-zoom map geometry | On upstream release |
| OpenStreetMap | Interactive maps, detailed boundaries and later points of interest | Provider tiles live; imported extracts daily or weekly |
| Wikidata | Cross-source identifiers, localized metadata and links to suitable media | Weekly enrichment |
| Wikimedia Commons | Optional place images with per-file creator and license metadata | Revalidate metadata every 30 days |
| Community observations | Costs, experiences and practical local knowledge | Append immediately; aggregate on a scheduled window |

GeoNames is a global geographical gazetteer assembled from many upstream
sources and community corrections. Its free dumps use CC BY 4.0 and are
provided without a guarantee of accuracy, timeliness or completeness.
`cities5000` contains populated places over 5,000 people plus selected
administrative seats. It is an excellent global bootstrap, not an authority
for every metric and not the final Atlas place taxonomy.

Country and city population require different follow-up sources. Country
population should use an authoritative statistical source such as the World
Bank or UN. City population must distinguish city proper, urban area and metro
area. Atlas must not present one of these as another.

Official references:

- [GeoNames data dumps](https://download.geonames.org/export/dump/)
- [Natural Earth terms](https://www.naturalearthdata.com/about/terms-of-use/)
- [OpenStreetMap planet and extracts](https://planet.openstreetmap.org/)
- [Wikidata data access](https://www.wikidata.org/wiki/Help:Data_access)
- [Wikimedia Commons reuse](https://commons.wikimedia.org/wiki/Commons:Reusing_content_outside_Wikimedia)

### Refresh policy

Public page requests do not fetch upstream data. Import jobs download outside
the database transaction, validate the complete input, calculate a checksum,
record a source snapshot and atomically publish the normalized result. A failed
job leaves the last good snapshot active.

Refresh frequency follows how quickly a fact changes and how expensive or
risky it is to obtain. The upstream publication frequency is a maximum, not a
requirement to import that often. GeoNames publishes daily files, but a weekly
full snapshot is enough for the initial Atlas place catalog. Daily
modification and deletion feeds become useful when the catalog importer can
reconcile removed and reclassified places safely.

Every public value eventually exposes:

- source and publisher;
- upstream record or dataset;
- license;
- observation or effective date when available;
- Atlas retrieval time;
- method, sample size and uncertainty when the value is aggregated.

Freshness is a product state, not a silent overwrite. Values may be current,
due for refresh, stale or unavailable. The application keeps serving the last
known value with its date unless policy marks it unsafe to publish.

OpenStreetMap is suitable for maps, geography and later points of interest.
Atlas must:

- show the required attribution;
- track ODbL-derived records and public uses;
- avoid bulk use of donated public tile and geocoding services;
- use switchable tile and geocoding providers;
- prefer scheduled extracts or a dedicated provider for production imports;
- keep data with incompatible licenses separate when necessary.

Country maps are not derived from city coordinates. The current GeoNames
country registry does not provide the boundary geometry used by Atlas, so
country map pages wait for the Natural Earth boundary importer. Guessing a
country extent from its imported cities would omit islands and sparsely
populated areas.

### Place images

An image is optional. A wrong, generic or illegally reused image is worse than
a strong map and useful data. The initial enrichment source is Wikimedia
Commons, usually discovered through a Wikidata entity. Atlas stores the file
identifier, original URL, creator, license, license URL, source page and
retrieval time. The interface renders that attribution with the image.

Atlas never scrapes image search engines or assumes that a Wikipedia image is
automatically reusable. Community uploads come later and require explicit
rights confirmation, moderation and takedown handling.

Weather, exchange rates and other machine-measured data should be imported
automatically. Asking people to report data that a reliable machine source can
provide wastes attention and reduces quality.

## Open contribution

People contribute through the product because structured forms can validate
units, context and privacy before data is stored. GitHub issues and pull
requests remain useful for methodology, schemas, importers and reproducible
bugs.

Every selected public data product includes provenance and an explicit license.
Transparent data means people can inspect sources, methodology, freshness and
aggregate history. It does not require publishing the application database,
private records or raw community submissions.
