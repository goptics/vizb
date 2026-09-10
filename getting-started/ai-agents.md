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

Type `/vizb` plus the file and what you want in natural language. The agent picks a chart, writes an HTML file, and returns the path — it does not paste the HTML into the chat.

```text
/vizb sales.csv as bar by region & product, show labels
```

Same conversion as the CLI. Every example in these docs has an **Agent** tab with a prompt like this:

### Agent

```text
/vizb sales.csv as bar by region & product, show labels
```

### CLI

```bash
vizb bar sales.csv -g region,product -p x,y -l -o sales.html
```

### HTTP

```json
{
  "input": "order_date,region,category,product,quantity,amount\n2024-01-01,Central,Hardware,Connector,16,2488.24\n2024-01-02,East,Tools,Gear,42,207.08\n2024-01-03,North,Mechanical,Sensor,20,3465.22\n2024-01-04,West,Electronics,Valve,24,7633.44\n2024-03-12,South,Industrial,Widget,31,5088.10\n2024-06-18,East,Electronics,Relay,12,2214.50\n2024-09-05,North,Hardware,Bolt,28,1724.27\n2024-12-20,West,Tools,Gadget,19,2938.82\n2025-02-08,Central,Mechanical,Valve,27,7350.26\n2025-04-14,South,Electronics,Widget,41,3278.77\n2025-07-22,East,Industrial,Connector,15,5093.42\n2025-10-03,North,Tools,Gear,33,4102.15\n2025-11-19,West,Hardware,Sensor,22,2890.40\n2025-12-28,South,Mechanical,Gadget,26,611.32\n",
  "parser": "csv",
  "grouping": {
    "columns": [
      "region",
      "product"
    ],
    "pattern": "x,y"
  },
  "charts": {
    "types": [
      "bar"
    ],
    "configs": [
      {
        "type": "bar",
        "showLabels": true
      }
    ]
  },
  "output": {
    "format": "html"
  }
}
```

### Action

```yaml
- uses: goptics/vizb@v0
  with:
    file: sales.csv
    group: region,product
    group-pattern: x,y
    charts: bar
    chart: "bar:labels"
    output-html: sales.html
```

Other prompts:

```text
/vizb sales.csv as pie by region
/vizb this benchmark as a line chart
/vizb bench.txt as bar
/vizb path/to/data.csv as bar
```

Or ask without the slash: “turn this benchmark output into an HTML chart.”

The agent should start with `vizb <chart> <file> -o out.html` and only add grouping or select flags if auto output is wrong.
