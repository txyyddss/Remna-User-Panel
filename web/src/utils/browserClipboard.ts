async function copyWithClipboardAPI(value: string): Promise<boolean> {
  try {
    const clipboard = globalThis.navigator?.clipboard
    if (typeof clipboard?.writeText !== 'function') return false
    await clipboard.writeText(value)
    return true
  } catch {
    return false
  }
}

function copyWithSelection(value: string): boolean {
  if (
    typeof document === 'undefined'
    || !document.body
    || typeof document.execCommand !== 'function'
  ) {
    return false
  }

  const active = document.activeElement as HTMLElement | null
  const textarea = document.createElement('textarea')
  textarea.value = value
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.top = '-9999px'
  textarea.style.left = '-9999px'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)

  try {
    textarea.focus()
    textarea.select()
    textarea.setSelectionRange?.(0, value.length)
    return document.execCommand('copy')
  } catch {
    return false
  } finally {
    textarea.parentNode?.removeChild(textarea)
    try {
      active?.focus({ preventScroll: true })
    } catch {
      try {
        active?.focus()
      } catch {
        // The focused control may have been detached.
      }
    }
  }
}

export async function copyText(value: string): Promise<boolean> {
  if (await copyWithClipboardAPI(value)) return true
  return copyWithSelection(value)
}
