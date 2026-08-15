<script setup lang="ts">
import type { RolloverProjection } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { formatBytes, formatMoney } from '@/utils/format'

defineProps<{ detail: RolloverProjection | null; loading: boolean; error: string | null }>()
const emit = defineEmits<{ back: []; retry: [] }>()
</script>

<template>
  <div class="home-ride__detail-face">
    <div class="home-ride__detail-heading">
      <UButton color="neutral" variant="ghost" icon="i-ph-arrow-left" :label="$t('home.rolloverBack')" data-haptic @click="emit('back')" />
      <span class="home-ride__detail-title">{{ $t('home.rolloverTitle') }}</span>
    </div>
    <template v-if="loading">
      <div class="home-ride__detail-loading" aria-live="polite"><SkeletonBlock height="5rem" /><SkeletonBlock height="4rem" /></div>
      <p class="sr-only">{{ $t('home.rolloverLoading') }}</p>
    </template>
    <template v-else-if="error">
      <div class="home-ride__detail-error"><InlineNotice tone="warning">{{ error }}</InlineNotice><UButton color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :label="$t('home.rolloverRetry')" data-haptic @click="emit('retry')" /></div>
    </template>
    <template v-else-if="detail">
      <InlineNotice v-if="detail.warningCode" tone="warning">{{ $t(`home.rolloverWarning.${detail.warningCode}`) }}</InlineNotice>
      <template v-else>
        <div v-if="!detail.predictedRollover" class="home-ride__detail-overview">
          <span class="home-ride__detail-icon"><UIcon name="i-ph-arrows-clockwise" aria-hidden="true" /></span>
          <div><span>{{ $t('home.rolloverReductionNeeded') }}</span><strong>{{ formatBytes(detail.requiredReductionBytes ?? '0') }}</strong></div>
        </div>
        <section class="home-ride__window">
          <h3>{{ $t('home.rolloverForecast') }}</h3>
          <dl class="home-ride__window-metrics">
            <div><dt>{{ $t('home.rolloverUsed') }}</dt><dd>{{ formatBytes(detail.actualUsedTrafficBytes ?? '0') }}</dd></div>
            <div><dt>{{ $t('home.rolloverProjected') }}</dt><dd>{{ formatBytes(detail.projectedFullTermUsageBytes ?? '0') }}</dd></div>
            <div><dt>{{ $t('home.rolloverMaximumUsage') }}</dt><dd>{{ formatBytes(detail.maximumAllowableUsageBytes ?? '0') }}</dd></div>
            <div v-if="detail.predictedRollover"><dt>{{ $t('home.rolloverForecastCredit') }}</dt><dd>{{ formatMoney(detail.predictedRollover) }}</dd></div>
          </dl>
          <p v-if="!detail.predictedRollover" class="home-ride__maximum-hint">{{ $t('home.rolloverReduceDaily', { amount: formatBytes(detail.requiredDailyReductionBytes ?? '0') }) }}</p>
        </section>
      </template>
    </template>
  </div>
</template>
