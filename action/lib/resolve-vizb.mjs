import { existsSync } from 'node:fs'
import { homedir } from 'node:os'
import { delimiter, join } from 'node:path'

/**
 * Resolve the vizb executable path for spawnSync.
 *
 * On Windows, Node's spawn without shell only finds names covered by PATHEXT
 * (e.g. vizb.exe). An extensionless file on PATH (legacy install layout) is
 * invisible to spawnSync — resolve an absolute path instead.
 *
 * @param {{ env?: NodeJS.ProcessEnv, platform?: string, homedir?: () => string, existsSync?: (p: string) => boolean }} [opts]
 * @returns {string}
 */
export function resolveVizbBin(opts = {}) {
  const env = opts.env ?? process.env
  const platform = opts.platform ?? process.platform
  const home = opts.homedir ?? homedir
  const exists = opts.existsSync ?? existsSync
  const isWin = platform === 'win32'

  if (env.VIZB_BIN && exists(env.VIZB_BIN)) {
    return env.VIZB_BIN
  }

  // Prefer the install location used by action.yml.
  const names = isWin ? ['vizb.exe', 'vizb'] : ['vizb']
  const localBin = join(home(), '.local', 'bin')
  for (const name of names) {
    const candidate = join(localBin, name)
    if (exists(candidate)) return candidate
  }

  // Search PATH for an absolute match (needed for extensionless Windows files).
  const pathEnv = env.PATH || env.Path || ''
  for (const dir of pathEnv.split(delimiter)) {
    if (!dir) continue
    for (const name of names) {
      const candidate = join(dir, name)
      if (exists(candidate)) return candidate
    }
  }

  // Last resort: bare name (works on Unix; on Windows expects PATHEXT match).
  return isWin ? 'vizb.exe' : 'vizb'
}
