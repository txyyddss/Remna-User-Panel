<script setup lang="ts">
import { computed } from 'vue'
import DOMPurify from 'dompurify'
import MarkdownIt from 'markdown-it'

const props = withDefaults(defineProps<{
  source: string
  compact?: boolean
}>(), {
  compact: false,
})

const markdown = new MarkdownIt({ breaks: true, html: false, linkify: false, typographer: false })
const allowedColors = new Set(['default', 'muted', 'accent', 'success', 'warning', 'danger'])
const allowedSizes = new Set(['sm', 'base', 'lg', 'xl'])

markdown.inline.ruler.before('link', 'safe_text_directive', (state, silent): boolean => {
  const tail = state.src.slice(state.pos)
  const match = /^\[([^\]\n]+)\]\{([^}\n]+)\}/.exec(tail)
  if (!match) return false
  const attributes = match[2].trim().split(/\s+/)
  let color = ''
  let size = ''
  for (const attribute of attributes) {
    const [key, value, ...extra] = attribute.split('=')
    if (extra.length || !value) return false
    if (key === 'color' && !color && allowedColors.has(value)) color = value
    else if (key === 'size' && !size && allowedSizes.has(value)) size = value
    else return false
  }
  if (!color && !size) return false
  if (!silent) {
    const open = state.push('span_open', 'span', 1)
    open.attrSet('class', [color && `md-color-${color}`, size && `md-size-${size}`].filter(Boolean).join(' '))
    state.push('text', '', 0).content = match[1]
    state.push('span_close', 'span', -1)
  }
  state.pos += match[0].length
  return true
})
const defaultLinkOpen = markdown.renderer.rules.link_open
markdown.renderer.rules.link_open = (tokens, index, options, environment, renderer) => {
  const token = tokens[index]
  const hrefIndex = token.attrIndex('href')
  const href = String(hrefIndex >= 0 ? token.attrs?.[hrefIndex]?.[1] ?? '' : '')
  if (!/^https:\/\//i.test(href)) {
    if (hrefIndex >= 0 && token.attrs) token.attrs[hrefIndex][1] = '#'
  } else {
    token.attrSet('target', '_blank')
    token.attrSet('rel', 'noopener noreferrer')
  }
  return defaultLinkOpen
    ? defaultLinkOpen(tokens, index, options, environment, renderer)
    : renderer.renderToken(tokens, index, options)
}

const html = computed(() => DOMPurify.sanitize(markdown.render(props.source), {
  ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 's', 'code', 'pre', 'ul', 'ol', 'li', 'blockquote', 'a', 'h1', 'h2', 'h3', 'span'],
  ALLOWED_ATTR: ['href', 'target', 'rel', 'class'],
  ALLOW_UNKNOWN_PROTOCOLS: false,
}))
</script>

<template>
  <!-- Content is rendered with HTML disabled in MarkdownIt and then allowlisted by DOMPurify. -->
  <!-- eslint-disable-next-line vue/no-v-html -->
  <div class="markdown-content" :class="{ 'markdown-content--compact': compact }" v-html="html" />
</template>

<style scoped>
.markdown-content {
  color: var(--text-muted);
  font-size: 0.86rem;
  line-height: 1.6;
}

.markdown-content :deep(> :first-child) { margin-top: 0; }
.markdown-content :deep(> :last-child) { margin-bottom: 0; }

.markdown-content :deep(a) {
  color: var(--accent);
  text-decoration: underline;
  text-underline-offset: 0.18em;
}

.markdown-content :deep(h1), .markdown-content :deep(h2), .markdown-content :deep(h3) {
  margin: 1.1em 0 0.45em;
  color: var(--text);
  font-weight: 700;
  line-height: 1.2;
}

.markdown-content :deep(h1) { font-size: 1.3em; }
.markdown-content :deep(h2) { font-size: 1.15em; }
.markdown-content :deep(h3) { font-size: 1.02em; }
.markdown-content :deep(ul), .markdown-content :deep(ol) { margin: 0.7em 0; padding-left: 1.25em; }
.markdown-content :deep(li + li) { margin-top: 0.28em; }
.markdown-content :deep(strong) { color: var(--text); }
.markdown-content :deep(blockquote) { margin: 0.8em 0; padding-left: 0.8em; border-left: 2px solid var(--accent); color: var(--text); }
.markdown-content :deep(code) { font-family: var(--font-mono); }
.markdown-content :deep(pre) { overflow-x: auto; padding: 0.7em; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface); }

.markdown-content :deep(.md-color-default) { color: var(--text); }
.markdown-content :deep(.md-color-muted) { color: var(--text-muted); }
.markdown-content :deep(.md-color-accent) { color: var(--accent); }
.markdown-content :deep(.md-color-success) { color: var(--success); }
.markdown-content :deep(.md-color-warning) { color: var(--warning); }
.markdown-content :deep(.md-color-danger) { color: var(--danger); }
.markdown-content :deep(.md-size-sm) { font-size: 0.78em; }
.markdown-content :deep(.md-size-base) { font-size: 1em; }
.markdown-content :deep(.md-size-lg) { font-size: 1.18em; }
.markdown-content :deep(.md-size-xl) { font-size: 1.38em; }

.markdown-content--compact {
  font-size: 0.76rem;
  line-height: 1.45;
}
</style>
