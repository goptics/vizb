---
title: "Line Chart"
description: "Trace trends across categories with connected lines — one series per Y value, or a full 3D polyline scene."
---

The **line chart** draws connected data points across your X-axis categories, making it easy to spot trends, regressions, and performance curves. It follows the same dimension model as the bar chart — run the `line` subcommand and your data renders as lines instead of bars.

Use line charts when the shape of the curve matters as much as individual values: latency vs. input size, throughput vs. concurrency, or any benchmark where you want to see how values change across a sweep.

## How vizb builds it

vizb maps your grouping columns to the chart's visual axes:

| Dimension | Role |
|-----------|------|
| **XAxis** | Category axis — each unique X value is a point on the line |
| **YAxis** | Series axis — each unique Y value becomes a separate line |
| **ZAxis** | Z series — each Z value gets its own 3D polyline across the X/Y grid (3D only, requires echarts-gl) |

## Dimensions

Sample orders below (copy → `sales.csv`). Full file: [`examples/csv/sales.csv`](https://github.com/goptics/vizb/blob/main/examples/csv/sales.csv) (10,000 rows, 2024–2025). Live: [tabular-data](https://vizb.goptics.org/examples/live/tabular-data/).

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

### 1D (X axis)

A single line tracing one value per X category. The simplest view: one grouping column, one continuous curve from left to right.

### Agent

```text
/vizb sales.csv as line by order_date
```

### CLI

```bash
vizb line sales.csv -g order_date -p x -o out.html
```

### HTTP

```json
{
  "input": "order_date,region,category,product,quantity,amount\n2024-01-01,Central,Hardware,Connector,16,2488.24\n2024-01-02,East,Tools,Gear,42,207.08\n2024-01-03,North,Mechanical,Sensor,20,3465.22\n2024-01-04,West,Electronics,Valve,24,7633.44\n2024-03-12,South,Industrial,Widget,31,5088.10\n2024-06-18,East,Electronics,Relay,12,2214.50\n2024-09-05,North,Hardware,Bolt,28,1724.27\n2024-12-20,West,Tools,Gadget,19,2938.82\n2025-02-08,Central,Mechanical,Valve,27,7350.26\n2025-04-14,South,Electronics,Widget,41,3278.77\n2025-07-22,East,Industrial,Connector,15,5093.42\n2025-10-03,North,Tools,Gear,33,4102.15\n2025-11-19,West,Hardware,Sensor,22,2890.40\n2025-12-28,South,Mechanical,Gadget,26,611.32\n",
  "parser": "csv",
  "grouping": {
    "columns": [
      "order_date"
    ],
    "pattern": "x"
  },
  "charts": {
    "types": [
      "line"
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
    group: order_date
    group-pattern: x
    charts: line
    output-html: out.html
```

### 2D (X + Y axes)

One line per Y value, all sharing the same X axis. The legend lists the Y values so you can show or hide individual lines. Great for comparing multiple algorithms or configurations across the same input sweep.

### Agent

```text
/vizb sales.csv as line by region & category
```

### CLI

```bash
vizb line sales.csv -g region,category -p x,y -o out.html
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
      "line"
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
    charts: line
    output-html: out.html
```

Add `--stack` to render the grouped lines as a stacked area chart, which emphasizes the total per X value and each Y series' contribution to that total.

### Agent

```text
/vizb sales.csv as line by region & category, stacked
```

### CLI

```bash
vizb line sales.csv -g region,category -p x,y --stack -o stacked-area.html
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
      "line"
    ],
    "configs": [
      {
        "type": "line",
        "stack": true
      }
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
    charts: line
    chart: "line:stack"
    output-html: stacked-area.html
```

### 3D (X + Y + Z axes)

WebGL `line3D` scene: each Z value becomes a separate 3D polyline with scatter markers at each vertex. The markers keep individual data points readable when lines converge or overlap. Rotate, zoom, and pan the scene in the browser. Requires echarts-gl (bundled automatically).

This example uses [Go benchmark output](https://github.com/goptics/vizb/blob/main/examples/go/worker-pools.txt) comparing worker pool implementations. Slash-separated names like `BenchmarkAllSleep10ms/1u-1Mt/Pond-Eager-8` map to `z/y/x` — benchmark suite on Z, workload config on Y, pool implementation on X.

### Agent

```text
/vizb worker-pools.txt as line, split into depth, y & x, log scale
```

### CLI

```bash
vizb line worker-pools.txt -p z/y/x --scale log -o out.html
```

### HTTP

```json
{
  "input": "<contents of worker-pools.txt>",
  "parser": "go",
  "grouping": {
    "pattern": "z/y/x"
  },
  "charts": {
    "types": [
      "line"
    ],
    "configs": [
      {
        "type": "line",
        "scale": "log"
      }
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
    file: worker-pools.txt
    group-pattern: z/y/x
    charts: line
    chart: "line:scale=log"
    output-html: out.html
```

> Log scale is supported in 2D mode. When enabled, gaps appear at X values where the value is zero or negative — those points are skipped rather than breaking the chart.

> The Z axis requires **both** X and Y to be present. Passing a Z without an X/Y floor will be rejected with an error.

## `--select` (csv/json)

`--select` is **repeatable** (csv/json only). With explicit `-g` / `-p` / `-r`, it picks which numeric columns get their own chart (optional `{label}` renames each chart). Without group, it switches to solo value / mixed / multi-stat modes that plot raw columns as coordinate axes.

### Agent

```text
/vizb sales.csv as line by region & product, quantity & amount
```

### CLI

```bash
vizb line sales.csv -g region,product -p x,y --select quantity,amount -o lines.html
```

### HTTP

```json
{
  "input": "order_date,region,category,product,quantity,amount\n2024-01-01,Central,Hardware,Connector,16,2488.24\n2024-01-02,East,Tools,Gear,42,207.08\n2024-01-03,North,Mechanical,Sensor,20,3465.22\n2024-01-04,West,Electronics,Valve,24,7633.44\n2024-03-12,South,Industrial,Widget,31,5088.10\n2024-06-18,East,Electronics,Relay,12,2214.50\n2024-09-05,North,Hardware,Bolt,28,1724.27\n2024-12-20,West,Tools,Gadget,19,2938.82\n2025-02-08,Central,Mechanical,Valve,27,7350.26\n2025-04-14,South,Electronics,Widget,41,3278.77\n2025-07-22,East,Industrial,Connector,15,5093.42\n2025-10-03,North,Tools,Gear,33,4102.15\n2025-11-19,West,Hardware,Sensor,22,2890.40\n2025-12-28,South,Mechanical,Gadget,26,611.32\n",
  "parser": "csv",
  "grouping": {
    "columns": [
      "region",
      "product"
    ],
    "pattern": "x,y"
  },
  "select": [
    "quantity,amount"
  ],
  "charts": {
    "types": [
      "line"
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
    group: region,product
    group-pattern: x,y
    select: quantity,amount
    charts: line
    output-html: lines.html
```

See [Select](/guides/select) for the full mode reference, and [Group vs Select](/guides/group-vs-select) for when to use each.

## Auto-value mode (all-numeric data)

On an all-numeric CSV/JSON file, vizb auto-detects the first 2–3 columns as coordinate axes and renders value-type lines — with 3+ columns it auto-enables 3D (`line3D`). See [Auto-value](/guides/group-vs-select#auto-value-all-numeric-data) for the inference rules; solo `--select` overrides it.

### Agent

```text
/vizb spiral-3d.csv as line, 3d visual map, rotate
```

### CLI

```bash
vizb line spiral-3d.csv -o line.html --3d-visualmap --3d-rotate
```

### HTTP

```json
{
  "input": "<contents of spiral-3d.csv>",
  "parser": "csv",
  "charts": {
    "types": [
      "line"
    ],
    "configs": [
      {
        "type": "line",
        "threeDVisualMap": true,
        "threeDRotate": true
      }
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
    file: spiral-3d.csv
    charts: line
    chart: "line:3d-visualmap;3d-rotate"
    output-html: line.html
```

## Settings

These settings apply to line charts:

| Setting | CLI flag | UI toggle | Notes |
|---------|----------|-----------|-------|
| Sort | `--sort asc\|desc` | Sort control | Sorts points along the X axis |
| Stack | `--stack` | Stack series | Renders 2D grouped lines as stacked areas; ignored for z data and log scale |
| Labels | `--show-labels` | Show labels | Displays value at each data point |
| Smooth lines | `--smooth` | Smooth lines | Curves segments between points — 2D only |
| Swap | `--swap` | Axis switcher | Rotates which column maps to X vs Y |
| Scale (log) | `--scale log` or `--scale 'type=log;axes=x'` | Scale toggle | `linear` / `log` (value axis), or bag for per-axis log. Gaps at ≤0 |
| 3D view | `--3d` | 3D view | Value 3D for x+y data (y → depth, metric → height) |
| 3D visual map | `--3d-visualmap` | 3D visual map | Metric gradient coloring — on by default with `--3d`; grouped/value 3D |
| Auto-rotate | `--3d-rotate` | Auto rotate | Spins the 3D scene — 3D only |

Override settings for just this chart type without affecting others:

```bash
# Log scale and labels
vizb line sales.csv -g region,category --scale log -l -o out.html

# Per-axis log bag (unwrapped; no braces on --scale)
vizb line sales.csv --scale 'type=log;axes=x' -o out.html

# Sort descending
vizb line sales.csv -g region,category --sort desc -o out.html

# Stacked 2D area chart
vizb line sales.csv -g region,category -p x,y --stack -o stacked-area.html
```

## Next Steps
