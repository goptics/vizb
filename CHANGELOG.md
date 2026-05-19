# Changelog

Notable changes to Vizb documented here.

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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