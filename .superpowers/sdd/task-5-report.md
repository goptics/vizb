# Task 5 Report: Documentation for `--select` axis mode

**Branch:** `feat/select-axis-mode`  
**BASE:** `3091693`  
**Status:** Complete

## Summary

Documented the `--select` axis mode feature across design spec, chart guides, grouping guide, root command reference, and examples. Docs build passes (`pnpm run build` in `docs/`).

## Deliverables

| Item | Path | Notes |
|------|------|-------|
| Design spec | `docs/superpowers/specs/2026-06-28-select-axis-mode.md` | Supersedes mixed-axis `--axes` design; force-added (gitignored `superpowers/`) |
| Bar chart docs | `docs/src/content/docs/charts/bar.mdx` | Three-role `--select` structure |
| Line chart docs | `docs/src/content/docs/charts/line.mdx` | Three-role `--select` structure |
| Scatter chart docs | `docs/src/content/docs/charts/scatter.mdx` | Three-role `--select` + mixed mode primary |
| Grouping guide | `docs/src/content/docs/guides/grouping.mdx` | Solo select axis mode + decision guide table |
| Root command | `docs/src/content/docs/commands/root.mdx` | Mode-at-a-glance matrix, repeatable `--select`, multi-dataset |
| Example CSV | `examples/csv/region-metrics.csv` | `region,latency,sales` mixed scatter demo |
| Examples README | `examples/csv/README.md` | `region-metrics.csv` section + quick reference row |

## Three-role `--select` structure (chart docs)

1. **Grouped + numeric select** — with `-g` / `-p` / `-r`: pick numeric stat columns
2. **Solo all-numeric value axes** — no grouping: 2–3 numeric cols → `x,y[,z]`
3. **Solo mixed category + value** — no grouping: category x + numeric y[,z]; scatter primary

## Mode-at-a-glance table

Present in:
- Design spec (`docs/superpowers/specs/2026-06-28-select-axis-mode.md`)
- Root command (`/commands/root` → `--select` mode matrix)
- Grouping guide (decision table: group+select vs solo select)

## `dataflags.go`

No change required — Usage string already documents repeatable `--select`, solo 2–3 col views, and grouped numeric select:

```
csv/json only: select columns (repeatable); solo mode: 2–3 cols per view as x,y[,z] axes (e.g. --select region,latency); grouped mode: numeric stat columns with optional {label}
```

## Verification

```bash
cd docs && pnpm install && pnpm run build
# ✓ 33 pages built successfully
```

## Commit

```
docs: document --select axis mode (solo/mixed/multi-view)
```