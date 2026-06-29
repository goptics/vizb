# select-axis-mode Pattern Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Collapse the 9 over-engineering/spaghetti findings from the PR review into 3 design patterns (Mode enum/State, RowReader Template Method, ChartBuilder Strategy + polymorphism) — net deletion of ~600 lines of duplication/branching while preserving all current behavior.

**Architecture:** Three layered tasks executed on the current `feat/select-axis-mode` branch. Task 1 refactors the Go side (Mode enum, RowReader interface, FlagBag dedup, dual-inference collapse, dead-code removal). Task 2 refactors the TS side (ChartBuilder strategies, chart-shape polymorphism, worker mode registry). Task 3 does cross-layer cleanup (useDataPoint simplification) and final verification. No behavior changes — pure structural refactor. Existing tests are the safety net (`go test ./...` and `vitest run` must stay green between every task).

**Tech Stack:** Go 1.24 (parser/cmd packages), Vue 3 + TypeScript + Vitest (UI), ECharts (chart rendering), pflag (CLI flags), cobra (command framework).

**Branch:** `feat/select-axis-mode` (refactor commits land on top of the current PR work).

---

## File Structure

### Task 1 — Go-side refactors

| File | Action | Responsibility |
|------|--------|----------------|
| `pkg/parser/registry.go` | Modify | Add `Mode` type + `Config.Mode` field |
| `pkg/parser/mode.go` | Create | `ResolveMode(cfg) Mode` + mode predicates (`IsGrouped`, `IsValue`, etc.) |
| `pkg/parser/tabular.go` | Create | `RowReader` interface + shared `parseMixedMode`/`parseSelectStatMode`/`parseValueMode`/`dispatchSelectMode` |
| `pkg/parser/csv/csv.go` | Modify | `csvRowReader` adapter; thin `ParseCSV` delegating to tabular dispatch |
| `pkg/parser/json/json.go` | Modify | `jsonRowReader` adapter; thin `ParseJSON` delegating to tabular dispatch |
| `pkg/parser/axes_spec.go` | Modify | Delete `inferSelectViewMixed` + `parseAxisFloat`; `DatasetAxesForSelectView` uses carried `AxisType` |
| `pkg/parser/select_view_spec.go` | Modify | Delete `IsSelectAxisMode`/`IsMultiSelectStatMode`/`IsMixedMode` (replaced by `cfg.Mode`); keep `HasSelect`/`IsExplicitGrouping`/parsing helpers |
| `pkg/parser/group_spec.go` | Modify | `NoExplicitGrouping` uses `HasSelect` (already done — verify only) |
| `cmd/cli/flagbag.go` | Modify | Unify `slices`+`arrays` into one `stringSlices map[string]*[]string`; branch only at `Bind` |
| `cmd/cli/pipeline.go` | Modify | `assembleDataset`/`buildDataset` switch on `cfg.Mode` instead of predicate ladder |

### Task 2 — TS-side refactors

| File | Action | Responsibility |
|------|--------|----------------|
| `ui/src/lib/builders/types.ts` | Create | `ChartBuilder` interface + `BuildContext` type |
| `ui/src/lib/builders/grouped.ts` | Create | `GroupedBuilder` (non-preserveRows path) |
| `ui/src/lib/builders/preserveRows.ts` | Create | `PreserveRowsBuilder` (preserveRows path, category-scatter sub-path) |
| `ui/src/lib/builders/value.ts` | Create | `ValueBuilder` (existing `buildValueModeChart` extracted) |
| `ui/src/lib/builders/mixed.ts` | Create | `MixedBuilder` (existing `buildMixedModeChart` extracted) |
| `ui/src/lib/builders/index.ts` | Create | `pickBuilder(chart)` registry + `builderForChart(chart)` accessor |
| `ui/src/lib/transform.ts` | Modify | `buildChartForSignature` delegates to builder; delete inlined branches |
| `ui/src/lib/utils.ts` | Modify | `chartHasPlottableData`/`chartAxisBadgeCount`/`computeChartGrandTotal`/`is3D`/`canOfferValue3D` delegate to builder |
| `ui/src/workers/transform.worker.ts` | Modify | Replace `valueMode`/`mixedMode`/normal triplication with `modes` registry |

### Task 3 — Cross-layer cleanup

| File | Action | Responsibility |
|------|--------|----------------|
| `ui/src/composables/useDataPoint.ts` | Modify | Replace 4 mirrored computeds with one `chartMode` computed |
| `ui/src/components/SettingsPanel.vue` | Modify | Consume `chartMode` instead of 3 separate flags |
| `.superpowers/sdd/refactor-report.md` | Create | SDD execution report |

---

## Task 1: Go-side refactors (Mode enum + RowReader + FlagBag + dead code)

**Files:**
- Create: `pkg/parser/mode.go`
- Create: `pkg/parser/tabular.go`
- Modify: `pkg/parser/registry.go:14-33` (add `Mode` type + field)
- Modify: `pkg/parser/axes_spec.go:122-162` (delete dual-inference helpers)
- Modify: `pkg/parser/select_view_spec.go:152-173` (delete mode predicates)
- Modify: `pkg/parser/csv/csv.go` (thin adapter)
- Modify: `pkg/parser/json/json.go` (thin adapter)
- Modify: `cmd/cli/flagbag.go:21-55,185-261` (unify slice stores)
- Modify: `cmd/cli/pipeline.go:444-492` (switch on `cfg.Mode`)
- Modify: `cmd/cli/flagbag.go:265-321` (`ParseConfig` resolves `cfg.Mode`)
- Delete: dead code `readCSVTable` (csv.go:445) and `readJSONArray` (json.go:478)

