---
title: "Dimensions"
description: "Vizb maps every row onto Name, X, Y, and Z (n/x/y/z). Learn what each dimension does and how patterns assign them."
---

Every Vizb chart is driven by the same four dimensions. In group mode, you assign them with `--group-pattern` / `-p` (or `--group-regex` / `-r`). For CSV and JSON you first pick source columns with `--group` / `-g`. For benchmarks, the benchmark name is already the label and `-p` splits it. In select mode, `--select` maps columns onto the same dimensions as coordinates.

| Dimension | Shorthand | Role |
|-----------|-----------|------|
| **Name** | `n` | Splits data into separate chart panels (one panel per unique name) |
| **XAxis** | `x` | Categories or values on the horizontal axis |
| **YAxis** | `y` | Series / variants inside a chart (or second coordinate in select mode) |
| **ZAxis** | `z` | Depth layer for 3D (or a third coordinate). Requires both `x` and `y` |

## Tiny example

Suppose this CSV (copy, save as `sales.csv`):

```csv
region,product,sales
Asia,Widget,100
Asia,Gadget,80
EU,Widget,60
```

Map `region` to Y and `product` to X:

### Agent

```text
/vizb sales.csv as bar by region & product
```

### CLI

```bash
vizb bar sales.csv -g region,product -p y,x -o out.html
```

### HTTP

```json
{
  "input": "region,product,sales\nAsia,Widget,100\nAsia,Gadget,80\nEU,Widget,60\n",
  "parser": "csv",
  "grouping": {
    "columns": [
      "region",
      "product"
    ],
    "pattern": "y,x"
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
    group: region,product
    group-pattern: y,x
    charts: bar
    output-html: out.html
```

| Row | Name | X | Y | sales (metric) |
|-----|------|---|---|----------------|
| 1 | (none) | Widget | Asia | 100 |
| 2 | (none) | Gadget | Asia | 80 |
| 3 | (none) | Widget | EU | 60 |

You get one bar chart for `sales`: X = product, series = region. Add `-p n,x,y` with a third grouping column if you want separate panels per name.

## Rules that always apply

- At least one of **`x` or `y`** is required for grouping to take effect.
- **`z` requires both `x` and `y`.**
- **`n` is optional.** Without it, all points share one unnamed chart panel.
- Pattern separators on CSV/JSON must **match** how you listed columns in `-g` (commas with comma lists, spaces with quoted space lists, and so on). See [Matching separators](/guides/group#matching-separators).

## Dimensionality

| Pattern (CSV style) | What you get |
|---------------------|--------------|
| `-p x` | 1D — single axis |
| `-p x,y` | 2D — series or groups under each X |
| `-p x,y,z` | 3D — WebGL bar/line/scatter when those charts are selected |
| `-p n,x,y` | Multiple chart panels, each 2D |

Benchmark names usually use slashes: `-p n/x/y` for names like `BenchmarkSort/1024/QuickSort`.

How each chart type interprets these axes is in the [Charts overview](/charts#dimensions-1d-2d-and-3d).

## Two ways to fill dimensions

Dimensions are shared. How you fill them depends on mode:

| Mode | Fills dimensions from | Metrics |
|------|----------------------|---------|
| **Group** | Category columns (or bench name segments) | Each remaining numeric column → its own chart |
| **Select** (solo) | Columns listed as `x,y[,z]` coordinates | Values sit on the axes; no per-column charts |

When you are unsure which mode you need, start with [Group vs Select](/guides/group-vs-select) in Guides.

## Auto paths (no `-g` / `-p` needed)

| File shape | What Vizb does |
|------------|----------------|
| Categorical + numeric columns | **Auto-group**: highest-cardinality non-numeric column becomes X |
| All-numeric columns | **Auto-value**: first 2-3 columns become continuous X/Y[/Z] |

The CLI logs the inference (`🧠 Auto-grouped…` / `🧠 Auto-valued…`). Pass explicit `-g` / `-p` or `--select` whenever the guess is wrong.

> Auto-group only sets `-p x`. Grouped 3D needs both `x` and `y` in the pattern, so supply them explicitly for 3D grouped charts.

## Next
