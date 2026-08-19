import { computed, readonly, shallowRef } from 'vue'

const minimumScale = 1
const maximumScale = 4
const zoomStep = 0.5

interface Point { x: number; y: number }

export function useImageZoom() {
  const scale = shallowRef(minimumScale)
  const offsetX = shallowRef(0)
  const offsetY = shallowRef(0)
  const pointers = new Map<number, Point>()
  let dragPoint: Point | undefined
  let pinchDistance = 0
  let pinchScale = minimumScale
  let moved = false
  let lastTap: Point & { at: number } | undefined

  const isZoomed = computed(() => scale.value > minimumScale)
  const imageStyle = computed(() => ({ transform: `translate3d(${offsetX.value}px, ${offsetY.value}px, 0) scale(${scale.value})` }))

  function onPointerDown(event: PointerEvent): void {
    const target = event.currentTarget as HTMLElement
    target.setPointerCapture(event.pointerId)
    pointers.set(event.pointerId, point(event))
    moved = false
    if (pointers.size === 1) dragPoint = point(event)
    if (pointers.size === 2) startPinch()
  }

  function onPointerMove(event: PointerEvent): void {
    if (!pointers.has(event.pointerId)) return
    const target = event.currentTarget as HTMLElement
    const current = point(event)
    pointers.set(event.pointerId, current)
    if (pointers.size === 2) {
      const distance = currentPointerDistance()
      if (pinchDistance > 0) scale.value = clampScale(pinchScale * distance / pinchDistance)
      clampOffset(target)
      moved = true
      return
    }
    if (!dragPoint || !isZoomed.value) return
    offsetX.value += current.x - dragPoint.x
    offsetY.value += current.y - dragPoint.y
    dragPoint = current
    clampOffset(target)
    moved = true
  }

  function onPointerUp(event: PointerEvent): void {
    const target = event.currentTarget as HTMLElement
    const wasGesture = moved || pointers.size > 1
    pointers.delete(event.pointerId)
    if (target.hasPointerCapture(event.pointerId)) target.releasePointerCapture(event.pointerId)
    if (pointers.size === 1) dragPoint = pointers.values().next().value
    if (!wasGesture && event.pointerType === 'touch') doubleTap(event, target)
    if (pointers.size === 0) dragPoint = undefined
  }

  function onDoubleClick(event: MouseEvent): void {
    toggle(event.currentTarget as HTMLElement)
  }

  function zoomIn(target?: HTMLElement): void { setScale(scale.value + zoomStep, target) }
  function zoomOut(target?: HTMLElement): void { setScale(scale.value - zoomStep, target) }
  function reset(): void { scale.value = minimumScale; offsetX.value = 0; offsetY.value = 0 }

  function doubleTap(event: PointerEvent, target: HTMLElement): void {
    const current = { ...point(event), at: Date.now() }
    if (lastTap && current.at - lastTap.at < 280 && distance(current, lastTap) < 24) {
      lastTap = undefined
      toggle(target)
      return
    }
    lastTap = current
  }

  function startPinch(): void {
    pinchDistance = currentPointerDistance()
    pinchScale = scale.value
    dragPoint = undefined
  }

  function toggle(target?: HTMLElement): void {
    if (isZoomed.value) reset()
    else setScale(2, target)
  }

  function setScale(next: number, target?: HTMLElement): void {
    scale.value = clampScale(next)
    if (scale.value === minimumScale) {
      offsetX.value = 0
      offsetY.value = 0
    } else if (target) clampOffset(target)
  }

  function clampOffset(target: HTMLElement): void {
    const bounds = target.getBoundingClientRect()
    const maxX = bounds.width * (scale.value - 1) / 2
    const maxY = bounds.height * (scale.value - 1) / 2
    offsetX.value = Math.min(maxX, Math.max(-maxX, offsetX.value))
    offsetY.value = Math.min(maxY, Math.max(-maxY, offsetY.value))
  }

  function currentPointerDistance(): number {
    const [first, second] = [...pointers.values()]
    return first && second ? distance(first, second) : 0
  }

  return { scale: readonly(scale), isZoomed, imageStyle, onPointerDown, onPointerMove, onPointerUp, onDoubleClick, zoomIn, zoomOut, reset }
}

function point(event: PointerEvent): Point { return { x: event.clientX, y: event.clientY } }
function clampScale(value: number): number { return Math.min(maximumScale, Math.max(minimumScale, value)) }
function distance(first: Point, second: Point): number { return Math.hypot(first.x - second.x, first.y - second.y) }
