# CI Group-Pattern Support

**Date:** 2026-05-12  
**Status:** Approved  
**Scope:** `vizb action` subcommand — CI mode benchmark grouping

## Problem

The `vizb action` subcommand hardcodes the benchmark name splitting (`splitBenchName` with `strings.SplitN(name, "/", 2)`) and always injects the git tag as `xAxis`. Users have no control over data orientation or grouping — the tag is always the x-axis dimension, and complex benchmark names cannot be mapped to different dimensions.

The root `vizb` command already has a flexible `--group-pattern` / `--group-regex` system that parses benchmark names into `name/xAxis/yAxis` dimensions. This system should be available in CI mode too.

## Design

### Core concept: tag as dedicated dimension

In CI mode, the tag/commit is always a known value from git context. It is never parsed from the benchmark name. A 2D group-pattern describes how the **benchmark name** is split — the **missing third dimension** always gets the tag.

**How it works:**
1. `parser.ParseBenchmarkData()` parses the bench file, applying the user's 2D pattern
2. The parser produces `BenchmarkData` entries with 2 of the 3 dimensions filled
3. `ci.InjectTag()` finds the dimension NOT in the pattern and fills it with the tag value
4. CI-specific logic (merge, prune, runtimes) runs on the completed data

### Valid patterns

Only **2D patterns** are valid in CI mode. 3D patterns (e.g., `n/y/x`) are rejected — the user's pattern describes the **bench name's** structure, and the tag always takes the third slot. The `ValidateGroupPattern` function checks for this.

All 6 permutations of 2 elements from `{n, x, y}`:

| Pattern | Bench: `Add/Queue` + tag `v1.0.0` | Tag fills |
|---------|-----------------------------------|-----------|
| `n/y`   | name=Add, yAxis=Queue             | **xAxis** |
| `n/x`   | name=Add, xAxis=Queue             | **yAxis** |
| `y/n`   | yAxis=Add, name=Queue             | **xAxis** |
| `x/n`   | xAxis=Add, name=Queue             | **yAxis** |
| `x/y`   | xAxis=Add, yAxis=Queue            | **name**  |
| `y/x`   | yAxis=Add, xAxis=Queue            | **name**  |

`--group-regex` overrides `--group-pattern` when set. The regex must define exactly 2 named capture groups from `{n, x, y, name, xAxis, yAxis}`. The missing dimension gets the tag.

**Default:** `n/y` — preserves current behavior (tag → xAxis, bench name → name + yAxis).

### Data depth handling

| Bench name depth | Pattern | Result |
|---|---|---|
| 1D (`Foo`) | `n/y` | name=Foo, yAxis="", **xAxis=tag** |
| 2D (`Add/Queue`) | `n/y` | name=Add, yAxis=Queue, **xAxis=tag** |
| 3D+ (`Add/Queue/Sub`) | `n/y` | name=Add, yAxis=Queue/Sub, **xAxis=tag** |

3D+ bench names work naturally: the parser's `SplitN(part, sep, 2)` logic lumps extra segments into the last pattern slot (same as local mode).

### Validation

CI patterns must satisfy:
1. Exactly 2 of `{n, x, y, name, xAxis, yAxis}` — no more, no less
2. At least one of xAxis or yAxis (inherited from `ValidateGroupPattern`)
3. No empty parts, no invalid identifiers

`ValidateGroupPattern` currently checks individual part validity but not part count. A new CI-specific function `ValidateCIPattern(pattern string) error` wraps `ValidateGroupPattern` and additionally checks that `len(parser.ParsePatternParts(pattern)) == 2`. `InjectTag` calls this before processing.

If `--group-regex` is set, it must define named capture groups for exactly 2 dimensions. Same rule: captured groups must be from `{name, xAxis, yAxis}`, the missing dimension gets the tag.

## Architecture

### Sharing parsing: `pkg/ci/` calls `pkg/parser/`

Currently `pkg/ci/action.go` duplicates benchfmt parsing (`parseBenchName`, `splitBenchName`, `resultToStats`, its own reader loop). We consolidate:

```
Old:  bench.txt → ci loop (benchfmt + splitBenchName + resultToStats)
New:  bench.txt → parser.ParseBenchmarkData() → ci.InjectTag() → ci.merge/prune
```

The root command's `ParseBenchmarkData` uses `shared.FlagState.GroupPattern`. For CI, we either:
- Temporarily set `shared.FlagState.GroupPattern` before calling `ParseBenchmarkData`
- Or pass the pattern through a new parameterized variant

The first option is simpler and avoids changing the parser's function signature.

### Files changed

| File | Change |
|---|---|
| `pkg/parser/parse_pattern.go` | Export `ParsePatternParts` (currently unexported) |
| `pkg/ci/action.go` | Delete `parseBenchName`, `splitBenchName`, `resultToStats`, benchfmt loop. Add `InjectTag()`. Call `parser.ParseBenchmarkData`. |
| `shared/action_state.go` | Add `GroupPattern string`, `GroupRegex string` |
| `cmd/action.go` | Add `--group-pattern` (default `n/y`) and `--group-regex` flags |
| `action.yml` | Add `group-pattern` and `group-regex` inputs |
| `cmd/action_test.go` | Update for new flags |
| `pkg/ci/action_test.go` | Update for `InjectTag`, pattern validation |

