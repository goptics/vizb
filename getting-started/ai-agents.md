---
title: "AI agents"
description: "Connect vizb to coding agents with npx skills add so /vizb can chart tables and benchmarks."
---

Coding agents do not know vizb unless you install the skill. The skill teaches the agent when to use vizb, how to install the binary if it is missing, and which docs page to fetch. The agent still **runs the vizb CLI** — this is not an MCP server.

## Install the skill

```bash
# this project
npx skills add goptics/skills --skill vizb

# every project on this machine
npx skills add goptics/skills --skill vizb -g

# confirm the CLI can see it
npx skills add goptics/skills --list
```

The [skills CLI](https://github.com/vercel-labs/skills) clones the small [goptics/skills](https://github.com/goptics/skills) pack (not this repo) and copies `skills/vizb` into the agents you have installed (Grok, Claude Code, Cursor, Codex, and others). After that, `/vizb` is available. Automatic invocation also stays on: “chart this CSV” or “plot the Go bench” should load the skill without typing `/vizb`.

Skill install is not binary install. If `vizb` is missing, the skill runs the [install](/getting-started/install) one-liners.

> Agents fetch extra guidance from [`/llms.txt`](/llms.txt) and per-page `.md` URLs such as [`/guides/group-vs-select.md`](/guides/group-vs-select.md). Those are generated from this site at build time.

## Use it

```text
/vizb path/to/data.csv
```

Or ask in natural language: “turn this benchmark output into an HTML chart.”

The agent should start with `vizb <chart> <file> -o out.html` and only add grouping or select flags if auto output is wrong.
