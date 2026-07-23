# CSV examples

Sample datasets for trying vizb chart types. Each file below shows a different data shape and how vizb maps it to charts.

Run any example from the repo root:

```bash
vizb bar examples/csv/sales.csv -o out.html
vizb scatter examples/csv/noise-grid.csv -o out.html
```

Use a chart subcommand (`bar`, `line`, `scatter`, `pie`, …) or the root command with `--charts`. CSV is auto-detected when you pass a `.csv` file.

---

## `sales.csv`

**Shape:** Mixed categorical + numeric columns (10,000 order rows).

| Column kind | Examples |
|-------------|----------|
| Categorical | `order_date`, `region`, `product`, `category`, `customer_name`, … |
| Numeric (metrics) | `quantity`, `amount`, `total`, `rating`, … |

**What vizb does:** With no `--group`, vizb **auto-groups** by the categorical column with the most distinct values (usually `order_date`) and plots numeric columns as series. With explicit `--group` / `--group-pattern`, you control name / x / y / z layout and 3D grouping.

**Good for:** 2D bar/line/scatter, grouped 3D bar, log scale, date splitting, swap.

| Goal | Command |
|------|---------|
| Auto-group (simplest) | `vizb bar examples/csv/sales.csv` |
| Grouped 3D bar (log) | `vizb bar examples/csv/sales.csv -g product,region,category -p n,x,y --3d --scale log` |
| Sales by date + category | `vizb bar examples/csv/sales.csv -g order_date,category -p "[n{Year}-y{Month}-x{Date}],z{Category}"` |
| Line / scatter on same data | `vizb line examples/csv/sales.csv` · `vizb scatter examples/csv/sales.csv` |

These variants are also built in CI as `00-sales-auto-group`, `01-sales-grouped`, and `02-sales-by-date` on the **tabular-data** dashboard (see `.github/workflows/tabular-data-examples.yml`).

---

## `spiral-3d.csv`

**Shape:** Three numeric columns only — `x`, `y`, `z` (25,000 points along a 3D spiral).

**What vizb does:** No categorical columns → **auto-value mode**. First three columns become continuous `x`, `y`, `z` axes; 3D is enabled automatically for bar / line / scatter.

**Good for:** Continuous `scatter3D` / `line3D` / `bar3D`, camera rotate.

| Goal | Command |
|------|---------|
| Auto-value 3D scatter | `vizb scatter examples/csv/spiral-3d.csv` |
| Auto-rotate | `vizb scatter examples/csv/spiral-3d.csv --3d-rotate` |
| 3D bar or line | `vizb bar examples/csv/spiral-3d.csv` · `vizb line examples/csv/spiral-3d.csv` |

---

## `noise-surface.csv`

**Shape:** Three numeric columns — `x`, `y`, `z` on a **21×21** grid (441 points). `z` is a simplex-noise height field over `(x, y)`.

**What vizb does:** Auto-value mode (2D coordinates + height on `z`). Renders as a continuous 3D surface-style bar chart.

**Good for:** 2D grid → 3D height (bar3D), quick noise demo (smaller than the grid examples).

| Goal | Command |
|------|---------|
| Noise height field | `vizb bar examples/csv/noise-surface.csv` |
| Scatter view of same points | `vizb scatter examples/csv/noise-surface.csv` |

---

## `noise-grid.csv`

**Shape:** Four numeric columns — `x`, `y`, `z`, `value` on a **21³** voxel grid (9,261 points, indices 0–20). `value` is the visual metric (simplex noise × 2 + 4).

