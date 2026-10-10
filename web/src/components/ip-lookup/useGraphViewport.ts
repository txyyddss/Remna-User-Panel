import { shallowRef } from 'vue'

export function useGraphViewport() {
  const scale = shallowRef(1)
  const offset = shallowRef({ x: 0, y: 0 })
  let drag: { id: number; x: number; y: number; origin: { x: number; y: number } } | null = null
  function reset() { scale.value = 1; offset.value = { x: 0, y: 0 } }
  function zoom(factor: number) { scale.value = Math.min(3, Math.max(0.35, scale.value * factor)) }
  function pointerDown(event: PointerEvent) {
    if ((event.target as Element).closest('[data-asn]')) return
    const target = event.currentTarget as HTMLElement
    target.setPointerCapture(event.pointerId)
    drag = { id: event.pointerId, x: event.clientX, y: event.clientY, origin: offset.value }
  }
  function pointerMove(event: PointerEvent) {
    if (!drag || drag.id !== event.pointerId) return
    offset.value = { x: drag.origin.x + event.clientX - drag.x, y: drag.origin.y + event.clientY - drag.y }
  }
  function pointerUp() { drag = null }
  function keyboard(event: KeyboardEvent) {
    const arrows: Record<string, [number, number]> = { ArrowLeft: [30, 0], ArrowRight: [-30, 0], ArrowUp: [0, 30], ArrowDown: [0, -30] }
    if (arrows[event.key]) { event.preventDefault(); const [x, y] = arrows[event.key]!; offset.value = { x: offset.value.x + x, y: offset.value.y + y } }
    if (event.key === '+' || event.key === '=') zoom(1.2)
    if (event.key === '-') zoom(1 / 1.2)
    if (event.key === '0') reset()
  }
  return { scale, offset, reset, zoom, pointerDown, pointerMove, pointerUp, keyboard }
}
