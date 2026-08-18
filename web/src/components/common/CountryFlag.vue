<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ code: string }>()
const flags = {
  TW: 'i-circle-flags-tw',
  CN: 'i-circle-flags-cn',
  DE: 'i-circle-flags-de',
  GB: 'i-circle-flags-gb',
  HK: 'i-circle-flags-hk',
  JP: 'i-circle-flags-jp',
  SG: 'i-circle-flags-sg',
  US: 'i-circle-flags-us',
} as const
const normalizedCode = computed(() => props.code.trim().toLowerCase())
const flagIcon = computed(() => flags[normalizedCode.value.toUpperCase() as keyof typeof flags])
const flagEmoji = computed(() => {
  const code = normalizedCode.value.toUpperCase()
  if (!/^[A-Z]{2}$/.test(code)) return '--'
  return String.fromCodePoint(...[...code].map((letter) => 0x1f1e6 + letter.charCodeAt(0) - 65))
})
</script>

<template>
  <span class="country-flag" :title="normalizedCode.toUpperCase()">
    <UIcon v-if="flagIcon" :name="flagIcon" aria-hidden="true" />
    <span v-else aria-hidden="true">{{ flagEmoji }}</span>
  </span>
</template>

<style scoped>
.country-flag { width: 28px; height: 20px; display: inline-grid; place-items: center; overflow: hidden; color: var(--text-muted); font-family: var(--font-mono); font-size: 1rem; line-height: 1; }
.country-flag :deep(svg) { width: 100%; height: 100%; }
</style>
