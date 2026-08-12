/**
 * Load and normalize GitHub Action inputs from environment variables.
 * Composite step must map each input as INPUT_<NAME> (see action.yml).
 */

/**
 * @param {NodeJS.ProcessEnv} [env]
 * @returns {Record<string, string>}
 */
export function readRawInputs(env = process.env) {
  const keys = [
    'cmd',
    'file',
    'bench-cmd',
    'bench-file',
    'name',
    'title',
    'description',
    'tag',
    'id',
    'group',
    'group-pattern',
    'group-regex',
    'sort',
    'filter',
    'mem-unit',
    'time-unit',
    'number-unit',
    'round',
    'select',
    'col-axis',
    'json-path',
    'stat',
    'chart',
    'charts',
    'parser',
    'show-labels',
    'enable-3d',
    'merge-files',
    'merge-dir',
    'tag-axis',
    'output-json',
    'output-html',
    'data-url',
    'vizb-binary',
  ]

  /** @type {Record<string, string>} */
  const raw = {}
  for (const key of keys) {
    raw[key] = env[`INPUT_${key}`] ?? ''
  }
  return raw
}

/**
 * @typedef {object} ResolvedInputs
 * @property {string} file
 * @property {string} cmd
 * @property {boolean} hasInput
 * @property {string} jsonFile
 * @property {string} name
 * @property {string} title
 * @property {string} description
 * @property {string} tag
 * @property {string} id
 * @property {string} group
 * @property {string} groupPattern
 * @property {string} groupRegex
 * @property {string} sort
 * @property {string} filter
 * @property {string} memUnit
 * @property {string} timeUnit
 * @property {string} numberUnit
 * @property {boolean} round
 * @property {string} select
 * @property {string} colAxis
 * @property {string} jsonPath
 * @property {string} stat
 * @property {string} chart
 * @property {string} charts
 * @property {string} parser
 * @property {boolean} showLabels
 * @property {boolean} enable3d
 * @property {string} mergeFiles
 * @property {string} mergeDir
 * @property {string} tagAxis
 * @property {string} outputJson
 * @property {string} outputHtml
 * @property {string} dataUrl
 */

/**
 * @param {Record<string, string>} raw
 * @returns {ResolvedInputs}
 */
export function resolveInputs(raw) {
  const file = raw.file || raw['bench-file'] || ''
  const cmd = raw.cmd || raw['bench-cmd'] || ''
  const mergeFiles = raw['merge-files'] || ''
  const mergeDir = raw['merge-dir'] || ''
  const dataUrl = raw['data-url'] || ''
  const outputJson = raw['output-json'] || ''
  const outputHtml = raw['output-html'] || ''

  const hasInput = Boolean(file || cmd)
  if (!hasInput && !mergeFiles && !mergeDir && !dataUrl) {
    throw new Error(
      'No input provided. Specify cmd, file, merge-files, merge-dir, or data-url.',
    )
  }

  return {
    file,
    cmd,
    hasInput,
    jsonFile: outputJson || 'bench.json',
    name: raw.name || '',
    title: raw.title || '',
    description: raw.description || '',
    tag: raw.tag || '',
    id: raw.id || '',
    group: raw.group || '',
    groupPattern: raw['group-pattern'] || '',
    groupRegex: raw['group-regex'] || '',
    sort: raw.sort || '',
    filter: raw.filter || '',
    memUnit: raw['mem-unit'] || '',
    timeUnit: raw['time-unit'] || '',
    numberUnit: raw['number-unit'] || '',
    round: raw.round === 'true',
    select: raw.select || '',
    colAxis: raw['col-axis'] || '',
    jsonPath: raw['json-path'] || '',
    stat: raw.stat || '',
    chart: raw.chart || '',
    charts: raw.charts || '',
    parser: raw.parser || '',
    showLabels: raw['show-labels'] === 'true',
    enable3d: raw['enable-3d'] === 'true',
    mergeFiles,
    mergeDir,
    tagAxis: raw['tag-axis'] || 'n',
    outputJson,
    outputHtml,
    dataUrl,
  }
}

/**
 * Split space-separated merge file paths (matches prior bash word-splitting).
 * @param {string} value
 * @returns {string[]}
 */
export function splitMergeFiles(value) {
  if (!value) return []
  return value.split(/\s+/).filter(Boolean)
}
