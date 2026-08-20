import { describe, expect, it } from 'vitest'

import { findClosestTrafficInteraction } from './useStatisticsTrafficScrub'

function rect(left: number, top: number, right: number, bottom: number): DOMRect {
  return { left, top, right, bottom, width: right - left, height: bottom - top, x: left, y: top, toJSON: () => ({}) }
}

function trafficElement(attributes: Record<string, string>, bounds: DOMRect): HTMLDivElement {
  const element = document.createElement('div')
  Object.entries(attributes).forEach(([name, value]) => element.setAttribute(name, value))
  element.getBoundingClientRect = () => bounds
  return element
}

describe('statistics traffic touch hit testing', () => {
  it('selects the closest rendered segment within the nearest day', () => {
    const plot = document.createElement('div')
    const firstDay = trafficElement({ 'data-statistics-traffic-day': '' }, rect(0, 0, 40, 180))
    const secondDay = trafficElement({ 'data-statistics-traffic-day': '' }, rect(48, 0, 88, 180))
    firstDay.append(
      trafficElement({ 'data-statistics-traffic-id': 'day-1:bottom' }, rect(0, 110, 40, 180)),
      trafficElement({ 'data-statistics-traffic-id': 'day-1:top' }, rect(0, 70, 40, 110)),
    )
    secondDay.append(trafficElement({ 'data-statistics-traffic-id': 'day-2:only' }, rect(48, 90, 88, 180)))
    plot.append(firstDay, secondDay)

    expect(findClosestTrafficInteraction(plot, 18, 84)).toBe('day-1:top')
    expect(findClosestTrafficInteraction(plot, 80, 20)).toBe('day-2:only')
  })

  it('returns no target for an empty traffic plot', () => {
    expect(findClosestTrafficInteraction(document.createElement('div'), 10, 10)).toBeUndefined()
  })
})
