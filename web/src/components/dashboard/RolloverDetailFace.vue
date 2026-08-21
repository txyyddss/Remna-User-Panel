<script setup lang="ts">
import { computed, toRefs } from 'vue'

import type { RolloverProjection } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { formatBytes, formatMoney } from '@/utils/format'

const props = defineProps<{ detail: RolloverProjection | null; loading: boolean; error: string | null }>()
const emit = defineEmits<{ back: []; retry: [] }>()
const { detail, loading, error } = toRefs(props)

type RolloverState = 'alreadyExceeded' | 'predictedExceeded' | 'predicted'
type RolloverTone = 'success' | 'warning' | 'danger'

const rolloverPresentationByState: Record<RolloverState, { icon: string; tone: RolloverTone }> = {
  alreadyExceeded: { icon: 'i-ph-x-circle-fill', tone: 'danger' },
  predictedExceeded: { icon: 'i-ph-warning-circle-fill', tone: 'warning' },
  predicted: { icon: 'i-ph-check-circle-fill', tone: 'success' },
}

function parseTrafficBytes(value: string | null): bigint | null {
  return value && /^\d+$/.test(value) ? BigInt(value) : null
}

const alreadyExceeded = computed(() => {
  const actual = parseTrafficBytes(detail.value?.actualUsedTrafficBytes ?? null)
  const maximum = parseTrafficBytes(detail.value?.maximumAllowableUsageBytes ?? null)
  return actual !== null && maximum !== null && actual > maximum
})

const rolloverState = computed<RolloverState>(() => {
  if (detail.value?.predictedRollover) return 'predicted'
  return alreadyExceeded.value ? 'alreadyExceeded' : 'predictedExceeded'
})

const rolloverStatusKey = computed(() => {
  if (rolloverState.value === 'predicted') return 'home.rolloverStatus.willGet'
  if (rolloverState.value === 'alreadyExceeded') return 'home.rolloverStatus.couldNotGet'
  return 'home.rolloverStatus.dailyReduction'
})

const rolloverMetricLabelKey = computed(() => rolloverState.value === 'predictedExceeded'
  ? 'home.rolloverMaximumDailyUsage'
  : 'home.rolloverForecastCredit')

const rolloverPresentation = computed(() => rolloverPresentationByState[rolloverState.value])
</script>

<template>
  <div class="home-ride__detail-face">
    <div class="home-ride__detail-heading">
      <UButton color="neutral" variant="ghost" icon="i-ph-arrow-left" :label="$t('home.rolloverBack')" data-haptic="navigate" @click="emit('back')" />
      <span class="home-ride__detail-title">{{ $t('home.rolloverTitle') }}</span>
    </div>
    <template v-if="loading">
      <div class="home-ride__detail-loading" aria-live="polite"><SkeletonBlock height="5rem" /><SkeletonBlock height="4rem" /></div>
      <p class="sr-only">{{ $t('home.rolloverLoading') }}</p>
    </template>
    <template v-else-if="error">
      <div class="home-ride__detail-error"><InlineNotice tone="warning">{{ error }}</InlineNotice><UButton color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :label="$t('home.rolloverRetry')" data-haptic="retry" @click="emit('retry')" /></div>
    </template>
    <template v-else-if="detail">
      <InlineNotice v-if="detail.warningCode" tone="warning">{{ $t(`home.rolloverWarning.${detail.warningCode}`) }}</InlineNotice>
      <template v-else>
        <div
          class="home-ride__detail-overview"
          :class="`home-ride__detail-overview--${rolloverPresentation.tone}`"
          role="status"
        >
          <span class="home-ride__detail-icon"><UIcon :name="rolloverPresentation.icon" aria-hidden="true" /></span>
          <strong>{{ $t(rolloverStatusKey) }}</strong>
        </div>
        <section class="home-ride__window">
          <h3>{{ $t('home.rolloverForecast') }}</h3>
          <dl class="home-ride__window-metrics">
            <div><dt>{{ $t('home.rolloverUsed') }}</dt><dd>{{ formatBytes(detail.actualUsedTrafficBytes ?? '0') }}</dd></div>
            <div><dt>{{ $t('home.rolloverProjected') }}</dt><dd>{{ formatBytes(detail.projectedFullTermUsageBytes ?? '0') }}</dd></div>
            <div><dt>{{ $t('home.rolloverMaximumUsage') }}</dt><dd>{{ formatBytes(detail.maximumAllowableUsageBytes ?? '0') }}</dd></div>
            <div>
              <dt>{{ $t(rolloverMetricLabelKey) }}</dt>
              <dd v-if="rolloverState === 'alreadyExceeded'">{{ $t('home.rolloverNotAvailable') }}</dd>
              <dd v-else-if="rolloverState === 'predictedExceeded'">{{ formatBytes(detail.maximumDailyUsageBytes ?? '0') }}</dd>
              <dd v-else>{{ formatMoney(detail.predictedRollover!) }}</dd>
            </div>
          </dl>
        </section>
      </template>
    </template>
    <InlineNotice v-else tone="warning">{{ $t('errors.rolloverFailed') }}</InlineNotice>
  </div>
</template>
