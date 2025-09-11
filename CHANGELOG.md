# Changelog

All notable changes to the Vizb project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.1] - 2025-09-11

### Fixed

- Group bench name order issue resolved
- Resolved grouping name when the split words len less then one

## [0.3.0] - 2025-06-21

### Added

- Added improved documentation for `--format` flag in README
- Added comprehensive test coverage for previously untested functions
- Added dedicated tests for temp file creation, stdin processing, and output generation
- Added example for JSON output format in documentation

### Changed

- Enhanced code testability with mock-friendly function variables
- Optimized sidebar display logic in chart templates to only show with sufficient chart count
- Standardized temporary file naming with consistent prefixes

### Fixed

- Fixed sidebar display in HTML output with better conditional rendering
- Fixed camelCase variable naming in chart template for consistency

## [0.2.0] - 2025-06-20

### Added

- Added benchmark indicator for easy navigation
- Added allocation unit conversion support for benchmark charts
- Added support for units lowercase input
- Enhanced grouping by adding bench name
- Added CPU count suffix to the headline

### Changed

- Adjusted chart bottom margin for better visualization
- Optimized benchmark parsing logic in chart.go with cleaner control flow
- Organized the post generation logs
- Renamed license file and added attribution request
- Updated documentation with bench group feature information
- Added Vizb attribution in footer

### Fixed

- Resolved extra line issue after pipe progress completed
- Fixed test count display on pipe and changed the status
- Corrected memory unit conversion from bytes to bits
- Removed build script as it is no longer needed

## [0.1.1] - 2025-06-19

### Fixed

- Enabled sorting of workloads in createChart function

## [0.1.0] - 2025-06-17

### Added

- Initial release of Vizb - Go Benchmark Visualization Tool
- Interactive HTML charts generation from Go benchmark results
- Support for multiple metrics:
  - Execution time visualization
  - Memory usage visualization
  - Allocation counts visualization
- Customizable units for metrics:
  - Time: ns, us, ms, s
  - Memory: B, KB, MB, GB
- Customizable chart titles and descriptions
- Responsive design for charts that work on any device
- Export capability to save charts as PNG images
- Simple CLI interface with helpful flags:
  - Custom output file name
  - Custom chart name and description
  - Custom units for time and memory
  - Custom separator for benchmark grouping
- Support for piped input to process benchmark data directly from stdin
- Benchmark grouping based on separator character (default: "/")
- Organized visualization with workload and subject grouping
- Development workflow using Task runner
