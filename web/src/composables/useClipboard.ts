import { shallowRef } from 'vue'

export function useClipboard() {
  const copied = shallowRef(false)
  let resetTimer: ReturnType<typeof setTimeout> | undefined

  async function copy(value: string): Promise<boolean> {
    try {
      await navigator.clipboard.writeText(value)
      copied.value = true
      if (resetTimer !== undefined) clearTimeout(resetTimer)
      resetTimer = setTimeout(() => { copied.value = false }, 1800)
      return true
    } catch {
      copied.value = false
      return false
    }
  }

  return { copied, copy }
}
