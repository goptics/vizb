<div align="center">
  <img src="assests/logo.png" alt="vizb logo" width="120" height="auto" />
  <h1>Vizb: Visualize Go Benchmarks in 4D</h1>

  <p>
    <a href="https://libs.tech/project/1003638795/vizb"><img src="https://libs.tech/project/1003638795/badge.svg" alt="libs.tech recommends" /></a>
    <a target="_blank" href="https://vizb-demo.netlify.app"><img src="https://img.shields.io/badge/Live-Demo-orange?style=for" alt="Live Demo" /></a>
    <a href="https://goreportcard.com/report/github.com/goptics/vizb"><img src="https://goreportcard.com/badge/github.com/goptics/vizb" alt="Go Report Card" /></a>
    <a href="https://github.com/goptics/vizb/actions/workflows/ci.yml"><img src="https://github.com/goptics/vizb/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
    <a href="https://codecov.io/gh/goptics/vizb"><img src="https://codecov.io/gh/goptics/vizb/branch/main/graph/badge.svg" alt="Codecov" /></a>
    <a href="https://golang.org/doc/devel/release.html"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for&logo=go" alt="Go Version" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=for" alt="License" /></a>
  </p>

  <p>
    Vizb is a CLI tool that transforms Go benchmark raw output into interactive <strong>4D visualizations</strong>. It allows you to <a href="#merging-multiple-benchmarks">merge multiple benchmark data</a>, apply <a href="#advance-usage">advanced grouping logic</a>, and explore performance across four dimensions: Source, Group, and two customizable axes (X and Y). All within a single, shareable HTML report.
  </p>
</div>

## Features

- **Modern Interactive UI**: Robust **Vue.js** application with a smooth and responsive experience.
- **Multi-Chart**: Supports multiple charts (`bar`, `line` and `pie`) in a single place.
- **Sorting**: Sort data (`asc`/`desc`) for comparison through UI settings or CLI flags.
- **Swap Axis**: Swap the `n`, `x` and `y` axes for diverse comparison through UI settings.
- **Multi-Dimensional Grouping**: Merge multiple benchmark data for deep comparative analysis.
- **Flexible Input**: Automatically processes raw `go test -bench` output and the standard JSON output of `go test -bench -json`.
- **Comprehensive Metrics**: Compare time, memory, and numbers with customizable units.
- **Smart Grouping**: Extract grouping logic from benchmark names using regex and group patterns.
- **Export Options**: Generate `single-file` HTML/JSON and options to save charts as `JPEG`.

## Installation

```bash
go install github.com/goptics/vizb
```

## Basic Usage

### Using raw benchmark output

Run your Go benchmarks and save the output:

```bash
go test -bench . > bench.txt
```

Generate charts from the benchmark:

```bash
vizb bench.txt -o output.html
```

### Direct piping

Pipe benchmark results directly to vizb:

```bash
# Raw output
go test -bench . | vizb -o output.html

# JSON output (automatically detected and converted)
go test -bench . -json | vizb -o output.html
```

### Using vizb standard JSON benchmark output

```bash
vizb bench.txt -f json -o output.json
```

Generate charts from the standard JSON benchmark data:

```bash
vizb output.json -o output.html
```

### Merging multiple benchmarks

You can combine multiple benchmark JSON files into a single html file using the `merge` command. This is useful for aggregating benchmark data from different runs, machines, or environments.

```bash
# Merge specific files
vizb merge output.json output2.json -o merged_report.html

# Merge all JSON files in a directory
vizb merge ./results/ -o all_results.html

# Mix and match files and directories
vizb merge ./old_results/ output.json -o comparison.html
```

Open the generated HTML file in your browser to view the interactive charts.

> [!Note]
> The `merge` command requires JSON files as input, which must be generated using `vizb bench.txt -f json`.

## Advance Usage

### How vizb groups your benchmark data

Vizb creates charts that make sense by putting your benchmark data into logical groupings and axes. It sees the data as `1D` (xAxis) by default, but if you have to deal with `2D` or `3D` data, you can use the `--group-pattern` and `--group-regex` flags to group your data.

### Understanding Group Patterns

A group pattern tells vizb how to dissect your benchmark names into three key components:

1.  **Name (n)**: The family or group the benchmark belongs to. Benchmarks with the same `Name` will be grouped together in the same chart. (optional)
2.  **XAxis (x)**: The category that goes on the X-axis (e.g., input size, concurrency level).
3.  **YAxis (y)**: The specific test case or variation (e.g., algorithm name, sub-test).

### Visualizing the Extraction

Imagine you have a benchmark named `BenchmarkSort/100/Ints`, which has `3D` data.

