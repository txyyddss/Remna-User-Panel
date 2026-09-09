import type { EditorToolbarItem } from '@nuxt/ui'

export function markdownToolbar(t: (key: string) => string): EditorToolbarItem[][] {
  const button = (key: string, icon: string) => ({ icon, 'aria-label': t(`markdown.${key}`), tooltip: { text: t(`markdown.${key}`) } })
  return [
    [{ kind: 'undo', ...button('undo', 'i-ph-arrow-counter-clockwise') }, { kind: 'redo', ...button('redo', 'i-ph-arrow-clockwise') }],
    [
      { kind: 'mark', mark: 'bold', ...button('bold', 'i-ph-text-b') },
      { kind: 'mark', mark: 'italic', ...button('italic', 'i-ph-text-italic') },
      { kind: 'mark', mark: 'strike', ...button('strike', 'i-ph-text-strikethrough') },
      { kind: 'heading', level: 2, ...button('heading', 'i-ph-text-h-two') },
      { kind: 'bulletList', ...button('bulletList', 'i-ph-list-bullets') },
      { kind: 'orderedList', ...button('orderedList', 'i-ph-list-numbers') },
      { kind: 'blockquote', ...button('blockquote', 'i-ph-quotes') },
      { kind: 'mark', mark: 'code', ...button('code', 'i-ph-code') },
    ],
  ]
}
