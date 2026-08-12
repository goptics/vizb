import assert from 'node:assert/strict'
import { join } from 'node:path'
import { describe, it } from 'node:test'
import {
  appendChartStatFlags,
  buildConvertArgs,
  resolveInputs,
  resolveVizbBin,
  runPipeline,
  splitMergeFiles,
} from './lib.mjs'

describe('appendChartStatFlags', () => {
  it('adds -c when charts is set', () => {
    assert.deepEqual(appendChartStatFlags([], 'bar,line', '', ''), ['-c', 'bar,line'])
  })

  it('parses multiline chart overrides, skips blanks and comments', () => {
    const chart = ['', '  # ignored', 'bar:scale=log', '  pie:labels', '# trailing'].join('\n')
    assert.deepEqual(appendChartStatFlags([], '', chart, ''), [
      '--chart',
      'bar:scale=log',
      '--chart',
      'pie:labels',
    ])
  })

  it('maps stat all/true to bare --stat and lists to --stat=value', () => {
    assert.deepEqual(appendChartStatFlags([], '', '', 'all'), ['--stat'])
    assert.deepEqual(appendChartStatFlags([], '', '', 'true'), ['--stat'])
    assert.deepEqual(appendChartStatFlags([], '', '', 'center,spread'), [
      '--stat=center,spread',
    ])
  })
})

describe('buildConvertArgs', () => {
  const base = {
    file: '',
    cmd: '',
    hasInput: true,
    jsonFile: 'bench.json',
    name: 'Comparisons',
    title: '',
    description: '',
    tag: '',
    id: '',
    group: '',
    groupPattern: 'x',
    groupRegex: '',
    sort: '',
    filter: '',
    memUnit: 'B',
    timeUnit: 'ns',
    numberUnit: '',
    round: false,
    select: '',
    colAxis: '',
    jsonPath: '',
    stat: '',
    chart: '',
    charts: '',
    parser: 'auto',
    showLabels: false,
    enable3d: false,
    mergeFiles: '',
    mergeDir: '',
    tagAxis: 'n',
    outputHtml: '',
    dataUrl: '',
  }

  it('includes defaults the action always passes', () => {
    assert.deepEqual(buildConvertArgs(base), [
      '-n',
      'Comparisons',
      '-p',
      'x',
      '-M',
      'B',
      '-T',
      'ns',
      '-P',
      'auto',
    ])
  })

  it('adds optional flags and chart/stat', () => {
    const args = buildConvertArgs({
      ...base,
      tag: 'v1',
      round: true,
      showLabels: true,
      charts: 'bar',
      chart: 'bar:scale=log',
      stat: 'true',
    })
    assert.ok(args.includes('--tag') && args.includes('v1'))
    assert.ok(args.includes('--round'))
    assert.ok(args.includes('-l'))
    assert.ok(args.includes('-c') && args.includes('bar'))
    assert.ok(args.includes('--chart') && args.includes('bar:scale=log'))
    assert.ok(args.includes('--stat'))
  })
})

describe('resolveInputs', () => {
  /** @param {Record<string, string>} map */
  function envFrom(map) {
    /** @type {NodeJS.ProcessEnv} */
    const env = {}
    for (const [k, v] of Object.entries(map)) env[`INPUT_${k}`] = v
    return env
  }

  it('prefers file/cmd over deprecated bench-* aliases', () => {
    const r = resolveInputs(
      envFrom({
        file: 'new.txt',
        'bench-file': 'old.txt',
        cmd: 'echo hi',
        'bench-cmd': 'echo old',
        name: 'Comparisons',
        'group-pattern': 'x',
        'mem-unit': 'B',
        'time-unit': 'ns',
        parser: 'auto',
      }),
    )
    assert.equal(r.file, 'new.txt')
    assert.equal(r.cmd, 'echo hi')
    assert.equal(r.hasInput, true)
    assert.equal(r.jsonFile, 'bench.json')
  })

  it('falls back to bench-* and custom output-json', () => {
    const r = resolveInputs(
      envFrom({ 'bench-file': 'old.txt', 'output-json': 'out.json' }),
    )
    assert.equal(r.file, 'old.txt')
    assert.equal(r.jsonFile, 'out.json')
  })

  it('allows merge-only and data-url without file/cmd', () => {
    assert.equal(
      resolveInputs(envFrom({ 'merge-files': 'a.json b.json' })).hasInput,
      false,
    )
    assert.equal(
      resolveInputs(envFrom({ 'data-url': 'https://example.com/d.json' })).dataUrl,
      'https://example.com/d.json',
    )
  })

  it('errors when no input source is provided', () => {
    assert.throws(() => resolveInputs({}), /No input provided/)
  })

  it('parses boolean-like inputs', () => {
    const r = resolveInputs(
      envFrom({
        file: 'a.txt',
        round: 'true',
        'show-labels': 'true',
        'enable-3d': 'true',
        'tag-axis': 'x',
      }),
    )
    assert.equal(r.round, true)
    assert.equal(r.showLabels, true)
    assert.equal(r.enable3d, true)
    assert.equal(r.tagAxis, 'x')
  })
})

