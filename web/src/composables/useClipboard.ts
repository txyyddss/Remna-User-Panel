import { shallowRef } from 'vue'

import { copyText } from '@/utils/browserCompatibility'

export function useClipboard() {
  const copied = shallowRef(false)
  let resetTimer: ReturnType<typeof setTimeout> | undefined

  async function copy(value: string): Promise<boolean> {
    try {
      if (!await copyText(value)) throw new Error('CLIPBOARD_UNAVAILABLE')
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
