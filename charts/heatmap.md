---
title: "Heatmap Chart"
description: "Compare values across two dimensions in a colored grid — optionally fold a z-axis into each cell."
---

The **heatmap** packs X and Y into a flat grid of colored cells. Each cell shows a value and a color — spot high/low combinations at a glance across the full matrix.

Use heatmaps when you have two categorical dimensions. A z-axis is optional: without z you get a single-gradient matrix; with z each cell shows the sum of its z-series values on the same gradient.

## How vizb builds it

| Dimension | Role |
|-----------|------|
| **XAxis** | Columns of the grid |
| **YAxis** | Rows of the grid |
| **ZAxis** | Folded into each cell: cell **value** = `Σ z`, colored on a shared gradient |

A cell with multiple z-series shows their summed value. Hover to see the per-z breakdown in the tooltip.

## Dimensions

Sample orders below (copy → `sales.csv`). Full file: [`examples/csv/sales.csv`](https://github.com/goptics/vizb/blob/main/examples/csv/sales.csv) (10,000 rows, 2024–2025).

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

### 2D (X + Y axes)

Classic heatmap: one value per X × Y cell, colored on a continuous gradient with a `visualMap` scale. No z-series — magnitude as color.

### CLI

```bash
vizb heatmap sales.csv -g region,category -p x,y -o out.html
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
      "heatmap"
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
    charts: heatmap
    output-html: out.html
```

### 3D (X + Y + Z axes)

Z values are folded into each cell rather than rendered as depth. Each cell displays `Σ z` and is colored on the same continuous gradient as the 2D heatmap. Hover a cell to see the individual z-series breakdown with colored dots. Toggle z-series in the legend — cells recompute live.

### CLI

```bash
vizb heatmap sales.csv -g region,category,product -p x,y,z -o out.html
```

### HTTP

```json
{
  "input": "order_date,region,category,product,quantity,amount\n2024-01-01,Central,Hardware,Connector,16,2488.24\n2024-01-02,East,Tools,Gear,42,207.08\n2024-01-03,North,Mechanical,Sensor,20,3465.22\n2024-01-04,West,Electronics,Valve,24,7633.44\n2024-03-12,South,Industrial,Widget,31,5088.10\n2024-06-18,East,Electronics,Relay,12,2214.50\n2024-09-05,North,Hardware,Bolt,28,1724.27\n2024-12-20,West,Tools,Gadget,19,2938.82\n2025-02-08,Central,Mechanical,Valve,27,7350.26\n2025-04-14,South,Electronics,Widget,41,3278.77\n2025-07-22,East,Industrial,Connector,15,5093.42\n2025-10-03,North,Tools,Gear,33,4102.15\n2025-11-19,West,Hardware,Sensor,22,2890.40\n2025-12-28,South,Mechanical,Gadget,26,611.32\n",
  "parser": "csv",
  "grouping": {
    "columns": [
      "region",
      "category",
      "product"
    ],
    "pattern": "x,y,z"
  },
  "charts": {
    "types": [
      "heatmap"
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
    group: region,category,product
    group-pattern: x,y,z
    charts: heatmap
    output-html: out.html
```

> The y-axis stays flat intentionally. Splitting each row by z made the grid tall and sparse — so z is folded into the cell instead.

## Settings

These settings apply to heatmap charts:

| Setting | CLI flag | UI toggle | Notes |
|---------|----------|-----------|-------|
| Sort | `--sort asc\|desc` | Sort control | Orders the matrix axes |
| Labels | `--show-labels` | Show labels | Shows the summed value inside each cell |
| Swap | `--swap yx` | Axis switcher | Swaps X and Y axes |

Override settings for just the heatmap without affecting other charts:

```bash
# Show per-cell labels
vizb heatmap sales.csv -g region,category,product -p x,y,z -l -o out.html

# Sort the axes descending
vizb heatmap sales.csv -g region,category,product -p x,y,z --sort desc -o out.html
```

> `scale` (log) and `3d-rotate` are not available for heatmap — those options are bar and line only.

## Next Steps
