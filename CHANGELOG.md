# Changelog

Notable changes to Vizb documented here.

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

# [v0.15.0]

### Added

- **`--smooth` for 2D line charts** — curved segments via CLI, `--chart line:smooth`, and UI toggle; unavailable for 3D and non-line charts ([#187](https://github.com/goptics/vizb/pull/187)).
- **`--horizontal` for 2D grouped bar charts** — layout-only (categories on Y, values on X); distinct from `--swap`; CLI, `--chart bar:horizontal`, UI toggle, and `?bar.h=` deep link ([#190](https://github.com/goptics/vizb/pull/190)).
- **Multi-select descriptive column picker** — stats panel lets viewers choose which descriptive columns drive metrics (select-all / reset) ([#191](https://github.com/goptics/vizb/pull/191)).
- **`--stack` for 2D bar and line charts** — stacked bars / stacked area lines for grouped x+y data; skipped when z is present or scale is log ([#194](https://github.com/goptics/vizb/pull/194)).
- **`--theme` color palette support** — 13 built-in series palettes plus custom comma-separated hex colors; theme stored on datasets and preserved through merge; UI palette selector with `localStorage` ([#195](https://github.com/goptics/vizb/pull/195)).

### Changed

- **Embed UI generation in CI** — stop tracking `pkg/template/vizb-ui.gen.go`; generate via `task init` / `task build:ui` and shared `setup-embed-ui` composite; drop gen-sync pre-commit/CI checks ([#197](https://github.com/goptics/vizb/pull/197)).

### Fixed

- **Grouped 3D z-axis sticky labels** — set empty z-axis `name` under option merge for grouped bar/line/scatter 3D so labels do not stick; value-mode still shows the metric name ([#192](https://github.com/goptics/vizb/pull/192)).

# [v0.14.1] - 2026-07-04

### Fixed

- **Same-tag merge no longer wipes accumulated history** — re-merging a dataset with an existing tag now replaces only that tag's data points on the inject axis (`-A`); older versions in `history[]` and data from other tags are preserved ([#181](https://github.com/goptics/vizb/pull/181)).
- **Tag-axis injection invisible in UI** — merge now adds the inject dimension to `axes` when missing so injected tag values appear in charts ([#182](https://github.com/goptics/vizb/pull/182)).

### Changed

- **Merge documentation** — updated merge command docs, merging guide, and stateful CI guide to reflect same-tag replacement and auto-axis behavior.
- **Stateful CI example storage** — replace GitHub artifact storage with R2 in the stateful CI example ([#180](https://github.com/goptics/vizb/pull/180)).

# [v0.14.0] - 2026-07-01

### Added

- **Solo `--select` axis mode, multi-stat, and mixed axes** — repeatable `--select x,y[,z]` without grouping; multi-stat mode merges repeatable `dim,metric` selects into one dataset with chart tabs split by stat type; mixed categorical+numeric axes now render on bar/line (2D and 3D) ([#173](https://github.com/goptics/vizb/pull/173)).
- **`--id` flag and `?id=` URL dataset selection** — stable top-level `id` field on datasets, set via `--id`; UI prefers `?id=` deep links over the legacy `?d=` index param ([#171](https://github.com/goptics/vizb/pull/171)).
- **`--visualmap` for 2D scatter** — opt-in gradient coloring for 2D scatter (metric → z → y fallback), separate from the existing `--3d-visualmap` path ([#169](https://github.com/goptics/vizb/pull/169)).
- **`--symbol` / `--symbol-size` flags** — custom marker shape and size for `vizb line` and `vizb scatter` (2D and 3D) ([#166](https://github.com/goptics/vizb/pull/166)).

### Changed

- **Applicability-rule pipeline** — declarative per-flag rules (`Keep`/`WarnKeep`/`Skip`/`Fatal`) replace ad-hoc eligibility checks like `WarnThreeDIfIneligible`; `config/` moved to `internal/{charts,flags}/` ([#170](https://github.com/goptics/vizb/pull/170)).
- **`dataZoom` and axis pointer behavior** improved for large categorical datasets on line/bar/scatter ([#174](https://github.com/goptics/vizb/pull/174)).

### Fixed

- **Sort not applied to 1-axis charts** — `--sort asc`/`desc` now affects solo-select bar/line charts; a legacy `preserveRows` override in `Dashboard.vue` was skipping the sort path ([#175](https://github.com/goptics/vizb/pull/175)).
- **2D scatter `--visualmap` ignored on large datasets** — ECharts skips `visualMap` in `large` mode; large-mode is now disabled when visualmap is on ([#172](https://github.com/goptics/vizb/pull/172)).
- **Pseudo-3D grid sizing** — value-mode `--3d` on 2D x+y data no longer forces a fixed 100×100×100 cube; box dimensions now size from category counts, and camera `viewControl.distance` is capped at 300 ([#161](https://github.com/goptics/vizb/pull/161)).
- **Y-axis no longer forced to zero** on line/scatter charts — `fitYAxisToData` scales the axis to the data range instead of wasting space including zero; bar charts keep the zero baseline ([#163](https://github.com/goptics/vizb/pull/163)).
- **Stats panel math corrections** — `computeProfiles` no longer drops `zAxis` for descriptive stats on 3D charts; 95% CI switched from a z-interval to a Student-t interval (correct for small `n`); `correlationMatrix` diagonal and `cqv` edge guards fixed for constant/negative-spanning columns ([#160](https://github.com/goptics/vizb/pull/160)).
- **Tooltip legend layout and sigma math** — legend rows flow into balanced multi-column grids past 10 entries; 3D tooltip Σ formulas corrected to avoid double-counting the hovered cell ([#159](https://github.com/goptics/vizb/pull/159)).
- **Metric preservation and auto-group aggregation** — `CollapseDataPointsByKey`/`AggregateDataPoints` now preserve `Metric`; tabular rows auto-grouped onto the X axis aggregate correctly ([#174](https://github.com/goptics/vizb/pull/174)).
- **GitHub Action merge-deploy stripped scatter settings** — `merge-deploy` called the action without a `charts` input, falling back to the default `bar,line,pie` and stripping scatter config on `vizb ui`; the action default is now empty ([#155](https://github.com/goptics/vizb/pull/155)).

# [v0.13.0]

### Added

- **Tabular visualization engine** — Vizb now charts any CSV or JSON table alongside Go, Rust, and JavaScript benchmark output. `--parser` defaults to `auto` and inspects input content to pick the right parser ([#123](https://github.com/goptics/vizb/pull/123)).
- **CSV parser** — comma-delimited tables; numeric columns become series; quoted fields, BOM strip, and ragged rows supported ([#123](https://github.com/goptics/vizb/pull/123)).
- **JSON parser** — array-of-objects input; nested objects flattened to dotted keys (`mem.alloc`); numeric strings accepted ([#123](https://github.com/goptics/vizb/pull/123)).
- **`--group` / `-g` flag** — map CSV/JSON columns to Name / X / Y / Z dimensions ([#123](https://github.com/goptics/vizb/pull/123)).
- **Bracket slot grouping** — single column cell encodes multiple dimensions via `[...]` slots in `--group-pattern` (e.g. `-p "[x-y-n],z"` for dates or slash paths); CSV/JSON only ([#123](https://github.com/goptics/vizb/pull/123)).
- **Column-value parser in group pattern** — parse cell contents inside bracket slots ([#131](https://github.com/goptics/vizb/pull/131)).
- **Trailing and consecutive separator skips** in pattern parsing ([#132](https://github.com/goptics/vizb/pull/132)).
- **`--select` flag** — pick specific value columns for CSV/JSON; optional `{label}` rename ([#140](https://github.com/goptics/vizb/pull/140)).
- **`--json-path` flag** — jq-like dot path to chart a nested array inside a JSON envelope (e.g. `.data.results`) ([#143](https://github.com/goptics/vizb/pull/143)).
- **Row aggregation** — when `--group` is active on csv/json, rows sharing the same `(Name, XAxis, YAxis, ZAxis)` key are summed before charting ([#123](https://github.com/goptics/vizb/pull/123)).
- **Auto-group mode** — with no `-g`/`-r`, infers the highest-cardinality non-numeric column as the X axis so `vizb data.csv` just works ([#145](https://github.com/goptics/vizb/pull/145)).
- **Auto-value mode** — when all columns are numeric, auto-assigns the first 2–3 columns as coordinate axes (`x`, `y`, `z`); a 4th column becomes a visualMap metric; works on bar, line, and scatter ([#145](https://github.com/goptics/vizb/pull/145)).
- **Scatter chart** — coordinate plotting in grouped or auto-value mode; 2D and 3D ([#142](https://github.com/goptics/vizb/pull/142)).
- **Heatmap chart** — X×Y colored grid; z-axis folded into per-cell sums ([#126](https://github.com/goptics/vizb/pull/126)).
- **Radar chart** — spider/web polygons for multi-metric profile comparison ([#128](https://github.com/goptics/vizb/pull/128)).
- **3D charts** — interactive WebGL bar3D, line3D, and scatter3D via echarts-gl; grouped (`-p n/x/y/z`), pseudo-3D (`--3d`), or continuous auto-value 3D ([#122](https://github.com/goptics/vizb/pull/122)).
- **3D flags** — `--3d`, `--3d-rotate`, and `--3d-visualmap` on bar, line, and scatter subcommands ([#122](https://github.com/goptics/vizb/pull/122), [#135](https://github.com/goptics/vizb/pull/135)).
- **3D tooltip totals** — per-legend Σ sums in 3D tooltip rows; chart total badge on hover ([#133](https://github.com/goptics/vizb/pull/133), [#141](https://github.com/goptics/vizb/pull/141)).
- **Statistics panel (`--stat`)** — opt-in per-chart analytics: 33 descriptive metrics across 7 groups, plus a correlation matrix (Pearson, Spearman, Kendall, distance correlation); computed off-thread in a Web Worker ([#126](https://github.com/goptics/vizb/pull/126)).
- **Stats panel UX** — sortable/searchable table with virtualization, 3D-aware per-z sub-rows, and CSV export ([#126](https://github.com/goptics/vizb/pull/126), [#136](https://github.com/goptics/vizb/pull/136), [#137](https://github.com/goptics/vizb/pull/137)).
- **Chart subcommands** — `vizb bar`, `vizb line`, `vizb scatter`, `vizb pie`, `vizb heatmap`, `vizb radar`; each accepts only flags valid for that chart type ([#129](https://github.com/goptics/vizb/pull/129)).
- **`--chart` per-chart overrides** — repeatable flag for per-type settings (e.g. `--chart bar:swap=yx,scale=log --chart pie:labels`) ([#129](https://github.com/goptics/vizb/pull/129)).
- **`--swap` flag** — axis permutation at generation time (e.g. `--swap yxn`) ([#127](https://github.com/goptics/vizb/pull/127)).
- **`vizb ui` command** — renamed from `vizb html`; `html` kept as a backward-compatible alias ([#121](https://github.com/goptics/vizb/pull/121)).
- **v0.12.0 settings migration** — old single-object `settings` shape auto-migrated to per-chart typed configs on JSON read ([#130](https://github.com/goptics/vizb/pull/130)).
- **Async chart pipeline** — transform Web Worker re-renders charts off the main thread; charts reveal progressively as each job completes ([#125](https://github.com/goptics/vizb/pull/125)).
- **Go-stage chunk pruning** — HTML bundle includes only JS chunks for active chart types ([#126](https://github.com/goptics/vizb/pull/126)).
- **Field-registry settings panel** — UI settings driven by per-chart field registry instead of a flat toggle list ([#130](https://github.com/goptics/vizb/pull/130)).
- **Per-chart URL sync** — query parameters target individual chart configs, not one global settings blob ([#127](https://github.com/goptics/vizb/pull/127)).
- **CSV example suite** — auto-group, auto-value, and 3D recipes ([#145](https://github.com/goptics/vizb/pull/145)).
- **Live GitHub contribution skylines example** — fetched from API at deploy time; demonstrates `--json-path`, `--select`, and `--stat` ([#148](https://github.com/goptics/vizb/pull/148)).
- **GitHub Action: `cmd` / `file` inputs** — replace `bench-cmd`/`bench-file`; support any tabular or benchmark input.
- **GitHub Action: `group` input** — forwards `-g`; omit to enable auto-group on csv/json.
- **GitHub Action: `select` input** — select value columns for CSV/JSON data (forwards `--select`).
- **GitHub Action: `json-path` input** — chart a nested JSON array via a jq-like dot path (forwards `--json-path`).
- **GitHub Action: `stat` input** — enable the stats panel (forwards `--stat`).
- **GitHub Action: `chart` input** — per-chart overrides as a multi-line string, one `--chart` per line (e.g. `bar:scale=log`).
- **GitHub Action: `enable-3d` input** — bundle the 3D renderer for `vizb ui` (forwards `--3d`).
- **GitHub Action: `vizb-binary` input** — path to a pre-built binary on the runner; skips release download/cache, useful for testing local builds or unreleased changes.

### Changed

- **Data model: `Benchmark` → `Dataset`** — `DataPoint` gains `zAxis` and serial `axes[]` metadata; per-chart typed `settings` array replaces the global settings blob ([#124](https://github.com/goptics/vizb/pull/124), [#127](https://github.com/goptics/vizb/pull/127), [#130](https://github.com/goptics/vizb/pull/130)).
- **Per-chart settings architecture** — scale, sort, labels, swap, and 3D options are scoped per chart type; bar/line/scatter carry scale and 3D fields; pie/heatmap/radar do not ([#127](https://github.com/goptics/vizb/pull/127), [#130](https://github.com/goptics/vizb/pull/130)).
- **`--scale` moved to per-chart scope** — removed from root command; available on `vizb bar`, `vizb line`, `vizb scatter`, and via `--chart` overrides ([#129](https://github.com/goptics/vizb/pull/129)).
- **Default chart types unchanged** — root still defaults to `bar,line,pie`; scatter, heatmap, and radar are opt-in via `-c` or subcommands.
- **GitHub Action: `parser` default** — changed from `go` to `auto` for content-based format detection.
- **GitHub Action: `charts` input** — now accepts `scatter`, `heatmap`, and `radar`.
- **GitHub Action: `tag-axis` input** — now accepts `z` in addition to `n`, `x`, and `y`.
- **CI: split CLI and UI workflows** — renamed `ci.yml` to `cli.yml` with layered lint → format → test → build gates; added `ui.yml` for the Vue app with the same pipeline shape ([#147](https://github.com/goptics/vizb/pull/147)).
- **CI: parallelized gates** — action tests centralized in CLI workflow ([#147](https://github.com/goptics/vizb/pull/147)).

### Removed

- **Root `--scale` flag** — use `--chart bar:scale=log` or a chart subcommand instead ([#129](https://github.com/goptics/vizb/pull/129)).
- **Global `FlagState`** — replaced by typed per-command options and per-chart config materialisation ([#129](https://github.com/goptics/vizb/pull/129)).
- **GitHub Action: `scale` input** — **BREAKING.** Use the `chart` input instead (e.g. `chart: 'bar:scale=log'`).

### Deprecated

- **Root `--sort` and `--show-labels`** — `vizb` (root) prints a stderr warning recommending the `--chart` equivalent (e.g. `--chart bar:sort=asc`, `--chart pie:labels`) ([#129](https://github.com/goptics/vizb/pull/129)).
- **GitHub Action: `sort` input** — use `chart` instead (e.g. `chart: 'bar:sort=asc'`).
- **GitHub Action: `show-labels` input** — use `chart` instead (e.g. `chart: 'pie:labels'`).
- **GitHub Action: `bench-cmd` / `bench-file` inputs** — use `cmd` / `file` instead.

# [0.12.0] - 2026-06-03

### Added

- **`--data-url` Flag**: New flag for `vizb html` to decouple benchmark data from the UI bundle — serves data from an external URL instead of embedding JSON inline ([#119](https://github.com/goptics/vizb/pull/119)).
- **CPU & OS in History Entries**: History popover now shows CPU and OS info per entry alongside tag and timestamp ([#118](https://github.com/goptics/vizb/pull/118)).
- **Multi-Language CI Examples**: Added CI pipeline generating live examples for Rust (Criterion, Divan) and JavaScript (Vitest, TinyBench) parsers ([#112](https://github.com/goptics/vizb/pull/112)).
- **Logo Animation**: Animated SVG logo using anime.js with OG image config ([#111](https://github.com/goptics/vizb/pull/111)).


### Fixed

- **Version Injection**: Version now injected via ldflags in GoReleaser build — binary reports correct version at runtime ([#116](https://github.com/goptics/vizb/pull/116)).
- **Line Chart Labels**: `showLabels` state was not applied to line charts — now consistent with bar/pie charts ([#114](https://github.com/goptics/vizb/pull/114)).
- **CI**: Resolved bench file path routing, write permissions, subdirectory handling, and group regex issues in the multi-lang examples pipeline.

### Changed

- **Version Embedding**: Refactored to use `init()` for embedding version flag instead of inline assignment ([#115](https://github.com/goptics/vizb/pull/115)).
- **Taskfile**: `act:test:stateless` and `act:test:stateful` now run containers as current user (`--user $(id -u):$(id -g)`) to avoid permission issues.

# [0.11.0] - 2026-05-23

### Added

- **Multi-Language Parser Registry**: Vizb now supports benchmark output from Rust and JavaScript via a new parser registry. Ships with parsers for Rust (Criterion, Divan) and JavaScript (Vitest, TinyBench) — making Vizb a cross-language benchmark visualization tool.
- **`--parser` Flag**: Explicitly select a parser for non-Go benchmark input (`--parser rs:criterion`, `--parser js:vitest`, etc.). Auto-detection based on file extension also supported.
- **Rust Parsers**: Added parsers for Criterion (`rs:criterion`) and Divan (`rs:divan`) benchmark output formats.
- **JavaScript Parsers**: Added parsers for Vitest (`js:vitest`) and TinyBench (`js:tinybench`) benchmark output formats.
- **Quick Installation**: One-command install via `curl` (macOS/Linux) and `winget` (Windows) — no manual binary download needed.

### Changed

- **Tests migrated to testify**: Test suite standardized on `testify/assert` for consistent assertions and better failure messages.

# [0.10.3] - 2026-05-20

### Fixed

- **Timestamp for Untagged Benchmarks**: Timestamp was only set when `--tag` flag was used. Now all benchmarks get a generation timestamp, enabling the UI timestamp badge for every run ([#95](https://github.com/goptics/vizb/pull/95)).
- **Draft Release Blocking Deploy**: GoReleaser was configured with `draft: true`, creating hidden releases. The `deploy-examples` workflow couldn't download binaries from draft releases (404). Removed the draft setting so releases publish immediately after approval ([#98](https://github.com/goptics/vizb/pull/98)).

### Changed

- **CI: Remove Redundant setup-go**: The `deploy-examples` workflow used `actions/setup-go@v6` in both `convert` and `merge-deploy` jobs, but the composite action downloads the vizb binary via curl — no Go toolchain needed ([#96](https://github.com/goptics/vizb/pull/96)).
- **CI: Go Version Consistency**: `release.yml` now uses `go-version-file: go.mod` instead of hardcoded `'^1.24'` for consistency with other workflows ([#96](https://github.com/goptics/vizb/pull/96)).
- **CI: Approval Issue Customization**: Added custom `issue-title` and `issue-body` to the manual-approval step — shows the release tag, build targets, and post-approval flow in the approval issue ([#98](https://github.com/goptics/vizb/pull/98)).

# [0.10.2] - 2026-05-20

### Added

- **Timestamp Badge**: Dashboard header now shows the benchmark's last update time with a calendar icon. Clicking reveals a popover with full tag history (all versions + timestamps) via radix-vue Popover ([#92](https://github.com/goptics/vizb/pull/92)).
- **Reusable Badge Component**: New `Badge.vue` component (icon + label + value) replaces inline CPU badge for consistent header styling ([#92](https://github.com/goptics/vizb/pull/92)).
- **Release Guard**: Manual approval gate via `trstringer/manual-approval` — tag push pauses the release workflow until explicitly approved in the Actions UI ([#93](https://github.com/goptics/vizb/pull/93)).
- **Auto-Deploy on Release**: `deploy-examples` workflow now triggers automatically when a release completes successfully, ensuring examples use the newly published binary ([#93](https://github.com/goptics/vizb/pull/93)).

### Changed

- **Benchmark TS Types**: Added `HistoryEntry`, `tag`, `timestamp`, and `history` fields to the TypeScript `Benchmark` type to match the Go data model ([#92](https://github.com/goptics/vizb/pull/92)).
- **PopoverContent**: Added optional `align` prop (`start` | `center` | `end`) to `PopoverContent.vue` for flexible popover positioning ([#92](https://github.com/goptics/vizb/pull/92)).
- **Deploy Trigger Fix**: Removed `pkg/template/ui/**` from `deploy-examples` push trigger — UI-only changes on main no longer redeploy examples using a stale binary ([#93](https://github.com/goptics/vizb/pull/93)).
- **Documents**: Updated README description and features to mention GitHub Action and release guard ([#93](https://github.com/goptics/vizb/pull/93)).

# [0.10.1] - 2026-05-19

### Fixed

- **Windows compatibility**: `unzip -o` in action to avoid interactive overwrite prompt on re-runs
- **Stateful test**: Fixed after `examples/jsons/` removal — generates previous JSON via the action itself
- **`mkdir -p`** in Generate JSON step to handle nested output paths (e.g., `results/prev/hash.json`)
- **Taskfile**: Removed stale `cp examples/jsons/hash.json` from `act:test:stateful` task

### Changed

- **CI: test-before-deploy pipeline**: `test-action` now gates `deploy-examples` via `workflow_call` — no deployment without passing tests on all OSes ([#87](https://github.com/goptics/vizb/pull/87))
- **CI: smarter triggers**: `ci.yml` only runs Go tests on source changes (`**.go`, `go.mod`, `go.sum`, `ci.yml`) instead of every push
- **CI: Go version**: `ci.yml` uses `go-version-file: go.mod` instead of hardcoded `1.24`
- **CI: deploy gating**: `deploy-examples` restricted to `main` branch only
- **Action: optional outputs**: `output-html` and `output-json` fully optional with separate gated generate steps
- **Action: simplified merge**: Always uses `vizb merge` (no intermediate `bench-new.json`)
- **Docs**: Added binary download option to installation section

# [0.10.0] - 2026-05-18

### Added

- **GitHub Action** (`action.yml`): Composite action for running vizb in CI
  - Auto-downloads vizb binary via curl with multi-OS support (linux, macos, windows)
  - Binary caching for pinned versions
  - Inputs: `bench-cmd`, `bench-file`, `tag`, `name`, `merge-files`, `merge-dir`, `tag-axis`, etc.
  - Outputs: `output-file-html`, `output-file-json`
  - Branding for GitHub Marketplace
- **`vizb html` subcommand**: Render benchmark JSON to interactive HTML
  - `vizb html <file.json> -o <output.html>`
  - Pure render — no merge or tag injection
  - Accepts single JSON object or array

### Changed

- **`vizb merge` JSON-only**: Merge command no longer generates HTML — outputs JSON only. Use `vizb html` for HTML rendering after merge.
- **Input priority**: File argument now takes precedence over piped stdin, matching standard Unix convention (`grep`, `cat`, `jq`). Heredocs in CI no longer need `< /dev/null` workaround.
- **`MustCreateFile`**: Auto-creates parent directories for output file paths.

# [0.9.5] - 2026-05-16

### Changed

- **Benchmark Timestamp**: Replaced `Runtimes map[string]string` with direct `Timestamp` field on each benchmark — each run carries its own timestamp instead of an indirect tag→timestamp map
- **Benchmark History**: Added `History []HistoryEntry` to track old tag+timestamp pairs from previous merges; latest tag stays on benchmark itself, history holds only older tags
- **Merge Internals**: Removed `benchGroup` struct and `taggedEntry` wrapper — replaced with `map[string]map[string]*Benchmark` for simpler tag dedup and insertion

### Removed

- `Runtimes` field from `Benchmark` struct (replaced by `Timestamp` + `History`)
- `latestRuntime` and `mergeRuntimes` helper functions (no longer needed)
- `benchGroup` and `taggedEntry` internal types

# [0.9.4] - 2026-05-15

### Added

- **Merge Dedup**: Benchmarks sharing same name and tag now deduplicated by latest runtime timestamp instead of merging both data sets ([#76](https://github.com/goptics/vizb/pull/76))
- **Chronological Tag Ordering**: Tags processed in chronological order during inner merge for deterministic data ordering ([#76](https://github.com/goptics/vizb/pull/76))
- **Latest Tag Preserved**: Merged output retains latest tag (by runtime timestamp) instead of clearing it ([#76](https://github.com/goptics/vizb/pull/76))

### Changed

- **Merge Internals**: Rewrote `MergeBenchmarks` with two-level map (`Name → Tag → Benchmark`) for cleaner dedup and deterministic output (removed 4 unused helpers) ([#76](https://github.com/goptics/vizb/pull/76))
- **Scale Selector Label**: Renamed "Y-Axis Scale" to "Data Scale" — scale applies to all chart data, not just Y-axis ([#77](https://github.com/goptics/vizb/pull/77))
- **Scale Selector Visibility**: Removed Y-axis data gating — scale selector now always visible for non-pie charts ([#77](https://github.com/goptics/vizb/pull/77))

### Fix

- Prevent tag re-injection into previously-merged benchmark data on incremental merges ([#73](https://github.com/goptics/vizb/pull/73))
- Trim CPU name whitespace when extracting from benchmark input

# [0.9.2] - 2026-05-14

### Fix

- Fix merge command to accept JSON files containing arrays of benchmarks ([#72](https://github.com/goptics/vizb/pull/72))

### Refactor

- Extract helpers from merge function, add typed `Dimension` constants for tag-axis values

# [0.9.1] - 2026-05-14

### Fix

- Fix json format as merged subcommand output ([#71](https://github.com/goptics/vizb/pull/71))

# [0.9.0] - 2026-05-14

### Added

- **Tag-Based Merging**: Benchmarks with same name but different tags deep-merged via `merge` subcommand — enables historical comparison across commits, releases, or environment variants ([#69](https://github.com/goptics/vizb/pull/69)).
- **`--tag -t` Flag**: New root command flag to assign label (e.g., commit hash, version number) to benchmark run. Auto-populates `runtimes` map with UTC timestamp.
- **`--tag-axis -A` Flag**: New merge subcommand flag controlling which data dimension receives tag annotation. Accepts `n` (name), `x` (xAxis), or `y` (yAxis). Defaults to `n`.
- **Runtimes Tracking**: `Tag` and `Runtimes` fields added to `Benchmark` struct for tracking provenance and timestamps.

### Breaking

- **Unit Shorthand Flags**: Shorthand flags for time (`-t`→`-T`), memory (`-m`→`-M`), and number (`-n`→`-N`) now capitalized.

# [0.8.0] - 2026-04-25

### Added

- **Logarithmic Scale**: Added `--scale` flag (`linear|log`) for bar and line charts — better visualization of benchmarks with high variance in values ([#66](https://github.com/goptics/vizb/pull/66)).
- **URL Routing for Scale**: Log scale synced with `sc` query parameter for shareable URLs.

# [0.7.1] - 2025-12-28

### Changed

- **Formatter**: Optimized `RoundToTwo` formatter and updated tests ([#62](https://github.com/goptics/vizb/pull/62)).
- **Documentation**: Updated examples with latest commands.
- **Internal**: Use signed tags for releases to ensure verification ([#61](https://github.com/goptics/vizb/pull/61)).

### Fixed

- **UI**: Resolved favicon xmlns attr issue ([#60](https://github.com/goptics/vizb/pull/60)).

# [0.7.0] - 2025-12-14

### Added

- **Filter Flag**: Added `-f/--filter` flag to filter benchmarks using regex ([#58](https://github.com/goptics/vizb/pull/58)).
- **Format Inference**: Output format auto-inferred from file extension (e.g., `.json` for JSON, others for HTML) ([#58](https://github.com/goptics/vizb/pull/58)).
- **URL Routing**: Lightweight URL router syncs UI state (sort, labels, chart type, selection) with query parameters ([#55](https://github.com/goptics/vizb/pull/55)).
- **Dynamic Title**: Browser tab title dynamically updates to show active benchmark name ([#56](https://github.com/goptics/vizb/pull/56)).

### Changed

- **Validation**: Enhanced validation logic, improved warning messages with specific failure reasons ([#57](https://github.com/goptics/vizb/pull/57)).
- **UI Architecture**: Refactored settings management to single reactive state object for consistency ([#53](https://github.com/goptics/vizb/pull/53)).
- **Internal**: Introduced generic `sortBy` function, removed unused dependencies ([#52](https://github.com/goptics/vizb/pull/52), [#54](https://github.com/goptics/vizb/pull/54)).

### Breaking Changes

- **Format Flag Removed**: `--format` flag removed in favor of automatic inference from output filename.
- **Short Flag Reassigned**: `-f` shorthand now used for `--filter` instead of `--format`.

# [0.6.0] - 2025-12-03

### Added

- **Standard Benchmark JSON Support**: Support for standard benchmark JSON files as input ([#47](https://github.com/goptics/vizb/pull/47)).
- **Axis Swapping**: Dynamic reordering of benchmark dimensions (name, xAxis, yAxis) with persistent state management ([#46](https://github.com/goptics/vizb/pull/46)).
- **Advanced Grouping**: Benchmark grouping via regex patterns ([#40](https://github.com/goptics/vizb/pull/40)).
- **Benchmark Environment**: Added benchmark environment details to results ([#39](https://github.com/goptics/vizb/pull/39)).
- **Metric Extraction**: Support for extracting and displaying benchmark iterations, throughput, and custom metrics ([#35](https://github.com/goptics/vizb/pull/35)).
- Added "Package Source" button to dashboard, updated docs with comprehensive examples.

### Changed

- **JSON Optimization**: Reduced JSON output size 30% by removing redundant `unit` and `per` properties, using `omitempty` ([#43](https://github.com/goptics/vizb/pull/43), [#45](https://github.com/goptics/vizb/pull/45)).
- **UI Performance**: Refactored UI components to reduce bundle size, improved dashboard layout ([#38](https://github.com/goptics/vizb/pull/38)).
- **CI/CD**: Updated CI runners to `slim` versions, improved GoReleaser config for optimized builds ([#37](https://github.com/goptics/vizb/pull/37), e673195).
- **Documentation**: Updated docs for upcoming release ([#50](https://github.com/goptics/vizb/pull/50)).
- **Build**: Updated favicon inline build script, added `format` task to Taskfile.yml.

### Fixed

- **Label Rotation**: Fixed x-axis label rotation based on character length ([#48](https://github.com/goptics/vizb/pull/48)).
- **Group Selection**: Fixed group selection state not preserved when switching benchmarks ([#41](https://github.com/goptics/vizb/pull/41), [#42](https://github.com/goptics/vizb/pull/42)).
- **Flag Validation**: Fixed flag validation rules and updated tests ([#36](https://github.com/goptics/vizb/pull/36)).

# [0.5.0] - 2025-11-23

### Added

- Completely new UI built with **Vue.js** — enhanced interactivity and modern design.
- **Dark and Light Mode** support.
- **Merge subcommand** for combining multiple benchmark JSON files into one report.
- **New CLI flags** for advanced sorting and label control.
- Introduced chart types **Bar, Line, and Pie** for different metrics.

### Changed

- Refactored UI embedding using vite in UI and html template in cli.
- CLI output visually enhanced with consistent styles for info, success, error, warning messages.
- Documentation expanded with improved examples and advanced group patterns.

### Fixed

- Progress bar finish line issue resolved in benchmark visualization.
- Improved chart rendering on high-DPI displays; images now save with correct chart titles.
- Test failures after migration resolved; noisy terminal output suppressed during tests.

### Breaking Changes

- Renamed `--alloc-unit` flag to `--number-unit` (`-u`).
- Updated pattern matching logic to use `x` and `y` axis terminology instead of `workload` and `subject`.

## [0.4.0] - 2025-09-18

### Added

- Advanced pattern-based grouping for extracting groups from benchmark names
- Skip functionality for selective benchmark processing
- Support for raw benchmark data processing
- Comprehensive flag validation with enhanced error handling
- Dynamic color assignment for chart subjects — consistent coloring across groups
- Enhanced temp file management
- Extensive test coverage for previously untested functions (temp file creation, stdin processing, output generation)

### Changed

- Enhanced flag validation to require subject parameter
- Improved chart template UX with dynamic legend resizing based on subject count
- Updated docs with advanced grouping examples
- Refined benchmark progress real-time logic
- Streamlined error handling and file ops with shared utility functions
- Updated README with shorthand patterns to reduce table width

### Fixed

- Hide CPU number display when value is 0
- Improved error handling in benchmark result parsing for group parsing
- Fixed inconsistent color assignment when groups have different subject list lengths

### Breaking changes

`--separator` flag replaced with `--group-pattern` for more flexibility.

## [0.3.2] - 2025-09-13

### Fixed

- Added dynamic versioning using debug module

## [0.3.1] - 2025-09-11

### Fixed

- Group bench name order issue resolved
- Fixed grouping name when split words len < 1

## [0.3.0] - 2025-06-21

### Added

- Improved docs for `--format` flag in README
- Comprehensive test coverage for previously untested functions
- Dedicated tests for temp file creation, stdin processing, output generation
- Example for JSON output format in docs

### Changed

- Enhanced code testability with mock-friendly function variables
- Optimized sidebar display logic — only show with sufficient chart count
- Standardized temp file naming with consistent prefixes

### Fixed

- Fixed sidebar display in HTML output with better conditional rendering
- Fixed camelCase variable naming in chart template for consistency

## [0.2.0] - 2025-06-20

### Added

- Benchmark indicator for easy navigation
- Allocation unit conversion support for benchmark charts
- Support for units lowercase input
- Enhanced grouping by adding bench name
- CPU count suffix added to headline

### Changed

- Adjusted chart bottom margin for better visualization
- Optimized benchmark parsing logic in chart.go with cleaner control flow
- Organized post-generation logs
- Renamed license file, added attribution request
- Updated docs with bench group feature info
- Added Vizb attribution in footer

### Fixed

- Resolved extra line issue after pipe progress completed
- Fixed test count display on pipe, changed status
- Corrected memory unit conversion from bytes to bits
- Removed build script — no longer needed

## [0.1.1] - 2025-06-19

### Fixed

- Enabled sorting of workloads in createChart function

## [0.1.0] - 2025-06-17

### Added

- Initial release of Vizb — Go Benchmark Visualization Tool
- Interactive HTML charts from Go benchmark results
- Multiple metric support:
  - Execution time visualization
  - Memory usage visualization
  - Allocation counts visualization
- Customizable units:
  - Time: ns, us, ms, s
  - Memory: B, KB, MB, GB
- Customizable chart titles and descriptions
- Responsive design for any device
- Export charts as PNG images
- Simple CLI interface with helpful flags:
  - Custom output file name
  - Custom chart name and description
  - Custom units for time and memory
  - Custom separator for benchmark grouping
- Piped input support — process benchmark data from stdin
- Benchmark grouping based on separator character (default: "/")
- Organized visualization with workload and subject grouping
- Development workflow using Task runner