**What vizb does:** Auto-value mode: `x,y,z` = position, 4th column = **metric**. Enables **3D** and **visualMap** automatically (point color and size by `value`), orthographic `scatter3D` like the [ECharts simplex-noise demo](https://echarts.apache.org/examples/en/editor.html?c=scatter3D-simplex-noise&gl=1).

**Good for:** `scatter3D` + visualMap on a manageable file size.

| Goal | Command |
|------|---------|
| 3D scatter + visualMap (default) | `vizb scatter examples/csv/noise-grid.csv` |
| Explicit visualMap flag | `vizb scatter examples/csv/noise-grid.csv --3d-visualmap` |

---

## `noise-grid-41.csv`

**Shape:** Same as `noise-grid.csv` but **41³** (68,921 points, indices 0–40) — full resolution matching the ECharts example loop (`i,j,k` from 0 to 40, `value = noise3D(i/20,j/20,k/20)*2+4`).

**What vizb does:** Same auto-value + metric + visualMap behavior as `noise-grid.csv`. Heavier dataset; same chart settings.

**Good for:** Full-density ECharts-style demo when you want every voxel.

| Goal | Command |
|------|---------|
| Full ECharts-scale grid | `vizb scatter examples/csv/noise-grid-41.csv` |

Use `noise-grid.csv` for faster loads; use this file when you need the complete grid.

---

## `region-metrics.csv`

**Shape:** Mixed categorical + numeric — `region` (category), `latency`, `sales` (8 rows).

**What vizb does:** Solo `--select` axis mode. `--select region,latency` enters **mixed mode** (category x + value y). Repeat `--select` for multi-view output (e.g. `region,latency` and `region,sales` in one HTML).

**Good for:** Mixed scatter demos, multi-dataset tabs, `--select` axis mode docs.

| Goal | Command |
|------|---------|
| Mixed scatter (region vs latency) | `vizb scatter examples/csv/region-metrics.csv --select region,latency` |
| Mixed 3D (region × latency × sales) | `vizb scatter examples/csv/region-metrics.csv --select region,latency,sales` |
| Multi-view (two datasets) | `vizb scatter examples/csv/region-metrics.csv --select region,latency --select region,sales` |

---

## `house-price-area2.csv`

**Shape:** `area`, `price` — 16,174 rows ([ECharts house-price scatter](https://echarts.apache.org/examples/en/editor.html?c=scatter-large)). Auto-value xy; add `--visualmap` for price gradient. CI id: `03-house-price-area2` (tabular-data dashboard).

---

## `clusters.csv`

**Shape:** `x`, `y` — 60 rows of 2D cluster coordinates. Auto-value xy; `--visualmap` and `--symbol-size 10` for sized, color-mapped points. CI id: `04-clusters` (math-and-3d dashboard).

| Goal | Command |
|------|---------|
| Cluster scatter with visualMap | `vizb scatter examples/csv/clusters.csv --visualmap --symbol-size 10 -o out.html` |

---

## `concurrency.csv`

**Shape:** Wide competitor table — one category column (`load`) and several numeric framework columns (`default`, `chi`, `echo`, `gin`, `goframe`, `httpz`). Throughput-style values at three load levels.

**What vizb does:** With `-g load -p y --col-axis x`, load becomes the Y series dimension and **framework column names land on X** as categories. All competitors share **one chart** (instead of one chart per numeric column). Chart title falls back to the dataset name (`-n`) because expanded stats omit `type`; use `--title` to override just that chart title.

**Good for:** Side-by-side library / framework comparison from wide CSV.

| Goal | Command |
|------|---------|
| Frameworks on X, load as series | `vizb bar examples/csv/concurrency.csv -g load -p y -A x` |
| Separate page and chart titles | `vizb bar examples/csv/concurrency.csv -g load -p y -A x -n 'Q1 release' --title 'Framework throughput'` |
| Load on X, frameworks as series | `vizb bar examples/csv/concurrency.csv -g load -p x -A y` |

CI id: `00-concurrency-frameworks` on the **comparisons** dashboard (see `.github/workflows/comparisons-examples.yml`).

---

## Quick reference

| File | Rows | Mode | Typical chart |
|------|------|------|----------------|
| `sales.csv` | 10,000 | Auto-group or explicit `--group` | Bar (2D / 3D), line, scatter |
| `spiral-3d.csv` | 25,000 | Auto-value (xyz) | Scatter3D, line3D, bar3D |
| `noise-surface.csv` | 441 | Auto-value (xyz grid) | Bar3D surface |
| `noise-grid.csv` | 9,261 | Auto-value (xyz + metric) | Scatter3D + visualMap |
| `noise-grid-41.csv` | 68,921 | Auto-value (xyz + metric) | Scatter3D + visualMap (full grid) |
| `region-metrics.csv` | 8 | Solo `--select` mixed | Scatter mixed (region × metric) |
| `house-price-area2.csv` | 16,174 | Auto-value (xy) | Scatter2D + visualMap |
| `clusters.csv` | 60 | Auto-value (xy) | Scatter2D + visualMap, symbol size 10 |
| `concurrency.csv` | 3 | Group + `--col-axis` | Bar/line competitor compare |

**Auto-group** applies when the file has categorical columns and you did not pass `--group`. **Auto-value** applies when every column is numeric — vizb assigns `x`, `y`, `z` (and optional 4th metric) without flags.

## More detail

- **Official site:** [vizb.goptics.org](https://vizb.goptics.org) — install, docs, and interactive dashboards
- **Live dashboards** (topic-split; switch charts with `?id=<id>`; numbered prefix matches `?d=` index within each page):

| Topic | Dashboard | Source files |
|-------|-----------|--------------|
| Tabular data | [tabular-data](https://vizb.goptics.org/examples/live/tabular-data/) | sales, house-price, life-expectancy |
| Math & 3D | [math-and-3d](https://vizb.goptics.org/examples/live/math-and-3d/) | spiral-3d, noise-surface, noise-grid, noise-grid-41, clusters |
| Comparisons | [comparisons](https://vizb.goptics.org/examples/live/comparisons/) | concurrency |

| Chart | Source file | Live |
|-------|-------------|------|
| Sales auto-group | `sales.csv` | [Open](https://vizb.goptics.org/examples/live/tabular-data/?id=00-sales-auto-group) |
| Sales grouped | `sales.csv` | [Open](https://vizb.goptics.org/examples/live/tabular-data/?id=01-sales-grouped) |
| Sales by date | `sales.csv` | [Open](https://vizb.goptics.org/examples/live/tabular-data/?id=02-sales-by-date) |
| House price vs area | `house-price-area2.csv` | [Open](https://vizb.goptics.org/examples/live/tabular-data/?id=03-house-price-area2) |
| Life expectancy vs income | `life-expectancy-income.csv` | [Open](https://vizb.goptics.org/examples/live/tabular-data/?id=04-life-expectancy-income) |
| Spiral 3D | `spiral-3d.csv` | [Open](https://vizb.goptics.org/examples/live/math-and-3d/?id=00-spiral-3d) |
| Noise surface | `noise-surface.csv` | [Open](https://vizb.goptics.org/examples/live/math-and-3d/?id=01-noise-surface) |
| Noise grid (21³) | `noise-grid.csv` | [Open](https://vizb.goptics.org/examples/live/math-and-3d/?id=02-noise-grid) |
| Noise grid (41³) | `noise-grid-41.csv` | [Open](https://vizb.goptics.org/examples/live/math-and-3d/?id=03-noise-grid-41) |
| Clusters | `clusters.csv` | [Open](https://vizb.goptics.org/examples/live/math-and-3d/?id=04-clusters) |
| HTTP framework throughput | `concurrency.csv` | [Open](https://vizb.goptics.org/examples/live/comparisons/?id=00-concurrency-frameworks) |

- **Docs:** [Tabular data](https://vizb.goptics.org/examples/tabular-data/) · [Math & 3D](https://vizb.goptics.org/examples/math-and-3d/) · [Comparisons](https://vizb.goptics.org/examples/comparisons/) · [Group guide](../../docs/src/content/docs/guides/group.mdx) · [3D charts](../../docs/src/content/docs/charts/3d.mdx)
