import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import { runPipeline } from './pipeline.mjs'

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
    outputJson: 'results/out.json',
    outputHtml: 'index.html',
    dataUrl: '',
    ...over,
  }
}

describe('runPipeline', () => {
  it('runs convert then ui for file + html', () => {
    /** @type {string[][]} */
    const calls = []
    runPipeline(baseInputs(), {
      run: (cmd, args) => {
        calls.push([cmd, ...args])
      },
      existsSync: () => false,
    })

    assert.equal(calls.length, 2)
    assert.equal(calls[0][0], 'vizb')
    assert.equal(calls[0][1], 'examples/go/hash.txt')
    assert.ok(calls[0].includes('-o'))
    assert.ok(calls[0].includes('results/out.json'))
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
        run: (cmd, args) => {
          calls.push([cmd, ...args])
        },
        runShellCapture: (command, path) => {
          shells.push(command, path)
        },
        tmpdir: () => '/tmp',
        existsSync: () => false,
      },
    )

    assert.equal(shells[0], 'echo hello')
    assert.ok(shells[1].startsWith('/tmp/vizb-data-input-'))
    assert.equal(calls.length, 1)
    assert.equal(calls[0][1], shells[1])
  })

  it('merges when merge-dir set and json exists', () => {
    /** @type {string[][]} */
    const calls = []
    runPipeline(
      baseInputs({
        mergeDir: 'results/prev',
        outputHtml: '',
      }),
      {
        run: (cmd, args) => {
          calls.push([cmd, ...args])
        },
        existsSync: (p) => p === 'results/out.json',
      },
    )

    assert.equal(calls.length, 2)
    assert.equal(calls[1][1], 'merge')
    assert.ok(calls[1].includes('results/out.json'))
    assert.ok(calls[1].includes('results/prev'))
    assert.ok(calls[1].includes('--tag-axis'))
    assert.ok(calls[1].includes('n'))
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
      }),
      {
        run: (cmd, args) => {
          calls.push([cmd, ...args])
        },
        existsSync: () => false,
      },
    )

    assert.equal(calls.length, 1)
    assert.deepEqual(calls[0].slice(0, 6), [
      'vizb',
      'ui',
      '-U',
      'https://example.com/d.json',
      '-o',
      'pages/index.html',
    ])
  })

  it('splits merge-files into separate args', () => {
    /** @type {string[][]} */
    const calls = []
    runPipeline(
      baseInputs({
        hasInput: false,
        file: '',
        mergeFiles: 'a.json b.json',
        outputHtml: '',
        jsonFile: 'merged.json',
      }),
      {
        run: (cmd, args) => {
          calls.push([cmd, ...args])
        },
        existsSync: () => false,
      },
    )

    assert.equal(calls.length, 1)
    assert.ok(calls[0].includes('a.json'))
    assert.ok(calls[0].includes('b.json'))
  })
})
