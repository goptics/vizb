import { describe, expect, it } from 'vitest'
import * as Index from './index.ts'
import * as Html from './html.ts'
import * as GoCodegen from './go-codegen.ts'
import * as Constants from './constants.ts'
import * as EmbedUi from './plugins/embed-ui.ts'
import * as InlineFavicon from './plugins/inline-favicon.ts'
// Pure type module: import for side-effect so coverage maps the file.
import './types.ts'

describe('embed-build barrels', () => {
  it('exports plugin factory and build options', () => {
    expect(typeof Index.createEmbedPlugins).toBe('function')
    expect(Index.embedBuildOptions.cssCodeSplit).toBe(false)
    expect(typeof Html.appendVizbDataScriptTag).toBe('function')
    expect(typeof GoCodegen.chunkKeyOf).toBe('function')
    expect(Constants.CHART_ROOT_PREFIX.ChartBar).toBe('bar')
    expect(typeof EmbedUi.embedUiPlugin).toBe('function')
    expect(typeof InlineFavicon.inlineFaviconPlugin).toBe('function')
  })
})
