<script setup lang="ts">
import {
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogOverlay,
  DialogPortal,
  DialogRoot,
  DialogTitle,
} from 'reka-ui'
import { PhCheckCircle, PhWarningCircle, PhX } from '@phosphor-icons/vue'

import type { ActivityResult } from '@/api/features'
import { formatMoney } from '@/utils/format'

defineProps<{ result: ActivityResult | null }>()
defineEmits<{ close: [] }>()
</script>

<template>
  <DialogRoot :open="Boolean(result)" @update:open="!$event && $emit('close')">
    <DialogPortal to="#overlays">
      <DialogOverlay class="dialog-overlay" />
      <DialogContent v-if="result" class="dialog-content activity-result-dialog">
        <header class="dialog-header">
          <span class="feature-icon" :class="{ 'feature-icon--warning': result.outcome === 'loss' }">
            <PhWarningCircle v-if="result.outcome === 'loss'" :size="25" weight="fill" />
            <PhCheckCircle v-else :size="25" weight="fill" />
          </span>
          <div>
            <DialogTitle class="dialog-title">{{ $t(`activity.result.${result.outcome}.title`) }}</DialogTitle>
            <DialogDescription class="dialog-description">{{ result.message }}</DialogDescription>
          </div>
          <DialogClose class="icon-button" :aria-label="$t('common.close')"><PhX :size="20" /></DialogClose>
        </header>
        <div class="result-balance">
          <span>{{ $t('activity.balanceAfter') }}</span>
          <strong>{{ formatMoney(result.balanceAfter) }}</strong>
        </div>
        <DialogClose class="button button--primary button--wide">{{ $t('common.close') }}</DialogClose>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>

<style scoped>
.activity-result-dialog { display: grid; gap: 1rem; }
.activity-result-dialog .dialog-header { display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: start; }
.result-balance { display: flex; align-items: baseline; justify-content: space-between; padding: 0.8rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.result-balance span { color: var(--text-muted); font-size: 0.74rem; }
.result-balance strong { font-size: 1rem; }
</style>
