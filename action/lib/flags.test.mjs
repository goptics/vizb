import assert from 'node:assert/strict'
import { describe, it } from 'node:test'
import { appendChartStatFlags, buildConvertArgs, buildUiArgs } from './flags.mjs'

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

  it('maps stat all/true to bare --stat', () => {
    assert.deepEqual(appendChartStatFlags([], '', '', 'all'), ['--stat'])
    assert.deepEqual(appendChartStatFlags([], '', '', 'true'), ['--stat'])
  })

  it('maps non-empty stat list to --stat=value', () => {
    assert.deepEqual(appendChartStatFlags([], '', '', 'center,spread'), [
      '--stat=center,spread',
    ])
  })

  it('combines charts, chart lines, and stat', () => {
    assert.deepEqual(
      appendChartStatFlags([], 'chord', 'chord:labels\n', 'all'),
      ['-c', 'chord', '--chart', 'chord:labels', '--stat'],
    )
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
    outputJson: '',
    outputHtml: '',
    dataUrl: '',
  }

  it('includes defaults that the action always passes', () => {
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
      id: 'hash',
      title: 'Title',
      description: 'desc',
      group: 'name',
      groupRegex: '.*',
      sort: 'asc',
      filter: 'Foo',
      numberUnit: 'M',
      round: true,
      select: 'price,count',
      colAxis: 'n',
      jsonPath: '.data',
      showLabels: true,
      charts: 'bar',
      chart: 'bar:scale=log',
      stat: 'true',
    })

    assert.ok(args.includes('--tag') && args.includes('v1'))
    assert.ok(args.includes('--id') && args.includes('hash'))
    assert.ok(args.includes('--title') && args.includes('Title'))
    assert.ok(args.includes('-d') && args.includes('desc'))
    assert.ok(args.includes('-g') && args.includes('name'))
    assert.ok(args.includes('-r') && args.includes('.*'))
    assert.ok(args.includes('--sort') && args.includes('asc'))
    assert.ok(args.includes('--filter') && args.includes('Foo'))
    assert.ok(args.includes('-N') && args.includes('M'))
    assert.ok(args.includes('--round'))
    assert.ok(args.includes('--select') && args.includes('price,count'))
    assert.ok(args.includes('--col-axis') && args.includes('n'))
    assert.ok(args.includes('--json-path') && args.includes('.data'))
    assert.ok(args.includes('-l'))
    assert.ok(args.includes('-c') && args.includes('bar'))
    assert.ok(args.includes('--chart') && args.includes('bar:scale=log'))
    assert.ok(args.includes('--stat'))
  })
})

describe('buildUiArgs', () => {
  it('adds chart flags and optional --3d', () => {
    assert.deepEqual(
      buildUiArgs({
        charts: 'pie',
        chart: 'pie:labels',
        stat: '',
        enable3d: true,
      }),
      ['-c', 'pie', '--chart', 'pie:labels', '--3d'],
    )
  })
})
