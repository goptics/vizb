import assert from 'node:assert/strict'
import { join } from 'node:path'
import { describe, it } from 'node:test'
import { resolveVizbBin } from './resolve-vizb.mjs'

describe('resolveVizbBin', () => {
  it('prefers VIZB_BIN when the path exists', () => {
    const bin = join('tools', 'vizb.exe')
    assert.equal(
      resolveVizbBin({
        env: { VIZB_BIN: bin },
        platform: 'win32',
        existsSync: (p) => p === bin,
        homedir: () => join('Users', 'runner'),
      }),
      bin,
    )
  })

  it('finds ~/.local/bin/vizb.exe on Windows before bare name', () => {
    const home = join('Users', 'runner')
    const expected = join(home, '.local', 'bin', 'vizb.exe')
    assert.equal(
      resolveVizbBin({
        env: { PATH: join('Windows', 'System32') },
        platform: 'win32',
        homedir: () => home,
        existsSync: (p) => p === expected,
      }),
      expected,
    )
  })

  it('falls back to extensionless ~/.local/bin/vizb on Windows', () => {
    const home = join('Users', 'runner')
    const expected = join(home, '.local', 'bin', 'vizb')
    assert.equal(
      resolveVizbBin({
        env: { PATH: '' },
        platform: 'win32',
        homedir: () => home,
        // Only extensionless file exists (legacy install).
        existsSync: (p) => p === expected,
      }),
      expected,
    )
  })

  it('searches PATH for an absolute hit', () => {
    assert.equal(
      resolveVizbBin({
        env: { PATH: '/opt/bin:/usr/bin' },
        platform: 'linux',
        homedir: () => '/home/u',
        existsSync: (p) => p === '/opt/bin/vizb',
      }),
      '/opt/bin/vizb',
    )
  })

  it('returns bare vizb.exe on Windows when nothing is found', () => {
    assert.equal(
      resolveVizbBin({
        env: { PATH: '' },
        platform: 'win32',
        homedir: () => 'C:\\Users\\runner',
        existsSync: () => false,
      }),
      'vizb.exe',
    )
  })

  it('returns bare vizb on Unix when nothing is found', () => {
    assert.equal(
      resolveVizbBin({
        env: { PATH: '' },
        platform: 'linux',
        homedir: () => '/home/u',
        existsSync: () => false,
      }),
      'vizb',
    )
  })
})
