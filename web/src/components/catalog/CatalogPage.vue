<script setup lang="ts">
import { computed, shallowRef } from 'vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useCatalog } from '@/composables/useCatalog'
import { useI18n } from '@/i18n'
import CatalogCheckout from './CatalogCheckout.vue'
import CatalogCouponStep from './CatalogCouponStep.vue'
import CatalogFlowControls from './CatalogFlowControls.vue'
import CatalogFlowProgress from './CatalogFlowProgress.vue'
import CatalogNodes from './CatalogNodes.vue'
import CatalogPaymentStep from './CatalogPaymentStep.vue'
import ComboOption from './ComboOption.vue'
import SquadSelector from './SquadSelector.vue'

const activeStep = shallowRef(1)
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
  selectedComboId,
  selectedSquadIds,
  selectedCouponGrantId,
  couponDiscarding,
  needsBalance,
  visibleCombos,
  visibleSquads,
  selectedCombo,
  selectedSquads,
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

const selectedCoupon = computed(() => eligibleCoupons.value.find((grant) => grant.id === selectedCouponGrantId.value) ?? null)
const nextDisabled = computed(() => {
  if (activeStep.value === 1 || activeStep.value === 2) return !selectedCombo.value
  return activeStep.value === 5 && !quote.value
})
const nextLabel = computed(() => activeStep.value === 5 ? t('catalog.continuePayment') : t('catalog.continue'))

function goBack(): void {
  if (activeStep.value > 1) activeStep.value -= 1
}

async function advance(): Promise<void> {
  if (!selectedCombo.value || activeStep.value >= 6) return
  if ((activeStep.value === 2 || activeStep.value === 4) && !(await refreshQuote())) return
  activeStep.value += 1
}

async function handleCouponRedeemed(grantId: string | null): Promise<void> {
  await load()
  if (grantId && eligibleCoupons.value.some((grant) => grant.id === grantId)) selectedCouponGrantId.value = grantId
}
</script>

<template>
  <div class="page page--catalog">
    <header class="page-header">
      <p class="eyebrow">{{ $t('catalog.eyebrow') }}</p>
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
          <CatalogCouponStep v-else-if="activeStep === 4" v-model:coupon-grant-id="selectedCouponGrantId" :coupons="couponGrants" :eligible-ids="eligibleCoupons.map((grant) => grant.id)" :discarding="couponDiscarding" :discard-coupon="discardCoupon" @redeemed="handleCouponRedeemed" />
          <CatalogCheckout v-else-if="activeStep === 5" :combo="selectedCombo" :squads="selectedSquads" :coupon="selectedCoupon" :balance="balance" :quote="quote" :quoting="quoting" :error="error" />
          <CatalogPaymentStep v-else :combo="selectedCombo" :balance="balance" :quote="quote" :purchase="purchase" :purchasing="purchasing" :needs-balance="needsBalance" :error="error" @confirm="confirmPurchase" />
        </section>
      </Transition>
      <CatalogFlowControls v-if="activeStep < 6" :show-back="activeStep > 1" :next-disabled="nextDisabled" :loading="quoting" :next-label="nextLabel" @back="goBack" @next="advance" />
    </template>
    <div v-else class="error-state">
      <h1>{{ $t('catalog.unavailable') }}</h1>
      <p>{{ error }}</p>
      <UButton :label="$t('common.tryAgain')" data-haptic @click="load" />
    </div>
  </div>
</template>

<style scoped>
.catalog-flow-step { min-height: 14rem; }
.combo-section { display: grid; gap: 0.8rem; }
</style>
