<script setup lang="ts">
import { computed } from 'vue'

import type { ActivityResult } from '@/api/features'
import { useI18n } from '@/i18n'
import { formatMoney } from '@/utils/format'
import BetSuccessFireworks from './BetSuccessFireworks.vue'
import { isSuccessfulBet } from './feedback'

const props = defineProps<{ result: ActivityResult | null }>()
defineEmits<{ close: [] }>()

const { t } = useI18n()
const showFireworks = computed(() => isSuccessfulBet(props.result))

const rewardLabel = computed(() => {
  const result = props.result
  const reward = result && (result.kind === 'draw' || result.kind === 'check_in') ? result.reward : null
  if (!reward || reward.kind === 'none') return t('activity.noReward')
  if (reward.kind === 'txb_delta') {
    const amount = formatMoney({ currency: 'TXB', minor: reward.txbDeltaMinor, display: '' })
    return result?.kind === 'check_in' ? t('activity.checkInReward', { amount }) : t('activity.rewardTxb', { amount })
  }
  if (reward.kind === 'coupon_grant') return t('activity.rewardCoupon')
  return t('activity.rewardSubscription', { days: reward.extensionDays })
})

const description = computed(() => {
  const result = props.result
  if (!result) return ''
  if (result.kind === 'check_in') {
    const state = result.reward.kind === 'none' ? 'checkInRecorded' : 'checkInRewarded'
    return t(`activity.resultDescription.${state}`)
  }
  if (result.kind === 'bet') return t(`activity.resultDescription.bet${result.outcome === 'win' ? 'Win' : 'Loss'}`)
  return t('activity.resultDescription.drawComplete')
})
</script>

<template>
  <UModal
    :open="Boolean(result)"
    :title="result ? $t(`activity.result.${result.outcome}.title`) : ''"
    :description="description"
    @update:open="!$event && $emit('close')"
  >
    <template v-if="result" #body>
      <div class="result-body">
        <BetSuccessFireworks v-if="showFireworks" :key="result.id" />
        <UIcon
          :name="result.outcome === 'loss' ? 'i-ph-warning-circle-fill' : 'i-ph-check-circle-fill'"
          class="feature-icon"
          :class="{ 'feature-icon--warning': result.outcome === 'loss' }"
          aria-hidden="true"
        />
        <div class="result-balance">
          <span>{{ $t('activity.balanceAfter') }}</span>
          <strong>{{ formatMoney(result.balanceAfter) }}</strong>
        </div>
        <div v-if="(result.kind === 'draw' || result.kind === 'check_in') && result.reward.kind !== 'none'" class="result-reward">
          <span>{{ $t('activity.reward') }}</span>
          <strong>{{ rewardLabel }}</strong>
        </div>
      </div>
    </template>
    <template #footer="{ close }">
      <UButton block :label="$t('common.close')" @click="close" />
    </template>
  </UModal>
</template>

<style scoped>
.result-body { position: relative; overflow: hidden; }
.result-body > :not(.bet-fireworks) { position: relative; z-index: 1; }
.result-balance { display: flex; align-items: baseline; justify-content: space-between; padding: 0.8rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.result-balance span { color: var(--text-muted); font-size: 0.74rem; }
.result-balance strong { font-size: 1rem; }
.result-reward { display: flex; align-items: baseline; justify-content: space-between; gap: 0.8rem; margin-top: 0.55rem; padding: 0.8rem; border: 1px solid #304138; border-radius: var(--radius-control); color: var(--accent); background: var(--accent-soft); }
.result-reward span { color: var(--text-muted); font-size: 0.74rem; }
.result-reward strong { text-align: right; }
</style>
