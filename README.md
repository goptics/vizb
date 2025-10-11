# Vizb - Go Benchmark Visualization Tool

[![libs.tech recommends](https://libs.tech/project/1003638795/badge.svg)](https://libs.tech/project/1003638795/vizb)
[![Go Report Card](https://goreportcard.com/badge/github.com/goptics/vizb)](https://goreportcard.com/report/github.com/goptics/vizb)
[![CI](https://github.com/goptics/vizb/actions/workflows/ci.yml/badge.svg)](https://github.com/goptics/vizb/actions/workflows/ci.yml)
[![Codecov](https://codecov.io/gh/goptics/vizb/branch/main/graph/badge.svg)](https://codecov.io/gh/goptics/vizb)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for&logo=go)](https://golang.org/doc/devel/release.html)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=for)](LICENSE)

Vizb is a powerful CLI tool for visualizing Go benchmark results as interactive HTML charts with advance grouping. It automatically processes both raw benchmark output and JSON-formatted benchmark data, helping you compare the performance of different implementations across various workloads.

## Features

- **Interactive Charts**: Generate beautiful, interactive HTML charts from Go benchmark results
- **Dual Input Support**: Process both raw benchmark output and JSON-formatted benchmark data automatically
- **Multiple Metrics**: Compare execution time, memory usage, and allocation counts
- **Customizable Units**: Display metrics in your preferred units (ns/us/ms/s for time, b/B/KB/MB/GB for memory)
- **Allocation Units**: Customize allocation count representation (K/M/B/T)
- **Flexible Grouping**: Use custom patterns to extract grouping information from benchmark names with `--group-pattern`
- **Multiple Output Formats**: Generate HTML charts or JSON data with the `--format` flag
- **Responsive Design**: Charts work well on any device or screen size
- **Export Capability**: Save charts as PNG images directly from the browser
- **Simple CLI Interface**: Easy-to-use command line interface with helpful flags
- **Piped Input Support**: Process benchmark data directly from stdin

## Overview

https://github.com/user-attachments/assets/5dad22b0-d21f-434f-ad6e-57f4ebc74981

## Installation

```bash
go install github.com/goptics/vizb
```

## Usage

### Basic Usage

**Option 1: Using raw benchmark output**

1. Run your Go benchmarks and save the output:

   ```bash
   go test -bench . -run=^$ > bench.txt
   ```

2. Generate a chart from the benchmark results:

   ```bash
   vizb bench.txt -o output.html
   ```

**Option 2: Using JSON benchmark output**

1. Run your Go benchmarks with JSON output:

   ```bash
   go test -bench . -run=^$ -json > bench.json
   ```

2. Generate a chart from the JSON benchmark results:

   ```bash
   vizb bench.json -o output.html
   ```

**Option 3: Direct piping (recommended)**

Pipe benchmark results directly to vizb:

```bash
# Raw output
go test -bench . -run=^$ | vizb -o output.html

# JSON output (automatically detected and converted)
go test -bench . -run=^$ -json | vizb -o output.html
```

Open the generated HTML file in your browser to view the interactive charts.

**Note**: Vizb automatically detects the input format (raw or JSON) and processes it accordingly. JSON files are automatically converted to the required text format for parsing.

## How vizb groups your benchmark results

Vizb organizes your benchmark results into logical groups to create meaningful charts by organizing related benchmarks together. You can control how benchmarks are grouped using the `--group-pattern` flag.

### Group Pattern Configuration

Use the `--group-pattern` (or `-p`) flag to specify how vizb should extract grouping information from your benchmark names:

```bash
vizb --group-pattern "name/workload/subject" results.txt
```

The pattern defines the order and separators for extracting:

- **Name**: The benchmark family/category name
- **Workload**: The test condition or data size
- **Subject**: The specific operation being benchmarked (required)

![vizb chart example](./assets/vizb-char-overview.png)

### Pattern Syntax

**Components**: Use `name`, `workload`, `subject` (or shorthand `n`, `w`, `s`)

**Separators**: Use `/` (slash) or `_` (underscore) to define how parts are split.

**Square Brackets**: Use `[...]` for PascalCase splitting when benchmark names don't have separators
- `[s,w,n]` - Split PascalCase and assign consecutive words
- `[,w]` - Skip first word, assign second word to `w`
- `[,,s]` - Skip first 2 words, assign third word to `s`

**Required**: Every pattern must include `subject` (the operation being measured)

### Examples

| Pattern    | Benchmark Name                        | Name        | Workload    | Subject        | Description                            |
| ---------- | ------------------------------------- | ----------- | ----------- | -------------- | -------------------------------------- |
| `s`        | `BenchmarkStringConcat`               | _(empty)_   | _(empty)_   | `StringConcat` | Default: treats entire name as subject |
| `n/s`      | `BenchmarkStringOps/Concat`           | `StringOps` | _(empty)_   | `Concat`       | Name and subject with slash            |
| `n/w/s`    | `BenchmarkStringOps/LargeData/Concat` | `StringOps` | `LargeData` | `Concat`       | All three components                   |
| `s/w/n`    | `BenchmarkConcat/LargeData/StringOps` | `StringOps` | `LargeData` | `Concat`       | Custom order                           |
| `n_s_w`    | `BenchmarkStringOps_Concat_LargeData` | `StringOps` | `LargeData` | `Concat`       | Underscore separator                   |
| `/n/s`     | `BenchmarkIgnored/StringOps/Concat`   | `StringOps` | _(empty)_   | `Concat`       | Skip first part                        |
| `[s,w,n]`  | `SubjectWorkloadName`                 | `Name`      | `Workload`  | `Subject`      | PascalCase splitting with square brackets |
| `s/[w]/[n]`| `Concat/LargeData/StringOps`          | `String`    | `Large`     | `Concat`       | Mixed separators and PascalCase        |
| `[,,s]`    | `FirstSecondThirdFourth`              | _(empty)_   | _(empty)_   | `Third`        | Skip first 2 words, take 3rd          |

> [!Note]
> The `workload` dimension only appears in charts when there are multiple workloads to compare. If all benchmarks have the same workload (or no workload), charts will be simplified to show just subjects.

### Command Line Options

```bash
Usage:
  vizb [target] [flags]

Arguments:
  target               Path to benchmark file (raw or JSON) (optional if using piped input)

Flags:
  -h, --help                Help for vizb
  -n, --name string         Name of the chart (default "Benchmarks")
  -d, --description string  Description of the benchmark
  -o, --output string       Output HTML file name
  -f, --format string       Output format (html, json) (default "html")
  -p, --group-pattern string Pattern to extract grouping information from benchmark names (default "subject")
  -m, --mem-unit string     Memory unit available: b, B, KB, MB, GB (default "B")
  -t, --time-unit string    Time unit available: ns, us, ms, s (default "ns")
  -a, --alloc-unit string   Allocation unit available: K, M, B, T (default: as-is)
  -v, --version             Show version information
```

#### Custom chart name and description

```bash
vizb bench.txt -n "String Comparison Benchmarks" -d "Comparing different string manipulation algorithms"
```

#### Custom units for time and memory

```bash
vizb bench.txt -t ms -m MB
```

## Chart Types

Vizb generates three types of charts:

1. **Execution Time**: Shows the execution time per operation for each subject across different workloads
2. **Memory Usage**: Shows the memory usage per operation for each subject across different workloads
3. **Allocations**: Shows the number of memory allocations per operation for each subject

## Development

This project uses [Task](https://taskfile.dev/) for managing development workflows.

### Setup Development Environment

```bash
# Install Task runner
go install github.com/go-task/task/v3/cmd/task@latest

# Install required development tools
task setup
```

### Available Tasks

```bash
# List all available tasks
task

# Generate templates
task generate

# Build the binary (run from ./bin/vizb)
task build

# Run tests
task test
```

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

### Attribution

While not required by the MIT license, we kindly ask that you preserve the attribution footer in charts generated by Vizb when displaying them publicly.
