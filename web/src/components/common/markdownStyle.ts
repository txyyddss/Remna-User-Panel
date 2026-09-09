import { Mark } from '@tiptap/core'

const colors = new Set(['default', 'muted', 'accent', 'success', 'warning', 'danger'])
const sizes = new Set(['sm', 'base', 'lg', 'xl'])

/** Preserve the existing allowlisted text directives in the Nuxt editor. */
export const markdownStyle = Mark.create({
  name: 'safeTextStyle',
  excludes: '_',
  addAttributes: () => ({ color: { default: null }, size: { default: null } }),
  parseHTML: () => [{ tag: 'span[data-text-style]' }],
  renderHTML: ({ HTMLAttributes }) => ['span', {
    'data-text-style': '',
    color: HTMLAttributes.color,
    size: HTMLAttributes.size,
    class: [colors.has(HTMLAttributes.color) && `md-color-${HTMLAttributes.color}`,
      sizes.has(HTMLAttributes.size) && `md-size-${HTMLAttributes.size}`].filter(Boolean).join(' '),
  }, 0],
  markdownTokenName: 'safeTextStyle',
  markdownTokenizer: {
    name: 'safeTextStyle', level: 'inline', start: '[',
    tokenize(source) {
      const match = /^\[([^\]\n]+)\]\{([^}\n]+)\}/.exec(source)
      if (!match) return
      const attrs: Record<string, string> = {}
      for (const attribute of match[2]!.trim().split(/\s+/)) {
        const [key, value, extra] = attribute.split('=')
        if (!key || !value || extra || attrs[key]) return
        if (!(key === 'color' && colors.has(value)) && !(key === 'size' && sizes.has(value))) return
        attrs[key] = value
      }
      return { type: 'safeTextStyle', raw: match[0], text: match[1], attrs }
    },
  },
  parseMarkdown: (token, helpers) => helpers.applyMark('safeTextStyle', [helpers.createTextNode(token.text ?? '')], token.attrs),
  renderMarkdown: (node, helpers) => {
    const attributes = [colors.has(node.attrs?.color) && `color=${node.attrs?.color}`,
      sizes.has(node.attrs?.size) && `size=${node.attrs?.size}`].filter(Boolean).join(' ')
    return `[${helpers.renderChildren(node)}]{${attributes}}`
  },
})
