<script setup lang="ts">
import type { IPLookupLiveMeta } from '@/api/ipLookup'
import { computed } from 'vue'
import { t } from '@/i18n'
const props = defineProps<{ meta: IPLookupLiveMeta }>()
const time = computed(() => {
  if (!props.meta.observedAt) return ''
  if (/^\d{4}-\d{2}-\d{2}$/.test(props.meta.observedAt)) return new Date(props.meta.observedAt).toLocaleDateString(undefined, { timeZone: 'UTC' })
  const raw = /^\d{4}-\d{2}-\d{2}T[\d:.]+$/.test(props.meta.observedAt) ? `${props.meta.observedAt}Z` : props.meta.observedAt
  const date = new Date(raw)
  return Number.isNaN(date.getTime()) ? props.meta.observedAt : date.toLocaleString()
})
const message = computed(() => props.meta.status === 'error' || props.meta.status === 'restricted' ? t(`ipLookupLive.errors.${props.meta.errorCode || 'IP_DETAILS_UNAVAILABLE'}`) : t(`ipLookupLive.status.${props.meta.status}`))
</script>

<template>
  <p class="live-source-note"><span>{{ meta.source }}</span><span v-if="meta.status !== 'success'">{{ message }}</span><time v-if="time" :datetime="meta.observedAt">{{ time }}</time></p>
</template>

<style scoped>
.live-source-note { display: flex; flex-wrap: wrap; gap: 0.25rem 0.65rem; margin: 0.4rem 0 0; font-size: 0.72rem; color: var(--text-muted); overflow-wrap: anywhere; }
</style>
