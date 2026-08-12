#!/usr/bin/env node
/**
 * Vizb GitHub Action pipeline runner.
 * Zero npm dependencies — Node builtins only.
 *
 * Expects INPUT_* env vars set by action.yml composite step.
 */
import { readRawInputs, resolveInputs } from './lib/inputs.mjs'
import { runPipeline } from './lib/pipeline.mjs'

function main() {
  try {
    const inputs = resolveInputs(readRawInputs())
    runPipeline(inputs)
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    console.error(`::error::${message}`)
    const code =
      err && typeof err === 'object' && 'exitCode' in err && typeof err.exitCode === 'number'
        ? err.exitCode
        : 1
    process.exit(code)
  }
}

main()
