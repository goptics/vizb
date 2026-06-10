import { CanvasRenderer } from 'echarts/renderers'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  ToolboxComponent,
} from 'echarts/components'

// Universal 2D ECharts modules shared by every 2D renderer (bar/line/pie).
// Importing this from each per-type component lets rollup hoist the shared
// echarts core / zrender / vue-echarts mass into one chunk, while each chart's
// own module body stays in its own lazily-parsed chunk.
export const BASE_2D = [
  CanvasRenderer,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  ToolboxComponent,
]
