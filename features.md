---
title: "Features"
description: "What Vizb can do — inputs, dimensions, charts, UI, merge, and CI — without writing chart code."
---

Vizb turns CSV, JSON, and benchmark output into interactive charts and stats **without writing chart or frontend code**. Tables and benchmarks share one dimension model; you run the same pipeline from the CLI, [GitHub Action](/ci-cd/github-action), or [REST API](/commands/serve).

New to the product? Start with [Getting Started](/getting-started).

## Inputs

  ### CSV and JSON tables

Chart numeric columns. Promote other columns to Name / X / Y / Z with `--group`. Nested JSON arrays via `--json-path`. See [Tabular data](/guides/data).

  ### Benchmark parsers

Go (`go test -bench`), Rust (Criterion, Divan), JavaScript (Vitest, Tinybench). Auto-detect from content, override with `-P`. See [Supported inputs](/guides/parsers).

  ### Auto detection

`--parser` defaults to `auto`. Vizb prints the chosen parser. File path or stdin pipe both work.

## Dimension model

  ### Name / X / Y / Z

Up to four dimensions on every chart. Patterns and regex split labels. See [Dimensions](/getting-started/dimensions).

  ### Group and select

Group: one chart per metric, categories on axes, sum duplicates. Select: columns as coordinates or metric picks. See [Group vs Select](/guides/group-vs-select).

  ### Auto-group and auto-value

No flags: categorical files auto-group; all-numeric files auto-value continuous axes. CLI logs the inference so you can override it.

  ### Wide competitor columns

`--col-axis` / `-A` places numeric column *names* on a free dimension so frameworks or versions share one chart.

## Charts and UI

  ### Eight chart types

Bar, line, scatter, pie, heatmap, radar, Sankey, Chord. Choose with `--charts` or a chart subcommand. See [Charts](/charts).

  ### 3D (WebGL)

Z-axis data (or auto-value with 3+ numeric columns) renders interactive bar3D / line3D / scatter3D. See [3D charts](/charts/3d).

  ### Self-contained HTML

Sort, zoom, swap axes, fullscreen, themes, JPEG export. One file, offline. See [UI](/ui).

  ### Statistics panel

Opt-in with `--stat`: 33 descriptive metrics per series plus Pearson, Spearman, Kendall, and distance correlation. See [Statistics](/ui/stats).

  ### Log scale and large data

Log scale on cartesian charts. DataZoom for wide X axes; aggregated grouped CSV/JSON stays responsive.

## Merge, CI, and API

  ### Tag-based merge

Tag datasets, merge JSON across runs, inject tags onto n/x/y/z. See [Merging](/guides/merging).

  ### GitHub Action

Composite action for cmd or file input, merge, HTML/JSON outputs. See [GitHub Action](/ci-cd/github-action).

  ### REST API

`vizb serve` exposes convert, merge, and UI generation. See [serve](/commands/serve) and [API](/api/).

  ### Units and export

Time, memory, and number units. HTML or JSON output. Chart JPEG from the UI.

## Learn more

> Prefer the learning path: [Install](/getting-started/install) → [Introduction](/getting-started) → [Dimensions](/getting-started/dimensions) → [Charts](/charts) → [Group vs Select](/guides/group-vs-select) → [Commands](/commands/root).
