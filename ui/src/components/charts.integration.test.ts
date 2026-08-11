import { describe, it, expect, vi, beforeEach } from 'vitest'
import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'

const useMock = vi.fn()
vi.mock('echarts/core', () => ({ use: (...args: unknown[]) => useMock(...args) }))
vi.mock('echarts/renderers', () => ({ CanvasRenderer: 'CanvasRenderer' }))
vi.mock('echarts/components', () => ({
  TitleComponent: 'TitleComponent',
  TooltipComponent: 'TooltipComponent',
  LegendComponent: 'LegendComponent',
  ToolboxComponent: 'ToolboxComponent',
  DataZoomComponent: 'DataZoomComponent',
  GridComponent: 'GridComponent',
  VisualMapComponent: 'VisualMapComponent',
  RadarComponent: 'RadarComponent',
}))
vi.mock('echarts/charts', () => ({
  BarChart: 'BarChart',
  LineChart: 'LineChart',
  ScatterChart: 'ScatterChart',
  PieChart: 'PieChart',
  HeatmapChart: 'HeatmapChart',
  RadarChart: 'RadarChart',
  SankeyChart: 'SankeyChart',
  ChordChart: 'ChordChart',
}))
vi.mock('echarts-gl/charts', () => ({
  Bar3DChart: 'Bar3DChart',
  Line3DChart: 'Line3DChart',
  Scatter3DChart: 'Scatter3DChart',
}))
vi.mock('echarts-gl/components', () => ({ Grid3DComponent: 'Grid3DComponent' }))
vi.mock('vue-echarts', () => ({
  default: defineComponent({
    name: 'VChart',
    props: ['option', 'initOptions', 'autoresize', 'updateOptions'],
    emits: ['legendselectchanged'],
    setup(props, { emit }) {
      return () =>
        h('div', {
          'data-testid': 'vchart',
          'data-not-merge': props.updateOptions?.notMerge === false ? '0' : '1',
          onClick: () => emit('legendselectchanged', { selected: { A: true } }),
          onDblclick: () => emit('legendselectchanged', { selected: { A: 'yes' } }),
        })
    },
  }),
}))

import ChartBar from './ChartBar.vue'
import ChartLine from './ChartLine.vue'
import ChartPie from './ChartPie.vue'
import ChartScatter from './ChartScatter.vue'
import ChartHeatmap from './ChartHeatmap.vue'
import ChartRadar from './ChartRadar.vue'
import ChartSankey from './ChartSankey.vue'
import ChartChord from './ChartChord.vue'
import Chart3D from './Chart3D.vue'
import { BASE_2D } from './charts/base'

describe('chart shells', () => {
  beforeEach(() => {
    useMock.mockClear()
  })

  const option = { title: { text: 't' } }
  const initOptions = { renderer: 'canvas' }

  it.each([
    ['ChartBar', ChartBar],
    ['ChartLine', ChartLine],
    ['ChartPie', ChartPie],
    ['ChartScatter', ChartScatter],
    ['ChartHeatmap', ChartHeatmap],
    ['ChartRadar', ChartRadar],
    ['ChartSankey', ChartSankey],
    ['ChartChord', ChartChord],
    ['Chart3D', Chart3D],
  ] as const)('%s mounts VChart and forwards legend event', async (name, Comp) => {
    const onLegendselectchanged = vi.fn()
    const w = mount(Comp, {
      props: { option, initOptions, onLegendselectchanged },
    })
    expect(w.find('[data-testid="vchart"]').exists()).toBe(true)
    await w.get('[data-testid="vchart"]').trigger('click')
    expect(onLegendselectchanged).toHaveBeenCalledWith({ selected: { A: true } })
    await w.get('[data-testid="vchart"]').trigger('dblclick')
    expect(onLegendselectchanged).toHaveBeenCalledTimes(1)
    expect(useMock).toHaveBeenCalled()
    expect(name).toBeTruthy()
  })

  it('exports BASE_2D modules', () => {
    expect(BASE_2D.length).toBeGreaterThan(0)
    expect(BASE_2D).toContain('CanvasRenderer')
  })

  it('Chart3D uses merge update options', () => {
    const w = mount(Chart3D, { props: { option, initOptions } })
    expect(w.get('[data-testid="vchart"]').attributes('data-not-merge')).toBe('0')
  })
})
