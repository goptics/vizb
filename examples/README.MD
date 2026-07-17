# Vizb examples

Sample inputs for the live dashboards on [vizb.goptics.org/examples/](https://vizb.goptics.org/examples/). CI turns each recipe into JSON, merges them into one HTML page **per topic**, and deploys to the `website` branch.

## Topics

| Topic | Source folder | Workflow | Live dashboard |
|-------|---------------|----------|----------------|
| Tabular data | `examples/csv/` | [prepare-tabular-data-examples.yml](../.github/workflows/prepare-tabular-data-examples.yml) | [tabular-data](https://vizb.goptics.org/examples/tabular-data/) |
| Math & 3D | `examples/csv/` | [prepare-math-and-3d-examples.yml](../.github/workflows/prepare-math-and-3d-examples.yml) | [math-and-3d](https://vizb.goptics.org/examples/math-and-3d/) |
| Comparisons | `examples/csv/` | [prepare-comparisons-examples.yml](../.github/workflows/prepare-comparisons-examples.yml) | [comparisons](https://vizb.goptics.org/examples/comparisons/) |
| GitHub Legends | _(API-fetched in CI)_ | [prepare-github-legends.yml](../.github/workflows/prepare-github-legends.yml) | [github-legends](https://vizb.goptics.org/examples/github-legends/) |
| Go | `examples/go/` | [prepare-go-examples.yml](../.github/workflows/prepare-go-examples.yml) | [go](https://vizb.goptics.org/examples/go/) |
| JavaScript | `examples/javascript/` | [prepare-javascript-examples.yml](../.github/workflows/prepare-javascript-examples.yml) | [javascript](https://vizb.goptics.org/examples/javascript/) |
| Rust | `examples/rust/` | [prepare-rust-examples.yml](../.github/workflows/prepare-rust-examples.yml) | [rust](https://vizb.goptics.org/examples/rust/) |

CSV input files stay under `examples/csv/`; live HTML is split by **purpose** (tabular vs math vs comparisons), not by parser name.

Tabular CSV recipes (auto-group, auto-value, 3D): [csv/README.md](csv/README.md).

Docs: [vizb.goptics.org/examples/](https://vizb.goptics.org/examples/).

Open a specific chart with `?id=<id>` — each matrix entry sets a numbered id via the action `id` input (forwarded as `--id`). Keep ids chronological (`00-`, `01-`, …) **within each topic dashboard** so the prefix stays in sync with `?d=0`, `?d=1`, … index in the same matrix order.

## Adding an example

### GitHub Legends (contribution skylines)

No repo file to add. Append one row to the matrix in [prepare-github-legends.yml](../.github/workflows/prepare-github-legends.yml) — only `id` and `user`:

```yaml
- id: 07-octocat        # numbered ?id= deep link; prefix matches ?d=7 in matrix order
  user: octocat         # chart title and API path; no other matrix fields needed
```

`id` must stay numbered (`00-`, `01-`, …) in matrix order. The workflow step already wires `curl`, `--json-path`, `--select`, date grouping, and 3D bar flags — copy a row, bump the number, set `user`. The dashboard rebuilds on release and every **Friday** (scheduled workflow).

Optionally add a row to [docs/src/content/docs/examples/github-legends.mdx](../docs/src/content/docs/examples/github-legends.mdx).

### CSV topics (tabular / math / comparisons)

1. **Add the file** — `examples/csv/my-data.csv`
2. **Add a matrix entry** — in the matching workflow:
   - Business / mixed tables → [prepare-tabular-data-examples.yml](../.github/workflows/prepare-tabular-data-examples.yml)
   - All-numeric 3D / surfaces → [prepare-math-and-3d-examples.yml](../.github/workflows/prepare-math-and-3d-examples.yml)
   - Wide competitor tables → [prepare-comparisons-examples.yml](../.github/workflows/prepare-comparisons-examples.yml)
3. **Push** — CI rebuilds that dashboard on [vizb.goptics.org](https://vizb.goptics.org)

### Go / JavaScript / Rust

1. **Add the file** — `examples/{go,javascript,rust}/my-data.txt`
2. **Add a matrix entry** — in `.github/workflows/prepare-{go,javascript,rust}-examples.yml`
3. **Push** — CI rebuilds that dashboard

```yaml
- id: 07-my-chart          # required — numbered ?id= deep link; prefix matches ?d=7
  file: my-data            # required — basename only (no extension)
  name: "My Chart"         # required — title in the dashboard
  description: "…"         # optional — subtitle / tooltip
  charts: bar              # optional — csv: bar | line | scatter | …
  chart: |                 # optional — per-chart overrides (see action.yml)
    bar:scale=log,sort=asc
    line:scale=log,sort=asc
  group: "a,b,c"           # optional — csv; omit for auto-group / auto-value
  group-pattern: "n,x,y"   # optional — with group (commas for csv, slashes for benchmarks)
  group-regex: "…"         # optional — benchmarks; use instead of group-pattern
  filter: "Decode"         # optional — benchmarks: filter rows by name
  parser: rs:criterion     # optional — when auto-detect is not enough
```

Only `id`, `file`, and `name` are required. Use `chart` for scale, sort, and 3D flags — not the deprecated `sort` input or a bare `scale` field (action.yml has no `scale` input). Copy a nearby matrix row and edit.

For a new **CSV** file, also add a short blurb to [csv/README.md](csv/README.md) and the matching topic docs page under `docs/src/content/docs/examples/`.

## Try locally

```bash
vizb bar examples/csv/sales.csv -o out.html
vizb -P go examples/go/hash.txt -o out.html
```

Docs and install: [vizb.goptics.org](https://vizb.goptics.org)
