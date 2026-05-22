<div align="center">
  <img src="assests/logo.png" alt="vizb logo" width="120" height="auto" />
  <h1>Vizb</h1>

  <p>
    <a href="https://github.com/avelino/awesome-go?tab=readme-ov-file#benchmarks"><img src="https://awesome.re/mentioned-badge-flat.svg" alt="Mentioned in Awesome Go" /></a>
    <a href="https://vizb.goptics.org"><img src="https://img.shields.io/badge/Docs-00ADD8?style=for&logo=readthedocs" alt="Docs" /></a>
    <a href="https://vizb.goptics.org/examples/"><img src="https://img.shields.io/badge/Live-Examples-orange?style=for" alt="Examples" /></a>
    <a href="https://goreportcard.com/report/github.com/goptics/vizb"><img src="https://goreportcard.com/badge/github.com/goptics/vizb" alt="Go Report Card" /></a>
    <a href="https://github.com/goptics/vizb/actions/workflows/ci.yml"><img src="https://github.com/goptics/vizb/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
    <a href="https://codecov.io/gh/goptics/vizb"><img src="https://codecov.io/gh/goptics/vizb/branch/main/graph/badge.svg" alt="Codecov" /></a>
    <a href="https://golang.org/doc/devel/release.html"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for&logo=go" alt="Go Version" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=for" alt="License" /></a>
  </p>

  <p>
    A CLI tool that transforms Go benchmark raw output into interactive <strong>4D visualizations</strong>. Merge multiple benchmark runs, apply advanced grouping logic, and explore performance across four dimensions — all within a single deployable HTML file. Available <strong><a href="#github-action">GitHub Action</a></strong> for seamless CI pipeline integration.
  </p>
</div>

## Features

- **Modern Interactive UI**: Robust **Vue.js** application with a smooth and responsive experience.
- **Multi-Chart**: Supports multiple charts (`bar`, `line` and `pie`) in a single place.
- **Sorting**: Sort data (`asc`/`desc`) for comparison through UI settings or CLI flags.
- **Swap Axis**: Swap the `n`, `x` and `y` axes for diverse comparison through UI settings.
- **Logarithmic Scale**: Use `--scale log` for bar and line charts to better visualize benchmarks with high variance in values.
- **Multi-Dimensional Grouping**: Merge multiple benchmark data for deep comparative analysis.
- **Tag-Based Merging**: Tag benchmarks with commit hashes or version labels to compare performance across releases with automatic data merging.
- **Flexible Input**: Automatically processes raw `go test -bench` output and the standard JSON output of `go test -bench -json`.
- **Comprehensive Metrics**: Compare time, memory, and numbers with customizable units.
- **Smart Grouping**: Extract grouping logic from benchmark names using regex and group patterns.
- **Filtering**: Filter benchmarks to include only those matching a regex pattern.
- **Export Options**: Generate `single-file` HTML/JSON and options to save charts as `JPEG`.
- **GitHub Action**: First-class CI support — run benchmarks, tag releases, merge history, and deploy visualizations directly from your workflows with a single composite action.

## Installation

### Go Toolchain

```bash
go install github.com/goptics/vizb@latest
```

### Download Binary

Pre-built binaries for Linux, macOS, and Windows are available on the [releases page](https://github.com/goptics/vizb/releases).

## Quick Start

### Direct piping

```bash
go test -bench . | vizb -o output.html
```

### From a saved file

```bash
go test -bench . > bench.txt
vizb bench.txt -o output.html
```

### JSON output

```bash
go test -bench . -json | vizb -o output.html
```

### Merging multiple benchmarks

```bash
# Merge specific files
vizb merge output.json output2.json -o merged.json

# Merge all JSON files in a directory
vizb merge ./results/ -o all.json

# Generate HTML from merged JSON
vizb html merged.json -o report.html
```

> [!Note]
> The `merge` command requires JSON files as input, generated with `-o output.json`.

## Documentation

Full documentation is available at **[vizb.goptics.org](https://vizb.goptics.org/)**:

- [CLI Commands](https://vizb.goptics.org/commands/root/)
- [Grouping Guide](https://vizb.goptics.org/guides/grouping/)
- [Merging Guide](https://vizb.goptics.org/guides/merging/)
- [CI/CD Integration](https://vizb.goptics.org/ci-cd/github-action/)
- [UI Features](https://vizb.goptics.org/ui/)

## GitHub Action

Vizb provides a composite GitHub Action to run benchmarks and generate visualizations in CI.

### Run bench and generate HTML

```yaml
- uses: actions/setup-go@v6
  with:
    go-version-file: go.mod

- uses: goptics/vizb@v0
  with:
    bench-cmd: "go test -bench=."
    output-html: pages/index.html
```

> [!Note]
> For full input reference, CI tutorials (stateless & stateful), and deployment guides, see the [CI/CD documentation](https://vizb.goptics.org/ci-cd/github-action/).

## Development

This project uses [Task](https://taskfile.dev/) for managing development workflows.

### Setup

```bash
go install github.com/go-task/task/v3/cmd/task@latest
task init
```

### Available Tasks

```bash
task dev:ui      # Run the UI in development mode
task dev:docs    # Run the docs site in development mode
task build:ui    # Build the UI
task build:cli   # Build the binary (run from ./bin/vizb)
task build:docs  # Build the docs for production
task build       # Build everything
task test        # Run tests
```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
