<script setup lang="ts">
import { shallowRef } from 'vue'
import { PhArrowRight, PhTicket } from '@phosphor-icons/vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import { useCoupons } from '@/composables/useCoupons'
import { formatDate } from '@/utils/format'

const { grants, loading, redeeming, error, message, redeem } = useCoupons()
const code = shallowRef('')

async function submit(): Promise<void> {
  if (await redeem(code.value)) code.value = ''
}
</script>

<template>
  <section class="section-block coupon-wallet">
    <div class="section-heading section-heading--stacked"><h2>Coupon wallet</h2><p>Redeem a code now, then choose one eligible grant during checkout.</p></div>
    <form class="coupon-redeem" @submit.prevent="submit"><label class="sr-only" for="coupon-code">Coupon code</label><span class="input-shell"><PhTicket :size="19" /><input id="coupon-code" v-model.trim="code" placeholder="ENTER-CODE" autocomplete="off" /></span><button class="button button--secondary" type="submit" :disabled="redeeming || !code"><PhArrowRight :size="18" />{{ redeeming ? 'Redeeming' : 'Redeem' }}</button></form>
    <InlineNotice v-if="message" tone="success">{{ message }}</InlineNotice>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <p v-if="loading" class="field-hint">Loading coupons</p>
    <div v-else-if="grants.length" class="coupon-list"><article v-for="grant in grants" :key="grant.id"><span class="feature-icon feature-icon--small"><PhTicket :size="18" /></span><div><strong>{{ grant.coupon.name }}</strong><small>{{ grant.coupon.code }} · {{ grant.coupon.kind.replaceAll('_', ' ') }} · {{ grant.coupon.expiresAt ? `expires ${formatDate(grant.coupon.expiresAt)}` : 'no expiry' }}</small></div><span>{{ grant.coupon.perUserUseLimit === null ? '∞' : Math.max(0, grant.coupon.perUserUseLimit - grant.useCount) }} uses</span></article></div>
    <div v-else class="empty-inline"><div><h3>No saved coupons</h3><p>Codes and lucky-draw coupon prizes appear here.</p></div></div>
  </section>
</template>

<style scoped>
.coupon-wallet { display: grid; gap: 0.8rem; }
.coupon-redeem { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 0.55rem; }
.coupon-list { display: grid; gap: 0.5rem; }
.coupon-list article { min-height: 58px; display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 0.65rem; padding: 0.55rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.coupon-list strong, .coupon-list small { display: block; }
.coupon-list strong { font-size: 0.78rem; }
.coupon-list small { margin-top: 0.2rem; color: var(--text-faint); font-size: 0.62rem; }
.coupon-list article > span:last-child { color: var(--accent); font-family: var(--font-mono); font-size: 0.65rem; }
</style>
