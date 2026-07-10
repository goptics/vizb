# Contributing to Vizb

Thanks for your interest in improving Vizb! This guide covers how to set up the
project, make a change, and get it merged. By participating you agree to abide
by our [Code of Conduct](CODE_OF_CONDUCT.md).

## Ways to contribute

- **Report a bug** — open an issue with input, command, and what you expected.
- **Request a feature** — describe the use case, not just the solution.
- **Add a parser** — bring a new benchmark framework or data format (see below).
- **Improve the docs** — the site lives in `docs/`.

## Prerequisites

- **Go 1.24+**
- **[Task](https://taskfile.dev/)** — the task runner used for all workflows
- **pnpm** — for the Vue UI and docs site

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

## Setup

```bash
task init    # install deps (Go, UI, docs) and generate the UI embed
```

`task init` runs `task build:ui` so `pkg/template/vizb-ui.gen.go` exists for
`go test` / `go build`. That file is **gitignored** (generated locally and in
CI on Node 22). Never commit it.

## Build & test

```bash
task build       # build UI + binary
task build:ui    # build the Vue UI only (writes pkg/template/vizb-ui.gen.go)
task build:cli   # build the Go binary to ./bin/vizb
task test        # go test -count=1 ./...
task lint        # golangci-lint run
task format      # gofmt + pnpm format
```

Run a single test:

```bash
go test -run TestName -v ./path/to/package
```

### Local example workflows

You can run the `deploy-examples` CI workflows on your machine with
[act](https://github.com/nektos/act) and Docker (GitHub Pages deploy is skipped).
Install act via `task act:install`, then:

```bash
task act:examples                              # all languages, opens browser preview
task act:examples -- --only csv,go             # subset of languages
task act:examples -- --reuse --no-open         # faster reruns, no browser
```

Output lands under `dist/examples/` with an overview at `dist/examples/index.html`.
Equivalent script: `./scripts/act-examples.sh` (same options).

> **Important:** `pkg/template/vizb-ui.gen.go` is generated from the Vue app
> (`EMBED_UI=True pnpm build` / `task build:ui`). Do **not** hand-edit it and
> do **not** commit it. After any change under `ui/`, run `task build:ui` before
> Go commands so the embedded UI is current.

## Project layout

```
main.go        entry point
cmd/           Cobra CLI commands (root, merge, ui, chart subcommands)
pkg/parser/    input parsing — CSV, JSON, and benchmark frameworks
pkg/template/  embeds the built Vue UI and generates the HTML output
shared/        cross-package types, merge logic, utilities
ui/            Vue 3 + TypeScript app (built into a single inline HTML bundle)
docs/          Astro Starlight documentation site
```

See [Internals → How It Works](https://vizb.goptics.org/internals/how-it-works/)
for the full data pipeline.

## Adding a parser

Vizb's biggest contribution surface is new input formats. The parser registry
lives in `pkg/parser/`. The end-to-end steps — choose a key, implement the parse
function, register it, and add tests — are documented in the
[Parser Guide → Add a New Parser](https://vizb.goptics.org/guides/parsers/#add-a-new-parser).
Please include a small sample input file with your tests.

## Submitting a change

1. Fork the repo and create a branch (`feat/…`, `fix/…`, or `docs/…`).
2. Make your change with tests where it makes sense.
3. Run `task test` and `task lint` — both must pass.
4. Use [Conventional Commit](https://www.conventionalcommits.org/) messages
   (e.g. `feat(parser): add Python pytest-benchmark support`), matching the
   existing history.
5. Open a pull request describing **what** changed and **why**. Link any
   related issue.

CI runs the test suite and coverage on every PR. Keep PRs focused — one logical
change per PR is easier to review and merge.

## Questions

Open an issue or start a discussion on
[GitHub](https://github.com/goptics/vizb). We're happy to help.
