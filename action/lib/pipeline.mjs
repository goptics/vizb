import { existsSync } from 'node:fs'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import { splitMergeFiles } from './inputs.mjs'
import { buildConvertArgs, buildUiArgs } from './flags.mjs'
import { run, runShellCapture } from './exec.mjs'
import { resolveVizbBin } from './resolve-vizb.mjs'

/**
 * @typedef {object} PipelineDeps
 * @property {(command: string, args: string[], options?: object) => void} [run]
 * @property {(command: string, stdoutPath: string, options?: object) => void} [runShellCapture]
 * @property {(path: string) => boolean} [existsSync]
 * @property {() => string} [tmpdir]
 * @property {string} [vizbBin]
 * @property {() => string} [resolveVizbBin]
 */

/**
 * Execute convert → merge → ui according to resolved inputs.
 * @param {import('./inputs.mjs').ResolvedInputs} inputs
 * @param {PipelineDeps} [deps]
 */
export function runPipeline(inputs, deps = {}) {
  const execRun = deps.run ?? run
  const execShell = deps.runShellCapture ?? runShellCapture
  const exists = deps.existsSync ?? existsSync
  const getTmp = deps.tmpdir ?? tmpdir
  const vizb = deps.vizbBin ?? (deps.resolveVizbBin ?? resolveVizbBin)()

  if (inputs.hasInput) {
    let inputPath = inputs.file
    if (!inputPath) {
      inputPath = join(getTmp(), `vizb-data-input-${process.pid}.txt`)
      execShell(inputs.cmd, inputPath)
    }

    const convertArgs = [
      inputPath,
      '-o',
      inputs.jsonFile,
      ...buildConvertArgs(inputs),
    ]
    execRun(vizb, convertArgs)
  }

  if (inputs.mergeFiles || inputs.mergeDir) {
    /** @type {string[]} */
    const files = []
    if (exists(inputs.jsonFile)) {
      files.push(inputs.jsonFile)
    }
    files.push(...splitMergeFiles(inputs.mergeFiles))
    if (inputs.mergeDir) {
      files.push(inputs.mergeDir)
    }

    execRun(vizb, [
      'merge',
      ...files,
      '-o',
      inputs.jsonFile,
      '--tag-axis',
      inputs.tagAxis,
    ])
  }

  if (inputs.outputHtml) {
    const uiFlags = buildUiArgs(inputs)
    if (inputs.dataUrl) {
      execRun(vizb, ['ui', '-U', inputs.dataUrl, '-o', inputs.outputHtml, ...uiFlags])
    } else {
      execRun(vizb, ['ui', inputs.jsonFile, '-o', inputs.outputHtml, ...uiFlags])
    }
  }
}
