<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import type { Editor } from '@tiptap/core'
import { useI18n } from '@/i18n'

const props = defineProps<{ editor: Editor }>()
const { t } = useI18n()
const color = shallowRef('accent')
const size = shallowRef('lg')
const colors = computed(() => ['default', 'muted', 'accent', 'success', 'warning', 'danger']
  .map(value => ({ value, label: t(`markdown.${value}`) })))
const sizes = computed(() => ['sm', 'base', 'lg', 'xl'].map((value, index) => ({
  value, label: t(`markdown.${['small', 'base', 'large', 'extraLarge'][index]}`),
})))
function apply(): void {
  const chain = props.editor.chain().focus()
  if (props.editor.state.selection.empty) {
    chain.insertContent({ type: 'text', text: t('markdown.text'), marks: [{ type: 'safeTextStyle', attrs: { color: color.value, size: size.value } }] }).run()
  } else chain.setMark('safeTextStyle', { color: color.value, size: size.value }).run()
}
</script>

<template>
  <div class="flex flex-wrap items-center gap-2 px-2 pb-2" role="toolbar" :aria-label="t('markdown.toolbar')">
    <USelect v-model="color" :items="colors" icon="i-ph-palette" :aria-label="t('markdown.color')" />
    <USelect v-model="size" :items="sizes" icon="i-ph-text-aa" :aria-label="t('markdown.size')" />
    <UButton :label="t('markdown.apply')" color="neutral" variant="soft" @click="apply" />
  </div>
</template>
