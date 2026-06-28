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

These variants are also built in CI as `00-sales-auto-group`, `01-sales-grouped`, and `02-sales-by-date` (see `.github/workflows/deploy-examples-csv.yml`).

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

## `house-price-area2.csv`

**Shape:** `area`, `price` — 16,174 rows ([ECharts house-price scatter](https://echarts.apache.org/examples/en/editor.html?c=scatter-large)). Auto-value xy; add `--visualmap` for price gradient. CI: `07-house-price-area2`.

---

## Quick reference

| File | Rows | Mode | Typical chart |
|------|------|------|----------------|
| `sales.csv` | 10,000 | Auto-group or explicit `--group` | Bar (2D / 3D), line, scatter |
| `spiral-3d.csv` | 25,000 | Auto-value (xyz) | Scatter3D, line3D, bar3D |
| `noise-surface.csv` | 441 | Auto-value (xyz grid) | Bar3D surface |
| `noise-grid.csv` | 9,261 | Auto-value (xyz + metric) | Scatter3D + visualMap |
| `noise-grid-41.csv` | 68,921 | Auto-value (xyz + metric) | Scatter3D + visualMap (full grid) |
| `house-price-area2.csv` | 16,174 | Auto-value (xy) | Scatter2D + visualMap |

**Auto-group** applies when the file has categorical columns and you did not pass `--group`. **Auto-value** applies when every column is numeric — vizb assigns `x`, `y`, `z` (and optional 4th metric) without flags.

## More detail

- **Official site:** [vizb.goptics.org](https://vizb.goptics.org) — install, docs, and interactive dashboards
- **Live CSV dashboards:** [vizb.goptics.org/examples/csv/](https://vizb.goptics.org/examples/csv/) — CI builds each recipe below from these files (switch charts with `?d=` or `?id=` when `--id` is set at build time)

| Dashboard | Source file | Live |
|-----------|-------------|------|
| Sales auto-group | `sales.csv` | [Open](https://vizb.goptics.org/examples/csv/) |
| Sales grouped 3D | `sales.csv` | [Open](https://vizb.goptics.org/examples/csv/?d=1) |
| Sales by date | `sales.csv` | [Open](https://vizb.goptics.org/examples/csv/?d=2) |
| Spiral 3D | `spiral-3d.csv` | [Open](https://vizb.goptics.org/examples/csv/?d=3) |
| Noise surface | `noise-surface.csv` | [Open](https://vizb.goptics.org/examples/csv/?d=4) |
| Noise grid (21³) | `noise-grid.csv` | [Open](https://vizb.goptics.org/examples/csv/?d=5) |
| Noise grid (41³) | `noise-grid-41.csv` | [Open](https://vizb.goptics.org/examples/csv/?d=6) |
| House price vs area | `house-price-area2.csv` | [Open](https://vizb.goptics.org/examples/csv/?d=7) |

- **Docs in repo:** [Grouping guide](../../docs/src/content/docs/guides/grouping.mdx) · [3D charts](../../docs/src/content/docs/charts/3d.mdx) · [All examples](https://vizb.goptics.org/examples/)
