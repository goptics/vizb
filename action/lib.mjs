import { spawnSync } from 'node:child_process'
import { closeSync, existsSync, openSync } from 'node:fs'
import { homedir, tmpdir } from 'node:os'
import { join } from 'node:path'

/** @param {string} key @param {NodeJS.ProcessEnv} [env] */
function input(key, env = process.env) {
  return env[`INPUT_${key}`] ?? ''
}

/**
 * @param {NodeJS.ProcessEnv} [env]
 */
export function resolveInputs(env = process.env) {
  const file = input('file', env) || input('bench-file', env)
  const cmd = input('cmd', env) || input('bench-cmd', env)
  const mergeFiles = input('merge-files', env)
  const mergeDir = input('merge-dir', env)
  const dataUrl = input('data-url', env)
  const outputJson = input('output-json', env)

  if (!file && !cmd && !mergeFiles && !mergeDir && !dataUrl) {
    throw new Error(
      'No input provided. Specify cmd, file, merge-files, merge-dir, or data-url.',
    )
  }

  return {
    file,
    cmd,
    cmdRetries: parseCmdRetries(input('cmd-retries', env)),
    hasInput: Boolean(file || cmd),
    jsonFile: outputJson || 'bench.json',
    name: input('name', env),
    title: input('title', env),
    description: input('description', env),
    tag: input('tag', env),
    id: input('id', env),
    group: input('group', env),
    groupPattern: input('group-pattern', env),
    groupRegex: input('group-regex', env),
    sort: input('sort', env),
    filter: input('filter', env),
    memUnit: input('mem-unit', env),
    timeUnit: input('time-unit', env),
    numberUnit: input('number-unit', env),
    round: input('round', env) === 'true',
    select: input('select', env),
    colAxis: input('col-axis', env),
    jsonPath: input('json-path', env),
    stat: input('stat', env),
    chart: input('chart', env),
    charts: input('charts', env),
    parser: input('parser', env),
    showLabels: input('show-labels', env) === 'true',
    enable3d: input('enable-3d', env) === 'true',
    mergeFiles,
    mergeDir,
    tagAxis: input('tag-axis', env) || 'n',
    outputHtml: input('output-html', env),
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

/** @param {number} seconds */
function sleepSync(seconds) {
  if (!(seconds > 0)) return
  Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, seconds * 1000)
}

/**
 * @param {() => void} fn
 * @param {{
 *   retries: number,
 *   sleep?: (sec: number) => void,
 *   log?: (msg: string) => void,
 * }} opts
 */
export function runWithRetries(fn, opts) {
  const sleep = opts.sleep ?? sleepSync
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
      sleep(waitSec)
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
 *   sleep?: (sec: number) => void,
 *   log?: (msg: string) => void,
 * }} [deps]
 */
export function runPipeline(inputs, deps = {}) {
  const execRun = deps.run ?? run
  const execShell = deps.runShellCapture ?? runShellCapture
  const vizb = deps.vizbBin ?? resolveVizbBin()

  if (inputs.hasInput) {
    let inputPath = inputs.file
    if (!inputPath) {
      inputPath = join(tmpdir(), `vizb-data-input-${process.pid}.txt`)
      runWithRetries(() => execShell(inputs.cmd, inputPath), {
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