### Step 1.1: Delete dead code `readCSVTable` and `readJSONArray`

- [ ] Delete `readCSVTable` function from `pkg/parser/csv/csv.go` (lines ~445-470). It is never called — `ParseCSV` does its own inline `os.Open`/`csv.NewReader`.

- [ ] Delete `readJSONArray` function from `pkg/parser/json/json.go` (lines ~478-530). It is never called — `ParseJSON` does its own inline `json.NewDecoder`.

- [ ] Run `go build ./...` to confirm no compile errors from the deletions.

- [ ] Run `go test ./pkg/parser/...` — all existing tests must stay green.

- [ ] Commit: `refactor(parser): delete unused readCSVTable/readJSONArray helpers`

### Step 1.2: Add `Mode` type and `ResolveMode`

- [ ] Add to `pkg/parser/registry.go` after the `Config` struct (after line 33):

```go
// Mode is the resolved parse mode for a Config. Set once in ParseConfig so
// every downstream call site switches on cfg.Mode instead of re-deriving it
// from overlapping predicates.
type Mode int

const (
    ModeAuto Mode = iota // no --select and no explicit grouping → auto-group/auto-value
    ModeGrouped          // explicit grouping (-g/-r/-p) + --select numeric stat columns
    ModeValue            // solo --select, all numeric columns → value axes x,y[,z]
    ModeMixed            // solo --select, one categorical x + numeric y[,z]
    ModeMultiStat        // repeatable solo --select (dim,metric) pairs merged into stats
)
```

Add `Mode Mode` field to the `Config` struct (after `ChartTypes`).

- [ ] Create `pkg/parser/mode.go`:

```go
package parser

// ResolveMode determines the parse Mode from the resolved Config. Call once
// after Select/SelectViews/Axes are populated (in ParseConfig) so downstream
// code switches on cfg.Mode instead of re-deriving it from predicates.
//
// Resolution order:
//   1. Explicit grouping + Select → ModeGrouped
//   2. Solo SelectViews (no explicit grouping):
//      a. len > 1 → ModeMultiStat (validated to 2-col dim,metric)
//      b. len == 1 → ModeValue or ModeMixed (caller resolves after type inference)
//   3. Otherwise → ModeAuto
//
// Mixed vs value for a solo single view is not known until ResolveAxesTypes
// runs (it needs the data). ResolveMode sets ModeValue for a single solo view;
// the parser sets ModeMixed on its local cfg copy after type inference. The
// dataset builder treats both identically for axes derivation.
func ResolveMode(cfg Config) Mode {
    if IsExplicitGrouping(cfg) && len(cfg.Select) > 0 {
        return ModeGrouped
    }
    if len(cfg.SelectViews) > 0 && !IsExplicitGrouping(cfg) {
        if len(cfg.SelectViews) > 1 {
            return ModeMultiStat
        }
        return ModeValue
    }
    return ModeAuto
}

// IsGrouped reports whether cfg is in grouped stat-column mode.
func (m Mode) IsGrouped() bool { return m == ModeGrouped }

// IsSelectAxis reports whether cfg is solo --select axis mode (value, mixed, or multi-stat).
func (m Mode) IsSelectAxis() bool { return m == ModeValue || m == ModeMixed || m == ModeMultiStat }

// IsMultiStat reports multi-stat solo --select mode.
func (m Mode) IsMultiStat() bool { return m == ModeMultiStat }
```

- [ ] In `pkg/parser/select_view_spec.go`, delete `IsSelectAxisMode` (lines 164-167) and `IsMultiSelectStatMode` (lines 169-173). Keep `HasSelect` and `IsExplicitGrouping` — they remain the inputs to `ResolveMode`. Callers will switch to `cfg.Mode.Is...()` methods.

- [ ] Run `go build ./...` — expect compile errors at the old call sites (csv.go, json.go, pipeline.go). Do not fix yet; those are fixed in steps 1.3-1.5.

### Step 1.3: Wire `ResolveMode` into `ParseConfig`

- [ ] In `cmd/cli/flagbag.go` `ParseConfig()`, after the `--select` routing block (after line 318), add:

```go
cfg.Mode = parser.ResolveMode(cfg)
```

- [ ] Run `go test ./cmd/cli/...` — `flagbag_test.go` should pass (it tests routing into `Select`/`SelectViews`, not mode predicates).

- [ ] Commit: `refactor(parser): add Mode enum resolved once in ParseConfig`

### Step 1.4: Introduce `RowReader` and collapse CSV/JSON parser duplication

- [ ] Create `pkg/parser/tabular.go`:

```go
package parser

import (
    "fmt"
    "strconv"

    "github.com/goptics/vizb/shared"
    "github.com/goptics/vizb/shared/utils"
)

// RowReader abstracts one input row for the shared tabular parse functions.
// CSV and JSON each provide an adapter; the shared parseMixedMode/parseValueMode/
// parseSelectStatMode implementations read cells through this interface so the
// ~250 lines of CSV↔JSON duplication collapse into one implementation each.
type RowReader interface {
    // Cell returns the raw string value of the named column on the current row.
    // present is false when the column is absent or the row is too short.
    Cell(source string) (raw string, present bool)
    // Numeric returns the finite float64 value of the named column, or false
    // when absent/empty/non-finite.
    Numeric(source string) (float64, bool)
    // AvailableColumns returns the column/field names in file order (for errors).
    AvailableColumns() []string
    // FlagLabel is "--select" or "--axes" — the flag name shown in error messages.
    FlagLabel() string
}

// ParseMixedMode maps one categorical column to x and numeric columns to y[,z];
// each row becomes a point with empty stats (no aggregation). cfg.Axes must have
// AxisType set (category/value) from ResolveAxesTypes.
func ParseMixedMode(rows []RowReader, cfg Config) []shared.DataPoint {
    type slot struct{ kind string }
    slots := make(map[string]slot, len(cfg.Axes))
    for _, spec := range cfg.Axes {
        slots[spec.AxisKey] = slot{kind: spec.AxisType}
    }

    var results []shared.DataPoint
    for _, row := range rows {
        var dp shared.DataPoint
        complete := true
        for key, sl := range slots {
            if sl.kind == "category" {
                cell, ok := row.Cell(specSourceForKey(cfg, key))
                if !ok || cell == "" {
                    complete = false
                    break
                }
                assignAxis(&dp, key, cell)
                continue
            }
            v, ok := row.Numeric(specSourceForKey(cfg, key))
            if !ok {
                complete = false
                break
            }
            formatted := strconv.FormatFloat(utils.FormatNumber(v, cfg.NumberUnit), 'g', -1, 64)
            assignAxis(&dp, key, formatted)
        }
        if !complete {
            continue
        }
        results = append(results, dp)
    }
    return results
}

// ParseValueMode implements value mode: each named numeric column becomes a
// coordinate on x, y[, z]; each row becomes a raw point with no stat series.
func ParseValueMode(rows []RowReader, cfg Config) []shared.DataPoint {
    var results []shared.DataPoint
    for _, row := range rows {
        var dp shared.DataPoint
        dst := []*string{&dp.XAxis, &dp.YAxis, &dp.ZAxis}
        complete := true
        for i, spec := range cfg.Axes {
            if i >= len(dst) {
                break
            }
            v, ok := row.Numeric(spec.Source)
            if !ok {
                complete = false
                break
            }
            *dst[i] = strconv.FormatFloat(utils.FormatNumber(v, cfg.NumberUnit), 'g', -1, 64)
        }
        if !complete {
            continue
        }
        if cfg.MetricColumn != "" {
            mv, ok := row.Numeric(cfg.MetricColumn)
            if ok {
                dp.Metric = strconv.FormatFloat(utils.FormatNumber(mv, cfg.NumberUnit), 'g', -1, 64)
            }
        }
        results = append(results, dp)
    }
    return results
}

// ParseSelectStatMode parses repeatable solo --select into one dataset. When
// every flag shares the same dimension column, each input row becomes one point
// with multiple stats; otherwise each (row × view) stays a separate point.
func ParseSelectStatMode(rows []RowReader, cfg Config) []shared.DataPoint {
    merge := MultiSelectSharedDim(cfg.SelectViews)
    var results []shared.DataPoint
    for _, row := range rows {
        AppendMultiSelectStatPoint(&results, cfg.SelectViews, cfg.NumberUnit, merge, func(view SelectView) (MultiSelectRowStat, bool) {
            if len(view.Columns) < 2 {
                return MultiSelectRowStat{}, false
            }
            dim, metric := view.Columns[0], view.Columns[1]
            dimVal, ok := row.Cell(dim.Source)
            if !ok || dimVal == "" {
                return MultiSelectRowStat{}, false
            }
            v, ok := row.Numeric(metric.Source)
            if !ok {
                return MultiSelectRowStat{}, false
            }
            return MultiSelectRowStat{DimVal: dimVal, Value: v}, true
        })
    }
    if len(results) == 0 {
        shared.ExitWithError("No dataSet data found", nil)
    }
    return results
}

// DispatchSelectMode routes a solo --select Config to the right parse function
// after running ResolveAxesTypes. Returns the parsed DataPoints. Called by the
// CSV/JSON entry points — they pass their RowReader slice and kindFn.
func DispatchSelectMode(rows []RowReader, cfg *Config, kindFn AxisColumnKind) []shared.DataPoint {
    if cfg.Mode.IsMultiStat() {
        return ParseSelectStatMode(rows, *cfg)
    }
    axesCfg := SelectViewAxesCfg(*cfg)
    flag := AxisColumnLabel(true)
    if err := ResolveAxesTypes(&axesCfg, kindFn); err != nil {
        shared.ExitWithError(err.Error(), nil)
    }
    if isMixedAxes(axesCfg) {
        return ParseMixedMode(rows, axesCfg)
    }
    return ParseValueMode(rows, axesCfg)
}

// isMixedAxes is the post-ResolveAxesTypes mixed check (replaces IsMixedMode
// predicate on the caller's cfg — operated on the local axesCfg copy).
func isMixedAxes(cfg Config) bool {
    hasCat, hasVal := false, false
    for _, s := range cfg.Axes {
        switch s.AxisType {
        case "category":
            hasCat = true
        case "value":
            hasVal = true
        }
    }
    return hasCat && hasVal
}

func specSourceForKey(cfg Config, key string) string {
    for _, s := range cfg.Axes {
        if s.AxisKey == key {
            return s.Source
        }
    }
    return ""
}

func assignAxis(dp *shared.DataPoint, key, value string) {
    switch key {
    case "x":
        dp.XAxis = value
    case "y":
        dp.YAxis = value
    case "z":
        dp.ZAxis = value
    }
}
```

