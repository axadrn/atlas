# Trust and governance

Atlas is community governed. The founder should not become the daily judge of
city ratings, contribution disputes or interpersonal arguments. That goal
requires explicit authority, review and appeal systems. It cannot be achieved
by removing moderation.

The operator still retains responsibility for system security, urgent safety
issues, legal requests and the integrity of the governance process. This is a
rare break-glass role, not routine content moderation.

## Threat model

Atlas assumes from the first public release that people will try to:

- promote their city, country, business or political position;
- damage a competing place or business;
- create many accounts;
- buy or automate accounts;
- submit plausible but false prices;
- coordinate ratings through private groups;
- farm XP and reputation;
- capture local moderation;
- harass contributors or expose their location;
- poison imports and external integrations.

No honest architecture can promise that bots or false data have no chance.
The engineering goal is stronger and measurable: untrusted activity has
bounded influence, suspicious activity is delayed, decisions are explainable
and damage is reversible.

## Trust is not one number

Trust is contextual and earned.

A person who consistently submits useful meal prices in Tokyo does not
automatically become a visa expert or a global moderator. Reputation is scoped
by contribution type, place or region, and review role.

Signals may include:

- account age and security;
- accepted contribution history;
- later confirmation by independent contributors;
- reviewer calibration;
- successful challenges and corrections;
- diversity of activity over time;
- validated abuse reports;
- conflicts of interest.

No signal grants unlimited weight. Influence has hard caps so one account or
coordinated cluster cannot dominate an aggregate.

Payment, nationality, wealth and public follower counts do not increase trust.

## Contribution pipeline

New contributions do not immediately control public aggregates.

1. Validate structure, units, bounds and required context.
2. Apply account, action and network rate limits.
3. Estimate risk using privacy-conscious behavioral and coordination signals.
4. Store the submission and give the contributor a receipt.
5. Make low-risk contributions eligible after policy checks.
6. Route uncertain or high-impact contributions to peer review.
7. Recalculate versioned aggregates from eligible inputs.
8. Continue monitoring for later coordination or contradiction.
9. Reverse eligibility and rewards when evidence changes.

The contributor sees whether a submission is pending, eligible, disputed or
rejected and receives a reason when disclosure would not help attackers evade
detection.

## Sybil and bot resistance

There is no private, inclusive and perfect one-person-one-account test. Atlas
therefore layers controls instead of pretending email or payment proves
personhood.

The baseline includes:

- passkeys as the preferred account credential;
- verified recovery channels;
- per-account, per-action and per-place limits;
- account aging before high-impact privileges;
- one active rating per place, dimension and period;
- delayed influence for new accounts;
- bounded daily reputation and XP;
- IP, network and device-cluster signals with limited retention;
- detection of bursts, repeated values, shared wording and coordinated voting;
- robust aggregates that reduce the effect of outliers;
- separate views for locals, residents and visitors;
- random audits of accepted contributions;
- reversible eligibility, reputation and moderation actions.

Exact thresholds remain private because publishing them would provide an
evasion manual. The principles, categories of signals and aggregate methods
remain public.

High-friction challenges are risk based. Normal people should not repeatedly
solve captchas because attackers exist.

## Aggregation

Raw averages are not trusted rankings.

- prices use robust distributions such as medians and percentile ranges;
- small samples are visibly uncertain and shrink toward a broader baseline;
- stale observations lose relevance according to a published policy;
- repeated submissions from one person cannot create extra weight;
- coordinated account clusters have bounded combined influence;
- subjective ratings remain segmented by relevant experience;
- Atlas shows sample size and freshness next to every aggregate;
- a global best-city score is not produced from unsegmented votes.

The public methodology is versioned. Aggregate releases record which version
created them.

## Community roles

### Contributor

Submits observations, experiences, claims and reports. Every valid account can
begin here.

### Reviewer

Reviews limited categories where they have earned relevant reputation.
Assignments include random and adversarial cases, not only cases the reviewer
chooses.

### Steward

Maintains a category, place or region. Stewards can resolve routine disputes,
improve definitions and propose policy changes. Their actions are public and
auditable.

### Council

A rotating group of trusted community members handles appeals, steward
conflicts and governance changes. High-impact decisions require more than one
person and members must recuse themselves from conflicts of interest.

### Operator

Maintains infrastructure, responds to security incidents and legal obligations
and protects the governance process. The operator can take emergency action,
but every non-sensitive emergency action receives an audit record and later
community review.

## Promotion and removal

Review power is earned through sustained, category-relevant work and can be
lost. Promotions use published eligibility rules plus community review. Roles
expire or rotate so inactive accounts do not retain permanent authority.

No user can purchase a role. No single steward can permanently ban another
user, rewrite an aggregate or close their own dispute.

## Appeals

Every material moderation action has:

- a reason code;
- a human-readable explanation;
- the evidence that can safely be disclosed;
- a defined appeal period;
- review by people who did not make the original decision;
- a final recorded outcome.

Automation can limit reach or queue content temporarily. Permanent account
removal and high-impact factual disputes require accountable human review,
except for narrowly defined security emergencies.

## Gamification safety

XP and reputation are separate.

XP recognizes participation and progression. Reputation determines limited,
domain-specific trust. Neither is money and neither creates ownership rights.

Rewards favor:

- contributions that remain eligible over time;
- corrections of stale or demonstrably wrong data;
- useful peer review;
- confirmed reports;
- completion of under-covered community goals.

Rewards are capped, delayed and reversible. Duplicate, coordinated or
low-effort activity produces no advantage. Public leaderboards use narrow,
time-boxed categories and never expose sensitive presence.

## Transparency

Atlas publishes:

- trust and aggregation methodology;
- policy version history;
- anonymized moderation statistics;
- aggregate manipulation attempts and responses;
- governance membership and conflicts;
- public data release provenance;
- known limitations and unresolved disputes.

Transparency does not include private risk signals, personal information or
exact abuse thresholds.

## Non-negotiable principles

- Facts are not decided by popularity.
- Opinions are not presented as facts.
- Money cannot buy influence.
- New users can contribute without immediately controlling aggregates.
- Moderation power is scoped, reviewable and temporary.
- Private location is never the price of participation.
- Every high-impact change is attributable and reversible.
- The community handles routine governance, with a narrow operator backstop.
