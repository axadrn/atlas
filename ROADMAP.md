# Roadmap

This roadmap follows the product loop. It is ordered to test whether travelers
will use and improve Atlas, not to maximize the number of features.

## 0. Foundations

- Record the product, data and governance decisions.
- Define privacy, contribution and moderation policies.
- Establish a lightweight decision log for future architectural changes.
- Keep the site out of search indexes until the real name and domain exist.

Exit condition: the first database and product work can be judged against
written principles instead of memory.

## 1. The world

- Add SQLite migrations and a typed Go storage layer.
- Curate global place identities and useful hierarchies with stable Atlas IDs.
- Track the source and license of every externally sourced fact.
- Build an interactive map using OpenStreetMap-derived data and replaceable
  tile infrastructure.
- Build local place search without depending on the public Nominatim service.
- Ship an installable PWA shell without caching private or live data responses.
- Create place routes with parent navigation, coverage and freshness states.
- Reserve `/api/v1` for authenticated first-party and external clients.
- Define the first public aggregate schema without exposing the application
  database or raw community submissions.

Exit condition: a visitor can explore the world, find a destination and
understand what Atlas knows and does not know about it.

## 2. Community alpha

- Add passkey-first accounts and secure sessions.
- Add optional profiles and private-by-default destination presence.
- Define the first structured observation categories.
- Build rotating quests for missing and stale data.
- Add contribution receipts and clear publication states.
- Add an append-only XP and reputation ledger.
- Show immediate personal feedback without immediately trusting the input.
- Add reporting, blocking, audit records and basic review queues.
- Add action-specific rate limits and the first coordination signals.

Exit condition: a new user can contribute in under a minute, understand what
happened to the contribution and return to continue building a destination.

## 3. Trust and useful aggregates

- Implement eligibility policies for each contribution type.
- Add peer confirmation and evidence-based disputes.
- Publish robust price and experience distributions.
- Show sample size, freshness, uncertainty and methodology everywhere.
- Add scoped reviewer reputation and random audits.
- Detect coordinated activity and cap cluster influence.
- Add reversible moderation and aggregate versions.
- Run recovery, abuse and privacy drills before increasing reach.

Exit condition: a burst of new or coordinated accounts cannot materially move
a destination aggregate without independent evidence and time.

## 4. Community governance

- Add reviewer promotion and expiration.
- Add place and category stewards.
- Add conflict declarations and recusal.
- Add independent appeals and rotating councils.
- Publish governance and moderation transparency reports.
- Limit operator activity to infrastructure, emergencies and governance
  integrity.

Exit condition: routine contribution and moderation disputes are resolved by
the community with reasons, appeals and no founder involvement.

## 5. Community depth

- Add saved places, change alerts and personal comparisons.
- Add guides and long-form posts with explicit sourcing.
- Add scoped discussions only where they improve decisions.
- Add safe, opt-in social discovery without exposing live location.
- Add contributor collaboration and local data campaigns.

Exit condition: community content improves retention without overwhelming the
review and safety systems.

## 6. Sustainable business and private tools

- Launch and validate focused Nomad Toolbox utilities.
- Keep basic access and corrections open.
- Add paid planning tools, alerts, personal exports and support.
- Offer selected public or paid API endpoints only where privacy, licensing and
  business incentives align.
- Add clearly labeled partner services with conflict disclosures.
- Reward sustained contributors with product access, not extra voting power.

Exit condition: revenue funds data quality and operations without purchasing
truth or governance influence.

## Build order inside every milestone

Each feature follows the same order:

1. Define the user value and abuse case.
2. Define public and private data boundaries.
3. Define the state transition and rollback path.
4. Implement the smallest complete vertical slice.
5. Test normal use, abuse, privacy and recovery.
6. Measure real behavior.
7. Keep, change or remove it.

Anything touching authentication, payments, private location, moderation or
user data requires manual security review before release.