- [ ] In `pkg/parser/csv/csv.go`, replace `csvAxisColumnKind`, `parseCSVMixedMode`, `parseCSVValueMode`, `parseCSVSelectStatMode` with a single `csvRowReader` adapter and a thin dispatch in `ParseCSV`. The adapter:

```go
// csvRowReader adapts one CSV row to the parser.RowReader interface.
type csvRowReader struct {
    row     []string
    colIdx  map[string]int
    flag    string
    headers []string
}

func (r csvRowReader) Cell(source string) (string, bool) {
    col, ok := r.colIdx[source]
    if !ok || col >= len(r.row) {
        return "", false
    }
    return strings.TrimSpace(r.row[col]), true
}

func (r csvRowReader) Numeric(source string) (float64, bool) {
    s, ok := r.Cell(source)
    if !ok || s == "" {
        return 0, false
    }
    return parseFinite(s)
}

func (r csvRowReader) AvailableColumns() []string { return nonEmpty(r.headers) }
func (r csvRowReader) FlagLabel() string          { return r.flag }

func csvKindFn(headers []string, dataRows [][]string, flag string) parser.AxisColumnKind {
    colIdx := map[string]int{}
    for i, h := range headers {
        if h != "" {
            colIdx[h] = i
        }
    }
    return func(source, axisKey string) (string, error) {
        col, ok := colIdx[source]
        if !ok {
            return "", fmt.Errorf("%s column %q not found; available: %v", flag, source, nonEmpty(headers))
        }
        anyNumeric, allNumeric, sawCell := false, true, false
        for _, row := range dataRows {
            if col >= len(row) {
                continue
            }
            cell := strings.TrimSpace(row[col])
            if cell == "" {
                continue
            }
            sawCell = true
            if _, ok := parseFinite(cell); ok {
                anyNumeric = true
            } else {
                allNumeric = false
            }
        }
        if !sawCell {
            return "", fmt.Errorf("%s column %q has no data", flag, source)
        }
        if axisKey == "x" {
            if allNumeric {
                return "value", nil
            }
            return "category", nil
        }
        if !anyNumeric {
            return "", fmt.Errorf("%s column %q is not numeric", flag, source)
        }
        return "value", nil
    }
}
```

Then `ParseCSV`'s select-routing block becomes:

```go
    if parser.HasSelect(cfg) {
        cfg.Mode = parser.ResolveMode(cfg)
        readers := make([]parser.RowReader, len(dataRows))
        colIdx := map[string]int{}
        for i, h := range headers {
            if h != "" {
                colIdx[h] = i
            }
        }
        flag := parser.AxisColumnLabel(true)
        for i, row := range dataRows {
            readers[i] = csvRowReader{row: row, colIdx: colIdx, flag: flag, headers: headers}
        }
        return parser.DispatchSelectMode(readers, &cfg, csvKindFn(headers, dataRows, flag))
    }
```

Delete the old `parseCSVMixedMode`/`parseCSVValueMode`/`parseCSVSelectStatMode`/`csvAxisColumnKind`. Keep the non-select grouped path (groupIdx/chartCols loop) unchanged.

- [ ] In `pkg/parser/json/json.go`, mirror the CSV change: add `jsonRowReader` and `jsonKindFn`, replace the select-routing block with `DispatchSelectMode`, delete `parseJSONMixedMode`/`parseJSONValueMode`/`parseJSONSelectStatMode`/`jsonAxisColumnKind`.

```go
type jsonRowReader struct {
    row     map[string]any
    seenCol map[string]bool
    colOrder []string
    flag     string
}

func (r jsonRowReader) Cell(source string) (string, bool) {
    v, ok := r.row[source]
    if !ok {
        return "", false
    }
    s := strings.TrimSpace(stringify(v))
    return s, s != ""
}

func (r jsonRowReader) Numeric(source string) (float64, bool) {
    v, ok := r.row[source]
    if !ok {
        return 0, false
    }
    return leafNumber(v)
}

func (r jsonRowReader) AvailableColumns() []string { return r.colOrder }
func (r jsonRowReader) FlagLabel() string          { return r.flag }

func jsonKindFn(rows []map[string]any, seenCol map[string]bool, colOrder []string, flag string) parser.AxisColumnKind {
    return func(source, axisKey string) (string, error) {
        if !seenCol[source] {
            return "", fmt.Errorf("%s field %q not found; available: %v", flag, source, colOrder)
        }
        anyNumeric, allNumeric, sawCell := false, true, false
        for _, row := range rows {
            v, ok := row[source]
            if !ok {
                continue
            }
            if strings.TrimSpace(stringify(v)) == "" {
                continue
            }
            sawCell = true
            if _, ok := leafNumber(v); ok {
                anyNumeric = true
            } else {
                allNumeric = false
            }
        }
        if !sawCell {
            return "", fmt.Errorf("%s field %q has no data", flag, source)
        }
        if axisKey == "x" {
            if allNumeric {
                return "value", nil
            }
            return "category", nil
        }
        if !anyNumeric {
            return "", fmt.Errorf("%s field %q is not numeric", flag, source)
        }
        return "value", nil
    }
}
```

