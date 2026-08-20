import { onBeforeUnmount, shallowRef, useTemplateRef } from 'vue'

const daySelector = '[data-statistics-traffic-day]'
const segmentSelector = '[data-statistics-traffic-id]'

interface TrafficScrubOptions {
  activate: (interactionId: string) => void
  deactivate: (interactionId: string) => void
  select: (interactionId: string) => void
}

function axisDistance(point: number, start: number, end: number): number {
  if (point < start) return start - point
  if (point > end) return point - end
  return 0
}

function closestElement(elements: HTMLElement[], point: number, axis: 'x' | 'y'): HTMLElement | undefined {
  return elements.reduce<HTMLElement | undefined>((closest, element) => {
    if (!closest) return element
    const rect = element.getBoundingClientRect()
    const closestRect = closest.getBoundingClientRect()
    const distance = axis === 'x'
      ? axisDistance(point, rect.left, rect.right)
      : axisDistance(point, rect.top, rect.bottom)
    const closestDistance = axis === 'x'
      ? axisDistance(point, closestRect.left, closestRect.right)
      : axisDistance(point, closestRect.top, closestRect.bottom)
    return distance < closestDistance ? element : closest
  }, undefined)
}

export function findClosestTrafficInteraction(root: HTMLElement, clientX: number, clientY: number): string | undefined {
  const days = Array.from(root.querySelectorAll<HTMLElement>(daySelector))
  const day = closestElement(days, clientX, 'x')
  if (!day) return undefined
  const segments = Array.from(day.querySelectorAll<HTMLElement>(segmentSelector))
  return closestElement(segments, clientY, 'y')?.dataset.statisticsTrafficId
}

export function useStatisticsTrafficScrub(options: TrafficScrubOptions) {
  const plot = useTemplateRef<HTMLElement>('trafficPlot')
  const activePointerId = shallowRef<number>()
  const isScrubbing = shallowRef(false)
  let previewedId: string | undefined
  let suppressClick = false
  let clickResetTimer: ReturnType<typeof setTimeout> | undefined

  function interactionAt(event: PointerEvent | MouseEvent): string | undefined {
    return plot.value
      ? findClosestTrafficInteraction(plot.value, event.clientX, event.clientY)
      : undefined
  }

  function preview(event: PointerEvent): string | undefined {
    const interactionId = interactionAt(event)
    if (!interactionId || interactionId === previewedId) return interactionId
    if (previewedId) options.deactivate(previewedId)
    options.activate(interactionId)
    previewedId = interactionId
    return interactionId
  }

  function clearPreview(): void {
    if (previewedId) options.deactivate(previewedId)
    previewedId = undefined
  }

  function endPointer(event: PointerEvent): void {
    if (event.pointerId !== activePointerId.value) return
    const interactionId = preview(event)
    if (interactionId) options.select(interactionId)
    clearPreview()
    activePointerId.value = undefined
    isScrubbing.value = false
    suppressClick = true
    if (clickResetTimer) clearTimeout(clickResetTimer)
    clickResetTimer = setTimeout(() => { suppressClick = false }, 400)
  }

  function onPointerDown(event: PointerEvent): void {
    if (event.pointerType === 'mouse' || !event.isPrimary) return
    activePointerId.value = event.pointerId
    isScrubbing.value = true
    plot.value?.setPointerCapture?.(event.pointerId)
    preview(event)
  }

  function onPointerMove(event: PointerEvent): void {
    if (event.pointerId === activePointerId.value) preview(event)
  }

  function onPointerCancel(event: PointerEvent): void {
    if (event.pointerId !== activePointerId.value) return
    clearPreview()
    activePointerId.value = undefined
    isScrubbing.value = false
  }

  function onClick(event: MouseEvent): void {
    if (suppressClick) {
      suppressClick = false
      event.preventDefault()
      event.stopPropagation()
      return
    }
    const interactionId = interactionAt(event)
    if (!interactionId) return
    options.select(interactionId)
    options.deactivate(interactionId)
  }

  onBeforeUnmount(() => {
    if (clickResetTimer) clearTimeout(clickResetTimer)
  })

  return { plot, isScrubbing, onPointerDown, onPointerMove, onPointerUp: endPointer, onPointerCancel, onClick }
}
