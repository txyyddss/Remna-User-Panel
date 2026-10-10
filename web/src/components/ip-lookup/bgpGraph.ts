import type { IPLookupBGPTopology } from '@/api/ipLookup'

export function graphLayout(topology: IPLookupBGPTopology, directLimit: number, depth: number) {
  const distance = new Map(topology.origins.map(asn => [asn, 0]))
  const frequency = new Map<number, number>()
  for (const path of topology.paths) for (const asn of new Set(path.asPath)) frequency.set(asn, (frequency.get(asn) ?? 0) + 1)
  const neighbors = new Map<number, number[]>()
  for (const edge of topology.edges) neighbors.set(edge.to, [...neighbors.get(edge.to) ?? [], edge.from])
  const queue = [...topology.origins]
  for (let index = 0; index < queue.length; index++) for (const neighbor of neighbors.get(queue[index]!) ?? []) {
    if (!distance.has(neighbor)) { distance.set(neighbor, distance.get(queue[index]!)! + 1); queue.push(neighbor) }
  }
  const direct = topology.nodes.filter(node => distance.get(node.asn) === 1).sort((a, b) => {
    const upstream = (asn: number) => topology.edges.some(edge => edge.from === asn && topology.origins.includes(edge.to) && edge.relationship === 'upstream') ? 1 : 0
    return upstream(b.asn) - upstream(a.asn) || (frequency.get(b.asn) ?? 0) - (frequency.get(a.asn) ?? 0) || a.asn - b.asn
  })
  const visible = new Set([...topology.origins, ...direct.slice(0, directLimit).map(node => node.asn)])
  for (let layer = 1; layer < depth; layer++) for (const edge of topology.edges) {
    if (visible.has(edge.to) && distance.get(edge.to) === layer && distance.get(edge.from) === layer + 1) visible.add(edge.from)
  }
  const nodes = topology.nodes.filter(node => visible.has(node.asn)).map(node => ({ ...node, layer: distance.get(node.asn) ?? 0, x: 0, y: 0 }))
  const maximumLayer = Math.max(1, ...nodes.map(node => node.layer))
  const largestColumn = Math.max(1, ...Array.from({ length: maximumLayer + 1 }, (_, layer) => nodes.filter(node => node.layer === layer).length))
  const height = Math.max(300, largestColumn * 74 + 40)
  for (let layer = 0; layer <= maximumLayer; layer++) {
    const column = nodes.filter(node => node.layer === layer).sort((a, b) => a.asn - b.asn)
    column.forEach((node, index) => { node.x = 24 + layer * 230; node.y = layer === 0 ? 120 + index * 74 : 24 + index * 74 })
  }
  const byASN = new Map(nodes.map(node => [node.asn, node]))
  const edges = topology.edges.filter(edge => visible.has(edge.from) && visible.has(edge.to)).map(edge => {
    const from = byASN.get(edge.from)!
    const to = byASN.get(edge.to)!
    const sign = from.x > to.x ? -1 : 1
    const sx = sign < 0 ? from.x : from.x + 184, sy = from.y + 24, tx = sign < 0 ? to.x + 184 : to.x, ty = to.y + 24
    const bend = Math.max(35, Math.abs(sx - tx) / 2)
    return { ...edge, id: `${edge.from}-${edge.to}`, d: `M ${sx} ${sy} C ${sx + sign * bend} ${sy}, ${tx - sign * bend} ${ty}, ${tx} ${ty}` }
  })
  return { nodes, edges, width: maximumLayer * 230 + 232, height, directCount: direct.length, hidden: topology.nodes.length - nodes.length, maxDepth: Math.max(1, ...distance.values()) }
}

export function pathEdges(path: number[]) {
  return new Set(path.slice(1).map((asn, index) => `${path[index]}-${asn}`))
}
