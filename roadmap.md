---
title: "Roadmap"
description: "Current capabilities and future plans for vizb."
---

## Current State

Vizb turns CSV/JSON tables and benchmark output into charts and stats without writing chart code: one dimension model (`n`/`x`/`y`/`z`), eight chart types (including 3D WebGL for bar/line/scatter), merge history, stats, CLI, GitHub Action, and REST.

Capability inventory lives on [Features](/features). Conceptual model: [Getting Started](/getting-started). Input adapters: [Supported inputs](/guides/parsers).

## Planned

Vizb currently focuses on linear, tabular data. The work below extends supported inputs and visual shapes.

### More Inputs

New benchmark adapters normalize another tool's output into the same table vizb already charts.

  ### Python

Integrate with `pytest-benchmark` output. Generate interactive visualizations for Python performance testing.

  ### C++

Parse output from Google Benchmark and similar C++ benchmarking frameworks. Bring multi-dimensional visualization to native performance testing.

  ### Dart

Support the Dart benchmark harness (`benchmark_harness` package). Visualize Dart VM performance data.

  ### Java

Parse output from JMH (Java Microbenchmark Harness) and similar Java benchmarking frameworks. Bring multi-dimensional visualization to JVM performance testing.

### Non-Linear Visualizations

We are exploring visualization types for data that is not well served by linear x/y/z charts:

- **Network / graph views** for relationship data.
- **Tree and hierarchy layouts** for nested or path-like data.
- **Geospatial maps** for location-based datasets.

These are research areas. They will land after the input-parser ecosystem is stable.

## Contributing

Interested in helping? Check out the [GitHub repository](https://github.com/goptics/vizb) for open issues and contribution guidelines.
