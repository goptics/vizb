import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useFullscreen } from './useFullscreen'

describe('useFullscreen', () => {
  let listeners: Record<string, EventListenerOrEventListenerObject[]>
  let addSpy: ReturnType<typeof vi.spyOn>
  let exitFullscreen: ReturnType<typeof vi.fn>
  let fullscreenEl: Element | null

  beforeEach(() => {
    listeners = {}
    fullscreenEl = null
    exitFullscreen = vi.fn().mockResolvedValue(undefined)

    Object.defineProperty(document, 'fullscreenElement', {
      configurable: true,
      get: () => fullscreenEl,
    })
    Object.defineProperty(document, 'exitFullscreen', {
      configurable: true,
      writable: true,
      value: exitFullscreen,
    })

    addSpy = vi.spyOn(document, 'addEventListener').mockImplementation((type, listener) => {
      ;(listeners[type] ??= []).push(listener)
    })
  })

  afterEach(() => {
    addSpy.mockRestore()
    vi.restoreAllMocks()
  })

  function fireFullscreenChange() {
    for (const listener of listeners.fullscreenchange ?? []) {
      if (typeof listener === 'function') listener(new Event('fullscreenchange'))
      else listener.handleEvent(new Event('fullscreenchange'))
    }
  }

  function featureOnclick(
    option: ReturnType<ReturnType<typeof useFullscreen>['withFullscreenToolbox']>
  ) {
    const toolbox = option.toolbox
    if (!toolbox || typeof toolbox !== 'object' || !('feature' in toolbox)) {
      throw new Error('missing toolbox.feature')
    }
    const feature = toolbox.feature as {
      myFullScreen?: { onclick?: () => void; title?: string; icon?: string; show?: boolean }
      saveAsImage?: { show?: boolean }
    }
    const my = feature.myFullScreen
    if (!my?.onclick) throw new Error('missing myFullScreen.onclick')
    return { feature, my, onclick: my.onclick }
  }

  it('registers fullscreenchange listener and starts not fullscreen', () => {
    const { isFullscreen, containerRef } = useFullscreen()
    expect(isFullscreen.value).toBe(false)
    expect(containerRef.value).toBeNull()
    expect(addSpy).toHaveBeenCalledWith('fullscreenchange', expect.any(Function))
  })

  it('toggleFullscreen is a no-op without a container', () => {
    const { withFullscreenToolbox } = useFullscreen()
    fullscreenEl = null
    const { onclick } = featureOnclick(withFullscreenToolbox({ toolbox: { feature: {} } } as never))
    onclick()
    expect(exitFullscreen).not.toHaveBeenCalled()
  })

  it('enters fullscreen when no element is fullscreen', () => {
    const { containerRef, withFullscreenToolbox } = useFullscreen()
    const el = document.createElement('div')
    const request = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(el, 'requestFullscreen', {
      configurable: true,
      value: request,
    })
    containerRef.value = el
    fullscreenEl = null

    featureOnclick(withFullscreenToolbox({} as never)).onclick()
    expect(request).toHaveBeenCalledOnce()
  })

  it('exits fullscreen when something is already fullscreen', () => {
    const { containerRef, withFullscreenToolbox } = useFullscreen()
    const el = document.createElement('div')
    Object.defineProperty(el, 'requestFullscreen', {
      configurable: true,
      value: vi.fn(),
    })
    containerRef.value = el
    fullscreenEl = el

    featureOnclick(withFullscreenToolbox({} as never)).onclick()
    expect(exitFullscreen).toHaveBeenCalledOnce()
  })

  it('updates isFullscreen from fullscreenchange against container', () => {
    const { containerRef, isFullscreen } = useFullscreen()
    const el = document.createElement('div')
    containerRef.value = el

    fullscreenEl = el
    fireFullscreenChange()
    expect(isFullscreen.value).toBe(true)

    fullscreenEl = document.createElement('span')
    fireFullscreenChange()
    expect(isFullscreen.value).toBe(false)

    fullscreenEl = null
    fireFullscreenChange()
    expect(isFullscreen.value).toBe(false)
  })

  it('merges toolbox feature and reflects enter/exit titles and icons', () => {
    const api = useFullscreen()
    const el = document.createElement('div')
    api.containerRef.value = el

    const enter = api.withFullscreenToolbox({
      title: { text: 't' },
      toolbox: { right: 8, feature: { saveAsImage: { show: true } } },
    } as never)

    expect(enter.title).toEqual({ text: 't' })
    const { feature, my } = featureOnclick(enter)
    expect(feature.saveAsImage).toEqual({ show: true })
    expect(my.show).toBe(true)
    expect(my.title).toBe('Fullscreen')
    expect(my.icon).toContain('path://M3 9V3H9')
    expect((enter.toolbox as { right: number }).right).toBe(8)

    fullscreenEl = el
    fireFullscreenChange()

    const exit = api.withFullscreenToolbox({ toolbox: undefined } as never)
    const exitMy = featureOnclick(exit).my
    expect(exitMy.title).toBe('Exit fullscreen')
    expect(exitMy.icon).toContain('path://M9 3V9H3')
  })
})
