import { describe, expect, it } from 'vitest'
import { extractPathDatasetId } from './pathRoute'

describe('extractPathDatasetId', () => {
  it.each([
    ['/', null],
    ['/index', null],
    ['/report.html', null],
    ['/assets/app.js', null],
    ['/go-1.25', 'go-1.25'],
    ['/apps/vizb/go-1.25%2Famd64', 'go-1.25/amd64'],
  ])('extracts the dataset identity from %s', (pathname, expected) => {
    expect(extractPathDatasetId(pathname)).toBe(expected)
  })
})
