#!/usr/bin/env node
import { resolveInputs, runPipeline } from './lib.mjs'

try {
  runPipeline(resolveInputs())
} catch (err) {
  const message = err instanceof Error ? err.message : String(err)
  console.error(`::error::${message}`)
  const code =
    err && typeof err === 'object' && 'exitCode' in err && typeof err.exitCode === 'number'
      ? err.exitCode
      : 1
  process.exit(code)
}
