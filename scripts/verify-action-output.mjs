import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

const { EXPECTED_JSON_KIND, EXPECTED_NAMES, HTML_FILE, JSON_FILE } = process.env

assert.ok(HTML_FILE, 'HTML_FILE is required')
assert.ok(JSON_FILE, 'JSON_FILE is required')
assert.ok(EXPECTED_NAMES, 'EXPECTED_NAMES is required')
assert.ok(
  EXPECTED_JSON_KIND === 'object' || EXPECTED_JSON_KIND === 'array',
  'EXPECTED_JSON_KIND must be "object" or "array"',
)

const expectedNames = JSON.parse(EXPECTED_NAMES)
assert.ok(Array.isArray(expectedNames), 'EXPECTED_NAMES must be a JSON array')

const json = JSON.parse(readFileSync(JSON_FILE, 'utf8'))
if (EXPECTED_JSON_KIND === 'array') {
  assert.ok(Array.isArray(json), `${JSON_FILE} must contain a dataset array`)
} else {
  assert.ok(!Array.isArray(json), `${JSON_FILE} must contain a single dataset`)
}

const datasets = Array.isArray(json) ? json : [json]
assert.deepEqual(
  datasets.map(({ name }) => name).sort(),
  [...expectedNames].sort(),
  `${JSON_FILE} contains unexpected datasets`,
)

for (const dataset of datasets) {
  assert.ok(Array.isArray(dataset.data), `${dataset.name} must contain data`)
  assert.ok(dataset.data.length > 0, `${dataset.name} data must not be empty`)
  assert.ok(Array.isArray(dataset.settings), `${dataset.name} must contain settings`)
  assert.ok(dataset.settings.length > 0, `${dataset.name} settings must not be empty`)
}

const html = readFileSync(HTML_FILE, 'utf8')
assert.match(html, /^<!DOCTYPE html>/, `${HTML_FILE} must be an HTML document`)
assert.ok(html.includes('window.VIZB_DATA = '), `${HTML_FILE} must embed vizb data`)

for (const name of expectedNames) {
  assert.ok(html.includes(JSON.stringify(name)), `${HTML_FILE} must contain ${name}`)
}

console.log(`Verified ${JSON_FILE} and ${HTML_FILE}`)