- [ ] In `pkg/parser/axes_spec.go`, delete `inferSelectViewMixed` (lines 141-154) and `parseAxisFloat` (lines 156-162). Rewrite `DatasetAxesForSelectView` to use the carried `AxisType`:

```go
func DatasetAxesForSelectView(view []ColumnSpec, _ []shared.DataPoint) []shared.Axis {
    cfg := Config{Axes: append([]ColumnSpec(nil), view...)}
    if isMixedAxes(cfg) {
        return MixedAxes(cfg)
    }
    return ValueAxes(cfg)
}
```

The `results` parameter is now unused but kept for API stability; rename to `_` to silence linters. The parser already sets `AxisType` via `ResolveAxesTypes` inside `DispatchSelectMode`, so the re-inference from `DataPoint.XAxis` strings is no longer needed.

- [ ] Run `go build ./...` — all old call sites must compile now.

- [ ] Run `go test ./...` — all existing tests must stay green. If a test references `IsSelectAxisMode`/`IsMultiSelectStatMode`/`IsMixedMode`, update it to use `cfg.Mode.Is...()` methods. The tests in `select_view_spec_test.go` for `TestIsSelectAxisMode`/`TestIsMultiSelectStatMode` should be rewritten to test `ResolveMode` instead:

```go
func TestResolveMode(t *testing.T) {
    if ResolveMode(Config{}) != ModeAuto {
        t.Fatal("empty config should be ModeAuto")
    }
    cfg := Config{SelectViews: []SelectView{{Columns: []ColumnSpec{{Source: "a"}, {Source: "b"}}}}}
    if m := ResolveMode(cfg); m != ModeValue {
        t.Fatalf("single solo view should be ModeValue, got %d", m)
    }
    cfg.SelectViews = append(cfg.SelectViews, SelectView{Columns: []ColumnSpec{{Source: "a"}, {Source: "c"}}})
    if m := ResolveMode(cfg); m != ModeMultiStat {
        t.Fatalf("two solo views should be ModeMultiStat, got %d", m)
    }
    cfg.Group = []string{"region"}
    cfg.Select = []ColumnSpec{{Source: "price"}}
    if m := ResolveMode(cfg); m != ModeGrouped {
        t.Fatalf("grouped + select should be ModeGrouped, got %d", m)
    }
}
```

- [ ] Commit: `refactor(parser): introduce RowReader, collapse CSV/JSON select-mode duplication`

### Step 1.5: Unify `FlagBag` slice stores

- [ ] In `cmd/cli/flagbag.go`, replace the two maps `slices map[string]*[]string` and `arrays map[string]*[]string` with a single `stringSlices map[string]*[]string`. Update `NewFlagBag` to allocate one entry per `KindStringSlice`/`KindStat`/`KindStringArray` flag.

- [ ] In `Bind`, branch on kind for the pflag call but point both at `b.stringSlices[f.Name]`:

```go
case flags.KindStringSlice:
    def, _ := f.Default.([]string)
    if f.Shorthand != "" {
        fs.StringSliceVarP(b.stringSlices[f.Name], f.Name, f.Shorthand, def, f.Usage)
    } else {
        fs.StringSliceVar(b.stringSlices[f.Name], f.Name, def, f.Usage)
    }
case flags.KindStat:
    sv := &statValue{value: b.stringSlices[f.Name]}
    if f.Shorthand != "" {
        fs.VarP(sv, f.Name, f.Shorthand, f.Usage)
    } else {
        fs.Var(sv, f.Name, f.Usage)
    }
    fs.Lookup(f.Name).NoOptDefVal = statFlagAll
case flags.KindStringArray:
    if f.Shorthand != "" {
        fs.StringArrayVarP(b.stringSlices[f.Name], f.Name, f.Shorthand, nil, f.Usage)
    } else {
        fs.StringArrayVar(b.stringSlices[f.Name], f.Name, nil, f.Usage)
    }
```

- [ ] Replace `StringSlice` and `StringArray` accessors with one backing store:

```go
func (b *FlagBag) StringSlice(name string) []string {
    if p := b.stringSlices[name]; p != nil {
        return *p
    }
    return nil
}

func (b *FlagBag) StringArray(name string) []string {
    return b.StringSlice(name) // same store; semantics differ only at Bind
}

func (b *FlagBag) StringSliceRef(name string) *[]string { return b.stringSlices[name] }
```

- [ ] Update `Reset` to use one case for all three slice kinds:

```go
case flags.KindStringSlice, flags.KindStat, flags.KindStringArray:
    def, _ := f.Default.([]string)
    *b.stringSlices[f.Name] = def
```

- [ ] Update `ChartSeed` and `applySoftRule` to use `b.stringSlices[f.Name]` instead of `b.slices[f.Name]`.

- [ ] Run `go test ./cmd/cli/...` — `flagbag_test.go` must stay green (it tests `StringSliceRef` and routing, both still work).

- [ ] Commit: `refactor(cli): unify FlagBag slice/array stores into one map`

### Step 1.6: Switch `pipeline.go` to `cfg.Mode`

- [ ] In `cmd/cli/pipeline.go` `assembleDataset` (line 444), replace the predicate chain:

