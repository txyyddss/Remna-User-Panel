<script setup lang="ts">
import type { DeepReadonly } from 'vue'
import { useRouter } from 'vue-router'

import type { Purchase, PurchaseAddonQuote, SquadProduct } from '@/api/types'
import { formatDate, formatMoney } from '@/utils/format'

defineProps<{
  squads: readonly SquadProduct[]
  quote: DeepReadonly<PurchaseAddonQuote> | null
  purchase: DeepReadonly<Purchase> | null
  quoting: boolean
  purchasing: boolean
  needsBalance: boolean
  error: string | null
}>()
const emit = defineEmits<{ back: []; confirm: []; home: [] }>()
const router = useRouter()

function goToBalance(): void {
  void router.push({ path: '/home', query: { topUp: '1' } })
}
</script>

<template>
  <section class="squad-addition-checkout" aria-live="polite">
    <template v-if="purchase">
      <div class="squad-addition-checkout__success">
        <UIcon name="i-ph-check-circle-fill" aria-hidden="true" />
        <div><strong>{{ $t('home.squadAddition.successTitle') }}</strong><p>{{ $t('home.squadAddition.successDescription') }}</p></div>
      </div>
      <UButton block trailing-icon="i-ph-house" :label="$t('catalog.returnHome')" data-haptic="navigate" @click="emit('home')" />
    </template>
    <template v-else>
      <UButton class="squad-addition-checkout__back" color="neutral" variant="ghost" icon="i-ph-arrow-left" :label="$t('catalog.back')" data-haptic="navigate" @click="emit('back')" />
      <div class="section-heading section-heading--stacked"><h2>{{ $t('home.squadAddition.checkoutTitle') }}</h2><p>{{ $t('home.squadAddition.checkoutDescription') }}</p></div>
      <div class="squad-addition-checkout__summary">
        <div><span>{{ $t('catalog.optionalSquads') }}</span><strong>{{ squads.map((squad) => squad.name).join($t('home.squadSeparator')) }}</strong></div>
        <div v-if="quote"><span>{{ $t('home.squadAddition.validUntil') }}</span><strong>{{ formatDate(quote.expiresAt) }}</strong></div>
        <div class="squad-addition-checkout__total"><span>{{ $t('home.squadAddition.proratedTotal') }}</span><strong>{{ quote ? formatMoney(quote.price) : quoting ? $t('catalog.quoting') : $t('common.notAvailable') }}</strong></div>
      </div>
      <UAlert v-if="error" color="warning" variant="soft" icon="i-ph-warning-circle" :description="error" />
      <USkeleton v-if="quoting" class="h-16" />
      <UButton v-if="needsBalance" block trailing-icon="i-ph-plus" :label="$t('catalog.addBalance')" data-haptic="navigate" @click="goToBalance" />
      <UButton v-else block :disabled="purchasing || !quote" :loading="purchasing" trailing-icon="i-ph-check" :label="purchasing ? $t('catalog.confirming') : $t('home.squadAddition.confirm')" data-haptic="confirm" @click="emit('confirm')" />
    </template>
  </section>
</template>

<style scoped>
.squad-addition-checkout, .squad-addition-checkout__summary { display: grid; gap: 0.8rem; }
.squad-addition-checkout__back { justify-self: start; padding-inline: 0; }
.squad-addition-checkout__summary { padding: 0.85rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }
.squad-addition-checkout__summary > div { display: grid; gap: 0.25rem; padding-bottom: 0.65rem; border-bottom: 1px solid var(--line); }
.squad-addition-checkout__summary > div:last-child { padding-bottom: 0; border-bottom: 0; }
.squad-addition-checkout__summary span { color: var(--text-faint); font-size: 0.68rem; }
.squad-addition-checkout__summary strong { overflow-wrap: anywhere; font-size: 0.8rem; }
.squad-addition-checkout__total strong { color: var(--accent); font-family: var(--font-mono); font-size: 1.35rem; }
.squad-addition-checkout__success { display: flex; align-items: flex-start; gap: 0.75rem; padding: 0.9rem; border: 1px solid var(--success); border-radius: var(--radius-panel); background: var(--success-soft); }
.squad-addition-checkout__success > :first-child { flex: 0 0 auto; color: var(--success); font-size: 1.5rem; }
.squad-addition-checkout__success p { margin: 0.3rem 0 0; color: var(--text-muted); font-size: 0.8rem; }
</style>
