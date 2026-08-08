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
  ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 's', 'code', 'pre', 'ul', 'ol', 'li', 'blockquote', 'a', 'h1', 'h2', 'h3'],
  ALLOWED_ATTR: ['href', 'target', 'rel'],
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

.markdown-content :deep(code) { font-family: var(--font-mono); }

.markdown-content--compact {
  font-size: 0.76rem;
  line-height: 1.45;
}
</style>
