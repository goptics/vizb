---
title: "Introduction"
description: "Your first interactive HTML chart from a table or benchmark — without writing chart code."
---

Vizb turns **CSV, JSON, and benchmark output** into interactive charts and stats **without writing chart code**. Point it at data (CLI, [GitHub Action](/ci-cd/github-action), or [REST](/commands/serve)), open one HTML report in any browser.

**Need the binary first?** [Install](/getting-started/install), then come back here.

## First chart

Given a CSV like this — use **Copy CSV**, save as `sales.csv`, then run the command:

```csv
order_date,region,category,product,quantity,amount
2024-01-01,Central,Hardware,Connector,16,2488.24
2024-01-02,East,Tools,Gear,42,207.08
2024-01-03,North,Mechanical,Sensor,20,3465.22
2024-01-04,West,Electronics,Valve,24,7633.44
2024-03-12,South,Industrial,Widget,31,5088.10
2024-06-18,East,Electronics,Relay,12,2214.50
2024-09-05,North,Hardware,Bolt,28,1724.27
2024-12-20,West,Tools,Gadget,19,2938.82
2025-02-08,Central,Mechanical,Valve,27,7350.26
2025-04-14,South,Electronics,Widget,41,3278.77
2025-07-22,East,Industrial,Connector,15,5093.42
2025-10-03,North,Tools,Gear,33,4102.15
2025-11-19,West,Hardware,Sensor,22,2890.40
2025-12-28,South,Mechanical,Gadget,26,611.32
```

### CLI

```bash
vizb bar sales.csv -g region,category -p x,y -o sales.html
```

### HTTP

```json
{
  "input": "order_date,region,category,product,quantity,amount\n2024-01-01,Central,Hardware,Connector,16,2488.24\n2024-01-02,East,Tools,Gear,42,207.08\n2024-01-03,North,Mechanical,Sensor,20,3465.22\n2024-01-04,West,Electronics,Valve,24,7633.44\n2024-03-12,South,Industrial,Widget,31,5088.10\n2024-06-18,East,Electronics,Relay,12,2214.50\n2024-09-05,North,Hardware,Bolt,28,1724.27\n2024-12-20,West,Tools,Gadget,19,2938.82\n2025-02-08,Central,Mechanical,Valve,27,7350.26\n2025-04-14,South,Electronics,Widget,41,3278.77\n2025-07-22,East,Industrial,Connector,15,5093.42\n2025-10-03,North,Tools,Gear,33,4102.15\n2025-11-19,West,Hardware,Sensor,22,2890.40\n2025-12-28,South,Mechanical,Gadget,26,611.32\n",
  "parser": "csv",
  "grouping": {
    "columns": [
      "region",
      "category"
    ],
    "pattern": "x,y"
  },
  "charts": {
    "types": [
      "bar"
    ]
  },
  "output": {
    "format": "html"
  }
}
```

### Action

```yaml
- uses: goptics/vizb@v0
  with:
    file: sales.csv
    group: region,category
    group-pattern: x,y
    charts: bar
    output-html: sales.html
```

Open `sales.html` in a browser. You get a grouped bar chart: region on X, category as series, one chart tab per numeric column (`quantity`, `amount`). The full repo sample is `examples/csv/sales.csv` (10,000 orders across 2024–2025).

| Flag | Role |
|------|------|
| `bar` | One chart type (smallest learning path) |
| `-g region,category` | Category columns → dimensions |
| `-p x,y` | First group column → X, second → Y series |
| `-o sales.html` | Self-contained HTML |

Zero flags (auto-group picks a categorical column when present):

```bash
vizb bar sales.csv -o sales.html
```

JSON works the same way (array of objects; nested objects flatten to dotted keys):

```bash
vizb bar data.json -o output.html
vizb bar data.json -g name -p x -o output.html
```

Nested arrays inside an envelope: [`--json-path`](/guides/data#selecting-a-nested-array-with---json-path).

> A `go test -bench -json` event stream is **not** treated as tabular JSON. It stays a Go benchmark. Full rules: [Tabular Data](/guides/data).

Detection is automatic. Force with `-P csv` or `-P json` when needed.

## How it works

Vizb is not a charting library you embed, and not a drag-and-drop dashboard. It is a **pipeline** (flags when you need them; auto-inference when you don’t):

```text
table or bench text
  → parser (auto-detected)
  → map rows to Name / X / Y / Z
  → Dataset
  → chart renderers + optional stats
  → HTML or JSON
```

| Idea | Meaning |
|------|---------|
| **Row → point** | Each data row becomes a point with dimensions and numeric stats |
| **Dimensions** | Up to four labels: Name / X / Y / Z — details in [Dimensions](/getting-started/dimensions) |
| **Group** | Categories become axes; each numeric column becomes its own chart; duplicate keys are summed |
| **Select** | Columns become coordinate axes (or picked metrics); rows stay separate points |
| **Charts** | Every chart type reads the same Dataset — see [Charts](/charts) |

**Outputs:** HTML (default interactive report), JSON (Dataset for merge/re-render), or the same conversions via [`vizb serve`](/commands/serve).

Deepen when you need to:

  
  
  
  

## Benchmarks (optional path)

Pipe benchmark output. The framework is detected from the content in most cases.

  ### Go

```bash
  go test -bench . | vizb -o output.html
  go test -bench . -json | vizb -o output.html
  ```

  From a saved file:

  ```bash
  go test -bench . > bench.txt
  vizb bench.txt -o output.html
  ```

  Split names such as `BenchmarkSort/1024/QuickSort`:

  ```bash
  go test -bench . | vizb -p n/x/y -o output.html
  ```

  ### Rust

Criterion and Divan from `cargo bench`:

  ```bash
  cargo bench | vizb -o output.html
  cargo bench > bench.txt && vizb bench.txt -o output.html
  ```

  ### JavaScript

Vitest (`npx vitest bench`) and Tinybench (`node bench.js`):

  ```bash
  npx vitest bench | vizb -o output.html
  npx vitest bench > bench.txt && vizb bench.txt -o output.html
  ```

  > Vitest benchmarking is experimental and does not follow SemVer. Output-format changes may break the parser. See the [Parser Guide](/guides/parsers).

Force a parser with `-P go`, `-P rs:criterion`, `-P rs:divan`, `-P js:vitest`, or `-P js:tinybench`.

> Prefer a single chart type while learning: `vizb bar …`. Use the root command with `--charts` when you want several renderers in one file. Save JSON to merge later: `vizb … -o data.json`, then `vizb ui data.json -o report.html`.

## Where to go next

| Goal | Start here |
|------|------------|
| Install | [Install](/getting-started/install) |
| Dimensions in depth | [Dimensions](/getting-started/dimensions) |
| Group vs select | [Group vs Select](/guides/group-vs-select) |
| CSV / JSON rules | [Tabular data](/guides/data) |
| Pattern / regex grouping | [Group](/guides/group) |
| Benchmarks | [Supported inputs](/guides/parsers) |
| Merge / CI | [Merging](/guides/merging), [GitHub Action](/ci-cd/github-action) |
| Live samples | [Examples](/examples) |