```go
func assembleDataset(results []shared.DataPoint, m RunMeta, configs []internal_charts.ChartConfig, cfg parser.Config) *shared.Dataset {
    var view []parser.ColumnSpec
    if cfg.Mode.IsSelectAxis() && !cfg.Mode.IsMultiStat() && len(cfg.SelectViews) == 1 {
        view = cfg.SelectViews[0].Columns
    }
    name := ""
    if len(view) > 0 {
        name = parser.SelectViewDatasetName(view, 0)
    }
    return buildDataset(results, m, configs, cfg, view, name)
}
```

- [ ] In `buildDataset` (line 468), replace the `if/else if` ladder:

```go
func buildDataset(results []shared.DataPoint, m RunMeta, configs []internal_charts.ChartConfig, cfg parser.Config, view []parser.ColumnSpec, viewName string) *shared.Dataset {
    var axes []shared.Axis
    switch cfg.Mode {
    case parser.ModeAuto:
        axes = deriveAxesFromData(results)
        autoEnableValueMode3D(configs, axes, valueModeHasMetric(cfg, results))
    case parser.ModeMultiStat:
        axes = parser.MultiSelectStatAxes(cfg.SelectViews)
    case parser.ModeValue, parser.ModeMixed:
        if len(view) > 0 {
            axes = parser.DatasetAxesForSelectView(view, results)
        } else {
            axes = parser.DatasetAxesForSelectView(cfg.SelectViews[0].Columns, results)
        }
        autoEnableValueMode3D(configs, axes, valueModeHasMetric(cfg, results))
    default: // ModeGrouped
        axes = parser.GroupAxes(cfg)
        if len(cfg.Axes) > 0 {
            if parser.IsMixedAxesPublic(cfg) {
                axes = parser.MixedAxes(cfg)
            } else {
                axes = parser.ValueAxes(cfg)
            }
        }
    }
    axes = appendMetricAxis(axes, cfg, results)
    // ... rest unchanged
```

Export `isMixedAxes` as `IsMixedAxesPublic` (or inline the check) since `buildDataset` is in `cmd/cli` and `isMixedAxes` is in `pkg/parser`. Add a tiny exported helper to `pkg/parser/mode.go`:

```go
// IsMixedAxes reports whether the Axes carry one category + at least one value
// (post ResolveAxesTypes). Used by the dataset builder for the --axes mixed path.
func IsMixedAxes(cfg Config) bool { return isMixedAxes(cfg) }
```

- [ ] Run `go test ./...` — all green.

- [ ] Commit: `refactor(cli): switch pipeline mode ladder to cfg.Mode switch`

### Step 1.7: Final Go-side verification

- [ ] Run `go vet ./...` — no warnings.

- [ ] Run `go test ./...` — all packages green.

- [ ] Run a manual smoke test: `go run ./cmd/vizb scatter examples/csv/region-metrics.csv --select region,latency -o /tmp/mixed.html` and confirm the HTML renders without error.

---

## Task 2: TS-side refactors (ChartBuilder Strategy + polymorphism + worker registry)

**Files:**
- Create: `ui/src/lib/builders/types.ts`
- Create: `ui/src/lib/builders/grouped.ts`
- Create: `ui/src/lib/builders/preserveRows.ts`
- Create: `ui/src/lib/builders/value.ts`
- Create: `ui/src/lib/builders/mixed.ts`
- Create: `ui/src/lib/builders/index.ts`
- Modify: `ui/src/lib/transform.ts:140-330` (delegate to builders)
- Modify: `ui/src/lib/utils.ts:171-280` (delegate to builders)
- Modify: `ui/src/workers/transform.worker.ts` (mode registry)

### Step 2.1: Create `ChartBuilder` interface and shape registry

- [ ] Create `ui/src/lib/builders/types.ts`:

```ts
import type { ChartData, DataPoint, AxisLabels, Sort, ScaleType, CanonicalAxisOrders, Axis } from '@/types'

export interface BuildContext {
  signature: string
  statTemplate: ChartData['statTemplate'] & object
  labels?: AxisLabels
  sort: Sort
  showLabels: boolean
  scale: ScaleType
  canonical?: CanonicalAxisOrders
  threeD: boolean
  preserveRows: boolean
}

export interface ChartBuilder {
  /** Build the ChartData for this chart shape from the raw data points. */
  build(data: DataPoint[], ctx: BuildContext): ChartData
  /** Whether the chart has any plottable data. */
  plottable(chart: ChartData): boolean
  /** Cardinality for an axis badge. */
  badgeCount(chart: ChartData, axis: 'x' | 'y' | 'z'): number
  /** Sum of every plotted metric value. */
  grandTotal(chart: ChartData, visibleZ?: Record<string, boolean>): number
  /** Whether this chart should render as 3D. */
  is3D(chart: ChartData, cfg?: { threeD?: boolean }, axes?: Axis[]): boolean
}

export const builderStatType = (chart: ChartData): string => chart.statType ?? 'grouped'
```

- [ ] Commit: `refactor(ui): add ChartBuilder interface`

### Step 2.2: Extract `GroupedBuilder` (non-preserveRows path)

- [ ] Create `ui/src/lib/builders/grouped.ts` containing the non-preserveRows branch of `buildChartForSignature` (the `dataMap`/`countMap` accumulation + averaging + series construction). The `build` method returns a `ChartData` with `statType: 'grouped'`. The `plottable`/`badgeCount`/`grandTotal`/`is3D` methods contain the existing grouped branches from `utils.ts`.

