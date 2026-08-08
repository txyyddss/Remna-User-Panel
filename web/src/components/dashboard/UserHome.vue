<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import { PhArrowClockwise } from '@phosphor-icons/vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useDashboard } from '@/composables/useDashboard'
import BalanceHero from './BalanceHero.vue'
import ComingSoonLinks from './ComingSoonLinks.vue'
import EntitlementSummary from './EntitlementSummary.vue'
import SubscriptionPanel from './SubscriptionPanel.vue'
import UsagePanel from './UsagePanel.vue'

const { dashboard, loading, refreshing, revoking, error, usageRatio, load, revokeSubscription } = useDashboard()
const revoked = shallowRef(false)
const firstName = computed(() => dashboard.value?.user.firstName || 'there')

watch(revoking, (next, previous) => {
  if (previous && !next) revoked.value = true
})

async function confirmRevoke(): Promise<void> {
  revoked.value = await revokeSubscription()
}
</script>

<template>
  <div class="page page--home">
    <template v-if="loading">
      <SkeletonBlock height="13rem" />
      <div class="content-grid">
        <SkeletonBlock height="18rem" />
        <SkeletonBlock height="18rem" />
      </div>
    </template>
    <template v-else-if="dashboard">
      <BalanceHero :balance="dashboard.balance" :first-name="firstName" />
      <div class="page-toolbar page-toolbar--end">
        <button class="text-button" type="button" :disabled="refreshing" @click="load({ quiet: true })">
          <PhArrowClockwise :size="17" :class="{ 'icon-spin': refreshing }" />
          Refresh
        </button>
      </div>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      <InlineNotice v-if="revoked" tone="success" title="Link replaced">Your previous subscription link is no longer valid.</InlineNotice>
      <div class="content-grid content-grid--dashboard">
        <EntitlementSummary :active="dashboard.activePurchase" :queued="dashboard.queuedPurchase" />
        <UsagePanel
          v-if="dashboard.statistics"
          :statistics="dashboard.statistics"
          :ratio="usageRatio"
          :stale="dashboard.statisticsStale"
          :warning="dashboard.statisticsWarning"
          :fetched-at="dashboard.fetchedAt"
        />
        <section v-else class="section-block empty-inline"><div><h3>No traffic statistics</h3><p>Usage appears after Remnawave provisioning completes.</p></div></section>
        <SubscriptionPanel
          :subscription-url="dashboard.subscriptionUrl"
          :revoking="revoking"
          @revoke="confirmRevoke"
        />
        <ComingSoonLinks />
      </div>
    </template>
    <div v-else class="error-state">
      <h1>Home is taking a pause.</h1>
      <p>{{ error ?? 'Your dashboard could not be loaded.' }}</p>
      <button class="button button--primary" type="button" @click="load()">Try again</button>
    </div>
  </div>
</template>
