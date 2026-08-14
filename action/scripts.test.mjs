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

function writeExec(path, body) {
  writeFileSync(path, `#!/usr/bin/env bash\n${body}`)
  chmodSync(path, 0o755)
}

function runInstallRelease({ runnerOs, runnerArch, tag }) {
  const dir = tempDir('vizb-install-')
  const home = join(dir, 'home')
  const cwd = join(dir, 'cwd')
  const bin = join(dir, 'bin')
  mkdirSync(home, { recursive: true })
  mkdirSync(cwd, { recursive: true })
  mkdirSync(bin, { recursive: true })

  const destName = runnerOs === 'Windows' ? 'vizb.exe' : 'vizb'
  const dest = join(home, '.local', 'bin', destName)
  const urlLog = join(dir, 'curl.url')
  const archiveLog = join(dir, 'curl.dest')

  writeExec(
    join(bin, 'curl'),
    `printf '%s' "$2" > '${urlLog}'\nprintf '%s' "$4" > '${archiveLog}'\nprintf archive > "$4"\n`,
  )
  writeExec(join(bin, 'tar'), `printf extracted > '${dest}'\n`)
  writeExec(join(bin, 'unzip'), `printf extracted > '${dest}'\n`)

  const result = runBash(
    installSh,
    {
      HOME: home,
      RUNNER_OS: runnerOs,
      RUNNER_ARCH: runnerArch,
      VIZB_TAG: tag,
      PATH: `${bin}:${process.env.PATH}`,
    },
    { cwd },
  )

  return {
    result,
    cwd,
    dest,
    url: existsSync(urlLog) ? readFileSync(urlLog, 'utf8') : '',
    archive: existsSync(archiveLog) ? readFileSync(archiveLog, 'utf8') : '',
  }
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
    const r = runInstallRelease({ runnerOs: 'macOS', runnerArch: 'X64', tag: 'v0.18.2' })
    assert.equal(r.result.status, 0, r.result.stderr)
    assert.ok(r.dest.endsWith(`${join('.local', 'bin', 'vizb')}`))
    assert.equal(
      r.url,
      'https://github.com/goptics/vizb/releases/download/v0.18.2/vizb@0.18.2-darwin-amd64.tar.gz',
    )
  })

  it('downloads to a unique temp archive and removes it', () => {
    const r = runInstallRelease({ runnerOs: 'Linux', runnerArch: 'X64', tag: 'v0.18.2' })
    assert.equal(r.result.status, 0, r.result.stderr)
    assert.equal(existsSync(join(r.cwd, 'vizb-archive')), false)
    assert.ok(r.archive.includes('vizb-archive.'))
    assert.equal(existsSync(r.archive), false)
  })

  it('maps windows/x64 to a zip url and vizb.exe dest', () => {
    const r = runInstallRelease({ runnerOs: 'Windows', runnerArch: 'X64', tag: 'v0.18.2' })
    assert.equal(r.result.status, 0, r.result.stderr)
    assert.ok(r.dest.endsWith(`${join('.local', 'bin', 'vizb.exe')}`))
    assert.equal(
      r.url,
      'https://github.com/goptics/vizb/releases/download/v0.18.2/vizb@0.18.2-windows-amd64.zip',
    )
  })
})
