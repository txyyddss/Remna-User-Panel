<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/i18n'
import { formatBytes } from '@/utils/format'

const props = defineProps<{
  name: string
  totalBytes: string
  categories: readonly string[]
  dailyBytes: readonly string[]
}>()

const { t } = useI18n()
const values = computed(() => props.dailyBytes.map((value) => {
  try { return BigInt(value) } catch { return 0n }
}))
const maximum = computed(() => values.value.reduce((current, value) => value > current ? value : current, 0n))
const bars = computed(() => values.value.map((value, index) => {
  const height = maximum.value === 0n ? 4 : Math.max(8, Number((value * 100n) / maximum.value))
  const date = props.categories[index] ?? ''
  return { date, value: formatBytes(props.dailyBytes[index] ?? '0'), height: `${height}%`, label: t('home.trafficChartDay', { date, value: formatBytes(props.dailyBytes[index] ?? '0') }) }
}))
</script>

<template>
  <div class="traffic-node-chart" role="img" :aria-label="t('home.trafficChartNode', { name, value: formatBytes(totalBytes) })">
    <div class="traffic-node-chart__bars">
      <span v-for="bar in bars" :key="bar.date" class="traffic-node-chart__bar" :style="{ height: bar.height }" :aria-label="bar.label" role="img" />
    </div>
    <div v-if="bars.length" class="traffic-node-chart__axis" aria-hidden="true"><span>{{ bars[0].date }}</span><span>{{ bars[bars.length - 1].date }}</span></div>
  </div>
</template>

<style scoped>
.traffic-node-chart { display: grid; min-width: 0; gap: 0.35rem; }
.traffic-node-chart__bars { min-height: 5.5rem; display: flex; align-items: end; gap: 3px; padding: 0.55rem 0.35rem 0.2rem; border-bottom: 1px solid var(--line); }
.traffic-node-chart__bar { min-width: 4px; flex: 1 1 0; border-radius: 3px 3px 0 0; background: var(--accent); opacity: 0.82; }
.traffic-node-chart__axis { display: flex; justify-content: space-between; gap: 0.5rem; color: var(--text-faint); font-family: var(--font-mono); font-size: 0.58rem; }
</style>
