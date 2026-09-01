<script setup lang="ts">
import { computed, onMounted, shallowRef, watch } from 'vue'
import { useRouter } from 'vue-router'

import type { SquadProduct } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useCatalog } from '@/composables/useCatalog'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { useSessionStore } from '@/stores/session'
import CatalogCheckout from './CatalogCheckout.vue'
import CatalogConfirmation from './CatalogConfirmation.vue'
import CatalogCouponStep from './CatalogCouponStep.vue'
import CatalogFlowControls from './CatalogFlowControls.vue'
import CatalogFlowProgress from './CatalogFlowProgress.vue'
import CatalogSquadStep from './CatalogSquadStep.vue'
import CatalogComboPricingTable from './CatalogComboPricingTable.vue'
import SquadActivationDialog from './SquadActivationDialog.vue'
import { useCatalogSquadPresentation } from './useCatalogSquadPresentation'

const activeStep = shallowRef(1)
const sessionStore = useSessionStore()
const router = useRouter()
const stepKey = () => sessionStore.user?.id ? `txc-catalog-step:v2:${sessionStore.user.id}` : null
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
  eligibleCoupons,
  load,
  selectCombo,
  toggleSquad,
  refreshQuote,
  discardCoupon,
  confirmPurchase,
} = useCatalog()
const { featuredIds: featuredSquadIds, orderedIds: orderedSquadIds } = useCatalogSquadPresentation(visibleSquads, includedSquadIds, selectedComboId)

async function restoreStepQuote(): Promise<void> {
  if (![3, 4].includes(activeStep.value) || loading.value || quoting.value || quoteUsable.value || purchase.value || !selectedCombo.value) return
  await refreshQuote()
}

onMounted(() => {
  try {
    const value = Number(globalThis.sessionStorage?.getItem(stepKey() ?? ''))
    if (value >= 1 && value <= 4) activeStep.value = value
  } catch { /* Storage is optional in restricted WebViews. */ }
  void restoreStepQuote()
})
watch(activeStep, (value) => {
  try { const key = stepKey(); if (key) globalThis.sessionStorage?.setItem(key, String(value)) } catch { /* Ignore unavailable storage. */ }
  scrollCatalogToTop()
})
watch([activeStep, loading, selectedCombo, selectedSquadIds, selectedCouponGrantId], () => { void restoreStepQuote() }, { deep: true })
watch(autoRenewalBlocked, (blocked) => {
  if (blocked) void router.replace({ path: '/home', query: { autoRenewBlocked: '1' } })
})

const selectedCoupon = computed(() => eligibleCoupons.value.find((grant) => grant.id === selectedCouponGrantId.value) ?? null)
const nextDisabled = computed(() => {
  if (activeStep.value === 1 || activeStep.value === 2) return !selectedCombo.value
  if (activeStep.value === 3) return !quoteUsable.value || !quote.value || quote.value.accessibleNodes.length === 0
  return true
})
const activationOpen = shallowRef(false)
const activationTarget = shallowRef<SquadProduct | null>(null)
const activationPrompting = shallowRef(false)
let activationResolver: ((code: string | null) => void) | null = null

function goBack(): void {
  if (activeStep.value > 1) activeStep.value -= 1
}

const showCatalogBack = computed(() => !purchase.value && activeStep.value > 1)
useTelegramBackButton(showCatalogBack, goBack)

function clearPersistedStep(): void {
  try {
    const key = stepKey()
    if (key) globalThis.sessionStorage?.removeItem(key)
  } catch { /* Storage is optional in restricted WebViews. */ }
}

const catalogScrollOptions = { top: 0, left: 0, behavior: 'auto' as const }
function scrollCatalogToTop(): void { globalThis.scrollTo?.(catalogScrollOptions); globalThis.document?.querySelector<globalThis.HTMLElement>('.app-frame__content')?.scrollTo(catalogScrollOptions) }

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
  if (!selectedCombo.value || activeStep.value >= 4) return
  if (activeStep.value === 2 && (!(await refreshQuote()) || !quote.value?.accessibleNodes.length)) return
  if (activeStep.value === 3 && (!quoteUsable.value || quoting.value)) return
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
        <InlineNotice v-if="error && activeStep < 4" tone="warning">{{ error }}</InlineNotice>
        <section :key="activeStep" class="catalog-flow-step">
          <div v-if="activeStep === 2" class="combo-section">
            <div class="section-heading"><h2>{{ $t('catalog.coreCombos') }}</h2></div>
            <CatalogComboPricingTable v-if="visibleCombos.length" :combos="visibleCombos" :selected-id="selectedComboId" @select="selectCombo" />
            <div v-else class="empty-inline"><div><h3>{{ $t('catalog.noCombos') }}</h3><p>{{ $t('catalog.noCombosHint') }}</p></div><UButton color="neutral" variant="outline" :label="$t('common.refresh')" data-haptic="refresh" @click="load" /></div>
          </div>
          <CatalogSquadStep v-else-if="activeStep === 1" :squads="visibleSquads" :selected-ids="selectedSquadIds" :included-ids="includedSquadIds" :featured-ids="featuredSquadIds" :ordered-ids="orderedSquadIds" @toggle="toggleSquad" />
          <CatalogCouponStep v-else-if="activeStep === 3" v-model:coupon-grant-id="selectedCouponGrantId" :coupons="eligibleCoupons" :eligible-ids="eligibleCoupons.map((grant) => grant.id)" :discarding="couponDiscarding" :discard-coupon="discardCoupon" :quoting="quoting" @redeemed="handleCouponRedeemed" />
          <CatalogCheckout v-else-if="activeStep === 4" :combo="selectedCombo" :squads="selectedSquads" :coupon="selectedCoupon" :quote="quote" :quoting="quoting" :error="error" :purchasing="purchasing || activationPrompting" :needs-balance="needsBalance" @back="goBack" @confirm="handlePurchase" />
        </section>
        <CatalogFlowControls v-if="activeStep < 4" :show-back="activeStep > 1" :next-disabled="nextDisabled" :loading="quoting" :next-label="$t('catalog.continue')" @back="goBack" @next="advance" />
      </template>
      <div v-else class="error-state">
        <h1>{{ $t('catalog.unavailable') }}</h1>
        <p>{{ error }}</p>
        <UButton :label="$t('common.tryAgain')" data-haptic="retry" @click="load" />
      </div>
    </template>
    <SquadActivationDialog v-model:open="activationOpen" :squad="activationTarget" @submit="resolveActivation" @cancel="resolveActivation(null)" />
  </div>
</template>

<style scoped>
.catalog-flow-step { min-height: 14rem; }
.combo-section { display: grid; gap: 0.8rem; }
</style>
