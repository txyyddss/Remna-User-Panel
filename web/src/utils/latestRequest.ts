export interface LatestRequest {
  begin: () => number
  isCurrent: (token: number) => boolean
  invalidate: () => void
  dispose: () => void
}

export function createLatestRequest(): LatestRequest {
  let generation = 0
  let disposed = false

  return {
    begin: () => {
      generation += 1
      return generation
    },
    isCurrent: (token) => !disposed && token === generation,
    invalidate: () => { generation += 1 },
    dispose: () => {
      disposed = true
      generation += 1
    },
  }
}
