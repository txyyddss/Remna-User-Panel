<script setup lang="ts">
import { computed, type DeepReadonly } from 'vue'
import { motion } from 'motion-v'

import type { Purchase } from '@/api/types'
import { motionDurations } from '@/composables/motionPresets'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { useI18n } from '@/i18n'
import { formatBytes, formatDate, formatMoney } from '@/utils/format'

const props = defineProps<{
  purchase: DeepReadonly<Purchase>
}>()

const emit = defineEmits<{
  home: []
}>()

const { t } = useI18n()
const { reducedMotion, offset } = useMotionPreferences()

function statusLabel(): string {
  return t(`catalog.purchaseStatus.${props.purchase.status}`)
}

function resetLabel(): string {
  return t(`home.reset.${props.purchase.resetStrategy}`)
}

const summaryLines = computed(() => [
  { label: t('catalog.coreCombos'), value: props.purchase.comboName },
  { label: t('catalog.purchaseCharged'), value: formatMoney(props.purchase.price) },
  { label: t('catalog.purchaseDiscount'), value: formatMoney(props.purchase.couponDiscount) },
  { label: t('catalog.purchaseStarts'), value: formatDate(props.purchase.validFrom) },
  { label: t('catalog.purchaseEnds'), value: formatDate(props.purchase.validUntil) },
  { label: t('catalog.purchaseStatusLabel'), value: statusLabel() },
  { label: t('catalog.purchaseTraffic'), value: formatBytes(props.purchase.trafficLimitBytes) },
  { label: t('catalog.purchaseReset'), value: resetLabel() },
])
</script>

<template>
  <section class="catalog-confirmation" data-test="catalog-confirmation" role="status" aria-live="polite">
    <motion.div
      class="catalog-confirmation__hero"
      :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: offset(6) }"
      :animate="{ opacity: 1, y: 0 }"
      :transition="{ duration: reducedMotion ? 0.08 : motionDurations.normal, ease: 'easeOut' }"
    >
      <motion.div
        class="catalog-confirmation__icon"
        aria-hidden="true"
        :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, scale: 0.8 }"
        :animate="{ opacity: 1, scale: 1 }"
        :transition="{ duration: reducedMotion ? 0.08 : 0.3, ease: 'easeOut' }"
      >
        <UIcon name="i-ph-check-circle-fill" />
      </motion.div>
      <div class="catalog-confirmation__copy">
        <p class="catalog-confirmation__eyebrow">{{ $t('catalog.purchaseConfirmed') }}</p>
        <h1>{{ $t('catalog.purchaseSuccessTitle') }}</h1>
        <p>{{ $t('catalog.purchaseScheduled', { name: purchase.comboName }) }}</p>
      </div>
    </motion.div>

    <motion.div layout class="catalog-confirmation__summary" :aria-label="$t('catalog.purchaseSummary')">
      <motion.div
        v-for="(line, index) in summaryLines"
        :key="line.label"
        class="catalog-confirmation__line"
        :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: offset(4) }"
        :animate="{ opacity: 1, y: 0 }"
        :transition="{ duration: reducedMotion ? 0.08 : motionDurations.fast, delay: reducedMotion ? 0 : 0.12 + index * 0.03, ease: 'easeOut' }"
      >
        <span>{{ line.label }}</span>
        <strong>{{ line.value }}</strong>
      </motion.div>
    </motion.div>

    <UButton
      block
      class="catalog-confirmation__home-action"
      :ui="{ trailingIcon: 'absolute end-3 top-1/2 -translate-y-1/2' }"
      trailing-icon="i-ph-house"
      :label="$t('catalog.returnHome')"
      data-haptic="navigate"
      @click="emit('home')"
    />
  </section>
</template>

<style scoped>
.catalog-confirmation { display: grid; gap: 1rem; max-width: 42rem; margin: 0 auto; }
.catalog-confirmation__hero { display: grid; justify-items: center; gap: 0.8rem; padding: 1.5rem 1rem 1rem; text-align: center; }
.catalog-confirmation__icon { display: inline-flex; align-items: center; justify-content: center; width: 4.5rem; height: 4.5rem; border: 1px solid var(--line-strong); border-radius: 1.5rem; color: var(--accent); background: var(--accent-soft); font-size: 2.75rem; }
.catalog-confirmation__copy { display: grid; gap: 0.45rem; }
.catalog-confirmation__copy h1, .catalog-confirmation__copy p { margin: 0; }
.catalog-confirmation__copy h1 { font-size: 1.75rem; letter-spacing: 0; }
.catalog-confirmation__copy p:last-child { color: var(--text-muted); font-size: 0.82rem; line-height: 1.5; }
.catalog-confirmation__eyebrow { color: var(--accent); font-size: 0.7rem; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; }
.catalog-confirmation__summary { display: grid; gap: 0.65rem; padding: 0.9rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }
.catalog-confirmation__line { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; padding-bottom: 0.65rem; border-bottom: 1px solid var(--line); }
.catalog-confirmation__line:last-child { padding-bottom: 0; border-bottom: 0; }
.catalog-confirmation__line span { color: var(--text-faint); font-size: 0.7rem; }
.catalog-confirmation__line strong { overflow-wrap: anywhere; color: var(--text); font-family: var(--font-mono); font-size: 0.78rem; text-align: right; }
.catalog-confirmation__home-action { position: relative; justify-content: center; }
</style>
