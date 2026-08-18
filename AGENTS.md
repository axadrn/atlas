# Agent rules

Rules for AI agents working in this repo.

## Dev workflow

- Treat installed shadcn-templ component sources as upstream code. Do not
  modify, trim or delete them. Compose them in application code, and install
  missing components through the shadcn-templ CLI.
- Never run `templ generate` manually. `task dev` runs the templ watcher and
  regenerates on every change. A manual run fights the watcher, breaks the
  live dev session and forces a restart.
- Same for Tailwind: `task dev` watches `assets/css/globals.css`. Do not run
  `tailwindcss` by hand while dev is running.
- Production builds happen in the Dockerfile, not locally.
- Need a one off compile check? `go build ./...` is fine, it does not touch
  generated files.

## Writing

- Everything in English: code comments, docs, commit messages, PR text.
- No em dashes, no AI boilerplate. Write like a human.
- No Co-Authored-By trailers in commits.
- Commit format: see the conventional commits rule in CONTRIBUTING.md.

## Review

- Anything touching auth, payments or user data gets manual review before
  merge. Do not merge such changes on your own.
