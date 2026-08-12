import { spawnSync } from 'node:child_process'
import { closeSync, openSync } from 'node:fs'

/**
 * Run a command with an argv array (no shell).
 * @param {string} command
 * @param {string[]} args
 * @param {{ cwd?: string, env?: NodeJS.ProcessEnv }} [options]
 */
export function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    stdio: 'inherit',
    encoding: 'utf8',
    ...options,
  })

  if (result.error) {
    throw result.error
  }
  if (result.status !== 0) {
    const code = result.status ?? 1
    const err = new Error(`${command} exited with code ${code}`)
    // @ts-expect-error attach exit code for callers
    err.exitCode = code
    throw err
  }
}

/**
 * Run a user shell command and write stdout to a file (stderr inherited).
 * @param {string} command
 * @param {string} stdoutPath
 * @param {{ cwd?: string, env?: NodeJS.ProcessEnv }} [options]
 */
export function runShellCapture(command, stdoutPath, options = {}) {
  const fd = openSync(stdoutPath, 'w')
  try {
    const result = spawnSync(command, {
      shell: true,
      stdio: ['ignore', fd, 'inherit'],
      encoding: 'utf8',
      ...options,
    })

    if (result.error) {
      throw result.error
    }
    if (result.status !== 0) {
      const code = result.status ?? 1
      const err = new Error(`Command failed with exit code ${code}: ${command}`)
      // @ts-expect-error attach exit code for callers
      err.exitCode = code
      throw err
    }
  } finally {
    closeSync(fd)
  }
}
