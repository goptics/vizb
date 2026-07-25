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
    ['/%E0%A4%A', null], // invalid URI sequence → catch
  ])('extracts the dataset identity from %s', (pathname, expected) => {
    expect(extractPathDatasetId(pathname)).toBe(expected)
  })

  it('returns null when no path segments remain', () => {
    expect(extractPathDatasetId('///')).toBe(null)
  })
})
