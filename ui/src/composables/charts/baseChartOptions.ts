import type { Ref } from 'vue'
import type { EChartsOption } from 'echarts'
import type { Axis, ChartData, Sort, ScaleType, ChartType } from '@/types'
import { createTooltipConfig, createToolboxConfig, getChartStyling } from './shared/chartConfig'
import { fontSize } from './shared/common'
import { is3D } from '@/lib/utils'

export interface BaseChartConfig {
  chartData: Ref<ChartData>
  sort: Ref<Sort>
  showLabels: Ref<boolean>
  isDark: Ref<boolean>
  // `scale` and `threeDRotate` only apply to bar/line (the chart types whose
  // config carries them). Pie/heatmap/radar configs don't produce a Ref for
  // these fields, so they're relaxed to optional here; chart composables that
  // access them default at the call site (e.g. `scale?.value ?? "linear"`).
  scale?: Ref<ScaleType>
  threeDRotate?: Ref<boolean>
  // Legend selection state for z series (3D); key = z name, false = hidden.
  visibleZ?: Ref<Record<string, boolean>>
  // Runtime value-3D toggle (x+y only, when threeD is baked via --3d).
  threeD?: Ref<boolean>
  threeDVisualMap?: Ref<boolean>
  visualMap?: Ref<boolean>
  /** CLI-baked marker overrides (--symbol / --symbol-size); line and scatter only. */
  symbol?: Ref<string | undefined>
  symbolSize?: Ref<number | undefined>
  /** 2D line-only curve toggle. Ignored by 3D renderers. */
  smooth?: Ref<boolean>
  /** 2D bar horizontal orientation. When true, bars grow rightward. */
  horizontal?: Ref<boolean>
  /** Active swap target (e.g. xyz) — scatter value-mode 3D is swap-driven. */
  arrangementTarget?: Ref<string>
  chartAxes?: Ref<Axis[] | undefined>
  chartType?: Ref<ChartType>
}

export const getBaseOptions = (config: BaseChartConfig): Partial<EChartsOption> => {
  const { isDark } = config
  const { textColor, backgroundColor } = getChartStyling(isDark.value)
  // zrender's getRenderedCanvas only composites WebGL (echarts-gl) layers when
  // saveAsImage's pixelRatio <= the chart dpr; a larger ratio falls back to a
  // 2D-only redraw that drops the 3D content. Cap 3D exports at the device dpr.
  const dpr = window.devicePixelRatio || 1
  const saveAsImagePixelRatio = is3D(
    config.chartData,
    config.threeD?.value,
    config.arrangementTarget?.value,
    config.chartAxes?.value,
    config.chartType?.value
  )
    ? dpr
    : 2
  return {
    backgroundColor,
    tooltip: createTooltipConfig(false, isDark.value) as EChartsOption['tooltip'],
    toolbox: createToolboxConfig(isDark.value, config.chartData.value.title, saveAsImagePixelRatio),
    legend: {
      show: true,
      left: 'center',
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { fontSize, color: textColor },
    },
    emphasis: {
      focus: 'series',
    },
  }
}
