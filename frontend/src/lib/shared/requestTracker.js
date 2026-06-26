/**
 * Creates a request tracker that generates monotonically increasing request
 * IDs and provides a stale-check helper. Useful for async store operations
 * where responses must be discarded if a newer request has already been sent.
 *
 * @returns {{ next(): number, isStale(id: number): boolean }}
 *
 * @example
 * const tracker = createRequestTracker();
 *
 * async function load() {
 *   const id = tracker.next();
 *   const data = await api("/data");
 *   if (tracker.isStale(id)) return; // a newer request superseded this one
 *   // … apply data …
 * }
 */
export function createRequestTracker() {
  let counter = 0;

  function next() {
    return ++counter;
  }

  function isStale(id) {
    return id !== counter;
  }

  return { next, isStale };
}
