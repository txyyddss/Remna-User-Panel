<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useDashboard } from '@/composables/useDashboard'
import BalanceHero from './BalanceHero.vue'
import ComingSoonLinks from './ComingSoonLinks.vue'
import EntitlementSummary from './EntitlementSummary.vue'
import SubscriptionPanel from './SubscriptionPanel.vue'
import UsagePanel from './UsagePanel.vue'

const { dashboard, loading, revoking, error, usageRatio, activeSquadNames, load, revokeSubscription } = useDashboard()
const revoked = shallowRef(false)
const route = useRoute()
const router = useRouter()
const reissueOrderId = computed(() => typeof route.query.reissue === 'string' && route.query.reissue ? route.query.reissue : undefined)
const topUpRequested = computed(() => route.query.topUp === '1' && !reissueOrderId.value)

watch(revoking, (next, previous) => {
  if (previous && !next) revoked.value = true
})

async function confirmRevoke(): Promise<void> {
  revoked.value = await revokeSubscription()
}

function consumeTopUpRequest(): void {
  if (!topUpRequested.value) return
  const query = { ...route.query }
  delete query.topUp
  void router.replace({ name: 'home', query })
}

function consumeReissueRequest(): void {
  if (!reissueOrderId.value) return
  const query = { ...route.query }
  delete query.reissue
  void router.replace({ name: 'home', query })
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
      <BalanceHero
        :balance="dashboard.balance"
        :open-top-up="topUpRequested"
        :reissue-order-id="reissueOrderId"
        @paid="load({ quiet: true })"
        @top-up-request-consumed="consumeTopUpRequest"
        @reissue-request-consumed="consumeReissueRequest"
      />
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      <div class="home-stack">
        <SubscriptionPanel
          :subscription-url="dashboard.subscriptionUrl"
          :revoking="revoking"
          @revoke="confirmRevoke"
        />
        <InlineNotice v-if="revoked" tone="success" :title="$t('dashboard.linkReplaced')">{{ $t('dashboard.previousLinkInvalid') }}</InlineNotice>
        <UsagePanel
          v-if="dashboard.statistics"
          :statistics="dashboard.statistics"
          :ratio="usageRatio"
          :stale="dashboard.statisticsStale"
          :fetched-at="dashboard.fetchedAt"
          :term="dashboard.activePurchase"
        />
        <section v-else class="section-block home-usage home-usage--empty empty-inline">
          <div><h3>{{ $t('dashboard.noStatistics') }}</h3><p>{{ $t('dashboard.statisticsPending') }}</p></div>
        </section>
        <EntitlementSummary :active="dashboard.activePurchase" :queued="dashboard.queuedPurchase" :squad-names="activeSquadNames" />
        <ComingSoonLinks />
      </div>
    </template>
    <div v-else class="error-state">
      <h1>{{ $t('dashboard.unavailable') }}</h1>
      <p>{{ error ?? $t('dashboard.loadFailed') }}</p>
      <UButton :label="$t('common.tryAgain')" @click="load()" />
    </div>
  </div>
</template>