### New exporter function

```go
// pkg/parser/parse_pattern.go
func ParsePatternParts(pattern string) []string
```

Used by `ci.InjectTag()` to determine which dimension is missing from the user's pattern.

### InjectTag

```go
// pkg/ci/action.go
func InjectTag(data []shared.BenchmarkData, tag, pattern, regex string) ([]shared.BenchmarkData, error)
```

Logic:
1. Determine missing dimension:
   - If `regex != ""`: compile regex, extract named capture group names from `{name, xAxis, yAxis}`, missing = set difference
   - If `pattern != ""`: `parser.ParsePatternParts(pattern)` → missing = set difference from `{name, xAxis, yAxis}`
2. Validate: exactly one dimension must be missing
3. For each `BenchmarkData`, set `d.<MissingDim> = tag`
4. If no tag provided, skip injection (no-op)

### RunAction simplified

```go
func RunAction(opts ActionOpts) (*shared.Benchmark, error) {
    // Set parser flags temporarily
    prevPattern := shared.FlagState.GroupPattern
    prevRegex := shared.FlagState.GroupRegex
    shared.FlagState.GroupPattern = opts.GroupPattern
    shared.FlagState.GroupRegex = opts.GroupRegex
    defer func() {
        shared.FlagState.GroupPattern = prevPattern
        shared.FlagState.GroupRegex = prevRegex
    }()

    // Parse using shared parser
    data := parser.ParseBenchmarkData(opts.Input)
    data, err := InjectTag(data, opts.Tag, opts.GroupPattern)
    // If GroupRegex was used, InjectTag determines missing dimension the same way

    // Build Benchmark, merge, prune, track runtimes
    // ...
}
```

Note: `RunAction` no longer returns `*shared.Run` — it was unused.

#### Unit conversion notes

`ParseBenchmarkData` applies unit formatting via `shared.FlagState.TimeUnit` (default `"ns"`), `MemUnit` (default `"B"`), and `NumberUnit` (default `""`). These defaults produce the same output as the current CI `resultToStats` function. The CI action command does not expose `--time-unit`, `--mem-unit`, or `--number-unit` flags — users who need custom units should use the root `vizb` command.

### Deleted code

From `pkg/ci/action.go`:
- `parseBenchName(name benchfmt.Name) (string, string)` (~13 lines)
- `splitBenchName(fullName string) (string, string)` (~8 lines)
- `resultToStats(br shared.BenchmarkResult) []shared.Stat` (~16 lines)
- The inline benchfmt reading loop (~30 lines)

Total: ~67 lines removed, ~20 lines added (`InjectTag` + wiring).

### action.yml updates

```yaml
inputs:
  group-pattern:
    description: 'Pattern to parse benchmark names in CI mode (n/y, n/x, x/y, etc.). The missing dimension gets the tag.'
    required: false
    default: 'n/y'
  group-regex:
    description: 'Regex to parse benchmark names in CI mode. Overrides group-pattern.'
    required: false
    default: ''
```

Passed to CLI:
```bash
vizb action bench.txt \
  --tag ${{ github.ref_name }} \
  --group-pattern ${{ inputs.group-pattern }} \
  --group-regex ${{ inputs.group-regex }} \
  --merge benchmarks.json \
  --output benchmarks.json \
  $HTML_FLAG
```

## Testing

### New test cases needed

1. **`InjectTag` unit tests:**
   - Pattern `n/y` → name/yAxis from bench, tag → xAxis
   - Pattern `n/x` → name/xAxis from bench, tag → yAxis
   - Pattern `x/y` → xAxis/yAxis from bench, tag → name
   - 1D bench name with `n/y` → name only, tag → xAxis
   - 3D+ bench name with `n/y` → yAxis gets merged tail, tag → xAxis
   - Invalid 3D pattern → error
   - Invalid 1D pattern → error

2. **CI action tests (updated):**
   - Basic run with default `n/y` pattern
   - Run with custom pattern `x/y`
   - Merge with pattern
   - Prune with pattern
   - Pattern validation (3D pattern rejected)

3. **CLI tests (updated):**
   - Pass `--group-pattern` flag
   - Pass `--group-regex` flag
   - Default pattern produces xAxis=tag

## Edge cases and limitations

1. **3D+ bench names**: Extra segments merge into the last pattern slot (parser `SplitN` behavior). Example: `Add/Queue/Sub` with `n/y` → name=Add, yAxis=Queue/Sub. This matches current `splitBenchName` behavior.
2. **Empty yAxis**: For 1D bench names with `n/y`, yAxis stays empty. The tag fills xAxis. Charts render fine with empty yAxis (single series).
3. **benchfmt strips "Benchmark" prefix**: No manual stripping needed. Names from benchfmt are already clean.
4. **Pattern vs Regex precedence**: `--group-regex` overrides `--group-pattern` when set (same as root command).
5. **Concurrent safety**: `shared.FlagState.GroupPattern` is temporarily mutated. Tests must save/restore it.

## Migration

This is a refactor of unreleased CI functionality. No backward compatibility concerns. Existing benchmark JSON files (with hardcoded xAxis=tag) will be re-created by the action on next run.
