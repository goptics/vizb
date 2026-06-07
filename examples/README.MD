# Vizb Examples

Benchmark output files used to generate the vizb examples dashboard, organized by language.

## How it works

Push changes to `examples/{go,javascript,rust}/` to trigger a per-language CI pipeline that:

1. Converts each `.txt` bench file to JSON with per-benchmark name, description, and group pattern
2. Merges all JSONs into a single HTML dashboard for that language
3. Deploys to GitHub Pages on the `website` branch

## Live dashboards

| Language | Workflow | Dashboard |
|----------|----------|-----------|
| Go       | `.github/workflows/deploy-examples-go.yml`       | [vizb.goptics.org/examples/go/][go-dash]       |
| JavaScript | `.github/workflows/deploy-examples-javascript.yml` | [vizb.goptics.org/examples/javascript/][js-dash] |
| Rust     | `.github/workflows/deploy-examples-rust.yml`     | [vizb.goptics.org/examples/rust/][rust-dash]     |

[go-dash]: https://vizb.goptics.org/examples/go/
[js-dash]: https://vizb.goptics.org/examples/javascript/
[rust-dash]: https://vizb.goptics.org/examples/rust/

## Adding a new benchmark

1. Drop a `.txt` bench output file into `examples/{language}/`
2. Add a matrix entry in the corresponding workflow (e.g. `.github/workflows/deploy-examples-go.yml`):

```yaml
- output: my-benchmark
  file: my-benchmark.txt
  name: "My Benchmark"
  description: "What this benchmark measures"
  group-pattern: "x/y"      # or group-regex for regex grouping
  scale: "linear"            # or "log"
```

3. Push — the CI rebuilds and deploys automatically.
