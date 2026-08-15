# Contributing

## Dev setup

```
go mod tidy
task dev
```

That is all. The templ watcher and Tailwind run inside `task dev`, so never
run `templ generate` or `tailwindcss` by hand while it is up.

## Rules

Vibecoded PRs are welcome. AI wrote half of this repo, banning it would be
weird. The rules:

1. You own your PR. You can explain every line when asked. "The AI did it"
   is not an answer, it is a self report.
2. Small PRs, one thing per PR. Drive-by refactors get closed without mercy.
3. Data changes need sources. A tax rate without a link to an official
   source does not get merged.
4. CI must be green.
5. Anything touching auth, payments or user data gets a slow, manual,
   grumpy review. Plan for it.
6. Write like a human. No AI boilerplate in commits, comments or docs.
   No em dashes.
7. Commits follow conventional commits, the simple version: `feat:`, `fix:`,
   `docs:`, `chore:`, `ci:`, `refactor:`. Lowercase, one line, no scopes,
   no body unless it really needs one.

By contributing you agree that your contribution is licensed like the rest
of the repo: MIT for code, CC BY 4.0 for data.
