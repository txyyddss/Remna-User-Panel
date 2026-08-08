<script setup lang="ts">
import { shallowRef } from 'vue'
import { PhPlus } from '@phosphor-icons/vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import CouponWalletPanel from '@/components/coupons/CouponWalletPanel.vue'
import { useBalance } from '@/composables/useBalance'
import { formatMoney } from '@/utils/format'
import BalancePaymentSheet from './BalancePaymentSheet.vue'
import LedgerList from './LedgerList.vue'

const paymentOpen = shallowRef(false)
const { balance, methods, ledger, loading, error, load } = useBalance()

function handlePaid(): void {
  void load()
}
</script>

<template>
  <div class="page page--balance">
    <header class="page-header">
      <p class="eyebrow">Wallet</p>
      <h1>Balance and activity.</h1>
      <p>Top up securely, then use TXB for combos and optional squads.</p>
    </header>

    <template v-if="loading">
      <SkeletonBlock height="14rem" />
      <SkeletonBlock height="20rem" />
    </template>
    <template v-else-if="balance">
      <section class="wallet-balance">
        <span>Available</span>
        <strong>{{ formatMoney(balance) }}</strong>
        <button class="button button--light" type="button" @click="paymentOpen = true">
          <PhPlus :size="19" weight="bold" />
          Add balance
        </button>
      </section>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      <CouponWalletPanel />
      <LedgerList :entries="ledger" />
      <BalancePaymentSheet v-model:open="paymentOpen" :methods="methods" @paid="handlePaid" />
    </template>
    <div v-else class="error-state">
      <h1>Wallet is unavailable.</h1>
      <p>{{ error }}</p>
      <button class="button button--primary" type="button" @click="load">Try again</button>
    </div>
  </div>
</template>
