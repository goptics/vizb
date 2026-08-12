import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import { readRawInputs, resolveInputs, splitMergeFiles } from './inputs.mjs'

describe('readRawInputs', () => {
  it('maps INPUT_* env keys', () => {
    const raw = readRawInputs({
      INPUT_file: 'a.txt',
      INPUT_cmd: '',
      INPUT_chart: 'bar:scale=log',
    })
    assert.equal(raw.file, 'a.txt')
    assert.equal(raw.cmd, '')
    assert.equal(raw.chart, 'bar:scale=log')
  })
})

describe('resolveInputs', () => {
  it('prefers file/cmd over deprecated bench-* aliases', () => {
    const r = resolveInputs({
      file: 'new.txt',
      'bench-file': 'old.txt',
      cmd: 'echo hi',
      'bench-cmd': 'echo old',
      'merge-files': '',
      'merge-dir': '',
      'data-url': '',
      'output-json': '',
      'output-html': '',
      name: 'Comparisons',
      'group-pattern': 'x',
      'mem-unit': 'B',
      'time-unit': 'ns',
      parser: 'auto',
      round: 'false',
      'show-labels': 'false',
      'enable-3d': 'false',
      'tag-axis': 'n',
    })
    assert.equal(r.file, 'new.txt')
    assert.equal(r.cmd, 'echo hi')
    assert.equal(r.hasInput, true)
    assert.equal(r.jsonFile, 'bench.json')
  })

  it('falls back to bench-* when new inputs empty', () => {
    const r = resolveInputs({
      file: '',
      'bench-file': 'old.txt',
      cmd: '',
      'bench-cmd': '',
      'merge-files': '',
      'merge-dir': '',
      'data-url': '',
      'output-json': 'out.json',
      'output-html': '',
      name: '',
      'group-pattern': '',
      'mem-unit': '',
      'time-unit': '',
      parser: '',
      round: 'false',
      'show-labels': 'false',
      'enable-3d': 'false',
      'tag-axis': 'n',
    })
    assert.equal(r.file, 'old.txt')
    assert.equal(r.hasInput, true)
    assert.equal(r.jsonFile, 'out.json')
  })

  it('allows merge-only and data-url without file/cmd', () => {
    const mergeOnly = resolveInputs({
      file: '',
      'bench-file': '',
      cmd: '',
      'bench-cmd': '',
      'merge-files': 'a.json b.json',
      'merge-dir': '',
      'data-url': '',
      'output-json': '',
      'output-html': '',
      name: '',
      'group-pattern': '',
      'mem-unit': '',
      'time-unit': '',
      parser: '',
      round: 'false',
      'show-labels': 'false',
      'enable-3d': 'false',
      'tag-axis': 'n',
    })
    assert.equal(mergeOnly.hasInput, false)

    const dataOnly = resolveInputs({
      file: '',
      'bench-file': '',
      cmd: '',
      'bench-cmd': '',
      'merge-files': '',
      'merge-dir': '',
      'data-url': 'https://example.com/data.json',
      'output-json': '',
      'output-html': 'index.html',
      name: '',
      'group-pattern': '',
      'mem-unit': '',
      'time-unit': '',
      parser: '',
      round: 'false',
      'show-labels': 'false',
      'enable-3d': 'false',
      'tag-axis': 'n',
    })
    assert.equal(dataOnly.hasInput, false)
    assert.equal(dataOnly.dataUrl, 'https://example.com/data.json')
  })

  it('errors when no input source is provided', () => {
    assert.throws(
      () =>
        resolveInputs({
          file: '',
          'bench-file': '',
          cmd: '',
          'bench-cmd': '',
          'merge-files': '',
          'merge-dir': '',
          'data-url': '',
          'output-json': '',
          'output-html': '',
          name: '',
          'group-pattern': '',
          'mem-unit': '',
          'time-unit': '',
          parser: '',
          round: 'false',
          'show-labels': 'false',
          'enable-3d': 'false',
          'tag-axis': 'n',
        }),
      /No input provided/,
    )
  })

  it('parses boolean-like inputs', () => {
    const r = resolveInputs({
      file: 'a.txt',
      'bench-file': '',
      cmd: '',
      'bench-cmd': '',
      'merge-files': '',
      'merge-dir': '',
      'data-url': '',
      'output-json': '',
      'output-html': '',
      name: '',
      'group-pattern': '',
      'mem-unit': '',
      'time-unit': '',
      parser: '',
      round: 'true',
      'show-labels': 'true',
      'enable-3d': 'true',
      'tag-axis': 'x',
    })
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
