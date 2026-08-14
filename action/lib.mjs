import { spawnSync } from 'node:child_process'
import { closeSync, existsSync, openSync } from 'node:fs'
import { homedir, tmpdir } from 'node:os'
import { join } from 'node:path'
import { setTimeout } from 'node:timers/promises'

/** @param {NodeJS.ProcessEnv} env */
function parseActionInputs(env) {
  const raw = env.VIZB_ACTION_INPUTS
  if (!raw) return {}
  try {
    return JSON.parse(raw)
  } catch {
    throw new Error('VIZB_ACTION_INPUTS is not valid JSON')
  }
}

/**
 * @param {NodeJS.ProcessEnv} [env]
 */
export function resolveInputs(env = process.env) {
  const blob = parseActionInputs(env)
  const input = (key) => String(blob[key] ?? '')

  const file = input('file') || input('bench-file')
  const cmd = input('cmd') || input('bench-cmd')
  const mergeFiles = input('merge-files')
  const mergeDir = input('merge-dir')
  const dataUrl = input('data-url')
  const outputJson = input('output-json')

  if (!file && !cmd && !mergeFiles && !mergeDir && !dataUrl) {
    throw new Error(
      'No input provided. Specify cmd, file, merge-files, merge-dir, or data-url.',
    )
  }

  return {
    file,
    cmd,
    cmdRetries: parseCmdRetries(input('cmd-retries')),
    hasInput: Boolean(file || cmd),
    jsonFile: outputJson || 'bench.json',
    name: input('name'),
    title: input('title'),
    description: input('description'),
    tag: input('tag'),
    id: input('id'),
    group: input('group'),
    groupPattern: input('group-pattern'),
    groupRegex: input('group-regex'),
    sort: input('sort'),
    filter: input('filter'),
    memUnit: input('mem-unit'),
    timeUnit: input('time-unit'),
    numberUnit: input('number-unit'),
    round: input('round') === 'true',
    select: input('select'),
    colAxis: input('col-axis'),
    jsonPath: input('json-path'),
    stat: input('stat'),
    chart: input('chart'),
    charts: input('charts'),
    parser: input('parser'),
    showLabels: input('show-labels') === 'true',
    enable3d: input('enable-3d') === 'true',
    mergeFiles,
    mergeDir,
    tagAxis: input('tag-axis') || 'n',
    outputHtml: input('output-html'),
    dataUrl,
  }
}

/** @param {string} value */
function parseCmdRetries(value) {
  if (!value) return 1
  const n = Number(value)
  if (!Number.isInteger(n) || n < 1) {
    throw new Error('cmd-retries must be an integer >= 1')
  }
  return n
}

const CMD_RETRY_DELAY_SEC = 2

/**
 * @param {() => void} fn
 * @param {{
 *   retries: number,
 *   sleep?: (sec: number) => void | Promise<void>,
 *   log?: (msg: string) => void,
 * }} opts
 */
export async function runWithRetries(fn, opts) {
  const wait = opts.sleep ?? ((sec) => setTimeout(sec * 1000))
  const log = opts.log ?? ((msg) => console.error(msg))

  for (let attempt = 1; attempt <= opts.retries; attempt++) {
    try {
      fn()
      return
    } catch (err) {
      if (attempt === opts.retries) throw err
      const waitSec = CMD_RETRY_DELAY_SEC * 2 ** (attempt - 1)
      log(
        `::warning::cmd failed (attempt ${attempt}/${opts.retries}, ${err.message ?? err}). Retrying in ${waitSec}s.`,
      )
      await wait(waitSec)
    }
  }
}

/** @param {string} value */
export function splitMergeFiles(value) {
  if (!value) return []
  return value.split(/\s+/).filter(Boolean)
}

/**
 * @param {string[]} args
 * @param {string} charts
 * @param {string} chartInput
 * @param {string} statVal
 */
