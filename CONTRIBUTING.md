# Contributing to Vizb

Vizb turns benchmark and tabular data into interactive HTML charts or JSON
datasets. Use this guide for local setup and pull requests. See the
[Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

## Ways to help

- Report bugs with the input, command, expected result, and actual result.
- Request features with the use case and desired outcome.
- Add parsers for benchmark frameworks or data formats.
- Improve documentation under `docs/`.

## Requirements

- Go 1.26+
- [Task](https://taskfile.dev/)
- pnpm

Install Task if needed:

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

The Dev Container is optional. Reopen the repository in the container; it
installs the toolchain and runs `task init`. Ports 5173 (UI) and 4321 (docs)
are forwarded.

## Setup and commands

```bash
task init            # install dependencies and generate the UI embed
task build           # build UI, CLI, and docs
task build:ui        # build UI; writes pkg/template/vizb-ui.gen.go
task build:cli       # build ./bin/vizb
task dev:ui          # run the UI dev server
task dev:docs        # run the docs dev server
task test            # run CLI and UI tests
task test:cli        # run Go tests
task test:ui         # run Vitest tests
task test:cover      # run Go coverage
task lint            # run CLI and UI linters
task lint:cli        # run golangci-lint
task lint:ui         # run Vue/TypeScript checks
task format          # format Go and UI files
task format:check    # check formatting without writing
```

Run one Go test:

```bash
go test -run 'TestSubjectSuite/TestCase' -v ./path/to/package
```

`pkg/template/vizb-ui.gen.go` is generated and gitignored. Never edit or commit
it. Run `task build:ui` after changing `ui/` and before Go builds or tests.

Run deploy-example workflows locally with `task act:install` and Docker:

```bash
task act:examples -- --only tabular-data,go
task act:examples -- --reuse --no-open
```

## Layout

```text
main.go        entry point
cmd/           Cobra commands and chart registration
pkg/parser/    CSV, JSON, and benchmark parsers
pkg/template/  embedded UI and HTML generation
shared/        shared types and utilities
ui/            Vue 3 + TypeScript app
docs/          Astro Starlight site
```

See [How It Works](https://vizb.goptics.org/internals/how-it-works/) for the
full data pipeline.

## Add a parser

Parser code lives in `pkg/parser/`. Follow the
[Parser Guide](https://vizb.goptics.org/guides/parsers/#add-a-new-parser): choose
an identifier, implement and register the parser, then add tests with a small
representative input fixture.

## Submit a change

1. Create a focused branch: `feat/...`, `fix/...`, `refactor/...`, or `docs/...`.
2. Add or update tests for affected behavior.
3. Check affected-feature coverage; add tests for gaps or describe remaining gaps.
4. Run `task test`, `task lint`, and `task format:check`.
5. Use [Conventional Commits](https://www.conventionalcommits.org/), such as
   `feat(parser): add Python pytest-benchmark support`.
6. Open a focused pull request describing what changed and why. Link related
   issues.

CI runs tests, coverage, linting, formatting, and builds on pull requests.

## Questions

Open an [issue](https://github.com/goptics/vizb/issues) or discussion on GitHub.
