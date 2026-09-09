---
title: "Examples"
description: "Topic-based sample datasets and live dashboards for tabular data, math 3D, comparisons, GitHub Legends, and benchmarks."
---

Sample inputs live under [`examples/`](https://github.com/goptics/vizb/tree/main/examples). CI converts each recipe to JSON, merges them into one HTML page **per topic**, and deploys to [vizb.goptics.org/examples/live/](https://vizb.goptics.org/examples/live/). Open a specific dataset with `?id=<id>` — each matrix entry sets a numbered id (`00-`, `01-`, …) via the action `id` input; the prefix stays in sync with `?d=0`, `?d=1`, … in the same matrix order.

## Browse by topic

  
  
  
  
  
  

## Live dashboards

| Topic | Workflow | Status | Dashboard |
|-------|----------|--------|-----------|
| Tabular data | [tabular-data-examples.yml](https://github.com/goptics/vizb/blob/main/.github/workflows/tabular-data-examples.yml) | [![Tabular Data examples workflow status](https://github.com/goptics/vizb/actions/workflows/tabular-data-examples.yml/badge.svg)](https://github.com/goptics/vizb/actions/workflows/tabular-data-examples.yml) | [Open](https://vizb.goptics.org/examples/live/tabular-data/) |
| Math & 3D | [math-and-3d-examples.yml](https://github.com/goptics/vizb/blob/main/.github/workflows/math-and-3d-examples.yml) | [![Math and 3D examples workflow status](https://github.com/goptics/vizb/actions/workflows/math-and-3d-examples.yml/badge.svg)](https://github.com/goptics/vizb/actions/workflows/math-and-3d-examples.yml) | [Open](https://vizb.goptics.org/examples/live/math-and-3d/) |
| Comparisons | [comparisons-examples.yml](https://github.com/goptics/vizb/blob/main/.github/workflows/comparisons-examples.yml) | [![Comparisons examples](https://github.com/goptics/vizb/actions/workflows/comparisons-examples.yml/badge.svg)](https://github.com/goptics/vizb/actions/workflows/comparisons-examples.yml) | [Open](https://vizb.goptics.org/examples/live/comparisons/) |
| GitHub Legends | [github-legends.yml](https://github.com/goptics/vizb/blob/main/.github/workflows/github-legends.yml) | [![GitHub Legends examples workflow status](https://github.com/goptics/vizb/actions/workflows/github-legends.yml/badge.svg)](https://github.com/goptics/vizb/actions/workflows/github-legends.yml) | [Open](https://vizb.goptics.org/examples/live/github-legends/) |
| Go | [go-examples.yml](https://github.com/goptics/vizb/blob/main/.github/workflows/go-examples.yml) | [![Go examples workflow status](https://github.com/goptics/vizb/actions/workflows/go-examples.yml/badge.svg)](https://github.com/goptics/vizb/actions/workflows/go-examples.yml) | [Open](https://vizb.goptics.org/examples/live/go/) |
| JavaScript | [javascript-examples.yml](https://github.com/goptics/vizb/blob/main/.github/workflows/javascript-examples.yml) | [![JavaScript examples workflow status](https://github.com/goptics/vizb/actions/workflows/javascript-examples.yml/badge.svg)](https://github.com/goptics/vizb/actions/workflows/javascript-examples.yml) | [Open](https://vizb.goptics.org/examples/live/javascript/) |
| Rust | [rust-examples.yml](https://github.com/goptics/vizb/blob/main/.github/workflows/rust-examples.yml) | [![Rust examples workflow status](https://github.com/goptics/vizb/actions/workflows/rust-examples.yml/badge.svg)](https://github.com/goptics/vizb/actions/workflows/rust-examples.yml) | [Open](https://vizb.goptics.org/examples/live/rust/) |

## Adding an example

Same steps for every topic — see [examples/README.MD](https://github.com/goptics/vizb/blob/main/examples/README.MD):

1. Add the file under `examples/{category}/` (CSV recipes stay under `examples/csv/`)
2. Add a matrix entry in the matching `.github/workflows/*-examples.yml` (or `github-legends.yml`)
3. Push to `main` — CI rebuilds that dashboard

For contribution skylines, only add a user row to the GitHub Legends matrix — no repo file required.

> Clone the repo and run any example locally:
> 
>   ```bash
  vizb bar examples/csv/sales.csv -o sales.html
  vizb -P go examples/go/hash.txt -o hash.html
  ```