export function appendChartStatFlags(args, charts, chartInput, statVal) {
  if (charts) args.push('-c', charts)

  for (const rawLine of (chartInput || '').split(/\r?\n/)) {
    const line = rawLine.replace(/^[ \t]+/, '')
    if (!line || line.startsWith('#')) continue
    args.push('--chart', line)
  }

  if (statVal === 'all' || statVal === 'true') args.push('--stat')
  else if (statVal) args.push(`--stat=${statVal}`)

  return args
}

/** @param {ReturnType<typeof resolveInputs>} inputs */
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
  return appendChartStatFlags(args, inputs.charts, inputs.chart, inputs.stat)
}

/**
 * Prefer absolute path: Node spawnSync on Windows only finds PATHEXT names.
 * @param {{ platform?: string, homedir?: () => string, existsSync?: (p: string) => boolean }} [opts]
 */
export function resolveVizbBin(opts = {}) {
  const platform = opts.platform ?? process.platform
  const home = (opts.homedir ?? homedir)()
  const exists = opts.existsSync ?? existsSync
  const name = platform === 'win32' ? 'vizb.exe' : 'vizb'
  const candidate = join(home, '.local', 'bin', name)
  return exists(candidate) ? candidate : name
}

/** @param {import('node:child_process').SpawnSyncReturns<string>} result @param {string} label */
function throwIfFailed(result, label) {
  if (result.error) throw result.error
  if (result.status !== 0) {
    const code = result.status ?? 1
    const err = new Error(`${label} exited with code ${code}`)
    // @ts-expect-error attach exit code
    err.exitCode = code
    throw err
  }
}

/** @param {string} command @param {string[]} args */
export function run(command, args) {
  throwIfFailed(
    spawnSync(command, args, { stdio: 'inherit', encoding: 'utf8' }),
    command,
  )
}

/** @param {string} command @param {string} stdoutPath */
export function runShellCapture(command, stdoutPath) {
  const fd = openSync(stdoutPath, 'w')
  try {
    throwIfFailed(
      spawnSync(command, {
        shell: true,
        stdio: ['ignore', fd, 'inherit'],
        encoding: 'utf8',
      }),
      command,
    )
  } finally {
    closeSync(fd)
  }
}

/**
 * @param {ReturnType<typeof resolveInputs>} inputs
 * @param {{
 *   run?: typeof run,
 *   runShellCapture?: typeof runShellCapture,
 *   vizbBin?: string,
 *   sleep?: (sec: number) => void | Promise<void>,
 *   log?: (msg: string) => void,
 * }} [deps]
 */
export async function runPipeline(inputs, deps = {}) {
  const execRun = deps.run ?? run
  const execShell = deps.runShellCapture ?? runShellCapture
  const vizb = deps.vizbBin ?? resolveVizbBin()

  if (inputs.hasInput) {
    let inputPath = inputs.file
    if (!inputPath) {
      inputPath = join(tmpdir(), `vizb-data-input-${process.pid}.txt`)
      await runWithRetries(() => execShell(inputs.cmd, inputPath), {
        retries: inputs.cmdRetries,
        sleep: deps.sleep,
        log: deps.log,
      })
    }
    execRun(vizb, [inputPath, '-o', inputs.jsonFile, ...buildConvertArgs(inputs)])
  }

  if (inputs.mergeFiles || inputs.mergeDir) {
    /** @type {string[]} */
    const files = []
    if (existsSync(inputs.jsonFile)) files.push(inputs.jsonFile)
    files.push(...splitMergeFiles(inputs.mergeFiles))
    if (inputs.mergeDir) files.push(inputs.mergeDir)
    execRun(vizb, ['merge', ...files, '-o', inputs.jsonFile, '--tag-axis', inputs.tagAxis])
  }

  if (inputs.outputHtml) {
    const uiFlags = appendChartStatFlags([], inputs.charts, inputs.chart, inputs.stat)
    if (inputs.enable3d) uiFlags.push('--3d')
    if (inputs.dataUrl) {
      execRun(vizb, ['ui', '-U', inputs.dataUrl, '-o', inputs.outputHtml, ...uiFlags])
    } else {
      execRun(vizb, ['ui', inputs.jsonFile, '-o', inputs.outputHtml, ...uiFlags])
    }
  }
}
