# Vision

Atlas is building the best platform for digital nomads and perpetual
travellers. Free for nomads, forever. Companies pay.

This file is the long list of what Atlas wants to become. The ROADMAP.md
order still rules: foundations, world, community, trust, governance, then
business. Nothing here jumps that queue; this is the map of where the queue
leads.

## Principles

- Free forever for individual nomads. No paywall on data, community or
  tools. Ever. The incumbent model paywalls community and hides
  methodology; Atlas wins by doing the opposite.
- Radical transparency: every number shows its source and freshness, the
  ranking methodology is published, all filters stay free.
- Companies are the customers: job posts, featured listings, clearly
  labeled partner placements, commercial API tiers.
- Open source and open data stay. The moat is community, freshness and
  trust, not locked-up rows.
- The web app is the free product and carries everything an individual
  nomad needs. Public pages stay crawlable, which is all the visibility
  LLMs and search engines require. The API is a commercial product from
  day one: companies that want fresh, structured data on demand pay for
  it, without exceptions or free tiers.
- Simple beats complete. Every feature ships as the smallest vertical
  slice that is actually useful.

## Product universe

### Destinations (live today, keeps growing)

- Curated top 50 ranking, later replaced by community-driven scores with a
  published formula.
- Place pages with sourced facts, freshness and honest coverage gaps.
- Costs, internet, safety, weather, visa basics per destination, all
  community-verified with sample size and dates shown.

### Community

- Passkey accounts, profiles, private-by-default destination presence.
- Observations and reviews with trust and freshness mechanics, free to
  write and read.
- Meetups, free to attend.
- Blogs and long-form guides with explicit sourcing.
- An index of the free city groups where nomads already talk (Telegram,
  WhatsApp), instead of a competing paid chat.
- A light social layer: short posts, follows, opt-in check-ins.

### Work

- Remote job board; companies pay to post, priced well under the
  incumbents.
- Coworking and coliving directory; spaces pay for featured placement,
  basic listings stay free and community-editable.

### Nomad Toolbox

Free tools that solve real nomad pain, each one small and sharp. The
current tool landscape is fragmented across paid single-purpose apps;
bundling them free is the wedge.

- Tax residence day tracker first: 183-day rules, Schengen 90/180, US
  state day counting, treaty thresholds.
- US expat form helpers: FEIE Form 2555, FBAR reminders and prefills.
- Visa run planner and stay-limit alerts.
- Cost of stay calculator and trip budgeter.
- eSIM and travel insurance comparison, clearly labeled if partnered.
- Time zone meeting planner for distributed teams.
- More as pain points surface; tools are the top of the funnel.

## Deliberately not building

Lessons from the incumbents' most criticized features:

- No paid chat and no dating. The real community lives in free groups;
  Atlas links them instead of walling a copy.
- No opaque composite scores. Rankings publish their inputs and formula
  or do not rank on soft factors at all.
- No vibes-crowdsourced sentiment metrics about groups of people.
- No pay-to-review. Selection bias is not a feature.
- No bans without recourse. Moderation follows the governance rules in
  ROADMAP.md.

## Monetization ladder (B2B only)

1. Bridge revenue: clearly labeled affiliate partners (insurance, eSIM,
   banking) with conflict disclosures.
2. Featured coworking and coliving listings; job posts priced to undercut
   the incumbents.
3. The commercial API: live structured data and aggregates where
   licensing allows, paid from day one.
4. Employer and space subscriptions once volume justifies it, and
   expansion beyond the niche (remote hiring, relocation services) as the
   long-term growth path.

The open cost structure is the strategic advantage: Atlas can be
sustainable at revenue levels that would kill a venture-backed competitor.

## Next up

Milestone 2 kickoff: passkey auth slice on feat/passkey-auth. Users,
credentials and sessions tables per the key rules in data.md (usr_, ses_
prefixes), go-webauthn, session middleware with rotating HttpOnly cookies,
minimal sign-in dialog, security tests, manual review before merge.
