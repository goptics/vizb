import { ref } from 'vue'
import type { EChartsOption } from 'echarts'

const ENTER_FULLSCREEN_ICON = 'path://M3 9V3H9 M21 9V3H15 M3 15V21H9 M21 15V21H15'
const EXIT_FULLSCREEN_ICON = 'path://M9 3V9H3 M15 3V9H21 M9 21V15H3 M15 21V15H21'

export function useFullscreen() {
  const containerRef = ref<HTMLElement | null>(null)
  const isFullscreen = ref(false)

  function toggleFullscreen() {
    if (!containerRef.value) return
    if (!document.fullscreenElement) {
      containerRef.value.requestFullscreen()
    } else {
      document.exitFullscreen()
    }
  }

  document.addEventListener('fullscreenchange', () => {
    isFullscreen.value = document.fullscreenElement === containerRef.value
  })

  function withFullscreenToolbox(option: EChartsOption): EChartsOption {
    const toolbox = option.toolbox as Record<string, unknown> | undefined
    const feature = (toolbox?.feature ?? {}) as Record<string, unknown>
    return {
      ...option,
      toolbox: {
        ...toolbox,
        feature: {
          ...feature,
          myFullScreen: {
            show: true,
            title: isFullscreen.value ? 'Exit fullscreen' : 'Fullscreen',
            icon: isFullscreen.value ? EXIT_FULLSCREEN_ICON : ENTER_FULLSCREEN_ICON,
            onclick: toggleFullscreen,
          },
        },
      },
    } as EChartsOption
  }

  return { containerRef, isFullscreen, withFullscreenToolbox }
}
