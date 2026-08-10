<script setup lang="ts">
import { computed, nextTick, shallowRef } from 'vue'

import MarkdownContent from '@/components/common/MarkdownContent.vue'
import { useI18n } from '@/i18n'

interface TextareaExpose {
  textareaRef?: globalThis.HTMLTextAreaElement
}

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

const textarea = shallowRef<TextareaExpose>()
const color = shallowRef('accent')
const size = shallowRef('lg')
const { t } = useI18n()
const colorItems = computed(() => ['default', 'muted', 'accent', 'success', 'warning', 'danger'].map((value) => ({
  value,
  label: t(`markdown.${value}`),
})))
const sizeItems = computed(() => [
  { value: 'sm', label: t('markdown.small') },
  { value: 'base', label: t('markdown.base') },
  { value: 'lg', label: t('markdown.large') },
  { value: 'xl', label: t('markdown.extraLarge') },
])

async function applyDirective(): Promise<void> {
  const field = textarea.value?.textareaRef
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
    <UFormField :label="label" :required="required">
      <UTextarea
        ref="textarea"
        v-model="model"
        :rows="4"
        :maxlength="props.maxlength"
        :placeholder="placeholder"
        :required="required"
        autoresize
      />
    </UFormField>
    <div class="markdown-field__toolbar" role="toolbar" :aria-label="t('markdown.toolbar')">
      <USelect v-model="color" :items="colorItems" value-key="value" icon="i-ph-palette" :aria-label="t('markdown.color')" />
      <USelect v-model="size" :items="sizeItems" value-key="value" icon="i-ph-text-aa" :aria-label="t('markdown.size')" />
      <UButton :label="t('markdown.apply')" color="neutral" variant="soft" @click="applyDirective" />
    </div>
    <div class="markdown-field__preview">
      <span>{{ t('markdown.preview') }}</span>
      <MarkdownContent :source="model || placeholder || t('markdown.descriptionPreview')" compact />
    </div>
  </div>
</template>

<style scoped>
.markdown-field { display: grid; gap: 0.55rem; }
.markdown-field__toolbar { display: flex; flex-wrap: wrap; gap: 0.45rem; align-items: center; }
.markdown-field__preview { min-height: 58px; padding: 0.7rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface); }
.markdown-field__preview > span { display: block; margin-bottom: 0.45rem; color: var(--text-faint); font-size: 0.68rem; font-weight: 700; }
</style>
