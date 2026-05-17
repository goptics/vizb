# GitHub Action: Binary Download via curl

## Problem

Two edge cases with the current composite GitHub Action (`action.yml`):

1. **Windows support** — The action relies on `go install`, bash arrays, `find`, and `/dev/null`,
   which have subtle compatibility issues on Windows runners despite Git Bash.

2. **`@latest` caching** — When users reference `goptics/vizb@latest`, the cache key includes
   the literal string "latest", so it never invalidates when new versions are published.

## Solution

Replace `go install` with downloading pre-built binaries from GitHub Releases via `curl`.
GoReleaser already publishes binaries for linux/darwin/windows on amd64/arm64.

## Design

### Version Resolution

| User reference | `action_ref` | Resolution |
|---|---|---|
| `goptics/vizb@v0.9.4` | `v0.9.4` | Use directly |
| `goptics/vizb@latest` | `latest` | Query GitHub redirect to resolve tag |
| `goptics/vizb@main` | `main` | Error — must use version tag or `latest` |

Latest resolution uses curl redirect:
```bash
TAG=$(curl -sL -o /dev/null -w '%{url_effective}' \
  https://github.com/goptics/vizb/releases/latest | sed 's|.*/tag/||')
```

### Cache Policy

- **Pinned versions:** Cache with key `vizb-{tag}-{os}-{arch}`
- **`@latest`:** Skip cache entirely (always fresh install)

### Binary Download

OS/Arch mapping from runner context:

| Runner variable | goreleaser value |
|---|---|
| `runner.os` = Linux | `linux` |
| `runner.os` = macOS | `darwin` |
| `runner.os` = Windows | `windows` |
| `runner.arch` = X64 | `amd64` |
| `runner.arch` = ARM64 | `arm64` |

Download URL format:
```
https://github.com/goptics/vizb/releases/download/{tag}/vizb@{version}-{os}-{arch}.{ext}
```

Where `{version}` = `{tag}` with leading `v` stripped, `{ext}` = `.tar.gz` (linux/darwin) or `.zip` (windows).

### Extraction

- Linux/macOS: `tar xzf`
- Windows: `unzip` (available in Git Bash on GitHub Windows runners)

No PowerShell calls — pure bash, avoiding cross-shell path issues.

### Binary Placement

Binary installed to `~/.local/bin/` (added to `$GITHUB_PATH`).
No longer depends on Go toolchain or `go env GOPATH`.

### Merge Step Simplification

The `find` loop for `merge-dir` is removed. The `vizb merge` command already
accepts directory paths natively (`collectJSONFiles` in `cmd/merge.go` handles
dirs via `filepath.Glob`). The action now passes `merge-dir` directly to `vizb merge`.

This also eliminates the Windows `find` compatibility concern.

### Error Handling

- Invalid `action_ref` (branch name): verify release exists via HTTP status, error with guidance
- `@latest` with no releases: detect empty tag resolution, error
- Download failure: `curl -f` returns non-zero on HTTP errors

## Changes Summary

| Area | Before | After |
|---|---|---|
| Install method | `go install` (compiles) | `curl` download pre-built binary |
| Go toolchain required | Yes | No |
| Install time | ~10-30s | ~2-3s |
| Windows support | Git Bash quirks | Native binary, `unzip` |
| `@latest` caching | Stale | Always fresh |
| Merge step | `find` + loop | Pass dirs directly |
| Binary path | `$(go env GOPATH)/bin/` | `~/.local/bin/` |

## Action Steps (final)

1. Resolve version (map `action_ref` to release tag, resolve `@latest` via API)
2. Cache binary (by tag+os+arch, skip for `@latest`)
3. Download & extract (curl + tar/unzip, only on cache miss)
4. Resolve input (unchanged)
5. Convert to JSON (unchanged)
6. Merge (simplified — no `find`, pass dirs to `vizb merge`)
7. Generate output (unchanged)
