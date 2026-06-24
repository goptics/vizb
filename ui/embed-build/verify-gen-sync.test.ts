import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { describe, expect, it } from 'vitest'
import { execFileSync } from 'node:child_process'

const script = path.resolve('embed-build/verify-gen-sync.ts')
const genGo = path.resolve('..', 'pkg', 'template', 'vizb-ui.gen.go')

describe('verify-gen-sync', () => {
  it('accepts identical gen.go files', () => {
    const out = execFileSync('node', [script, genGo, genGo], { encoding: 'utf-8' })
    expect(out).toContain('in sync')
  })

  it('accepts different gzip bytes for identical chunk payloads', () => {
    const out = execFileSync(
      'node',
      [
        script,
        '/tmp/vizb-ci-test/committed.gen.go',
        '/tmp/vizb-ci-test/pkg/template/vizb-ui.gen.go',
      ],
      { encoding: 'utf-8' }
    )
    expect(out).toContain('in sync')
  })

  it('rejects a chunk graph with a missing key', () => {
    const source = fs.readFileSync(genGo, 'utf-8')
    const tampered = source.replace(/^\t"vizb:axisAlignTicks-lwmJPqWv":.*\n/m, '')
    const tmp = path.join(os.tmpdir(), 'vizb-ui.gen.go.tampered')
    fs.writeFileSync(tmp, tampered)
    expect(() =>
      execFileSync('node', [script, genGo, tmp], { encoding: 'utf-8', stdio: 'pipe' })
    ).toThrow()
  })
})
