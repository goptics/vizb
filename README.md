<div align="center">
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./assets/logo-dark.gif">
  
  <source media="(prefers-color-scheme: light)" srcset="./assets/logo-light.gif">
  
  <img alt="Vizb" width="100px" src="./assets/logo-light.gif">
</picture>

  <h1>Vizb</h1>

  <p>
    <a href="https://github.com/avelino/awesome-go?tab=readme-ov-file#benchmarks"><img src="https://awesome.re/mentioned-badge-flat.svg" alt="Mentioned in Awesome Go" /></a>
    <a href="https://vizb.goptics.org"><img src="https://img.shields.io/badge/Docs-00ADD8?style=for&logo=readthedocs" alt="Docs" /></a>
    <a href="https://vizb.goptics.org/examples"><img src="https://img.shields.io/badge/Live-Examples-orange?style=for" alt="Examples" /></a>
    <a href="https://goreportcard.com/report/github.com/goptics/vizb"><img src="https://goreportcard.com/badge/github.com/goptics/vizb" alt="Go Report Card" /></a>
    <a href="https://github.com/goptics/vizb/actions/workflows/cli.yml"><img src="https://github.com/goptics/vizb/actions/workflows/cli.yml/badge.svg" alt="CLI" /></a>
    <a href="https://github.com/goptics/vizb/actions/workflows/ui.yml"><img src="https://github.com/goptics/vizb/actions/workflows/ui.yml/badge.svg" alt="UI" /></a>
    <a href="https://codecov.io/gh/goptics/vizb"><img src="https://codecov.io/gh/goptics/vizb/branch/main/graph/badge.svg" alt="Codecov" /></a>
    <a href="https://github.com/goptics/vizb/releases"><img src="https://img.shields.io/github/downloads/goptics/vizb/total?color=green&label=downloads" alt="Downloads" /></a>
    <a href="https://golang.org/doc/devel/release.html"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for&logo=go" alt="Go Version" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=for" alt="License" /></a>
  </p>

  <p>
    A tabular visualization engine for <strong>CSV, JSON, and benchmark output</strong>. Turns numeric rows into interactive charts and descriptive statistics in one self-contained HTML file — no server, no dependencies, no build step.
  </p>

  <p>
    <a href="https://vizb.goptics.org/getting-started/">Getting Started</a> ·
    <a href="https://vizb.goptics.org/guides/parsers/">Parser Guide</a> ·
    <a href="https://vizb.goptics.org/guides/data/">Tabular Data</a> ·
    <a href="https://vizb.goptics.org/commands/root/">CLI Commands</a> ·
    <a href="https://vizb.goptics.org/guides/grouping/">Grouping</a> ·
    <a href="https://vizb.goptics.org/guides/merging/">Merging</a> ·
    <a href="https://vizb.goptics.org/ci-cd/github-action/">CI/CD</a>
    <br />
    <sub>Full documentation at <a href="https://vizb.goptics.org/"><strong>vizb.goptics.org</strong></a></sub>
  </p>
</div>

## Installation

### Quick Install

```bash
# Linux / macOS
curl -fsSL https://vizb.goptics.org/install.sh | bash

# Windows
irm https://vizb.goptics.org/install.ps1 | iex
```

### Go Toolchain

```bash
go install github.com/goptics/vizb@latest
```

### Download Binary

Pre-built binaries for Linux, macOS, and Windows are available on the [releases page](https://github.com/goptics/vizb/releases).

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for setup, build/test commands, and how to add a parser.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
