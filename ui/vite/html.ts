import { parseHTML } from 'linkedom'

export const appendVizbDataScriptTag = (html: string): string => {
  const { document } = parseHTML(html)

  const script = document.createElement('script')
  script.type = 'text/javascript'
  // Custom Go-template delimiters: echarts-gl's clay.gl GLSL shaders use {{ }}
  // for loop unrolling, which would otherwise collide with html/template parsing.
  script.textContent = `window.VIZB_VERSION = [[VIZB .Version VIZB]]; window.VIZB_DATA = [[VIZB .Data VIZB]]; window.VIZB_DATA_URL = [[VIZB .DataURL VIZB]]; window.VIZB_CHARTS = [[VIZB .ChartList VIZB]];`

  document.head.appendChild(script)

  return '<!DOCTYPE html>' + document.documentElement.outerHTML
}
