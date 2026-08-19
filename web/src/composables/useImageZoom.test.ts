import { describe, expect, it } from 'vitest'

import { useImageZoom } from './useImageZoom'

describe('useImageZoom', () => {
  it('clamps zoom controls and resets the view state', () => {
    const zoom = useImageZoom()
    zoom.zoomOut()
    expect(zoom.scale.value).toBe(1)
    zoom.zoomIn()
    zoom.zoomIn()
    expect(zoom.scale.value).toBe(2)
    expect(zoom.isZoomed.value).toBe(true)
    zoom.reset()
    expect(zoom.scale.value).toBe(1)
    expect(zoom.isZoomed.value).toBe(false)
  })

  it('updates zoom from a two-finger gesture', () => {
    const zoom = useImageZoom()
    const captured = new Set<number>()
    const target = {
      setPointerCapture: (pointerId: number) => captured.add(pointerId),
      hasPointerCapture: (pointerId: number) => captured.has(pointerId),
      releasePointerCapture: (pointerId: number) => captured.delete(pointerId),
      getBoundingClientRect: () => new DOMRect(0, 0, 200, 100),
    } as unknown as HTMLElement
    const pointer = (pointerId: number, clientX: number) => ({
      pointerId, clientX, clientY: 0, pointerType: 'touch', currentTarget: target,
    }) as unknown as PointerEvent

    zoom.onPointerDown(pointer(1, 0))
    zoom.onPointerDown(pointer(2, 100))
    zoom.onPointerMove(pointer(2, 200))

    expect(zoom.scale.value).toBe(2)
    expect(zoom.isZoomed.value).toBe(true)
    zoom.onPointerUp(pointer(2, 200))
  })
})
