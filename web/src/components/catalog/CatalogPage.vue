<script setup lang="ts">
import { PhArrowRight, PhCheckCircle } from '@phosphor-icons/vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useCatalog } from '@/composables/useCatalog'
import { formatMoney } from '@/utils/format'
import CatalogCheckout from './CatalogCheckout.vue'
import ComboOption from './ComboOption.vue'
import SquadSelector from './SquadSelector.vue'

const {
  catalog,
  balance,
  loading,
  purchasing,
  error,
  purchase,
  selectedComboId,
  selectedSquadIds,
  selectedCouponGrantId,
  checkoutOpen,
  needsBalance,
  visibleCombos,
  visibleSquads,
  selectedCombo,
  selectedSquads,
  includedSquadIds,
  eligibleCoupons,
  load,
  selectCombo,
  toggleSquad,
  reviewPurchase,
  confirmPurchase,
} = useCatalog()
</script>

<template>
  <div class="page page--catalog">
    <header class="page-header">
      <p class="eyebrow">{{ $t('catalog.eyebrow') }}</p>
      <h1>{{ $t('catalog.title') }}</h1>
      <p>{{ $t('catalog.copy') }}</p>
    </header>

    <template v-if="loading">
      <div class="catalog-grid">
        <SkeletonBlock height="18rem" />
        <SkeletonBlock height="15rem" />
      </div>
    </template>
    <template v-else-if="catalog">
      <InlineNotice v-if="purchase" tone="success" :title="$t('catalog.purchaseConfirmed')">
        {{ $t('catalog.purchaseScheduled', { name: purchase.comboName }) }}
      </InlineNotice>
      <InlineNotice v-else-if="error && !checkoutOpen" tone="warning">{{ error }}</InlineNotice>

      <section class="combo-section">
        <div class="section-heading">
          <h2>{{ $t('catalog.coreCombos') }}</h2>
          <span v-if="balance" class="section-heading__meta">{{ $t('activity.balance', { amount: formatMoney(balance) }) }}</span>
        </div>
        <div v-if="visibleCombos.length" class="combo-grid">
          <ComboOption
            v-for="combo in visibleCombos"
            :key="combo.id"
            :combo="combo"
            :selected="selectedComboId === combo.id"
            @select="selectCombo"
          />
        </div>
        <div v-else class="empty-inline">
          <div>
            <h3>{{ $t('catalog.noCombos') }}</h3>
            <p>{{ $t('catalog.noCombosHint') }}</p>
          </div>
          <button class="button button--secondary" type="button" @click="load">{{ $t('common.refresh') }}</button>
        </div>
      </section>

      <SquadSelector :squads="visibleSquads" :selected-ids="selectedSquadIds" :included-ids="includedSquadIds" @toggle="toggleSquad" />

      <div class="catalog-action">
        <span v-if="selectedCombo">
          <PhCheckCircle :size="20" weight="fill" />
          {{ $t('catalog.selected', { name: selectedCombo.name }) }}
        </span>
        <button class="button button--primary" type="button" :disabled="!selectedCombo" @click="reviewPurchase">
          {{ $t('catalog.reviewPurchase') }}
          <PhArrowRight :size="19" />
        </button>
      </div>

      <CatalogCheckout
        v-if="balance"
        v-model:open="checkoutOpen"
        :combo="selectedCombo"
        :squads="selectedSquads"
        :coupons="eligibleCoupons"
        v-model:coupon-grant-id="selectedCouponGrantId"
        :balance="balance"
        :purchasing="purchasing"
        :needs-balance="needsBalance"
        :error="checkoutOpen ? error : null"
        @confirm="confirmPurchase"
      />
    </template>
    <div v-else class="error-state">
      <h1>{{ $t('catalog.unavailable') }}</h1>
      <p>{{ error }}</p>
      <button class="button button--primary" type="button" @click="load">{{ $t('common.tryAgain') }}</button>
    </div>
  </div>
</template>
