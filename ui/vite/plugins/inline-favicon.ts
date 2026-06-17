import fs from 'node:fs'
import path from 'path'
import { parseHTML } from 'linkedom'
import type { PluginOption } from 'vite'

export const inlineFaviconPlugin = (rootDir: string): PluginOption => {
  return {
    name: 'inline-favicon',
    enforce: 'pre',
    transformIndexHtml(html: string) {
      const { document } = parseHTML(html)
      const link = document.querySelector('link[rel="icon"]')

      if (!link) {
        return html
      }

      const href = link.getAttribute('href')
      if (!href) return html

      // Strip the leading "/" so the href resolves relative to the project
      // (path.resolve treats a leading slash as an absolute path and discards
      // the base), checking the public directory if needed.
      const relHref = href.replace(/^\//, '')
      let faviconPath = path.resolve(rootDir, relHref)

      if (!fs.existsSync(faviconPath)) {
        faviconPath = path.resolve(rootDir, 'public', relHref)
      }

      if (!fs.existsSync(faviconPath)) {
        console.warn(`[inline-favicon] Could not locate favicon: ${href}`)
        return html
      }

      const source = fs.readFileSync(faviconPath, 'utf-8')
      link.setAttribute('href', `data:${link.getAttribute('type')},${encodeURIComponent(source)}`)

      return '<!DOCTYPE html>' + document.documentElement.outerHTML
    },
  }
}