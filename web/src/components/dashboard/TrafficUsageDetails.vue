<script setup lang="ts">
import { computed } from 'vue'

import type { DashboardNodeUsage, Purchase } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { t } from '@/i18n'
import { formatBytes, formatDate } from '@/utils/format'

const props = defineProps<{
  startDate: string
  endDate: string
  term?: Purchase | null
  usage: DashboardNodeUsage | null
  loading: boolean
  error: string | null
}>()

const emit = defineEmits<{
  load: []
  'update:startDate': [value: string]
  'update:endDate': [value: string]
}>()

const maximumRangeMilliseconds = 30 * 24 * 60 * 60 * 1000
const rangeValid = computed(() => {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(props.startDate) || !/^\d{4}-\d{2}-\d{2}$/.test(props.endDate)) return false
  const start = Date.parse(`${props.startDate}T00:00:00.000Z`)
  const end = Date.parse(`${props.endDate}T00:00:00.000Z`)
  return Number.isFinite(start) && Number.isFinite(end) && start <= end && end - start <= maximumRangeMilliseconds
})
const nodeItems = computed(() => props.usage?.nodes.map((node) => ({
  ...node,
  value: node.uuid,
  label: t('home.trafficNode', { name: node.name, country: node.countryCode }),
})) ?? [])

function updateStart(value: unknown): void {
  emit('update:startDate', typeof value === 'string' ? value : '')
}

function updateEnd(value: unknown): void {
  emit('update:endDate', typeof value === 'string' ? value : '')
}

function load(): void {
  if (rangeValid.value) emit('load')
}
</script>

<template>
  <div class="traffic-usage">
    <div v-if="term" class="home-usage__term">
      <span>{{ $t('home.currentTerm') }}</span>
      <strong>{{ formatDate(term.validFrom) }} - {{ formatDate(term.validUntil) }}</strong>
    </div>
    <form class="traffic-usage__form" @submit.prevent="load">
      <div class="traffic-usage__range">
        <UFormField class="traffic-usage__field" :label="$t('home.trafficStart')">
          <UInput type="date" :model-value="startDate" @update:model-value="updateStart" />
        </UFormField>
        <UFormField class="traffic-usage__field" :label="$t('home.trafficEnd')">
          <UInput type="date" :model-value="endDate" @update:model-value="updateEnd" />
        </UFormField>
      </div>
      <UButton class="traffic-usage__apply" type="submit" :label="$t('home.trafficApply')" :disabled="!rangeValid" :loading="loading" />
      <p class="traffic-usage__hint">{{ $t('home.trafficRangeHint') }}</p>
    </form>

    <InlineNotice v-if="error" class="traffic-usage__notice" tone="warning">{{ error }}</InlineNotice>
    <p v-else-if="loading" class="traffic-usage__status">{{ $t('common.loading') }}</p>
    <template v-else-if="usage">
      <p class="traffic-usage__period">{{ $t('home.trafficPeriod', { start: usage.startDate, end: usage.endDate }) }}</p>
      <p class="traffic-usage__hint">{{ $t('home.trafficTopNodes', { count: 20 }) }}</p>
      <p v-if="!nodeItems.length" class="traffic-usage__status">{{ $t('home.noNodeUsage') }}</p>
      <UAccordion v-else class="traffic-usage__nodes" :items="nodeItems" type="multiple" :collapsible="true">
        <template #body="{ item }">
          <div class="traffic-usage__node-body">
            <p class="traffic-usage__total"><span>{{ $t('home.trafficTotal') }}</span><strong>{{ formatBytes(item.totalBytes) }}</strong></p>
            <dl class="traffic-usage__days">
              <template v-for="(date, index) in usage.categories" :key="`${item.uuid}-${date}`">
                <dt>{{ date }}</dt>
                <dd>{{ formatBytes(item.dailyBytes[index] ?? '0') }}</dd>
              </template>
            </dl>
          </div>
        </template>
      </UAccordion>
    </template>
    <p v-else class="traffic-usage__status">{{ $t('home.trafficLoadHint') }}</p>
  </div>
</template>

<style scoped>
.traffic-usage { width: min(26rem, calc(100vw - 2rem)); max-height: min(34rem, calc(100vh - 4rem)); overflow-y: auto; padding: 0.9rem; }
.traffic-usage__form, .traffic-usage__nodes { display: grid; gap: 0.7rem; }
.traffic-usage__range { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.55rem; }
.traffic-usage__field { min-width: 0; }
.traffic-usage__apply { justify-self: start; }
.traffic-usage__hint, .traffic-usage__period, .traffic-usage__status { margin: 0; color: var(--text-muted); font-size: 0.74rem; line-height: 1.45; }
.traffic-usage__period { color: var(--text-faint); font-family: var(--font-mono); }
.traffic-usage__notice { margin-top: 0.75rem; }
.traffic-usage__nodes { margin-top: 0.75rem; }
.traffic-usage__node-body { display: grid; gap: 0.6rem; padding: 0 0.2rem 0.25rem; }
.traffic-usage__total { display: flex; align-items: baseline; justify-content: space-between; gap: 0.75rem; margin: 0; font-size: 0.78rem; }
.traffic-usage__total span { color: var(--text-faint); }
.traffic-usage__total strong, .traffic-usage__days dd { font-family: var(--font-mono); }
.traffic-usage__days { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 0.35rem 0.75rem; margin: 0; font-size: 0.72rem; }
.traffic-usage__days dt { min-width: 0; color: var(--text-muted); }
.traffic-usage__days dd { margin: 0; text-align: right; }
</style>