- [ ] Commit: `refactor(ui): extract GroupedBuilder`

### Step 2.3: Extract `PreserveRowsBuilder` (preserveRows path)

- [ ] Create `ui/src/lib/builders/preserveRows.ts` containing the preserveRows branch (the `yOrder`/`ySeen` accumulation + `useCategoryScatter` sub-branch). `statType: 'preserveRows'`. The mixed-tuples sub-path sets `chart.mixedTuples`/`chart.xCategories`; the series sub-path builds `series` with null-padded values.

- [ ] Commit: `refactor(ui): extract PreserveRowsBuilder`

### Step 2.4: Extract `ValueBuilder` and `MixedBuilder`

- [ ] Create `ui/src/lib/builders/value.ts` — move `buildValueModeChart` (transform.ts:470-519) into the builder's `build` method. `statType: 'value'`. The `plottable`/`badgeCount`/`grandTotal`/`is3D` methods contain the existing value-mode branches.

- [ ] Create `ui/src/lib/builders/mixed.ts` — move `buildMixedModeChart` (transform.ts:522-595) into the builder's `build` method. `statType: 'mixed'`. The `plottable`/`badgeCount`/`grandTotal`/`is3D` methods contain the mixed-mode branches currently inlined in `utils.ts`.

- [ ] Commit: `refactor(ui): extract ValueBuilder and MixedBuilder`

### Step 2.5: Create builder registry

- [ ] Create `ui/src/lib/builders/index.ts`:

```ts
import type { ChartBuilder } from './types'
import { GroupedBuilder } from './grouped'
import { PreserveRowsBuilder } from './preserveRows'
import { ValueBuilder } from './value'
import { MixedBuilder } from './mixed'
import type { ChartData } from '@/types'

const grouped = new GroupedBuilder()
const preserveRows = new PreserveRowsBuilder()
const value = new ValueBuilder()
const mixed = new MixedBuilder()

/** Pick the builder for a chart based on its statType and flags. */
export function builderForChart(chart: ChartData): ChartBuilder {
  if (chart.statType === 'value') return value
  if (chart.statType === 'mixed') return mixed
  if (chart.statType === 'preserveRows') return preserveRows
  return grouped
}

/** Pick the builder for the build phase based on context flags. */
export function pickBuilder(ctx: { preserveRows?: boolean; mixedMode?: boolean; valueMode?: boolean }): ChartBuilder {
  if (ctx.valueMode) return value
  if (ctx.mixedMode) return mixed
  if (ctx.preserveRows) return preserveRows
  return grouped
}

export { grouped, preserveRows, value, mixed }
```

- [ ] Commit: `refactor(ui): add builder registry`

### Step 2.6: Slim `buildChartForSignature` to delegate

- [ ] In `ui/src/lib/transform.ts`, replace the body of `buildChartForSignature` (lines 140-330) with:

```ts
export function buildChartForSignature(
  data: DataPoint[],
  signature: string,
  statTemplate: Stat,
  labels: AxisLabels | undefined,
  sort: Sort,
  showLabels = false,
  scale: ScaleType = 'linear',
  canonical?: CanonicalAxisOrders,
  threeD = false,
  preserveRows = false
): ChartData {
  const builder = pickBuilder({ preserveRows })
  return builder.build(data, { signature, statTemplate, labels, sort, showLabels, scale, canonical, threeD, preserveRows })
}
```

Move the shared tail (sort, labels, render3D dispatch, statType assignment) into each builder's `build` method — or extract a `finalizeChart(chart, ctx)` helper in `types.ts` that all builders call. Prefer the helper to avoid duplicating the tail.

- [ ] Commit: `refactor(ui): delegate buildChartForSignature to ChartBuilder`

### Step 2.7: Delegate `utils.ts` shape queries to builders

- [ ] In `ui/src/lib/utils.ts`, replace the inlined branches in `chartHasPlottableData`, `chartAxisBadgeCount`, `computeChartGrandTotal`, `is3D`, `canOfferValue3D` with `builderForChart(chart).<method>(...)`. For `is3D` and `canOfferValue3D` which take `axes` not `chart`, keep a small mode check that picks the builder, then delegate.

Example for `chartHasPlottableData`:

```ts
export const chartHasPlottableData = (chart: ChartData): boolean =>
  builderForChart(chart).plottable(chart)
```

- [ ] Run `vitest run` — `utils.test.ts`, `transform.test.ts`, `transform.worker.test.ts` must stay green.

- [ ] Commit: `refactor(ui): delegate utils shape queries to ChartBuilder`

### Step 2.8: Replace worker triplication with mode registry

- [ ] In `ui/src/workers/transform.worker.ts`, replace the `valueMode`/`mixedMode`/normal branches with a registry. Define a `ModeHandler` interface and three implementations:

```ts
interface ModeHandler {
  ready(s: State): ReadyMessage
  compute(s: State, msg: ComputeMessage): void
}
```

- `GroupedHandler`: the existing normal path (`applyArrangement` + `readyReply` + `buildChartForSignature`).
- `ValueHandler`: the existing `valueModeReadyReply` + `buildValueModeChart` path.
- `MixedHandler`: the existing `mixedModeReadyReply` + `buildMixedModeChart` path.

Pick the handler once at init based on `isValueMode(axes)` / `isMixedMode(axes)` / default. The `init`, `setArrangement`, and `compute` message handlers all delegate to `state.handler.<method>`.

