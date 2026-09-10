---
title: "Pie Chart"
description: "Break data into proportional slices — vizb renders one pie per dimension, side-by-side."
---

The **pie chart** shows how a whole breaks into parts. Unlike bar and line charts, pie has no cartesian axes — vizb maps your dimensions to multiple pie charts rendered side-by-side, one per axis. Each slice carries its name and percentage, so the composition of your dataset is immediately readable. The default is a **filled pie**; `--donut` cuts a hole.

Use pie charts when you care about proportion and share rather than raw magnitude: which dataset name dominates, which configuration contributes the most, or how Z variants divide the total.

## How vizb builds it

Pie ignores the X/Y/Z grid model. Instead, each axis becomes its own standalone pie:

| Dimension | Role |
|-----------|------|
| **XAxis** | Each unique X value becomes a slice in the "By X-Axis" pie |
| **YAxis** | Each unique Y value becomes a slice in the "By Y-Axis" pie |
| **ZAxis** | Each unique Z value becomes a slice in the "By Z-Axis" pie |

Slice sizes are proportional to the sum of values for that axis entry. Labels show the slice name and its percentage of the total.

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

One pie chart: each unique X value is a slice. The slice size is proportional to its total value. This is the clearest way to see which X category dominates.

### Agent

```text
/vizb sales.csv as pie by region
```

### CLI

```bash
vizb pie sales.csv -g region -o out.html
```

### HTTP

```json
{
  "input": "order_date,region,category,product,quantity,amount\n2024-01-01,Central,Hardware,Connector,16,2488.24\n2024-01-02,East,Tools,Gear,42,207.08\n2024-01-03,North,Mechanical,Sensor,20,3465.22\n2024-01-04,West,Electronics,Valve,24,7633.44\n2024-03-12,South,Industrial,Widget,31,5088.10\n2024-06-18,East,Electronics,Relay,12,2214.50\n2024-09-05,North,Hardware,Bolt,28,1724.27\n2024-12-20,West,Tools,Gadget,19,2938.82\n2025-02-08,Central,Mechanical,Valve,27,7350.26\n2025-04-14,South,Electronics,Widget,41,3278.77\n2025-07-22,East,Industrial,Connector,15,5093.42\n2025-10-03,North,Tools,Gear,33,4102.15\n2025-11-19,West,Hardware,Sensor,22,2890.40\n2025-12-28,South,Mechanical,Gadget,26,611.32\n",
  "parser": "csv",
  "grouping": {
    "columns": [
      "region"
    ]
  },
  "charts": {
    "types": [
      "pie"
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
    group: region
    charts: pie
    output-html: out.html
```

### 2D (X + Y axes)

Two pies side-by-side: **By X-Axis** and **By Y-Axis**. Each pie breaks the total differently — the first by X category, the second by Y column. Comparing the two lets you see which axis has more evenly-distributed weight.

### Agent

```text
/vizb sales.csv as pie by region & category
```

### CLI

```bash
vizb pie sales.csv -g region,category -p x,y -o out.html
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
      "pie"
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
    charts: pie
    output-html: out.html
```

### 3D (X + Y + Z axes)

Three pies in a row: **By X-Axis**, **By Y-Axis**, and **By Z-Axis**. Each pie slices the total along a different axis, giving you a three-way breakdown of the same dataset in a single view.

This example splits `order_date` into month and day slots (same grouping as the [Sales by Date](https://vizb.goptics.org/examples/live/tabular-data/?id=02-sales-by-date) tabular-data dashboard) and maps `category` to the Z axis.

### Agent

```text
/vizb sales.csv as pie, split order_date into month & date (skip the year), category as depth
```

### CLI

```bash
vizb pie sales.csv -g order_date,category -p "[-y{Month}-x{Date}],z{Category}" -o out.html
```

### HTTP

```json
{
  "input": "order_date,region,category,product,quantity,amount\n2024-01-01,Central,Hardware,Connector,16,2488.24\n2024-01-02,East,Tools,Gear,42,207.08\n2024-01-03,North,Mechanical,Sensor,20,3465.22\n2024-01-04,West,Electronics,Valve,24,7633.44\n2024-03-12,South,Industrial,Widget,31,5088.10\n2024-06-18,East,Electronics,Relay,12,2214.50\n2024-09-05,North,Hardware,Bolt,28,1724.27\n2024-12-20,West,Tools,Gadget,19,2938.82\n2025-02-08,Central,Mechanical,Valve,27,7350.26\n2025-04-14,South,Electronics,Widget,41,3278.77\n2025-07-22,East,Industrial,Connector,15,5093.42\n2025-10-03,North,Tools,Gear,33,4102.15\n2025-11-19,West,Hardware,Sensor,22,2890.40\n2025-12-28,South,Mechanical,Gadget,26,611.32\n",
  "parser": "csv",
  "grouping": {
    "columns": [
      "order_date",
      "category"
    ],
    "pattern": "[-y{Month}-x{Date}],z{Category}"
  },
  "charts": {
    "types": [
      "pie"
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
    group: order_date,category
    group-pattern: "[-y{Month}-x{Date}],z{Category}"
    charts: pie
    output-html: out.html
```

## Settings

Pie charts expose four runtime settings — sort, labels, donut, and axis swap.

| Setting | CLI flag | UI toggle | Notes |
|---------|----------|-----------|-------|
| Sort | `--sort asc\|desc` | Sort control | Orders slices by value |
| Labels | `--show-labels` | Show labels | Shows name + percentage on each slice |
| Donut | `--donut` | Donut | Cut a hole in the pie. Off (default) is a filled pie. |
| Swap | `--swap` | Axis switcher | Reorders which axis maps to which pie |

Override settings for just this chart type without affecting others:

```bash
# Sort slices descending and show labels
vizb pie sales.csv -g region,category -p x,y --sort desc -l -o out.html

# Filled pie is the default; --donut cuts a hole
vizb pie sales.csv -g region --donut -o donut.html
```

## Next Steps
