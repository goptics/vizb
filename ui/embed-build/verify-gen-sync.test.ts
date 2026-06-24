import { execFileSync } from 'node:child_process'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import zlib from 'node:zlib'
import { describe, expect, it } from 'vitest'
import { removeChunkMapLine, replaceChunkB64, sampleChunkEntry } from './gen-go-parse'

const script = path.resolve('embed-build/verify-gen-sync.ts')
const genGo = path.resolve('..', 'pkg', 'template', 'vizb-ui.gen.go')

describe('verify-gen-sync', () => {
  it('accepts identical gen.go files', () => {
    const out = execFileSync('node', [script, genGo, genGo], { encoding: 'utf-8' })
    expect(out).toContain('in sync')
  })

  it('accepts different gzip bytes for identical chunk payloads', () => {
    const source = fs.readFileSync(genGo, 'utf-8')
    const { key, b64 } = sampleChunkEntry(source)
    const payload = zlib.gunzipSync(Buffer.from(b64, 'base64'))
    const altB64 = zlib.gzipSync(payload, { level: 1 }).toString('base64')
    expect(altB64).not.toBe(b64)

    const leftPath = path.join(os.tmpdir(), 'vizb-ui.gen.go.left')
    const rightPath = path.join(os.tmpdir(), 'vizb-ui.gen.go.right')
    fs.writeFileSync(leftPath, source)
    fs.writeFileSync(rightPath, replaceChunkB64(source, key, altB64))

    const out = execFileSync('node', [script, leftPath, rightPath], { encoding: 'utf-8' })
    expect(out).toContain('in sync')
  })

  it('rejects a chunk graph with a missing key', () => {
    const source = fs.readFileSync(genGo, 'utf-8')
    const { key } = sampleChunkEntry(source)
    const tampered = removeChunkMapLine(source, key)
    const tmp = path.join(os.tmpdir(), 'vizb-ui.gen.go.tampered')
    fs.writeFileSync(tmp, tampered)
    expect(() =>
      execFileSync('node', [script, genGo, tmp], { encoding: 'utf-8', stdio: 'pipe' })
    ).toThrow()
  })
})
