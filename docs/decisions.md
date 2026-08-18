# Decisions

This is the lightweight decision log for Atlas. It records choices that shape
multiple parts of the product. A decision can change, but not silently.

Each entry states the choice, why it currently wins, what it costs and what
evidence should reopen it.

## 1. International from day one

**Status:** accepted

Atlas uses a global place model, world map, currencies, units and time zones
from the first release. No country or region is the default worldview.

This wins because global participation is part of the product loop, not a
later translation project. Uneven data depth is acceptable and becomes visible
through coverage and freshness quests.

The cost is uneven initial coverage. Localization and disputed-geography
features are added when the product can present them correctly.

Reconsider only if global scope prevents a usable public alpha from shipping.
The fallback would reduce catalog depth, not replace the global data model.

## 2. Database-first with selective public data

**Status:** accepted

SQLite is the source of truth for live application data. Git contains code,
migrations, definitions, policies and documentation. The live
database and raw community data are never published as a dump.

This wins because high-volume observations, review state, privacy and
moderation are database workflows. A repository full of YAML would be easy to
diff but difficult to query, validate at scale, protect and update through the
product.

The cost is that a data correction is not automatically a pull request and
external reuse is not automatically available for every data category. Atlas
must provide transparent contribution history and audit records in the
product. Selected aggregate or reference datasets can become explicit public
data products after privacy, license and business review.

Reconsider if the community demonstrates a strong need to edit a small,
stable reference dataset through Git. That dataset may become a text-based
source without moving community or private data out of SQLite.

## 3. Structured contribution before general publishing

**Status:** accepted

The first community feature is structured observations, experiences and
claims. A general blog and unrestricted comment system come later.

This wins because structured contributions create comparable data, immediate
feedback and focused quests. General publishing creates large moderation costs
before proving the core data loop.

The cost is less freedom of expression in the first alpha.

Reconsider when contributors repeatedly need context that cannot fit into an
observation, experience or sourced claim.

## 4. Gamification rewards quality, not submission

**Status:** accepted

Submitting data can provide immediate visual feedback, but XP, reputation and
lasting rewards depend on later eligibility and usefulness. Rewards are
bounded and reversible.

This wins because immediate points for raw submissions create a direct spam
incentive. Delayed rewards align progression with data that survives review.

The cost is a less instant reward loop. The interface must clearly show pending
progress so honest contributors do not feel ignored.

Reconsider the timing when real retention data exists. Never reconsider the
rule that raw volume alone cannot create influence.

## 5. Community moderation with an operator backstop

**Status:** accepted

Reviewers, stewards and rotating councils handle routine review, disputes and
appeals. The operator handles infrastructure, security incidents, urgent
safety matters, legal obligations and failures of the governance system.

This wins because a founder cannot sustainably judge a global community and
should not become its permanent ruler. Removing all moderation would instead
hand power to coordinated attackers and the loudest groups.

The cost is substantial governance engineering and the need to recruit,
measure and remove community authority safely.

Reconsider individual roles and promotion rules as the community develops.
Do not remove the accountable emergency path without a real successor.

## 6. Bounded influence instead of perfect personhood

**Status:** accepted

Atlas does not claim to guarantee one human per account. It combines account
security, time, scoped reputation, action limits, cluster detection, robust
statistics, peer review and reversible decisions.

This wins because perfect personhood would require intrusive identity checks
and would still be vulnerable. Payment is also not proof of honesty and would
exclude legitimate contributors.

The cost is that some abuse enters the system and must be contained after
submission. Publication and aggregate updates therefore cannot always be
instant.

Reconsider individual defenses whenever attack evidence changes. Preserve the
principles of data minimization, bounded influence and reversibility.

## 7. OpenStreetMap with replaceable infrastructure

**Status:** accepted

Atlas uses OpenStreetMap-derived data for the map and later points of interest,
with required attribution and license tracking. Public tile, geocoding and
query servers may support development and modest prototypes within their
policies, but are not permanent production dependencies.

Place pages use a locally pinned MapLibre GL JS renderer with OpenFreeMap as
the initial vector tile and style provider. The provider URL is isolated from
place identity and camera data, so Atlas can switch providers or self-host
without changing place records. Country camera bounds are derived from stored
destination coordinates. They frame useful destination
coverage rather than claiming to be exact political geometry.

