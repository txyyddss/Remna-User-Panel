/**
 * Focuses an element without allowing an older or restricted WebView to
 * reject the focus options object during a route transition.
 */
export function focusWithoutScrolling(element: HTMLElement): void {
  try {
    element.focus({ preventScroll: true })
  } catch {
    try {
      element.focus()
    } catch {
      // The element may have been detached while the WebView was changing views.
    }
  }
}
