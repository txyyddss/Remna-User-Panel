<script setup lang="ts">
import type { RolloverProjection, RolloverWindow } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useI18n } from '@/i18n'
import { formatBPS, formatBytes, formatDateTime, formatMoney } from '@/utils/format'

defineProps<{
  detail: RolloverProjection | null
  loading: boolean
  error: string | null
}>()
const emit = defineEmits<{ back: []; retry: [] }>()
const { t } = useI18n()

function rangeLabel(window: RolloverWindow): string {
  return `${formatDateTime(window.start)} ${t('common.rangeSeparator')} ${formatDateTime(window.end)}`
}
</script>

<template>
  <div class="home-ride__detail-face">
    <div class="home-ride__detail-heading">
      <UButton color="neutral" variant="ghost" icon="i-ph-arrow-left" :label="$t('home.rolloverBack')" data-haptic @click="emit('back')" />
      <span class="home-ride__detail-title">{{ $t('home.rolloverTitle') }}</span>
    </div>

    <template v-if="loading">
      <div class="home-ride__detail-loading" aria-live="polite">
        <SkeletonBlock height="2.2rem" />
        <SkeletonBlock height="5rem" />
        <SkeletonBlock height="5rem" />
      </div>
      <p class="sr-only">{{ $t('home.rolloverLoading') }}</p>
    </template>

    <template v-else-if="error">
      <div class="home-ride__detail-error">
        <InlineNotice tone="warning">{{ error }}</InlineNotice>
        <UButton color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :label="$t('home.rolloverRetry')" data-haptic @click="emit('retry')" />
      </div>
    </template>

    <template v-else-if="detail">
      <div class="home-ride__detail-stats">
        <div class="home-ride__detail-stat home-ride__detail-stat--accent">
          <span>{{ $t('home.rolloverTermAmount') }}</span>
          <strong>{{ formatMoney(detail.term.rollover) }}</strong>
        </div>
        <div class="home-ride__detail-stat">
          <span>{{ $t('home.rolloverResetAmount') }}</span>
          <strong>{{ formatMoney(detail.lastResetPeriod.rollover) }}</strong>
        </div>
        <div class="home-ride__detail-stat">
          <span>{{ $t('home.rolloverPaid') }}</span>
          <strong>{{ formatMoney(detail.paid) }}</strong>
        </div>
        <div class="home-ride__detail-stat">
          <span>{{ $t('home.rolloverSaved') }}</span>
          <strong>{{ $t('home.rolloverSavedValue', { value: formatBPS(detail.savedBps) }) }}</strong>
        </div>
      </div>

      <div class="home-ride__windows">
        <section v-for="item in [{ key: 'term', value: detail.term }, { key: 'reset', value: detail.lastResetPeriod }]" :key="item.key" class="home-ride__window">
          <div class="home-ride__window-heading">
            <div>
              <h3>{{ item.key === 'term' ? $t('home.rolloverTermWindow') : $t('home.rolloverResetWindow') }}</h3>
              <p>{{ rangeLabel(item.value) }}</p>
            </div>
            <strong>{{ formatMoney(item.value.rollover) }}</strong>
          </div>
          <dl class="home-ride__window-metrics">
            <div><dt>{{ $t('home.rolloverAllocated') }}</dt><dd>{{ formatBytes(item.value.allocatedTrafficBytes) }}</dd></div>
            <div><dt>{{ $t('home.rolloverUsed') }}</dt><dd>{{ formatBytes(item.value.usedTrafficBytes) }}</dd></div>
            <div><dt>{{ $t('home.rolloverRemaining') }}</dt><dd>{{ formatBytes(item.value.remainingTrafficBytes) }}</dd></div>
            <div><dt>{{ $t('home.rolloverEligible') }}</dt><dd>{{ formatBytes(item.value.eligibleUnusedBytes) }}</dd></div>
          </dl>
          <p class="home-ride__maximum-hint">
            <span>{{ $t('home.rolloverMaximum', { amount: formatMoney(detail.maximum) }) }}</span>
            <strong v-if="item.value.maximumReachable">{{ $t('home.rolloverTrafficToMaximum', { amount: formatBytes(item.value.trafficToMaximumBytes ?? '0') }) }}</strong>
            <strong v-else>{{ $t('home.rolloverMaximumUnreachable') }}</strong>
          </p>
        </section>
      </div>
      <p class="home-ride__detail-footnote">{{ $t('home.rolloverFetchedAt', { date: formatDateTime(detail.fetchedAt) }) }}</p>
    </template>
  </div>
</template>
