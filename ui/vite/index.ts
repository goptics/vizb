import type { PluginOption } from 'vite'
import { inlineFaviconPlugin } from './plugins/inline-favicon.ts'
import { embedUiPlugin } from './plugins/embed-ui.ts'

// In embed mode we own the inlining (embedUiPlugin): keep CSS in one file to
// inline, drop modulepreload (its dep arrays would defy the import map), and
// inline any static assets so nothing external is referenced. Dynamic imports
// are intentionally NOT inlined — that's what keeps echarts-gl in its own
// lazily-parsed chunk.
export const embedBuildOptions = {
  cssCodeSplit: false,
  modulePreload: false as const,
  assetsInlineLimit: 100_000_000,
}

export const createEmbedPlugins = (rootDir: string): PluginOption[] => [
  inlineFaviconPlugin(rootDir),
  embedUiPlugin(rootDir),
]
