<script setup lang="ts">
import { computed } from 'vue'
import type { AffiliateOverview } from '@/api/features'
import { useI18n } from '@/i18n'
import { txbInputFromMinor } from '@/utils/format'

const props = defineProps<{ progress: AffiliateOverview['tierProgress'] }>()
const { t } = useI18n()
const value = computed(() => props.progress.successful - props.progress.current.threshold)
const maximum = computed(() => Math.max(1, (props.progress.next?.threshold ?? props.progress.successful) - props.progress.current.threshold))
function commission(bps: number, enabled: boolean): string { return enabled ? `${(bps / 100).toFixed(2)}%` : t('affiliates.off') }
function reward(): string {
  const next = props.progress.next?.reward
  if (!next || next.kind === 'none') return t('affiliates.noReward')
  if (next.kind === 'coupon') return next.couponName || t('affiliates.couponReward')
  if (next.kind === 'txb') return t('affiliates.txbReward', { value: txbInputFromMinor(String(next.txbMinor)) })
  return t('affiliates.extensionReward', { days: next.extensionDays })
}
</script>

<template>
  <section class="affiliate-section affiliate-tier">
    <div class="affiliate-section__heading"><div><span>{{ $t('affiliates.currentTier') }}</span><h2>{{ progress.current.name }}</h2></div><strong>{{ commission(progress.current.commissionBps, progress.current.commissionEnabled) }}</strong></div>
    <div v-if="progress.topTier" class="affiliate-top-tier"><UIcon name="i-ph-crown" /><strong>{{ $t('affiliates.topTierTitle') }}</strong><span>{{ $t('affiliates.topTierCopy') }}</span></div>
    <template v-else-if="progress.next">
      <UProgress :model-value="value" :max="maximum" :get-value-text="() => $t('affiliates.progressText', { current: value, total: maximum })" />
      <div class="affiliate-tier__facts">
        <div><span>{{ $t('affiliates.toUpgrade') }}</span><strong>{{ $t('affiliates.referralsLeft', { count: progress.remaining }) }}</strong></div>
        <div><span>{{ $t('affiliates.nextTier', { name: progress.next.name }) }}</span><strong>{{ commission(progress.next.commissionBps, progress.next.commissionEnabled) }}</strong><small>{{ reward() }}</small></div>
      </div>
    </template>
  </section>
</template>
