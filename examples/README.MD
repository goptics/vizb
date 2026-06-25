# Vizb examples

Sample inputs for the live dashboards on [vizb.goptics.org/examples/](https://vizb.goptics.org/examples/). CI turns each file into JSON, merges them into one HTML page per category, and deploys to the `website` branch.

## Categories

| Category | Folder | File | Workflow |
|----------|--------|------|----------|
| CSV | `examples/csv/` | `.csv` | [deploy-examples-csv.yml](../.github/workflows/deploy-examples-csv.yml) |
| JSON | _(API-fetched in CI)_ | jogruber contributions API | [deploy-examples-json.yml](../.github/workflows/deploy-examples-json.yml) |
| Go | `examples/go/` | `.txt` | [deploy-examples-go.yml](../.github/workflows/deploy-examples-go.yml) |
| JavaScript | `examples/javascript/` | `.txt` | [deploy-examples-javascript.yml](../.github/workflows/deploy-examples-javascript.yml) |
| Rust | `examples/rust/` | `.txt` | [deploy-examples-rust.yml](../.github/workflows/deploy-examples-rust.yml) |

Tabular CSV recipes (auto-group, auto-value, 3D): [csv/README.md](csv/README.md).

## Live dashboards

| Dashboard | URL |
|-----------|-----|
| CSV | [vizb.goptics.org/examples/csv/](https://vizb.goptics.org/examples/csv/) |
| JSON | [vizb.goptics.org/examples/json/](https://vizb.goptics.org/examples/json/) |
| Go | [vizb.goptics.org/examples/go/](https://vizb.goptics.org/examples/go/) |
| JavaScript | [vizb.goptics.org/examples/javascript/](https://vizb.goptics.org/examples/javascript/) |
| Rust | [vizb.goptics.org/examples/rust/](https://vizb.goptics.org/examples/rust/) |

Open a specific chart with `?d=0`, `?d=1`, … — index matches matrix order in the workflow file.

## Adding an example

### JSON (GitHub contributions)

No repo file to add. Append one row to the matrix in [deploy-examples-json.yml](../.github/workflows/deploy-examples-json.yml) — only `output` and `user`, in order:

```yaml
- output: 07-octocat    # next index + username — sets ?d= index and artifact name
  user: octocat         # chart title and API path; no other matrix fields needed
```

`output` must stay numbered (`00-`, `01-`, …) so merge order matches `?d=0`, `?d=1`, …. The workflow step already wires `curl`, `--json-path`, `--select`, date grouping, and 3D bar flags — copy a row, bump the number, set `user`.

Optionally add a row to [docs/src/content/docs/examples.mdx](../docs/src/content/docs/examples.mdx) under **JSON examples**.

### CSV / Go / JavaScript / Rust

1. **Add the file** — `examples/{category}/my-data.{csv,txt}`
2. **Add a matrix entry** — in `.github/workflows/deploy-examples-{category}.yml`, under `strategy.matrix.include`
3. **Push** — CI rebuilds that dashboard on [vizb.goptics.org](https://vizb.goptics.org)

```yaml
- output: 07-my-chart       # required — numbered id + sort order + ?d= index
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

Only `output`, `file`, and `name` are required. Use `chart` for scale, sort, and 3D flags — not the deprecated `sort` input or a bare `scale` field (action.yml has no `scale` input). Copy a nearby matrix row and edit.

For a new **CSV** file, also add a short blurb to [csv/README.md](csv/README.md).

## Try locally

```bash
vizb bar examples/csv/sales.csv -o out.html
vizb -P go examples/go/01-hash.txt -o out.html
```

Docs and install: [vizb.goptics.org](https://vizb.goptics.org)