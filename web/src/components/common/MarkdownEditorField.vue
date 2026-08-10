<script setup lang="ts">
import { nextTick, shallowRef } from 'vue'
import { PhPalette, PhTextAa } from '@phosphor-icons/vue'

import MarkdownContent from '@/components/common/MarkdownContent.vue'
import { useI18n } from '@/i18n'

const model = defineModel<string>({ required: true })
const props = withDefaults(defineProps<{
  label: string
  placeholder?: string
  required?: boolean
  maxlength?: number
}>(), {
  placeholder: '',
  required: false,
  maxlength: 1000,
})

const textarea = shallowRef<HTMLTextAreaElement>()
const color = shallowRef('accent')
const size = shallowRef('lg')
const { t } = useI18n()

async function applyDirective(): Promise<void> {
  const field = textarea.value
  if (!field) return
  const start = field.selectionStart
  const end = field.selectionEnd
  const selected = model.value.slice(start, end) || t('markdown.text')
  const directive = `[${selected}]{color=${color.value} size=${size.value}}`
  model.value = `${model.value.slice(0, start)}${directive}${model.value.slice(end)}`
  await nextTick()
  field.focus()
  field.setSelectionRange(start + 1, start + 1 + selected.length)
}
</script>

<template>
  <div class="markdown-field">
    <label>
      <span>{{ label }}</span>
      <textarea
        ref="textarea"
        v-model="model"
        rows="4"
        :maxlength="props.maxlength"
        :placeholder="placeholder"
        :required="required"
      />
    </label>
    <div class="markdown-field__toolbar" role="toolbar" :aria-label="t('markdown.toolbar')">
      <label><PhPalette :size="15" aria-hidden="true" /><span class="sr-only">{{ t('markdown.color') }}</span><select v-model="color" :aria-label="t('markdown.color')"><option value="default">{{ t('markdown.default') }}</option><option value="muted">{{ t('markdown.muted') }}</option><option value="accent">{{ t('markdown.accent') }}</option><option value="success">{{ t('markdown.success') }}</option><option value="warning">{{ t('markdown.warning') }}</option><option value="danger">{{ t('markdown.danger') }}</option></select></label>
      <label><PhTextAa :size="15" aria-hidden="true" /><span class="sr-only">{{ t('markdown.size') }}</span><select v-model="size" :aria-label="t('markdown.size')"><option value="sm">{{ t('markdown.small') }}</option><option value="base">{{ t('markdown.base') }}</option><option value="lg">{{ t('markdown.large') }}</option><option value="xl">{{ t('markdown.extraLarge') }}</option></select></label>
      <button class="button button--quiet" type="button" @click="applyDirective">{{ t('markdown.apply') }}</button>
    </div>
    <div class="markdown-field__preview"><span>{{ t('markdown.preview') }}</span><MarkdownContent :source="model || placeholder || t('markdown.descriptionPreview')" compact /></div>
  </div>
</template>

<style scoped>
.markdown-field { display: grid; gap: 0.55rem; }
.markdown-field label { display: grid; gap: 0.35rem; }
.markdown-field__toolbar { display: flex; flex-wrap: wrap; gap: 0.45rem; align-items: center; }
.markdown-field__toolbar label { display: flex; grid-template-columns: auto 1fr; align-items: center; gap: 0.3rem; }
.markdown-field__toolbar select { min-height: 36px; }
.markdown-field__preview { min-height: 58px; padding: 0.7rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface); }
.markdown-field__preview > span { display: block; margin-bottom: 0.45rem; color: var(--text-faint); font-size: 0.68rem; font-weight: 700; }
</style>
