# Vizb Agent Guide

## Project

Vizb is a Go CLI that converts CSV, JSON arrays, structured data, and Go, Rust,
or JavaScript benchmark output into self-contained interactive HTML charts or
JSON datasets. It embeds a Vue 3 UI and also ships as a GitHub composite action.

## Commands

Use [Task](https://taskfile.dev/) from the repository root.

```bash
task init        # install Go, UI, and docs dependencies
task build       # build docs, embedded UI, and CLI
task build:ui    # build UI and regenerate pkg/template/vizb-ui.gen.go
task build:cli   # build ./bin/vizb
task dev:ui      # run Vue dev server
task dev:docs    # run docs dev server
task test        # run CLI and UI tests
task test:cli    # run Go CLI tests without cache
task test:ui     # run Vitest UI tests
task test:cover  # run cross-package Go coverage
task lint        # run CLI and UI linters
task lint:cli    # run Go linter
task lint:ui     # run Vue/TypeScript type checking
task format      # format CLI and UI files
task format:cli  # format Go files (skips vizb-ui.gen.go)
task format:ui   # format UI files
task format:check     # check all formatting without writing
task format:check:cli # check Go formatting (skips vizb-ui.gen.go)
task format:check:ui  # check UI formatting
```

### Embedded UI gen file

`pkg/template/vizb-ui.gen.go` is the Vue chart UI embedded as Go source so
`go install` / `go build` work with **only the Go toolchain**. End users and
Go-only contributors do **not** need Node.

| Rule | Detail |
|------|--------|
| Tracked on main | Committed so pure-Go install works |
| Never hand-edit | Header is `// Code generated ... DO NOT EDIT.` |
| Regenerated with Node | `task build:ui` (Vue/Vite + embed plugin) — maintainer/UI work only |
| gofmt | Skipped for this file in Taskfile and CI format jobs |
| golangci-lint | Still runs on it (no special exclude) |

After a fresh clone, the tracked gen file is enough for CLI builds and tests.
After changing `ui/`, run `task build:ui` locally to regenerate before Go
builds. Prefer not to land gen churn on feature PRs that only touch Go or docs;
ongoing sync of the committed snapshot when `ui/` lands on main is owned by the
follow-up CI bot work (see #319).

## Architecture

```text
main.go -> cmd/                 Cobra root, merge, ui, and chart commands
           cmd/charts/<type>/   chart registration and Cobra metadata
           cmd/cli/             FlagBag, command builder, linear pipeline
           internal/charts/     chart configs, flags, rules, materialisation
           pkg/parser/          format detection and parsers
           pkg/template/        Vue embed and HTML generation
           shared/              datasets, merge logic, common utilities
           ui/                  Vue 3 + TypeScript application
```

Root flow: read file or stdin -> detect/parse format -> group data -> build
`shared.Dataset` -> emit JSON or self-contained HTML.

## Engineering Rules

- Keep changes scoped. Read nearby implementation and tests before editing.
- Bind CLI flags through `cli.FlagBag`; define chart flags in
  `internal/charts` and compose them in `cmd/charts/<type>`.
- New chart types must register their factory, flags, and Cobra metadata, then
  be blank-imported by `cmd/root.go`.
- Manage temporary files through `shared.TempFiles` and preserve cleanup on all
  exit paths.
- Preserve output compatibility unless the task explicitly changes it. `.json`
  output is JSON; other extensions produce HTML. Stdin format detection uses
  content, not extension.
- Do not hand-edit generated files or unrelated user changes.

## Test Design

- New or materially expanded Go tests use `testify/suite`. Name suites
  `<Subject>Suite`, write `TestX` methods, and add one `suite.Run` entry point
  per suite.
- Put shared fixtures and setup on the suite. Use `s.Run` for subcases;
  table-driven cases may live inside suite methods.
- UI tests group related behavior with Vitest `describe`, use `it` for cases,
  and keep hooks at the narrowest useful scope. Layers (see `ui/TESTING.md`):
  - unit: `src/**/*.test.ts` (Vitest node)
  - integration: `src/**/*.integration.test.ts` (Vitest happy-dom)
  - e2e: `ui/e2e/**/*.spec.ts` (Playwright)
  Prefer the lowest layer that fails for the right reason. Share fixtures from
  `ui/src/test-utils`. Settings visibility lives only in `fieldRegistry` tests.
- Keep tests deterministic and beside tested code. Parser additions include a
  small representative fixture.
- Do not migrate unrelated existing tests solely to satisfy these conventions.
- After every change, assess affected-feature coverage, add or update tests when
  needed, run focused checks, and call out remaining coverage gaps.

Run the narrowest relevant tests while developing. Before handoff, run checks
matching the touched areas:

```bash
go test -count=1 ./path/to/package   # focused Go change
task test:cli                        # all CLI tests
task test:ui                         # UI unit + integration
task test:ui:coverage                # UI tests with 100% coverage gates
task test:ui:e2e                     # UI Playwright smokes
task lint:cli                        # Go linting
task lint:ui                         # UI type checking
task format:check                    # all formatting checks
task test && task lint               # repository-wide test and lint validation
```
