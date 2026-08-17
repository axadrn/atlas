# Product

This document records the current product thesis. It is a decision, not a
promise that can never change. When a better idea wins, this document changes
with it.

## Why Atlas should exist

Travel information is fragmented across closed databases, SEO pages, private
groups and social posts. A useful fact is often missing its source, date,
context or methodology. A city score hides who voted and what they cared
about. Legal claims are repeated long after the rule changed.

Atlas should make travel knowledge inspectable. It should show what was
observed, what is opinion, what is a sourced claim, how fresh it is and how
much evidence supports it.

The goal is not another list of the best cities. The goal is a living world
map that travelers improve together.

## Product thesis

Atlas is the open source home for perpetual travelers and digital nomads. A
visitor explores the world, opens a city, contributes a small piece of current
knowledge and sees the city improve because of it.

The core loop is:

1. Discover a place on the world map.
2. See what the community currently knows and where the gaps are.
3. Optionally check in at city level.
4. Complete one short data quest.
5. See how the contribution compares with the current distribution.
6. Earn reputation after the contribution proves useful.
7. Return to track places, complete quests and help review other contributions.

The product must provide value before it asks for work. Contributions are
optional and small. A forced questionnaire would reduce participation and
encourage careless answers.

## International from day one

International means:

- the world map and place model are global;
- countries, regions and cities have stable identifiers;
- currencies, units, time zones, languages and local names are first-class;
- a person can contribute anywhere;
- no nationality or region is the default worldview.

It does not mean every place starts with equal depth. Missing and stale data
become visible community quests instead of being hidden.

## Four kinds of knowledge

Atlas never mixes these into one generic rating or fact.

### Observation

Something a person measured or paid at a time and place. Examples include an
apartment price, a meal price, a mobile plan or an internet speed test.

Observations require context. A rent observation without currency, period,
accommodation type, rough area and date is not useful.

### Experience

A subjective report such as perceived safety, noise, friendliness or
walkability. Experiences remain opinions. They are segmented by relevant
context and shown as distributions, not declared true or false.

### Claim

A falsifiable statement about the world, including a visa condition, tax rule
or transport policy. A claim can be supported or disputed with evidence. Votes
alone never turn a claim into a fact.

### Content

Profiles, guides, articles, discussions and comments. Content can provide
context and community, but it does not silently become structured data.

## The interactive map

The map is the main navigation surface, not decoration. It shows places,
freshness, coverage and active community quests. It should make both knowledge
and missing knowledge visible.

OpenStreetMap can provide map and place data. Its data, rendered tiles,
geocoding services and query services are different products with different
operational limits. Atlas will use replaceable providers and preserve required
attribution instead of depending permanently on donated public servers.

## Gamification

Gamification exists to improve data quality and reward stewardship. It must
not reward raw activity.

Good rewards include:

- explorer levels;
- country, region and city badges;
- first useful contributor badges;
- city coverage goals;
- quests for missing or stale observations;
- category-specific reputation;
- profile maps of contributed places;
- product benefits for sustained, accepted contributions.

Points are awarded after a contribution becomes eligible, not merely after it
is submitted. Rewards may be delayed, capped or reversed if the contribution
is later rejected. Paid membership never increases influence.

Global volume leaderboards are deliberately avoided. They reward spam, favor
people with more time and turn contribution into a race. Local and time-boxed
community goals are healthier.

## First public alpha

The first public alpha should contain:

- a global interactive map;
- country, region and city identities;
- place search from Atlas's own database;
- city pages with coverage and freshness;
- accounts and optional public profiles;
- private-by-default city-level presence;
- structured observations and experiences;
- rotating data quests;
- an append-only XP and reputation ledger;
- basic place comparison;
- contribution review, reporting and rollback;
- visible source, date, sample size and uncertainty.

The first alpha does not need:

- direct messages;
- a general blogging platform;
- comments on every object;
- paid membership;
- a stable public API;
- a single global city ranking;
- unsourced visa or tax claims presented as advice.

These features can be added when the core loop produces useful data and people
return to use it.

## Privacy defaults

Presence is sensitive. Atlas stores only the precision required by the
feature. City presence is optional, private by default and automatically
expires. Exact live location is never required for a contribution and is not
included in public data releases.

Public profiles are opt-in. Travel history, nationality and residency are not
public unless a person explicitly chooses to share them. Private information
must never become a trust shortcut merely because it is easy to collect.

## Business guardrails

Atlas is allowed to make money. Revenue cannot buy truth, moderation power,
rating weight or better placement disguised as community consensus.

Possible paid value includes advanced comparisons, alerts, planning tools,
exports, higher API limits, support and clearly labeled partner services. The
public provenance of data and the right to correct it remain available to
everyone.

## Nomad Toolbox

The Nomad Toolbox is a general paid product surface, not one predetermined
tool. It can contain small, focused utilities for planning, tracking,
documentation, logistics, finances, administration and compliance.

Individual tools are chosen from observed user problems. A tax presence
tracker or form preparation assistant may be an example, but neither defines
the toolbox. Every tool receives its own value, privacy, security and liability
review before implementation.

Community data helps people discover and understand places. Private tools help
an individual act on that knowledge. Private toolbox inputs never become
community data unless the person makes a separate, explicit contribution.

## What success looks like

The early product is working when:

- visitors find useful place information before being asked to contribute;
- contributors complete small quests and return;
- accepted observations remain useful after review;
- cities become fresher through community activity;
- manipulation attempts have bounded and reversible impact;
- people inspect and cite Atlas data outside Atlas;
- community reviewers resolve routine cases without founder intervention.