describe('splitMergeFiles', () => {
  it('splits on whitespace', () => {
    assert.deepEqual(splitMergeFiles('a.json  b.json\tc.json'), [
      'a.json',
      'b.json',
      'c.json',
    ])
    assert.deepEqual(splitMergeFiles(''), [])
  })
})

describe('resolveVizbBin', () => {
  it('returns ~/.local/bin/vizb(.exe) when present', () => {
    const home = join('Users', 'runner')
    const expected = join(home, '.local', 'bin', 'vizb.exe')
    assert.equal(
      resolveVizbBin({
        platform: 'win32',
        homedir: () => home,
        existsSync: (p) => p === expected,
      }),
      expected,
    )
  })

  it('falls back to bare name', () => {
    assert.equal(
      resolveVizbBin({
        platform: 'linux',
        homedir: () => '/home/u',
        existsSync: () => false,
      }),
      'vizb',
    )
    assert.equal(
      resolveVizbBin({
        platform: 'win32',
        homedir: () => join('Users', 'runner'),
        existsSync: () => false,
      }),
      'vizb.exe',
    )
  })
})

describe('runPipeline', () => {
  function baseInputs(over = {}) {
    return {
      file: 'examples/go/hash.txt',
      cmd: '',
      hasInput: true,
      jsonFile: 'results/out.json',
      name: 'Comparisons',
      title: '',
      description: '',
      tag: '',
      id: '',
      group: '',
      groupPattern: 'x',
      groupRegex: '',
      sort: '',
      filter: '',
      memUnit: 'B',
      timeUnit: 'ns',
      numberUnit: '',
      round: false,
      select: '',
      colAxis: '',
      jsonPath: '',
      stat: '',
      chart: '',
      charts: '',
      parser: 'auto',
      showLabels: false,
      enable3d: false,
      mergeFiles: '',
      mergeDir: '',
      tagAxis: 'n',
      outputHtml: 'index.html',
      dataUrl: '',
      ...over,
    }
  }

  it('runs convert then ui for file + html', () => {
    /** @type {string[][]} */
    const calls = []
    runPipeline(baseInputs(), {
      vizbBin: 'vizb',
      run: (cmd, args) => {
        calls.push([cmd, ...args])
      },
    })
    assert.equal(calls.length, 2)
    assert.equal(calls[0][0], 'vizb')
    assert.equal(calls[0][1], 'examples/go/hash.txt')
    assert.deepEqual(calls[1].slice(0, 4), ['vizb', 'ui', 'results/out.json', '-o'])
  })

  it('captures cmd stdout then converts', () => {
    /** @type {string[][]} */
    const calls = []
    /** @type {string[]} */
    const shells = []
    runPipeline(
      baseInputs({ file: '', cmd: 'echo hello', hasInput: true, outputHtml: '' }),
      {
        vizbBin: 'vizb',
        run: (cmd, args) => {
          calls.push([cmd, ...args])
        },
        runShellCapture: (command, path) => {
          shells.push(command, path)
        },
      },
    )
    assert.equal(shells[0], 'echo hello')
    assert.ok(shells[1].includes('vizb-data-input-'))
    assert.equal(calls.length, 1)
  })

  it('merges when merge-dir set and json exists', () => {
    /** @type {string[][]} */
    const calls = []
    // Pre-create path check via real existsSync is fine if we only convert+merge
    // without relying on file presence — inject by writing nothing and only merge-files.
    runPipeline(
      baseInputs({
        hasInput: false,
        file: '',
        mergeFiles: 'a.json b.json',
        mergeDir: 'results/prev',
        outputHtml: '',
        jsonFile: 'merged.json',
      }),
      {
        vizbBin: 'vizb',
        run: (cmd, args) => {
          calls.push([cmd, ...args])
        },
      },
    )
    assert.equal(calls.length, 1)
    assert.equal(calls[0][1], 'merge')
    assert.ok(calls[0].includes('a.json'))
    assert.ok(calls[0].includes('b.json'))
    assert.ok(calls[0].includes('results/prev'))
  })

  it('uses data-url for ui without convert', () => {
    /** @type {string[][]} */
    const calls = []
    runPipeline(
      baseInputs({
        file: '',
        hasInput: false,
        dataUrl: 'https://example.com/d.json',
        outputHtml: 'pages/index.html',
        enable3d: true,
        charts: 'pie',
      }),
      {
        vizbBin: 'vizb',
        run: (cmd, args) => {
          calls.push([cmd, ...args])
        },
      },
    )
    assert.deepEqual(calls[0].slice(0, 6), [
      'vizb',
      'ui',
      '-U',
      'https://example.com/d.json',
      '-o',
      'pages/index.html',
    ])
    assert.ok(calls[0].includes('-c') && calls[0].includes('pie'))
    assert.ok(calls[0].includes('--3d'))
  })
})
