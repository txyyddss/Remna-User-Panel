<script setup lang="ts">
import { shallowRef } from 'vue'

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
    <div class="section-heading section-heading--stacked">
      <h2>{{ $t('coupons.wallet') }}</h2>
      <p>{{ $t('coupons.walletHint') }}</p>
    </div>
    <form class="coupon-redeem" @submit.prevent="submit">
      <span id="coupon-code-label" class="sr-only">{{ $t('coupons.code') }}</span>
      <UInput
        id="coupon-code"
        v-model.trim="code"
        icon="i-ph-ticket"
        :placeholder="$t('coupons.codePlaceholder')"
        autocomplete="off"
        aria-labelledby="coupon-code-label"
      />
      <UButton
        type="submit"
        color="neutral"
        variant="outline"
        trailing-icon="i-ph-arrow-right"
        :disabled="redeeming || !code"
        :loading="redeeming"
        :label="redeeming ? $t('coupons.redeeming') : $t('coupons.redeem')"
      />
    </form>
    <InlineNotice v-if="message" tone="success">{{ message }}</InlineNotice>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <USkeleton v-if="loading" class="h-10" />
    <div v-else-if="grants.length" v-auto-animate class="coupon-list">
      <article v-for="grant in grants" :key="grant.id">
        <span class="feature-icon feature-icon--small"><UIcon name="i-ph-ticket" /></span>
        <div>
          <strong>{{ grant.coupon.name }}</strong>
          <small>
            {{ grant.coupon.code }} · {{ $t(`coupons.kind.${grant.coupon.kind}`) }} ·
            {{ grant.coupon.expiresAt ? $t('coupons.expires', { date: formatDate(grant.coupon.expiresAt) }) : $t('coupons.noExpiry') }}
          </small>
        </div>
        <span>{{ $t('coupons.uses', { count: grant.coupon.perUserUseLimit === null ? '∞' : Math.max(0, grant.coupon.perUserUseLimit - grant.useCount) }) }}</span>
      </article>
    </div>
    <div v-else class="empty-inline"><div><h3>{{ $t('coupons.empty') }}</h3><p>{{ $t('coupons.emptyHint') }}</p></div></div>
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