This wins because the map data is global, open and community maintained. A
replaceable provider boundary protects Atlas and the donated OSM
infrastructure as traffic grows.

The cost is license-aware exports and eventual tile or geocoding hosting cost.

Reconsider the provider, rendering stack or hosting model when traffic and map
features are measurable. Do not hide attribution or build a closed copy of OSM
data.

Official operational policies:

- [OpenStreetMap tile usage](https://operations.osmfoundation.org/policies/tiles/)
- [OpenStreetMap Nominatim usage](https://operations.osmfoundation.org/policies/nominatim/)
- [OpenStreetMap attribution](https://osmfoundation.org/wiki/Licence/Attribution_Guidelines)
- [MapLibre GL JS](https://maplibre.org/maplibre-gl-js/docs/)
- [OpenFreeMap](https://openfreemap.org/)

## 8. No universal truth score

**Status:** accepted

Observations are aggregated, experiences are segmented and claims are reviewed
against evidence. Atlas does not reduce all three to a single confidence or
city score.

This wins because a price, an opinion and a legal statement have different
meanings and failure modes. One score would obscure context and make rating
manipulation disproportionately valuable.

The cost is a more complex interface than a simple ranked list.

Reconsider presentation and personalization after usage research. Preserve the
separation between measurements, opinions and sourced claims.

## 9. The Nomad Toolbox is a product surface

**Status:** accepted, revised by decision 17

Atlas builds a changing collection of focused private tools for nomad
planning, tracking, documentation, logistics, finances, administration and
compliance. Specific examples are opportunities to validate, not commitments
that define the platform. Since decision 17 the tools are free for
individual users and their role is single-player value, not direct revenue.

This wins because private tools solve high-value individual problems and
give the platform utility before the community exists.

The cost is a wider product surface and different security requirements for
each tool. Toolbox data must remain separate from community contributions and
requires explicit review before any secondary use.

Reconsider individual tools based on demand, risk and maintenance cost. Keep
the separation between public community knowledge and private user workflows.

## 10. Store external facts instead of adding live page dependencies

**Status:** accepted

Atlas stores externally sourced facts in its normalized read model. Public
page requests do not depend on GeoNames, Wikidata or other catalog APIs being
available at that moment.

This wins because stored facts with source links and retrieval dates are fast
and inspectable, and an upstream outage does not take Atlas pages down.

The cost is bounded staleness. Every value must expose its source and relevant
date instead of implying that stored data is live.

Use live requests only for genuinely fast-changing data where a bounded cache
and graceful fallback exist, such as weather. Preserve source attribution and
atomic database writes.

## 11. Atlas owns place identity

**Status:** accepted

Every place has a stable Atlas ID. `place_sources` links that identity to
replaceable GeoNames, OpenStreetMap, Wikidata and future records. Community
records reference only the Atlas ID.

This wins because changing a source must not break ratings, contributions,
saved places or URLs.

The cost is a careful matching step for each new provider. Ambiguous name and
location matches require review. The launch schema keeps one source registry
and one place-to-source link instead of snapshots or field-level lineage.

Reconsider matching and merge support only when real duplicates appear. Keep
Atlas IDs stable.

## 12. Keep database access direct until generation pays for itself

**Status:** accepted

Atlas uses Go's `database/sql` package with `modernc.org/sqlite`. Queries live
in the catalog store and use explicit scanning. Atlas does not add `sqlx` or
`sqlc` to the initial SQLite application.

This wins because the current query surface is small and direct. `sqlx` mainly
reduces scanning boilerplate without checking SQL at compile time. `sqlc`
provides stronger generated types, but its official Go support for SQLite is
still marked beta. Adding either now would create another abstraction and
workflow before it solves a measured problem.

The cost is some repetitive `Scan` code and runtime discovery of malformed
queries. Focused store tests and migration tests cover that risk at the current
size.

Reconsider `sqlc` before the community read and write model becomes large, when
repetitive mappings or query defects become common, or when Atlas moves to
PostgreSQL. A PostgreSQL version should use `pgx/v5`; evaluate `sqlc` with it
instead of adding `sqlx` as an intermediate migration.

Official references:

- [`database/sql`](https://pkg.go.dev/database/sql)
- [`sqlc` database and language support](https://docs.sqlc.dev/en/latest/reference/language-support.html)
- [`sqlx`](https://jmoiron.github.io/sqlx/)
- [`pgx/v5`](https://pkg.go.dev/github.com/jackc/pgx/v5)

## 13. Keep the live place catalog only in SQLite

**Status:** accepted

Atlas curates its own place selection and stable IDs directly in SQLite. A
temporary offline process may initialize an empty development catalog, but it
does not remain as a production pipeline or a second source of truth.

This wins because Atlas needs destinations such as cities, islands and areas,
not a mirror of a geographical provider. Application actions can add or change
places without synchronizing Git files or rerunning a global catalog job.

The cost is that a fresh database needs an intentional initialization or a
backup restore. Externally sourced facts retain their provider, external record,
contribution description, retrieval date and license. Atlas owns the curation,
not every upstream fact.

Reconsider a durable import or synchronization system only after repetitive
manual maintenance becomes a measured problem.

## 14. Use one sparse place hierarchy

**Status:** accepted

Atlas keeps countries, regions, localities, cities, towns, islands and
neighborhoods in one `places` table. An optional `parent_id` forms the
hierarchy. `is_destination` is a separate product flag and is not a geographic
type.

This wins because ratings, observations, URLs and source links can reference
one stable Atlas ID at every level. Source records and classifications can
change without moving community data. The model supports country to city to
neighborhood without requiring every administrative level in between.

The cost is deliberate curation. Atlas does not pretend that one global source
models traveler-facing places consistently. Missing hierarchy levels are valid,
and ambiguous places keep the neutral `locality` type until Atlas has evidence
for something more precise.

Reconsider separate tables only if different place types develop materially
different identity or lifecycle rules. Do not add a full administrative
registry without a product feature that needs it.

## 15. Keep SSR direct and external APIs explicit

**Status:** accepted

Server-rendered pages call the Go store directly. Atlas does not expose a JSON
API merely to move data between code in the same process. Browser interactions
receive narrow HTML fragments or page-specific embedded data.

This wins because there is no internal HTTP layer to operate and no accidental
public bulk-data endpoint. It cannot prevent collection of facts that Atlas
shows publicly, so paid API value must come from structured access, licensed
datasets, richer aggregates, freshness, quotas and service quality.

Future first-party and external endpoints use `/api/v1` and require
authentication unless a route is deliberately public. Mobile clients use user
authentication. External integrations use revocable, scoped and rate-limited
API keys whose raw secrets are never stored.

Reconsider a separate API service only after independent scaling or operational
requirements appear. Do not put a shared secret into browser or mobile code.

## 16. Ship an installable PWA before native clients

**Status:** accepted

Atlas starts as a responsive, installable PWA. The manifest, Home Screen icons,
standalone display mode and offline fallback work with the existing Go and
templ application. The service worker caches static assets only.

This wins because mobile users get an app-like entry point immediately without
a second frontend, duplicated features or store-release workflow. The install
control uses the browser prompt where available and explains the iOS Home Screen
flow where a programmatic prompt is unavailable.

Build native iOS and Android clients together with Expo and React Native when
push notifications, background work, camera access, reliable offline workflows
or platform integrations create measured value. The native clients reuse the
versioned API rather than embedding the website or duplicating business rules.

## 17. Free forever for individuals, organizations pay

**Status:** accepted (2026-08-18)

Everything an individual user touches is free, including the Nomad Toolbox
and future features. Revenue comes from organizations: structured data
access, job postings, reports for cities and tourism boards, clearly
labeled sponsorship and public interest funding during the early years.
Sponsorship buys visibility, never truth: sponsors appear in community
data, rankings and comparisons exactly as if the deal did not exist, and
negative community data about a sponsor is not a termination event.

This wins because every platform that became the biggest of its kind is
free for its users, and network effects are the only path to that goal.
Paid memberships would cap growth, and payment proves nothing about
honesty, so a paywall would not even help against manipulation (decision 6).
Free tools are also the answer to the cold start problem: they are useful
to one person before any community exists.

The cost is that revenue arrives late and must be built actively on the
organization side. The early years run on savings, grants and patience.

Reconsider the organization-side mix freely as evidence arrives. Do not
reconsider charging individual users; that promise is load-bearing for
growth, positioning and trust. See [Growth](growth.md) for the full
reasoning.
