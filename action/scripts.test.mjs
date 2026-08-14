import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import { chmodSync, existsSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join } from 'node:path'
import { describe, it } from 'node:test'
import { fileURLToPath } from 'node:url'

const actionDir = dirname(fileURLToPath(import.meta.url))
const resolveSh = join(actionDir, 'resolve.sh')
const installSh = join(actionDir, 'install.sh')

function runBash(script, env, extra = {}) {
  return spawnSync('bash', [script], {
    encoding: 'utf8',
    env: { ...process.env, ...env },
    ...extra,
  })
}

function tempDir(prefix) {
  return mkdtempSync(join(tmpdir(), prefix))
}

function writeGitStub(binDir, body) {
  mkdirSync(binDir, { recursive: true })
  const git = join(binDir, 'git')
  writeFileSync(git, `#!/usr/bin/env bash\n${body}`)
  chmodSync(git, 0o755)
  return git
}

describe('resolve.sh', () => {
  it('uses a pinned ref as the tag without calling git', () => {
    const dir = tempDir('vizb-resolve-')
    const out = join(dir, 'out')
    writeFileSync(out, '')
    const gitLog = join(dir, 'git.log')
    writeGitStub(
      join(dir, 'bin'),
      `echo called >> '${gitLog}'\nexit 1\n`,
    )

    const result = runBash(resolveSh, {
      GITHUB_ACTION_REF: 'v0.18.2',
      GITHUB_OUTPUT: out,
      PATH: `${join(dir, 'bin')}:${process.env.PATH}`,
    })

    assert.equal(result.status, 0, result.stderr)
    assert.equal(readFileSync(out, 'utf8'), 'tag=v0.18.2\n')
    assert.equal(existsSync(gitLog), false)
  })

  it('resolves a major ref from stubbed git ls-remote', () => {
    const dir = tempDir('vizb-resolve-')
    const out = join(dir, 'out')
    writeFileSync(out, '')
    writeGitStub(
      join(dir, 'bin'),
      'printf "abc\\trefs/tags/v0.19.0\\n"\n',
    )

    const result = runBash(resolveSh, {
      GITHUB_ACTION_REF: 'v0',
      GITHUB_OUTPUT: out,
      PATH: `${join(dir, 'bin')}:${process.env.PATH}`,
    })

    assert.equal(result.status, 0, result.stderr)
    assert.equal(readFileSync(out, 'utf8'), 'tag=v0.19.0\n')
  })

  it('fails when a major ref has no tags', () => {
    const dir = tempDir('vizb-resolve-')
    const out = join(dir, 'out')
    writeFileSync(out, '')
    writeGitStub(join(dir, 'bin'), 'exit 0\n')

    const result = runBash(resolveSh, {
      GITHUB_ACTION_REF: 'v0',
      GITHUB_OUTPUT: out,
      PATH: `${join(dir, 'bin')}:${process.env.PATH}`,
    })

    assert.notEqual(result.status, 0)
    assert.match(result.stderr + result.stdout, /No release tags found for v0/)
  })
})

describe('install.sh', () => {
  it('copies vizb-binary to ~/.local/bin/vizb', () => {
    const dir = tempDir('vizb-install-')
    const home = join(dir, 'home')
    const src = join(dir, 'vizb')
    writeFileSync(src, 'bin')
    mkdirSync(home, { recursive: true })

    const result = runBash(installSh, {
      HOME: home,
      RUNNER_OS: 'Linux',
      VIZB_BINARY: src,
    })

    assert.equal(result.status, 0, result.stderr)
    assert.equal(readFileSync(join(home, '.local', 'bin', 'vizb'), 'utf8'), 'bin')
  })

  it('falls back to vizb-binary.exe and installs vizb.exe on Windows', () => {
    const dir = tempDir('vizb-install-')
    const home = join(dir, 'home')
    const src = join(dir, 'vizb')
    writeFileSync(`${src}.exe`, 'winbin')
    mkdirSync(home, { recursive: true })

    const result = runBash(installSh, {
      HOME: home,
      RUNNER_OS: 'Windows',
      VIZB_BINARY: src,
    })

    assert.equal(result.status, 0, result.stderr)
    assert.equal(readFileSync(join(home, '.local', 'bin', 'vizb.exe'), 'utf8'), 'winbin')
  })

  it('maps macos/x64 to a darwin/amd64 tar.gz url', () => {
    const dir = tempDir('vizb-install-')
    const result = runBash(installSh, {
      HOME: dir,
      RUNNER_OS: 'macOS',
      RUNNER_ARCH: 'X64',
      VIZB_TAG: 'v0.18.2',
      VIZB_DRY_RUN: '1',
    })

    assert.equal(result.status, 0, result.stderr)
    assert.match(result.stdout, /dest=.*\/\.local\/bin\/vizb$/m)
    assert.match(
      result.stdout,
      /url=https:\/\/github.com\/goptics\/vizb\/releases\/download\/v0\.18\.2\/vizb@0\.18\.2-darwin-amd64\.tar\.gz/,
    )
  })

  it('maps windows/x64 to a zip url and vizb.exe dest', () => {
    const dir = tempDir('vizb-install-')
    const result = runBash(installSh, {
      HOME: dir,
      RUNNER_OS: 'Windows',
      RUNNER_ARCH: 'X64',
      VIZB_TAG: 'v0.18.2',
      VIZB_DRY_RUN: '1',
    })

    assert.equal(result.status, 0, result.stderr)
    assert.match(result.stdout, /dest=.*\/\.local\/bin\/vizb\.exe$/m)
    assert.match(
      result.stdout,
      /url=https:\/\/github.com\/goptics\/vizb\/releases\/download\/v0\.18\.2\/vizb@0\.18\.2-windows-amd64\.zip/,
    )
  })
})
