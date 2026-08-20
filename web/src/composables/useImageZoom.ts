import { computed, readonly, shallowRef } from 'vue'

const minimumScale = 1
const maximumScale = 4
const zoomStep = 0.5

interface Point { x: number; y: number }

interface PinchState {
  distance: number
  contentX: number
  contentY: number
}

export function useImageZoom() {
  const scale = shallowRef(minimumScale)
  const offsetX = shallowRef(0)
  const offsetY = shallowRef(0)
  const isInteracting = shallowRef(false)
  const pointers = new Map<number, Point>()
  let dragPoint: Point | undefined
  let pinch: PinchState | undefined
  let moved = false
  let lastTap: Point & { at: number } | undefined

  const isZoomed = computed(() => scale.value > minimumScale)
  const imageStyle = computed(() => ({ transform: `translate3d(${offsetX.value}px, ${offsetY.value}px, 0) scale(${scale.value})` }))

  function onPointerDown(event: PointerEvent): void {
    const target = event.currentTarget as HTMLElement
    if (event.cancelable) event.preventDefault()
    target.setPointerCapture(event.pointerId)
    const current = point(event)
    pointers.set(event.pointerId, current)
    isInteracting.value = true
    if (pointers.size === 1) {
      moved = false
      dragPoint = current
    } else if (pointers.size === 2) {
      moved = true
      startPinch(target)
    }
  }

  function onPointerMove(event: PointerEvent): void {
    if (!pointers.has(event.pointerId)) return
    const target = event.currentTarget as HTMLElement
    if (event.cancelable) event.preventDefault()
    const current = point(event)
    pointers.set(event.pointerId, current)
    if (pointers.size >= 2) {
      const distance = currentPointerDistance()
      if (!pinch || pinch.distance <= 0) startPinch(target)
      else {
        const midpoint = relativePoint(currentPointerMidpoint(), target)
        const nextScale = clampScale(scale.value * distance / pinch.distance)
        scale.value = nextScale
        offsetX.value = midpoint.x - pinch.contentX * nextScale
        offsetY.value = midpoint.y - pinch.contentY * nextScale
        clampOffset(target)
        startPinch(target)
      }
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
    if (pointers.size >= 2) startPinch(target)
    else if (pointers.size === 1) {
      pinch = undefined
      dragPoint = pointers.values().next().value
    }
    if (!wasGesture && event.pointerType === 'touch') doubleTap(event, target)
    if (pointers.size === 0) {
      dragPoint = undefined
      pinch = undefined
      isInteracting.value = false
    }
  }

  function onDoubleClick(event: MouseEvent): void {
    toggle(event.currentTarget as HTMLElement, point(event))
  }

  function onWheel(event: WheelEvent): void {
    const factor = Math.exp(-event.deltaY * 0.002)
    setScale(scale.value * factor, event.currentTarget as HTMLElement, point(event))
  }

  function zoomIn(target?: HTMLElement): void { setScale(scale.value + zoomStep, target) }
  function zoomOut(target?: HTMLElement): void { setScale(scale.value - zoomStep, target) }
  function reset(): void {
    scale.value = minimumScale
    offsetX.value = 0
    offsetY.value = 0
    pointers.clear()
    dragPoint = undefined
    pinch = undefined
    isInteracting.value = false
  }

  function doubleTap(event: PointerEvent, target: HTMLElement): void {
    const current = { ...point(event), at: Date.now() }
    if (lastTap && current.at - lastTap.at < 280 && distance(current, lastTap) < 24) {
      lastTap = undefined
      toggle(target, current)
      return
    }
    lastTap = current
  }

  function startPinch(target: HTMLElement): void {
    const midpoint = relativePoint(currentPointerMidpoint(), target)
    pinch = {
      distance: currentPointerDistance(),
      contentX: (midpoint.x - offsetX.value) / scale.value,
      contentY: (midpoint.y - offsetY.value) / scale.value,
    }
    dragPoint = undefined
  }

  function toggle(target?: HTMLElement, focus?: Point): void {
    if (isZoomed.value) reset()
    else setScale(2, target, focus)
  }

  function setScale(next: number, target?: HTMLElement, focus?: Point): void {
    const previous = scale.value
    const normalized = clampScale(next)
    if (target && normalized !== previous) {
      const bounds = target.getBoundingClientRect()
      const relative = focus
        ? { x: focus.x - bounds.left - bounds.width / 2, y: focus.y - bounds.top - bounds.height / 2 }
        : { x: 0, y: 0 }
      const contentX = (relative.x - offsetX.value) / previous
      const contentY = (relative.y - offsetY.value) / previous
      offsetX.value = relative.x - contentX * normalized
      offsetY.value = relative.y - contentY * normalized
    }
    scale.value = normalized
    if (scale.value === minimumScale) {
      offsetX.value = 0
      offsetY.value = 0
    } else if (target) clampOffset(target)
  }

  function clampOffset(target: HTMLElement): void {
    const bounds = target.getBoundingClientRect()
    const image = target.querySelector('img')
    const baseWidth = image?.offsetWidth || bounds.width
    const baseHeight = image?.offsetHeight || bounds.height
    const maxX = Math.max(0, (baseWidth * scale.value - bounds.width) / 2)
    const maxY = Math.max(0, (baseHeight * scale.value - bounds.height) / 2)
    offsetX.value = Math.min(maxX, Math.max(-maxX, offsetX.value))
    offsetY.value = Math.min(maxY, Math.max(-maxY, offsetY.value))
  }

  function currentPointerDistance(): number {
    const [first, second] = [...pointers.values()]
    return first && second ? distance(first, second) : 0
  }

  function currentPointerMidpoint(): Point {
    const [first, second] = [...pointers.values()]
    return first && second ? { x: (first.x + second.x) / 2, y: (first.y + second.y) / 2 } : { x: 0, y: 0 }
  }

  return {
    scale: readonly(scale), isZoomed, isInteracting: readonly(isInteracting), imageStyle,
    onPointerDown, onPointerMove, onPointerUp, onDoubleClick, onWheel, zoomIn, zoomOut, reset,
  }
}

function point(event: Pick<MouseEvent, 'clientX' | 'clientY'>): Point { return { x: event.clientX, y: event.clientY } }
function relativePoint(value: Point, target: HTMLElement): Point {
  const bounds = target.getBoundingClientRect()
  return { x: value.x - bounds.left - bounds.width / 2, y: value.y - bounds.top - bounds.height / 2 }
}
function clampScale(value: number): number { return Math.min(maximumScale, Math.max(minimumScale, value)) }
function distance(first: Point, second: Point): number { return Math.hypot(first.x - second.x, first.y - second.y) }
