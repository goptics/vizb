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

- **Go 1.26+**
- **[Task](https://taskfile.dev/)** — the task runner used for all workflows
- **pnpm** — for the Vue UI and docs site

```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

### Dev Container (no local Go/Task/pnpm)

If you use [VS Code](https://code.visualstudio.com/), [Cursor](https://cursor.com/),
or [GitHub Codespaces](https://github.com/features/codespaces) with Docker, you
can skip installing Go, Task, and pnpm on the host:

1. Install the [Dev Containers](https://containers.dev/) extension (VS Code/Cursor)
   and a Docker-compatible runtime.
2. Open the repo and choose **Reopen in Container** (or create a Codespace).
3. On first create, the container installs the toolchain and runs `task init`
   automatically (deps + UI embed).

After that, use the same commands as a local setup (`task test`, `task build`,
`task dev:ui`, etc.). Ports **5173** (UI) and **4321** (docs) are forwarded.

> **Note:** `task act:*` (local GitHub Actions via [act](https://github.com/nektos/act))
> still needs Docker access from inside the container; that is not enabled by
> default in this Dev Container.

#### Host agent configs (Claude / Grok / Codex / OpenCode / …)

The container user is **`vizber`** (UID 1000) — the Microsoft image’s default
`vscode` login is renamed in the Dockerfile so nothing in the shell identity
uses that name. (`customizations.vscode` in `devcontainer.json` is only the
Dev Containers schema key for editor extensions; it is not the OS user.)

To reuse **your** machine’s agent setup (auth, skills, MCP, history), the
container bind-mounts common host directories from `$HOME` **twice**:

1. Under `/home/vizber/…` — what agents see via `$HOME`
2. Under the same absolute host path (e.g. `/home/<you>/…`) — so hardcoded
   hooks/settings and **path-keyed** MCP/project state still resolve

The repo is also mounted at the **same absolute host path**
(`workspaceFolder` = host path, not `/workspaces/vizb`). Agents key skills/MCP
by that path; using `/workspaces/…` would look like a different project.

| Host path | Inside container |
|-----------|------------------|
| repo checkout | **same absolute path** as on the host |
| `~/.claude`, `~/.agents`, `~/.grok`, `~/.codex`, … | `/home/vizber/…` **and** `$HOME/…` (host absolute) |
| `~/.orca`, `~/.junie` | same dual mount (hooks / extra skills) |
| `~/.local/bin` | `/home/vizber/.host-local-bin` **and** host `~/.local/bin` |
| `~/.local/share`, `~/.local/opt` | host absolute paths only (resolves symlinks like `claude` → `…/share/claude/versions/…`) |

Host CLIs are on `PATH`. Optional API keys (`ANTHROPIC_API_KEY`,
`OPENAI_API_KEY`, `XAI_API_KEY`) are forwarded from the host when set.

After changing mounts or the user rename, **fully rebuild** the Dev Container
(delete any old container first). Host binaries that depend on host-only libraries
may still fail inside the container; install the CLI in the container or run the
agent on the host against the mounted workspace.

If `pnpm` fails with `attempt to write a readonly database`, a root-owned
`.pnpm-store/` is left in the repo (often from an old Docker-as-root run). On the
**host**:

```bash
sudo rm -rf .pnpm-store
# or without sudo:
docker run --rm -v "$PWD":/w -w /w alpine rm -rf .pnpm-store
```

The Dev Container pins **pnpm 10.x** and forces the store under
`/home/vizber/.local/share/pnpm/store` so installs do not use a repo-local store.

## Setup

```bash
task init    # install deps (Go, UI, docs) and generate the UI embed
```

`task init` runs `task build:ui` so `pkg/template/vizb-ui.gen.go` exists for
`go test` / `go build`. That file is **gitignored** (generated locally and in
CI on Node 22). Never commit it. (In the Dev Container this runs once on first create.)

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
