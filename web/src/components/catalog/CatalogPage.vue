<script setup lang="ts">
import { computed, onMounted, shallowRef, watch } from 'vue'
import { useRouter } from 'vue-router'

import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useCatalog } from '@/composables/useCatalog'
import { useI18n } from '@/i18n'
import { useSessionStore } from '@/stores/session'
import CatalogCheckout from './CatalogCheckout.vue'
import CatalogConfirmation from './CatalogConfirmation.vue'
import CatalogCouponStep from './CatalogCouponStep.vue'
import CatalogFlowControls from './CatalogFlowControls.vue'
import CatalogFlowProgress from './CatalogFlowProgress.vue'
import CatalogNodes from './CatalogNodes.vue'
import ComboOption from './ComboOption.vue'
import SquadSelector from './SquadSelector.vue'
import SquadActivationDialog from './SquadActivationDialog.vue'
import type { SquadProduct } from '@/api/types'

const activeStep = shallowRef(1)
const sessionStore = useSessionStore()
const router = useRouter()
const stepKey = () => sessionStore.user?.id ? `txc-catalog-step:${sessionStore.user.id}` : null
const { t } = useI18n()
const {
  catalog,
  balance,
  loading,
  purchasing,
  quoting,
  error,
  purchase,
  quote,
  quoteUsable,
  autoRenewalBlocked,
  selectedComboId,
  selectedSquadIds,
  selectedCouponGrantId,
  couponDiscarding,
  needsBalance,
  visibleCombos,
  visibleSquads,
  selectedCombo,
  selectedSquads,
  activationSquads,
  includedSquadIds,
  couponGrants,
  eligibleCoupons,
  load,
  selectCombo,
  toggleSquad,
  refreshQuote,
  discardCoupon,
  confirmPurchase,
} = useCatalog()

async function restoreStepQuote(): Promise<void> {
  if (![3, 4, 5].includes(activeStep.value) || loading.value || quoting.value || quoteUsable.value || purchase.value || !selectedCombo.value) return
  await refreshQuote()
}

onMounted(() => {
  try {
    const value = Number(globalThis.sessionStorage?.getItem(stepKey() ?? ''))
    if (value >= 1 && value <= 5) activeStep.value = value
  } catch { /* Storage is optional in restricted WebViews. */ }
  void restoreStepQuote()
})
watch(activeStep, (value) => {
  try { const key = stepKey(); if (key) globalThis.sessionStorage?.setItem(key, String(value)) } catch { /* Ignore unavailable storage. */ }
})
watch([activeStep, loading, selectedCombo, selectedSquadIds, selectedCouponGrantId], () => { void restoreStepQuote() }, { deep: true })
watch(autoRenewalBlocked, (blocked) => {
  if (blocked) void router.replace({ path: '/home', query: { autoRenewBlocked: '1' } })
})

const selectedCoupon = computed(() => eligibleCoupons.value.find((grant) => grant.id === selectedCouponGrantId.value) ?? null)
const nextDisabled = computed(() => {
  if (activeStep.value === 1 || activeStep.value === 2) return !selectedCombo.value
  if (activeStep.value === 3) return !quoteUsable.value || !quote.value || quote.value.accessibleNodes.length === 0
  if (activeStep.value === 4) return !quoteUsable.value
  return true
})
const nextLabel = computed(() => t('catalog.continue'))
const activationOpen = shallowRef(false)
const activationTarget = shallowRef<SquadProduct | null>(null)
const activationPrompting = shallowRef(false)
let activationResolver: ((code: string | null) => void) | null = null

function goBack(): void {
  if (activeStep.value > 1) activeStep.value -= 1
}

function clearPersistedStep(): void {
  try {
    const key = stepKey()
    if (key) globalThis.sessionStorage?.removeItem(key)
  } catch { /* Storage is optional in restricted WebViews. */ }
}

async function handlePurchase(): Promise<void> {
  if (activationPrompting.value || purchasing.value) return
  activationPrompting.value = true
  try {
    const codes: Record<string, string> = {}
    for (const squad of activationSquads.value) {
      const code = await requestActivationCode(squad)
      if (!code) return
      codes[squad.remnaSquadUuid] = code
    }
    if (await confirmPurchase(codes)) clearPersistedStep()
  } finally {
    activationPrompting.value = false
  }
}

