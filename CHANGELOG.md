# Changelog

All notable changes to the Vizb project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2025-09-18

### Added

- Advanced pattern-based grouping system for extracting groups from benchmark names
- Skip functionality for selective benchmark processing during data analysis
- Support for raw benchmark data processing
- Comprehensive flag validation rules with enhanced error handling
- Dynamic color assignment for chart subjects to ensure consistent coloring across benchmark groups
- Enhanced temporary file management system
- Extensive test coverage for previously untested functions including temp file creation, stdin processing, and output generation

### Changed

- Enhanced flag validation to require subject parameter specification
- Improved chart template UX with dynamic legend section resizing based on subject numbers
- Updated documentation with advanced grouping examples and usage patterns
- Refined benchmark progress real-time logic
- Streamlined error handling and file operations with shared utility functions
- Updated README with shorthand patterns to reduce table width

### Fixed

- Improved chart template UX by hiding CPU number display when value is 0
- Enhanced error handling in benchmark result parsing for group parsing
- Resolved inconsistent color assignment issue when benchmark groups have different subject list lengths

### Breaking changes

In this release the `--separator` flag is been replaced with `--group-pattern` to brings more flexibility.

## [0.3.2] - 2025-09-13

### Fixed

- Added dynamic versioning using debug module

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
