<script setup lang="ts">
import { computed } from 'vue'

import type { ActivityResult } from '@/api/features'
import { useI18n } from '@/i18n'
import { formatMoney } from '@/utils/format'

const props = defineProps<{ result: ActivityResult | null }>()
defineEmits<{ close: [] }>()

const { t } = useI18n()

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
    </template>
    <template #footer="{ close }">
      <UButton block :label="$t('common.close')" @click="close" />
    </template>
  </UModal>
</template>

<style scoped>
.result-balance { display: flex; align-items: baseline; justify-content: space-between; padding: 0.8rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.result-balance span { color: var(--text-muted); font-size: 0.74rem; }
.result-balance strong { font-size: 1rem; }
</style>
