# Changelog

All notable changes to the Vizb project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