function requestActivationCode(squad: SquadProduct): Promise<string | null> {
  activationTarget.value = squad
  activationOpen.value = true
  return new Promise((resolve) => { activationResolver = resolve })
}

function resolveActivation(code: string | null): void {
  const resolve = activationResolver
  activationResolver = null
  activationOpen.value = false
  activationTarget.value = null
  resolve?.(code)
}

function goHome(): void {
  void router.push('/home')
}

async function advance(): Promise<void> {
  if (!selectedCombo.value || activeStep.value >= 5) return
  if (activeStep.value === 2 && !(await refreshQuote())) return
  if (activeStep.value === 4 && (!quoteUsable.value || quoting.value)) return
  activeStep.value += 1
}

async function handleCouponRedeemed(grantId: string | null): Promise<void> {
  await load()
  if (grantId && eligibleCoupons.value.some((grant) => grant.id === grantId)) selectedCouponGrantId.value = grantId
}
</script>

<template>
  <div class="page page--catalog">
    <template v-if="purchase">
      <CatalogConfirmation :purchase="purchase" @home="goHome" />
    </template>
    <template v-else>
      <header class="page-header">
        <h1>{{ $t('catalog.title') }}</h1>
      </header>

      <CatalogFlowProgress v-model="activeStep" />
      <template v-if="loading">
        <div class="catalog-grid">
          <SkeletonBlock height="18rem" />
          <SkeletonBlock height="15rem" />
        </div>
      </template>
      <template v-else-if="catalog && balance">
        <InlineNotice v-if="error && activeStep < 5" tone="warning">{{ error }}</InlineNotice>
        <Transition name="route" mode="out-in">
          <section :key="activeStep" class="catalog-flow-step">
            <div v-if="activeStep === 1" class="combo-section">
              <div class="section-heading"><h2>{{ $t('catalog.coreCombos') }}</h2></div>
              <div v-if="visibleCombos.length" v-auto-animate class="combo-grid">
                <ComboOption v-for="combo in visibleCombos" :key="combo.id" :combo="combo" :selected="selectedComboId === combo.id" @select="selectCombo" />
              </div>
              <div v-else class="empty-inline"><div><h3>{{ $t('catalog.noCombos') }}</h3><p>{{ $t('catalog.noCombosHint') }}</p></div><UButton color="neutral" variant="outline" :label="$t('common.refresh')" data-haptic @click="load" /></div>
            </div>
            <SquadSelector v-else-if="activeStep === 2" :squads="visibleSquads" :selected-ids="selectedSquadIds" :included-ids="includedSquadIds" @toggle="toggleSquad" />
            <CatalogNodes v-else-if="activeStep === 3" :quote="quote" :loading="quoting" />
            <CatalogCouponStep v-else-if="activeStep === 4" v-model:coupon-grant-id="selectedCouponGrantId" :coupons="couponGrants" :eligible-ids="eligibleCoupons.map((grant) => grant.id)" :discarding="couponDiscarding" :discard-coupon="discardCoupon" :quoting="quoting" @redeemed="handleCouponRedeemed" />
            <CatalogCheckout v-else-if="activeStep === 5" :combo="selectedCombo" :squads="selectedSquads" :coupon="selectedCoupon" :quote="quote" :quoting="quoting" :error="error" :purchasing="purchasing || activationPrompting" :needs-balance="needsBalance" @confirm="handlePurchase" />
          </section>
        </Transition>
        <CatalogFlowControls v-if="activeStep < 5" :show-back="activeStep > 1" :next-disabled="nextDisabled" :loading="quoting" :next-label="nextLabel" @back="goBack" @next="advance" />
      </template>
      <div v-else class="error-state">
        <h1>{{ $t('catalog.unavailable') }}</h1>
        <p>{{ error }}</p>
        <UButton :label="$t('common.tryAgain')" data-haptic @click="load" />
      </div>
    </template>
    <SquadActivationDialog v-model:open="activationOpen" :squad="activationTarget" @submit="resolveActivation" @cancel="resolveActivation(null)" />
  </div>
</template>

<style scoped>
.catalog-flow-step { min-height: 14rem; }
.combo-section { display: grid; gap: 0.8rem; }
</style>