- [ ] Run `vitest run` — `transform.worker.test.ts` must stay green.

- [ ] Run `cd ui && npm run typecheck` — no type errors.

- [ ] Commit: `refactor(ui): replace worker mode branches with ModeHandler registry`

### Step 2.9: Final TS-side verification

- [ ] Run `vitest run` — all UI tests green.

- [ ] Run `cd ui && npm run typecheck` — no type errors.

- [ ] Run `cd ui && npm run build` — build succeeds.

---

## Task 3: Cross-layer cleanup + final verification

**Files:**
- Modify: `ui/src/composables/useDataPoint.ts:116-132,204-208`
- Modify: `ui/src/components/SettingsPanel.vue:48-50,120-127`
- Create: `.superpowers/sdd/refactor-report.md`

### Step 3.1: Replace mirrored `useDataPoint` computeds with one `chartMode`

- [ ] In `ui/src/composables/useDataPoint.ts`, replace the 4 computeds (`isValueModeActive`, `isMixedModeActive`, `isValueModeDataset`, `isMixedModeDataset`) with one:

```ts
type ChartMode = 'grouped' | 'value' | 'mixed'

const chartMode = computed<ChartMode>(() => {
  const axes = activeDataSet.value?.axes
  if (!axes?.length) return 'grouped'
  if (isValueMode(axes)) return 'value'
  if (isMixedMode(axes)) return 'mixed'
  return 'grouped'
})
```

- [ ] Export `chartMode` and keep `isValueMode`/`isMixedMode` as computed aliases for backward compatibility with `SettingsPanel.vue`/`SwapControl.vue`:

```ts
return {
  // ...
  chartMode,
  isValueMode: computed(() => chartMode.value === 'value'),
  isMixedMode: computed(() => chartMode.value === 'mixed'),
  isValueModeDataset: computed(() => chartMode.value === 'value'),
  isMixedModeDataset: computed(() => chartMode.value === 'mixed'),
}
```

- [ ] In `ui/src/components/SettingsPanel.vue`, replace `isValueModeDataset.value || isValueMode.value || isMixedModeDataset.value` with `chartMode.value !== 'grouped'` (or keep the existing aliases if the alias approach is lower-risk — prefer the alias for this task to minimize UI churn).

- [ ] Run `vitest run` — green.

- [ ] Commit: `refactor(ui): collapse useDataPoint mode computeds into chartMode`

### Step 3.2: Write SDD execution report

- [ ] Create `.superpowers/sdd/refactor-report.md` following the format of the existing task reports (task-1-report.md etc.). Document:
  - Status: DONE
  - Summary of the 3 tasks
  - Lines deleted vs added (run `git diff main...HEAD --stat` before/after)
  - Test results (`go test ./...` and `vitest run` output)
  - Acceptance criteria table (all 9 findings addressed)

### Step 3.3: Final full verification

- [ ] Run `go vet ./...` — no warnings.

- [ ] Run `go test ./...` — all packages green.

- [ ] Run `cd ui && npm run typecheck` — no type errors.

- [ ] Run `cd ui && vitest run` — all UI tests green.

- [ ] Run `cd ui && npm run build` — build succeeds.

- [ ] Manual smoke test: `go run ./cmd/vizb scatter examples/csv/region-metrics.csv --select region,latency -o /tmp/mixed.html` and `go run ./cmd/vizb scatter examples/csv/region-metrics.csv --select region,latency --select region,sales -o /tmp/multi.html` — both render without error.

- [ ] Commit the report: `docs(sdd): refactor pattern cleanup report`

---

## Acceptance Criteria

| Finding | Resolution | Task |
|---------|------------|------|
| F1: Dead code `readCSVTable`/`readJSONArray` | Deleted in Step 1.1 | 1 |
| F2: CSV↔JSON parser duplication | `RowReader` interface + shared `ParseMixedMode`/`ParseValueMode`/`ParseSelectStatMode` | 1 |
| F3: Mode-predicate explosion (9 sites) | `Mode` enum resolved once in `ParseConfig`; `cfg.Mode` switch everywhere | 1 |
| F4: `buildChartForSignature` dual code path | `ChartBuilder` strategies: `GroupedBuilder`/`PreserveRowsBuilder`/`ValueBuilder`/`MixedBuilder` | 2 |
| F5: `utils.ts` shotgun surgery per shape | Per-shape `plottable`/`badgeCount`/`grandTotal`/`is3D` methods on builders | 2 |
| F6: `FlagBag` parallel arrays | Unified `stringSlices` map; branch only at `Bind` | 1 |
| F7: Dual type-inference paths | `AxisType` carried from parser; `inferSelectViewMixed`/`parseAxisFloat` deleted | 1 |
| F8: Worker dispatch triplication | `ModeHandler` registry with `ready`/`compute` per mode | 2 |
| F9: `useDataPoint` mirrored computeds | One `chartMode` computed + aliases | 3 |

## Risk Mitigation

- **No behavior changes** — pure structural refactor. If any existing test goes red, the refactor is wrong, not the test.
- **Existing tests are the safety net** — no new tests written; the PR already added comprehensive coverage for all 5 modes.
- **`go test ./...` and `vitest run` run between every step** — failures caught immediately, not at the end.
- **Each task is independently committable** — if Task 2 stalls, Task 1's commits are still valid and reviewable.
