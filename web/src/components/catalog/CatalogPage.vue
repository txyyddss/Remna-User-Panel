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
  checkoutOpen,
  needsBalance,
  visibleCombos,
  visibleSquads,
  selectedCombo,
  selectedSquads,
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
      <p class="eyebrow">Plans and regions</p>
      <h1>Choose your combo.</h1>
      <p>One account-wide traffic budget. Optional squads follow the same term.</p>
    </header>

    <template v-if="loading">
      <div class="catalog-grid">
        <SkeletonBlock height="18rem" />
        <SkeletonBlock height="15rem" />
      </div>
    </template>
    <template v-else-if="catalog">
      <InlineNotice v-if="purchase" tone="success" title="Purchase confirmed">
        {{ purchase.comboName }} was scheduled. Remnawave activation may take a moment.
      </InlineNotice>
      <InlineNotice v-else-if="error && !checkoutOpen" tone="warning">{{ error }}</InlineNotice>

      <section class="combo-section">
        <div class="section-heading">
          <h2>Core combos</h2>
          <span v-if="balance" class="section-heading__meta">Balance {{ formatMoney(balance) }}</span>
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
            <h3>No combos are available</h3>
            <p>The administrator has not published a plan yet.</p>
          </div>
          <button class="button button--secondary" type="button" @click="load">Refresh</button>
        </div>
      </section>

      <SquadSelector :squads="visibleSquads" :selected-ids="selectedSquadIds" @toggle="toggleSquad" />

      <div class="catalog-action">
        <span v-if="selectedCombo">
          <PhCheckCircle :size="20" weight="fill" />
          {{ selectedCombo.name }} selected
        </span>
        <button class="button button--primary" type="button" :disabled="!selectedCombo" @click="reviewPurchase">
          Review purchase
          <PhArrowRight :size="19" />
        </button>
      </div>

      <CatalogCheckout
        v-if="balance"
        v-model:open="checkoutOpen"
        :combo="selectedCombo"
        :squads="selectedSquads"
        :balance="balance"
        :purchasing="purchasing"
        :needs-balance="needsBalance"
        :error="checkoutOpen ? error : null"
        @confirm="confirmPurchase"
      />
    </template>
    <div v-else class="error-state">
      <h1>Plans are unavailable.</h1>
      <p>{{ error }}</p>
      <button class="button button--primary" type="button" @click="load">Try again</button>
    </div>
  </div>
</template>
