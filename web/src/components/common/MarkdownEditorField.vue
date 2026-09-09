<script setup lang="ts">
import { computed } from 'vue'
import MarkdownContent from './MarkdownContent.vue'
import MarkdownStyleControls from './MarkdownStyleControls.vue'
import { markdownStyle } from './markdownStyle'
import { markdownToolbar } from './markdownToolbar'
import { useI18n } from '@/i18n'

const model = defineModel<string>({ required: true })
const props = withDefaults(defineProps<{
  label: string
  placeholder?: string
  required?: boolean
  maxlength?: number
}>(), { placeholder: '', required: false, maxlength: 1000 })
const { t } = useI18n()
const toolbar = computed(() => markdownToolbar(t))
const error = computed(() => model.value.length > props.maxlength ? t('markdown.tooLong', { max: props.maxlength }) : undefined)
</script>

<template>
  <div class="markdown-field">
    <UFormField :label="label" :required="required" :error="error">
      <UEditor
        v-slot="{ editor }" v-model="model" content-type="markdown"
        :extensions="[markdownStyle]" :placeholder="placeholder" :aria-label="label"
        :aria-required="required" :aria-invalid="Boolean(error)" :mention="false" :image="false"
        :starter-kit="{ underline: false, heading: { levels: [1, 2, 3] } }"
        :ui="{ base: 'min-h-36 px-3 py-3 text-base' }"
        class="markdown-field__editor w-full rounded-lg border border-default"
      >
        <UEditorToolbar :editor="editor" :items="toolbar" class="flex-wrap border-b border-default p-1" />
        <MarkdownStyleControls :editor="editor" />
      </UEditor>
    </UFormField>
    <div class="markdown-field__preview">
      <span>{{ t('markdown.preview') }}</span>
      <MarkdownContent :source="model || placeholder || t('markdown.descriptionPreview')" compact />
    </div>
  </div>
</template>

<style scoped>
.markdown-field { display: grid; min-width: 0; gap: 0.55rem; }
.markdown-field__editor { overflow-wrap: anywhere; }
.markdown-field__preview { min-height: 58px; padding: 0.7rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface); }
.markdown-field__preview > span { display: block; margin-bottom: 0.45rem; color: var(--text-faint); font-size: 0.68rem; font-weight: 700; }
.markdown-field__editor :deep(.md-color-default) { color: var(--text); }
.markdown-field__editor :deep(.md-color-muted) { color: var(--text-muted); }
.markdown-field__editor :deep(.md-color-accent), .markdown-field__editor :deep(.md-color-success) { color: var(--accent); }
.markdown-field__editor :deep(.md-color-warning) { color: var(--warning); }
.markdown-field__editor :deep(.md-color-danger) { color: var(--danger); }
.markdown-field__editor :deep(.md-size-sm) { font-size: 0.85em; }
.markdown-field__editor :deep(.md-size-base) { font-size: 1em; }
.markdown-field__editor :deep(.md-size-lg) { font-size: 1.15em; }
.markdown-field__editor :deep(.md-size-xl) { font-size: 1.3em; }
</style>
