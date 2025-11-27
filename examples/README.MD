# Vizb Examples

This guide demonstrates how to process benchmark results and generate a unified dashboard using `vizb`.

## Prerequisites

Ensure you have `vizb` installed and available in your PATH.

## Step 1: Process Benchmark JSON Files

Use the `vizb` command to parse your benchmark output files (`.txt`) and generate JSON reports. You can define how benchmarks are grouped using patterns (`-p`) or regular expressions (`-r`).

### Benchmark Patterns

| Benchmark File | Pattern / Regex | Flag |
| :--- | :--- | :--- |
| `bench/http-routing.txt` | `x_n` | `-p` |
| `bench/worker-pools.txt` | `n/y/x` | `-p` |
| `bench/varmq.txt` | `x/y` | `-p` |
| `bench/hash.txt` | `'Hashing64(?<x>.*)'` | `-r` |
| `bench/serialization.txt` | `'(?<x>.*)By(?<y>.*)'` | `-r` |
| `bench/validator.txt` | `'(?<y>.+?)(?:(?<x>Success\|Failure)(?<n>.*)\|$)'` | `-r` |
| `bench/decode-encode.txt` | `'(?<n>[^/]+)/text=(?<x>[^/]+)/level=(?<y>.*)'` | `-r` |

### Template Command

Use the following template to process any of the benchmarks listed above:

```bash
vizb <bench-file> <flag> <pattern> -o <output-json> -f json
```

For example:
```bash
vizb bench/http-routing.txt -p x_n -o jsons/http-routing -f json
```

## Step 2: Generate the Dashboard (`index.html`)

After generating the JSON reports, use the `merge` command to combine them into a single, interactive HTML dashboard.

```bash
# Merge all JSON files from the 'jsons/' directory
vizb merge jsons/ -o index.html
```

This command will create an `index.html` file. Open it in your web browser to view and analyze your benchmark results.
