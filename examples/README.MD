# Vizb Examples

This directory contains benchmark output files used to generate the vizb examples dashboard.

## How it works

Pushing changes to `examples/bench/` triggers a CI workflow (`.github/workflows/deploy-examples.yml`) that:

1. Converts each `.txt` bench file to JSON with per-benchmark name, description, and group pattern
2. Merges all JSONs into a single HTML dashboard
3. Deploys to GitHub Pages on the orphan `examples` branch

## Live dashboard

**[goptics.github.io/vizb/examples](https://goptics.github.io/vizb/examples/)**

## Adding a new benchmark

1. Drop a `.txt` bench output file into `examples/bench/`
2. Add a matrix entry in `.github/workflows/deploy-examples.yml`:

```yaml
- output: my-benchmark
  bench-file: examples/bench/my-benchmark.txt
  name: "My Benchmark"
  description: "What this benchmark measures"
  group-pattern: "x/y"      # or group-regex for regex grouping
  scale: "linear"            # or "log"
```

3. Push — the CI rebuilds and deploys automatically.
