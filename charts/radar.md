---
title: "Radar Chart"
description: "Compare multi-axis profiles as spider/web polygons — one polygon per data point, one spoke per Y value."
---

The **radar chart** (also called a spider or web chart) draws filled polygons on a circular grid of spokes. Each spoke is a metric, and each polygon is a subject being compared across all those metrics at once. It's the chart to reach for when you want to see a multi-dimensional profile — not just "which is biggest" but "where does each variant win and where does it lose."

Use radar charts when your Y axis holds multiple meaningful metrics (throughput, latency, memory, etc.) and you want to compare several subjects across all of them simultaneously.

## How vizb builds it

Radar has no cartesian axes. Instead, vizb maps your dimensions to the radial layout:

| Dimension | Role |
|-----------|------|
| **YAxis** | Spoke indicators — each unique Y value becomes one spoke radiating from the center |
| **XAxis** | Data points — each unique X value becomes a polygon on the web |
| **ZAxis** | Series grouping — each Z value becomes a set of polygons; X values become multiple points within each Z polygon (3D only) |

Each polygon is filled with a semi-transparent area so overlapping shapes stay visible. Hover any polygon to see a tooltip listing every spoke name alongside its value.

## Dimensions

Long-form scores (subject × metric × value). Copy → `profiles.csv`:

```csv
subject,metric,score,tier
Pond,throughput,92,eager
Pond,latency,70,eager
Pond,memory,80,eager
Pond,stability,88,eager
Chi,throughput,85,eager
Chi,latency,78,eager
Chi,memory,74,eager
Chi,stability,82,eager
Gin,throughput,90,lazy
Gin,latency,65,lazy
Gin,memory,88,lazy
Gin,stability,79,lazy
Echo,throughput,88,lazy
Echo,latency,72,lazy
Echo,memory,81,lazy
Echo,stability,84,lazy
```

### 1D (X axis)

One polygon on the radar. The X values become the spokes, and the single data point is plotted as a stat-total shape. Best for a quick profile of one subject.

### CLI

```bash
vizb radar profiles.csv -g metric -o out.html
```

### HTTP

```json
{
  "input": "subject,metric,score,tier\nPond,throughput,92,eager\nPond,latency,70,eager\nPond,memory,80,eager\nPond,stability,88,eager\nChi,throughput,85,eager\nChi,latency,78,eager\nChi,memory,74,eager\nChi,stability,82,eager\nGin,throughput,90,lazy\nGin,latency,65,lazy\nGin,memory,88,lazy\nGin,stability,79,lazy\nEcho,throughput,88,lazy\nEcho,latency,72,lazy\nEcho,memory,81,lazy\nEcho,stability,84,lazy\n",
  "parser": "csv",
  "grouping": {
    "columns": [
      "metric"
    ]
  },
  "charts": {
    "types": [
      "radar"
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
    file: profiles.csv
    group: metric
    charts: radar
    output-html: out.html
```

### 2D (X + Y axes)

Multiple polygons, one per X value, all sharing the same set of spokes (Y values). The legend lists the X values so you can show or hide individual polygons. This is the core use case: compare how several subjects score across the same set of metrics.

### CLI

```bash
vizb radar profiles.csv -g subject,metric -p x,y -o out.html
```

### HTTP

```json
{
  "input": "subject,metric,score,tier\nPond,throughput,92,eager\nPond,latency,70,eager\nPond,memory,80,eager\nPond,stability,88,eager\nChi,throughput,85,eager\nChi,latency,78,eager\nChi,memory,74,eager\nChi,stability,82,eager\nGin,throughput,90,lazy\nGin,latency,65,lazy\nGin,memory,88,lazy\nGin,stability,79,lazy\nEcho,throughput,88,lazy\nEcho,latency,72,lazy\nEcho,memory,81,lazy\nEcho,stability,84,lazy\n",
  "parser": "csv",
  "grouping": {
    "columns": [
      "subject",
      "metric"
    ],
    "pattern": "x,y"
  },
  "charts": {
    "types": [
      "radar"
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
    file: profiles.csv
    group: subject,metric
    group-pattern: x,y
    charts: radar
    output-html: out.html
```

### 3D (X + Y + Z axes)

Multiple polygon sets, one set per Z value. Within each Z series, X values become multiple data points and Y values remain the spokes. The legend lists the Z values. To keep smaller polygons hoverable, vizb draws the **largest Z series first** so smaller polygons are rendered on top and their vertices stay reachable.

Spoke vertices are rendered as circle symbols, so you can hover individual points on each polygon to inspect a specific metric value.

### CLI

```bash
vizb radar profiles.csv -g subject,metric,tier -p x,y,z -o out.html
```

### HTTP

```json
{
  "input": "subject,metric,score,tier\nPond,throughput,92,eager\nPond,latency,70,eager\nPond,memory,80,eager\nPond,stability,88,eager\nChi,throughput,85,eager\nChi,latency,78,eager\nChi,memory,74,eager\nChi,stability,82,eager\nGin,throughput,90,lazy\nGin,latency,65,lazy\nGin,memory,88,lazy\nGin,stability,79,lazy\nEcho,throughput,88,lazy\nEcho,latency,72,lazy\nEcho,memory,81,lazy\nEcho,stability,84,lazy\n",
  "parser": "csv",
  "grouping": {
    "columns": [
      "subject",
      "metric",
      "tier"
    ],
    "pattern": "x,y,z"
  },
  "charts": {
    "types": [
      "radar"
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
    file: profiles.csv
    group: subject,metric,tier
    group-pattern: x,y,z
    charts: radar
    output-html: out.html
```

> **Scale (log) and auto-rotate are not available for radar charts.** Radar has no cartesian axis to scale logarithmically, and there is no 3D scene to rotate. These controls are hidden from the UI when only a radar chart is active.

## Settings

These settings apply to radar charts:

| Setting | CLI flag | UI toggle | Notes |
|---------|----------|-----------|-------|
| Sort | `--sort asc\|desc` | Sort control | Orders polygons / spokes |
| Labels | `--show-labels` | Show labels | Displays values at spoke vertices |
| Swap | `--swap` | Axis switcher | Swaps which column maps to spokes vs polygons |

Override settings for just this chart type without affecting others:

```bash
# Sort ascending and show labels
vizb radar profiles.csv -g subject,metric -p x,y --sort asc -l -o out.html
```

## Next Steps
