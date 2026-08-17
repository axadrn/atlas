# Decisions

This is the lightweight decision log for Atlas. It records choices that shape
multiple parts of the product. A decision can change, but not silently.

Each entry states the choice, why it currently wins, what it costs and what
evidence should reopen it.

## 1. International from day one

**Status:** accepted

Atlas uses a global place model, world map, currencies, units, time zones,
languages and localized names from the first release. No country or region is
the default worldview.

This wins because global participation is part of the product loop, not a
later translation project. Uneven data depth is acceptable and becomes visible
through coverage and freshness quests.

The cost is more careful place identity, localization and disputed-geography
modeling from the beginning.

Reconsider only if global scope prevents a usable public alpha from shipping.
The fallback would reduce imported depth, not replace the global data model.

## 2. Database-first with selective public data

**Status:** accepted

SQLite is the source of truth for live application data. Git contains code,
migrations, definitions, policies, importers and documentation. The live
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

**Status:** accepted

Atlas can monetize a changing collection of focused private tools for nomad
planning, tracking, documentation, logistics, finances, administration and
compliance. Specific examples are opportunities to validate, not commitments
that define the platform.

This wins because public place data can acquire users while private tools solve
high-value individual problems and support recurring revenue.

The cost is a wider product surface and different security requirements for
each tool. Toolbox data must remain separate from community contributions and
requires explicit review before any secondary use.

Reconsider individual tools based on demand, risk and maintenance cost. Keep
the separation between public community knowledge and private user workflows.
