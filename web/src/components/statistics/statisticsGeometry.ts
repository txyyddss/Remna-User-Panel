import type { StatisticSegment } from './statisticsFormat'

export interface RingSegment extends StatisticSegment {
  dasharray: string
  dashoffset: number
}

export interface PieSlice extends StatisticSegment {
  path: string
  labelLine: string
  labelX: number
  labelY: number
  labelAnchor: 'start' | 'middle' | 'end'
  showLabel: boolean
}

interface Point { x: number, y: number }

function polarPoint(center: number, radius: number, angle: number): Point {
  const radians = angle * Math.PI / 180
  return {
    x: center + radius * Math.cos(radians),
    y: center + radius * Math.sin(radians),
  }
}

function pointText(point: Point): string {
  return `${point.x.toFixed(3)} ${point.y.toFixed(3)}`
}

function piePath(center: number, radius: number, startAngle: number, endAngle: number): string {
  const start = polarPoint(center, radius, startAngle)
  const end = polarPoint(center, radius, endAngle)
  const largeArc = endAngle - startAngle > 180 ? 1 : 0
  return `M ${center} ${center} L ${pointText(start)} A ${radius} ${radius} 0 ${largeArc} 1 ${pointText(end)} Z`
}

export function ringSegments(segments: readonly StatisticSegment[]): RingSegment[] {
  let cursor = 0
  return segments.map((segment) => {
    const dashoffset = -cursor
    cursor += segment.percentage
    return {
      ...segment,
      dasharray: `${segment.percentage} ${Math.max(0, 100 - segment.percentage)}`,
      dashoffset,
    }
  })
}

export function pieSlices(segments: readonly StatisticSegment[]): PieSlice[] {
  let cursor = -90
  return segments.map((segment) => {
    const sweep = Math.min(359.999, segment.percentage * 3.6)
    const end = cursor + sweep
    const middle = cursor + sweep / 2
    const lineStart = polarPoint(100, 65, middle)
    const lineEnd = polarPoint(100, 80, middle)
    const label = polarPoint(100, 86, middle)
    const slice = {
      ...segment,
      path: piePath(100, 62, cursor, end),
      labelLine: `M ${pointText(lineStart)} L ${pointText(lineEnd)}`,
      labelX: label.x,
      labelY: label.y,
      labelAnchor: label.x > 104 ? 'start' as const : label.x < 96 ? 'end' as const : 'middle' as const,
      showLabel: segment.percentage >= 6,
    }
    cursor = end
    return slice
  })
}
