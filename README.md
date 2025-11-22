# Vizb - Go Benchmark Visualization Tool

[![libs.tech recommends](https://libs.tech/project/1003638795/badge.svg)](https://libs.tech/project/1003638795/vizb)
[![Go Report Card](https://goreportcard.com/badge/github.com/goptics/vizb)](https://goreportcard.com/report/github.com/goptics/vizb)
[![CI](https://github.com/goptics/vizb/actions/workflows/ci.yml/badge.svg)](https://github.com/goptics/vizb/actions/workflows/ci.yml)
[![Codecov](https://codecov.io/gh/goptics/vizb/branch/main/graph/badge.svg)](https://codecov.io/gh/goptics/vizb)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for&logo=go)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=for)](LICENSE)

Vizb is a powerful CLI tool for visualizing Go benchmark results as interactive HTML charts with advance grouping. It automatically processes both raw benchmark output and JSON-formatted benchmark data, helping you compare the performance of different implementations across various workloads.

## Features

- **Modern Interactive UI**: Robust **Vue.js** application with a smooth, responsive experience.
- **Dark & Light Mode**: Built-in support for both themes.
- **Sorting**: Sort data (asc/desc) via UI or CLI flags.
- **Multi-layer Grouping**: Merge multiple benchmark results for deep comparative analysis.
- **Flexible Input**: Automatically processes raw `go test` output or JSON.
- **Comprehensive Metrics**: Compare time, memory, and number of allocations with customizable units.
- **Smart Grouping**: Extract grouping logic from benchmark names using custom patterns.
- **Export Options**: Generate HTML/JSON reports or save charts as PNG.
- **Developer Friendly**: Simple CLI with piped input support (`| vizb`).

## Overview

https://github.com/user-attachments/assets/a6dbe3b9-f5aa-4643-8e19-d91e710c78fb

## Installation

```bash
go install github.com/goptics/vizb
```

## Usage

### Basic Usage

#### Option 1: Using raw benchmark output

Run your Go benchmarks and save the output:

```bash
go test -bench . > bench.txt
```

Generate a chart from the benchmark results:

```bash
vizb bench.txt -o output.html
```

#### Option 2: Using JSON benchmark output

Run your Go benchmarks with JSON output:

```bash
go test -bench . -json > bench.json
```

Generate a chart from the JSON benchmark results:

```bash
vizb bench.json -o output.html
```

#### Option 3: Direct piping (recommended)

Pipe benchmark results directly to vizb:

```bash
# Raw output
go test -bench . | vizb -o output.html

# JSON output (automatically detected and converted)
go test -bench . -json | vizb -o output.html
```

#### Option 4: Merging multiple benchmarks

You can combine multiple benchmark JSON files into a single report using the `merge` command. This is useful for aggregating results from different runs, machines, or environments.

```bash
# Merge specific files
vizb merge bench1.json bench2.json -o merged_report.html

# Merge all JSON files in a directory
vizb merge ./results/ -o all_results.html

# Mix and match files and directories
vizb merge ./old_results/ new_run.json -o comparison.html
```

Open the generated HTML file in your browser to view the interactive charts.

> [!Note]
> The `merge` command requires JSON files as input, which must be generated using `vizb bench.txt -f json`.

## How vizb groups your benchmark results

Vizb organizes your benchmark results into logical groups to create meaningful charts. By default, it tries to be smart, but you can control exactly how benchmarks are grouped using the `--group-pattern` flag.

### Understanding Group Patterns

A group pattern tells vizb how to dissect your benchmark names into three key components:

1.  **Name (n)**: The family or group the benchmark belongs to. Benchmarks with the same `Name` will be grouped together in the same chart.
2.  **XAxis (x)**: The category that goes on the X-axis (e.g., input size, concurrency level).
3.  **YAxis (y)**: The specific test case or variation (e.g., algorithm name, sub-test).

### Visualizing the Extraction

Imagine you have a benchmark named `BenchmarkMatrix/1024x1024/Parallel`.

If you use the pattern `name/xAxis/yAxis` (or `n/x/y`), vizb splits the name wherever it finds a `/`:

```text
Benchmark Name:  BenchmarkMatrix  /  1024x1024  /  Parallel
                     │                 │             │
Pattern:           [Name]           [XAxis]        [YAxis]
                     │                 │             │
Result:           "Matrix"        "1024x1024"     "Parallel"
```

### Pattern Syntax

- **Components**: Use `name`, `xAxis`, `yAxis` (or shorthands `n`, `x`, `y`).
- **Separators**: Use `/` (slash) or `_` (underscore) to match the separators in your benchmark names.
- **Skipping parts**: You can leave parts empty in the pattern to ignore sections of the benchmark name.

### Common Scenarios & Examples

Here are some common benchmark naming conventions and the patterns to use:

#### Scenario 1: Standard Go Benchmarks (Slash Separated)

Format: `Benchmark<Group>/<InputSize>/<Variant>`

**Pattern:** `n/x/y`

| Benchmark Name                 | Extracted Data                                            |
| :----------------------------- | :-------------------------------------------------------- |
| `BenchmarkSort/1024/QuickSort` | **Name:** `Sort` **XAxis:** `1024` **YAxis:** `QuickSort` |
| `BenchmarkSort/1024/MergeSort` | **Name:** `Sort` **XAxis:** `1024` **YAxis:** `MergeSort` |

#### Scenario 2: Underscore Separated

Format: `Benchmark<Group>_<Variant>_<InputSize>`

**Pattern:** `n_y_x`

| Benchmark Name             | Extracted Data                                        |
| :------------------------- | :---------------------------------------------------- |
| `BenchmarkHash_SHA256_1KB` | **Name:** `Hash` **YAxis:** `SHA256` **XAxis:** `1KB` |
| `BenchmarkHash_MD5_1KB`    | **Name:** `Hash` **YAxis:** `MD5` **XAxis:** `1KB`    |

#### Scenario 3: Simple Grouping (No X-Axis)

Format: `Benchmark<Group>/<Variant>`

**Pattern:** `n/y`

| Benchmark Name            | Extracted Data                                               |
| :------------------------ | :----------------------------------------------------------- |
| `BenchmarkJSON/Marshal`   | **Name:** `JSON` **XAxis:** _(empty)_ **YAxis:** `Marshal`   |
| `BenchmarkJSON/Unmarshal` | **Name:** `JSON` **XAxis:** _(empty)_ **YAxis:** `Unmarshal` |

#### Scenario 4: Ignoring Prefixes

Sometimes you might want to ignore a common prefix or a specific part of the name.

**Pattern:** `/n/y` (Starts with a separator to skip the first part)

| Benchmark Name           | Extracted Data                                                              |
| :----------------------- | :-------------------------------------------------------------------------- |
| `BenchmarkTest/JSON/Marshal` | **Name:** `JSON` **YAxis:** `Marshal` _(First part "BenchmarkTest" is ignored)_ |

#### Custom chart name and description

```bash
vizb bench.txt -n "String Comparison Benchmarks" -d "Comparing different string manipulation algorithms"
```

#### Custom units for time and memory

```bash
vizb bench.txt -t ms -m MB
```

#### Sorting and Chart Selection

```bash
vizb bench.txt --sort asc --charts bar,pie --show-labels
```

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
