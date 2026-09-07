// Preserve edits made while a cached form is being refreshed from the server.
export function mergeRefreshedDraft<T extends object>(draft: T, next: T, previous?: T): void {
  for (const key of Object.keys(next) as (keyof T)[]) {
    if (!previous || JSON.stringify(draft[key]) === JSON.stringify(previous[key])) draft[key] = next[key]
  }
}
