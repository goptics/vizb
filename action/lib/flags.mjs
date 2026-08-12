/**
 * Build vizb CLI argv fragments from action inputs.
 */

/**
 * Append -c / --chart / --stat flags (matches prior bash append_chart_stat_flags).
 * @param {string[]} args
 * @param {string} charts
 * @param {string} chartInput multiline chart overrides
 * @param {string} statVal
 * @returns {string[]}
 */
export function appendChartStatFlags(args, charts, chartInput, statVal) {
  if (charts) {
    args.push('-c', charts)
  }

  for (const rawLine of (chartInput || '').split(/\r?\n/)) {
    // Leading-only trim, matching bash ${line#"${line%%[![:space:]]*}"}
    const line = rawLine.replace(/^[ \t]+/, '')
    if (!line) continue
    if (line.startsWith('#')) continue
    args.push('--chart', line)
  }

  if (statVal === 'all' || statVal === 'true') {
    args.push('--stat')
  } else if (statVal) {
    args.push(`--stat=${statVal}`)
  }

  return args
}

/**
 * Flags for `vizb <input> -o <json>` convert path.
 * @param {import('./inputs.mjs').ResolvedInputs} inputs
 * @returns {string[]}
 */
export function buildConvertArgs(inputs) {
  /** @type {string[]} */
  const args = []

  if (inputs.tag) args.push('--tag', inputs.tag)
  if (inputs.id) args.push('--id', inputs.id)
  if (inputs.name) args.push('-n', inputs.name)
  if (inputs.title) args.push('--title', inputs.title)
  if (inputs.description) args.push('-d', inputs.description)
  if (inputs.group) args.push('-g', inputs.group)
  if (inputs.groupPattern) args.push('-p', inputs.groupPattern)
  if (inputs.groupRegex) args.push('-r', inputs.groupRegex)
  if (inputs.sort) args.push('--sort', inputs.sort)
  if (inputs.filter) args.push('--filter', inputs.filter)
  if (inputs.memUnit) args.push('-M', inputs.memUnit)
  if (inputs.timeUnit) args.push('-T', inputs.timeUnit)
  if (inputs.numberUnit) args.push('-N', inputs.numberUnit)
  if (inputs.round) args.push('--round')
  if (inputs.select) args.push('--select', inputs.select)
  if (inputs.colAxis) args.push('--col-axis', inputs.colAxis)
  if (inputs.jsonPath) args.push('--json-path', inputs.jsonPath)
  if (inputs.showLabels) args.push('-l')
  if (inputs.parser) args.push('-P', inputs.parser)

  appendChartStatFlags(args, inputs.charts, inputs.chart, inputs.stat)
  return args
}

/**
 * Flags for `vizb ui`.
 * @param {import('./inputs.mjs').ResolvedInputs} inputs
 * @returns {string[]}
 */
export function buildUiArgs(inputs) {
  /** @type {string[]} */
  const args = []
  appendChartStatFlags(args, inputs.charts, inputs.chart, inputs.stat)
  if (inputs.enable3d) args.push('--3d')
  return args
}
