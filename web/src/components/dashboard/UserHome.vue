<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

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
const firstName = computed(() => dashboard.value?.user.firstName || '')

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
      <BalanceHero :balance="dashboard.balance" :first-name="firstName || $t('dashboard.friend')" />
      <ComingSoonLinks />
      <div class="page-toolbar page-toolbar--end">
        <UButton
          class="text-button"
          color="neutral"
          variant="ghost"
          icon="i-ph-arrow-clockwise"
          :label="$t('common.refresh')"
          :loading="refreshing"
          @click="load({ quiet: true })"
        />
      </div>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      <InlineNotice v-if="revoked" tone="success" :title="$t('dashboard.linkReplaced')">{{ $t('dashboard.previousLinkInvalid') }}</InlineNotice>
      <div class="content-grid content-grid--dashboard">
        <EntitlementSummary :active="dashboard.activePurchase" :queued="dashboard.queuedPurchase" />
        <UsagePanel
          v-if="dashboard.statistics"
          :statistics="dashboard.statistics"
          :ratio="usageRatio"
          :stale="dashboard.statisticsStale"
          :fetched-at="dashboard.fetchedAt"
        />
        <section v-else class="section-block empty-inline">
          <div><h3>{{ $t('dashboard.noStatistics') }}</h3><p>{{ $t('dashboard.statisticsPending') }}</p></div>
        </section>
        <SubscriptionPanel
          :subscription-url="dashboard.subscriptionUrl"
          :revoking="revoking"
          @revoke="confirmRevoke"
        />
      </div>
    </template>
    <div v-else class="error-state">
      <h1>{{ $t('dashboard.unavailable') }}</h1>
      <p>{{ error ?? $t('dashboard.loadFailed') }}</p>
      <UButton :label="$t('common.tryAgain')" @click="load()" />
    </div>
  </div>
</template>