If you use the pattern `name/xAxis/yAxis` (or `n/x/y`), vizb splits the name wherever it finds a `/`:

```text
Benchmark Name:  BenchmarkSort  /  100  /  Ints
                     │              │        │
Pattern:           [Name]        [XAxis]   [YAxis]
                     │              │        │
Result:            "Sort"         "100"    "Ints"
```

### Group Pattern Syntax (`--group-pattern`)

- **Components**: Use `name`, `xAxis`, `yAxis` (or shorthands `n`, `x`, `y`).
- **Separators**: Use `/` (slash) or `_` (underscore) to match the separators in your benchmark names.
- **Skipping parts**: You can leave parts empty in the pattern to ignore sections of the benchmark name.

#### Standard Go Benchmarks (Slash Separated)

Format: `Benchmark<Group>/<InputSize>/<Variant>`

**Pattern:** `n/x/y`

| Benchmark Name                 | Extracted Data                                            |
| :----------------------------- | :-------------------------------------------------------- |
| `BenchmarkSort/1024/QuickSort` | **Name:** `Sort` **XAxis:** `1024` **YAxis:** `QuickSort` |
| `BenchmarkSort/1024/MergeSort` | **Name:** `Sort` **XAxis:** `1024` **YAxis:** `MergeSort` |

#### Underscore Separated

Format: `Benchmark<Group>_<Variant>_<InputSize>`

**Pattern:** `n_y_x`

| Benchmark Name             | Extracted Data                                        |
| :------------------------- | :---------------------------------------------------- |
| `BenchmarkHash_SHA256_1KB` | **Name:** `Hash` **YAxis:** `SHA256` **XAxis:** `1KB` |
| `BenchmarkHash_MD5_1KB`    | **Name:** `Hash` **YAxis:** `MD5` **XAxis:** `1KB`    |

#### Simple Grouping (No X-Axis)

Format: `Benchmark<Group>/<Variant>`

**Pattern:** `n/y`

| Benchmark Name            | Extracted Data                                               |
| :------------------------ | :----------------------------------------------------------- |
| `BenchmarkJSON/Marshal`   | **Name:** `JSON` **XAxis:** _(empty)_ **YAxis:** `Marshal`   |
| `BenchmarkJSON/Unmarshal` | **Name:** `JSON` **XAxis:** _(empty)_ **YAxis:** `Unmarshal` |

#### Ignoring Prefixes

Sometimes you might want to ignore a common prefix or a specific part of the name.

**Pattern:** `/n/y` (Starts with a separator to skip the first part)

| Benchmark Name               | Extracted Data                                                         |
| :--------------------------- | :--------------------------------------------------------------------- |
| `BenchmarkTest/JSON/Marshal` | **Name:** `JSON` **YAxis:** `Marshal` _(First part "Test" is ignored)_ |

### Group Regex Syntax (`--group-regex`)

For more complex benchmark names where simple patterns aren't enough, you can use Regular Expressions with named groups.

- **Named Groups**: Use `(?<name>...)`, `(?<xAxis>...)`, `(?<yAxis>...)` (or shorthands `(?<n>...)`, `(?<x>...)`, `(?<y>...)`) to capture parts of the benchmark name.
- **Flexibility**: Regex allows you to match specific characters, ignore parts, and handle irregular formats.

#### Examples

| Benchmark Name                            | Regex                                   | Extracted Data                                            | Dimensions |
| :---------------------------------------- | :-------------------------------------- | :-------------------------------------------------------- | :--------- |
| `BenchmarkHashing64MD5`                   | `Hashing64(?<x>.*)`                     | **XAxis:** `MD5`                                          | 1D         |
| `BenchmarkJSONByMarshal`                  | `(?<x>.*)By(?<y>.*)`                    | **XAxis:** `JSON` **YAxis:** `Marshal`                    | 2D         |
| `BenchmarkDecode/text=digits/level=speed` | `(?<n>.*)/text=(?<x>.*)/level=(?<y>.*)` | **Name:** `Decode` **XAxis:** `digits` **YAxis:** `speed` | 3D         |

> [!Note]
> You must specify at least one of the `x` and `y` axes when you use the `--group-[pattern|regex]` command. the `n` is optional.

## Development

This project uses [Task](https://taskfile.dev/) for managing development workflows.

### Setup Development Environment

```bash
# Install Task runner
go install github.com/go-task/task/v3/cmd/task@latest
```

### Available Tasks

```bash
# List all available tasks
task

# Run the UI in development mode
task dev:ui

# Build The UI
task build:ui

# Build the binary (run from ./bin/vizb)
task build:cli

# Build everything
task build

# Run tests
task test
```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
