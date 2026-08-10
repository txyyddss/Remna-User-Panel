<script setup lang="ts">
import { computed, onMounted, reactive, shallowRef } from 'vue'
import { PhArrowClockwise, PhX } from '@phosphor-icons/vue'

import type { AdminStatistics, StatisticsQuery } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useI18n } from '@/i18n'
import { txbInputFromMinor } from '@/utils/format'

const props = defineProps<{ title: string; load: (query: StatisticsQuery) => Promise<AdminStatistics> }>()
defineEmits<{ close: [] }>()

const statistics = shallowRef<AdminStatistics | null>(null)
const loading = shallowRef(true)
const error = shallowRef<string | null>(null)
const today = new Date().toISOString().slice(0, 10)
const monthAgo = new Date(Date.now() - 29 * 86_400_000).toISOString().slice(0, 10)
const filters = reactive<Required<Pick<StatisticsQuery, 'from' | 'to' | 'bucket'>>>({ from: monthAgo, to: today, bucket: 'daily' })
const maximumCount = computed(() => Math.max(1, ...(statistics.value?.series.map((point) => point.count) ?? [1])))
const { t } = useI18n()

async function refresh(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    statistics.value = await props.load({ ...filters })
  } catch (caught) {
    error.value = t('adminStatistics.error')
  } finally {
    loading.value = false
  }
}

onMounted(() => void refresh())
</script>

<template>
  <section class="statistics-panel">
    <header class="statistics-panel__header"><div><h3>{{ title }}</h3><p>{{ t('adminStatistics.serverCalculated') }}</p></div><button class="icon-button" type="button" :aria-label="t('adminStatistics.close')" @click="$emit('close')"><PhX :size="19" /></button></header>
    <form class="statistics-filters" @submit.prevent="refresh"><label><span>{{ t('adminStatistics.from') }}</span><input v-model="filters.from" type="date" required /></label><label><span>{{ t('adminStatistics.to') }}</span><input v-model="filters.to" type="date" required /></label><label><span>{{ t('adminStatistics.grouping') }}</span><select v-model="filters.bucket"><option value="daily">{{ t('adminStatistics.daily') }}</option><option value="weekly">{{ t('adminStatistics.weekly') }}</option></select></label><button class="button button--secondary" type="submit" :disabled="loading"><PhArrowClockwise :size="17" />{{ t('adminStatistics.refresh') }}</button></form>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <div v-if="statistics" class="statistics-content" :aria-busy="loading">
      <dl class="statistics-summary"><div><dt>{{ t('adminStatistics.participation') }}</dt><dd>{{ statistics.count }}</dd></div><div><dt>{{ t('adminStatistics.uniqueUsers') }}</dt><dd>{{ statistics.uniqueUsers }}</dd></div><div><dt>{{ t('adminStatistics.input') }}</dt><dd>{{ txbInputFromMinor(statistics.inputTxbMinor) }} TXB</dd></div><div><dt>{{ t('adminStatistics.output') }}</dt><dd>{{ txbInputFromMinor(statistics.outputTxbMinor) }} TXB</dd></div><div><dt>{{ t('adminStatistics.net') }}</dt><dd>{{ txbInputFromMinor(statistics.netTxbMinor) }} TXB</dd></div><div><dt>{{ t('adminStatistics.discountAddons') }}</dt><dd>{{ txbInputFromMinor(statistics.discountTxbMinor) }} / {{ txbInputFromMinor(statistics.addonTxbMinor) }} TXB</dd></div></dl>
      <div class="statistics-chart" role="img" :aria-label="t('adminStatistics.chartLabel', { bucket: statistics.bucket, from: statistics.from, to: statistics.to })">
        <span v-for="point in statistics.series" :key="point.periodStart" class="statistics-chart__bar" :style="{ height: `${Math.max(4, point.count / maximumCount * 100)}%` }" :title="`${point.periodStart}: ${point.count}`" />
      </div>
      <div class="statistics-table" tabindex="0">
        <table><caption>{{ t('adminStatistics.caption') }}</caption><thead><tr><th scope="col">{{ t('adminStatistics.period') }}</th><th scope="col">{{ t('adminStatistics.count') }}</th><th scope="col">{{ t('adminStatistics.unique') }}</th><th scope="col">{{ t('adminStatistics.inputTxb') }}</th><th scope="col">{{ t('adminStatistics.outputTxb') }}</th><th scope="col">{{ t('adminStatistics.netTxb') }}</th></tr></thead><tbody><tr v-for="point in statistics.series" :key="point.periodStart"><th scope="row">{{ point.periodStart }}</th><td>{{ point.count }}</td><td>{{ point.uniqueUsers }}</td><td>{{ txbInputFromMinor(point.inputTxbMinor) }}</td><td>{{ txbInputFromMinor(point.outputTxbMinor) }}</td><td>{{ txbInputFromMinor(point.netTxbMinor) }}</td></tr></tbody></table>
      </div>
      <ul v-if="statistics.distribution.length" class="statistics-distribution"><li v-for="slice in statistics.distribution" :key="slice.id"><span>{{ slice.label }}</span><strong>{{ slice.count }}</strong></li></ul>
    </div>
    <p v-else-if="loading" class="admin-loading">{{ t('adminStatistics.loading') }}</p>
  </section>
</template>

<style scoped>
.statistics-panel { display: grid; gap: 0.8rem; margin: 0.8rem; padding: 0.9rem; border: 1px solid var(--line); border-radius: var(--radius-card); background: var(--surface-raised); }
.statistics-panel__header { display: flex; justify-content: space-between; gap: 0.8rem; }
.statistics-panel__header h3, .statistics-panel__header p { margin: 0; }
.statistics-panel__header p { margin-top: 0.25rem; color: var(--text-muted); font-size: 0.72rem; }
.statistics-filters { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.6rem; }
.statistics-filters label { display: grid; gap: 0.3rem; }
.statistics-filters label span { color: var(--text-faint); font-size: 0.66rem; font-weight: 700; }
.statistics-summary { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.45rem; margin: 0; }
.statistics-summary div { padding: 0.6rem; border: 1px solid var(--line); border-radius: var(--radius-control); }
.statistics-summary dt { color: var(--text-faint); font-size: 0.62rem; }
.statistics-summary dd { margin: 0.2rem 0 0; font-size: 0.8rem; font-weight: 750; }
.statistics-chart { height: 150px; display: flex; align-items: end; gap: 3px; padding: 0.6rem; border: 1px solid var(--line); border-radius: var(--radius-control); }
.statistics-chart__bar { min-width: 4px; flex: 1; border-radius: 3px 3px 0 0; background: var(--accent); opacity: 0.78; }
.statistics-table { overflow: auto; }
.statistics-table table { width: 100%; min-width: 600px; border-collapse: collapse; font-size: 0.68rem; }
.statistics-table caption { padding: 0.4rem; text-align: left; color: var(--text-faint); }
.statistics-table th, .statistics-table td { padding: 0.45rem; border-bottom: 1px solid var(--line); text-align: left; }
.statistics-distribution { display: grid; gap: 0.35rem; margin: 0; padding: 0; list-style: none; }
.statistics-distribution li { display: flex; justify-content: space-between; padding: 0.45rem 0.6rem; background: var(--surface); }
@media (min-width: 760px) { .statistics-filters { grid-template-columns: repeat(3, minmax(150px, 1fr)) auto; align-items: end; }.statistics-summary { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
</style>
