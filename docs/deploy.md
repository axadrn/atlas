# Deploy

Atlas ships as a Docker image built by GitHub Actions. Dokploy only runs
the image; it never builds.

## The pipeline

1. Push to `main`.
2. The `docker` workflow builds the image and pushes
   `ghcr.io/axadrn/atlas:latest` plus a commit-sha tag.
3. The workflow calls the Dokploy deploy webhook. Dokploy pulls the fresh
   image and restarts the compose service from `docker-compose.yml`.
4. Migrations run automatically on start. The `atlas_data` named volume
   keeps `/data/atlas.db` across deployments.

## One-time setup

### Dokploy

- Create a Compose application pointing at this repository's
  `docker-compose.yml` (or paste the file). Attach the domain in Dokploy;
  route it to the `atlas` service on port 8090.
- Keep the `atlas_data` named volume. Do not replace it with an absolute
  host bind mount. Named volumes survive deployments and work with
  Dokploy's volume backups.
- Copy the application's deploy webhook URL from Dokploy. The URL embeds
  the deploy token, so treat the full URL as a secret.

### GitHub

- Repository Settings, Secrets and variables, Actions, New repository
  secret: name `DOKPLOY_WEBHOOK_URL`, value is the URL copied from
  Dokploy. Nothing else is needed; pushing to GHCR uses the built-in
  `GITHUB_TOKEN`.

## Data lifecycle

There are four separate paths for data changes. Keeping them separate makes
deployments repeatable and prevents a local database from replacing live
community data.

### Schema and initial catalog

Schema changes are append-only SQL migrations applied by the application at
startup. The initial catalog will also become a versioned SQL migration:
sources, places, source links, time zones and the initial curated ranking. A
fresh database must become a complete catalog by running migrations alone.

The initial catalog is the one allowed bulk data migration. Later provider
refreshes do not become migrations.

### Curated changes

Small reviewed changes such as ranking edits or a deliberately added place
travel as data migrations. They keep stable Atlas IDs and never rewrite
community records.

### Provider refreshes

Provider refreshes are explicit operations, not part of application startup
or deployment. A refresh follows one sequence:

1. Create a production backup.
2. Run the importer in dry-run mode.
3. Reject suspicious source versions, row counts or match counts.
4. Update whitelisted provider fields in one transaction.
5. Record the source version, checksum, counts and result.

The GeoNames refresh remains update-only. It matches the external ID of a
known place and may update its provider-managed name, coordinates,
population, time zone and retrieval time. It never inserts or deletes a
place and never changes Atlas-owned identity, hierarchy, destination or
ranking fields.

Community submissions do not write provider-managed place fields directly.
They remain separate reviewable records until Atlas deliberately accepts a
canonical correction. Field-level provenance is added when Atlas actually
starts combining different sources for the same field, not before.

### Community data

Once production accepts user-facing writes, production is the only source of
truth for accounts, sessions, reviews, votes and observations. A local or
staging database is never copied over production. Schema migrations may
transform user data in place, but seeds and provider refreshes do not own it.

## Backups and restore

Configure a daily Dokploy volume backup of `atlas_data` to an S3-compatible
destination. Stop the application container during the volume snapshot so
the SQLite database and its WAL sidecars are captured consistently. Test a
restore before relying on the schedule.

An online backup command may later use SQLite's backup API or `VACUUM INTO`
to create a consistent single-file snapshot without stopping the app. Never
copy a live `atlas.db` file on its own while WAL mode is active.

## SQLite operating boundary

SQLite stays on local storage attached to one server. WAL does not belong on
a network filesystem and there must be only one application instance writing
the database. Multiple Go connections inside that instance are fine; SQLite
still serializes writers.

The catalog can use `synchronous=NORMAL`. Before community writes launch,
durability should be reviewed and production should normally use
`synchronous=FULL` so a host power loss is less likely to lose recent commits.

## Open before launch

These items are decisions already made but not all are implemented yet:

- Export the current 2,435-place catalog into a versioned initial data
  migration and verify its licenses and attribution.
- Prove in a test that a fresh database contains the expected sources,
  places, source links, time zones and top-50 ranks after migrations.
- Ship the provider importer in the runtime image as an Atlas command. The
  current runtime image has neither the Go toolchain nor repository source,
  so `go run ./cmd/import` is not a production procedure.
- Add import-run auditing and defensive count checks before scheduling any
  provider refresh in Dokploy.
- Switch production durability to `synchronous=FULL` before accepting
  accounts, reviews, votes or other community writes.

## References

- [Dokploy Docker Compose storage](https://docs.dokploy.com/docs/core/docker-compose)
- [Dokploy volume backups](https://docs.dokploy.com/docs/core/volume-backups)
- [Dokploy scheduled jobs](https://docs.dokploy.com/docs/core/schedule-jobs)
- [SQLite online backup](https://www.sqlite.org/backup.html)
- [SQLite WAL](https://www.sqlite.org/wal.html)
