<script setup lang="ts">
import { subscriptionApi } from '@/api/subscription'
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useI18n } from '@/i18n'
import UptimeBar from './UptimeBar.vue'
import { useSubscriptionData } from './useSubscriptionData'

const { value: summary, loading, error, refresh } = useSubscriptionData(subscriptionApi.summary)
const { t } = useI18n()
</script>

<template>
  <section v-if="!summary || summary.activeCombo" class="section-block home-uptime">
    <SkeletonBlock v-if="loading" height="6rem" />
    <UptimeBar v-else-if="summary" :timeline="summary" :label="t('uptime.summary')" />
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <p v-else-if="summary?.errorCode" class="home-uptime__hint">{{ t('uptime.notAvailable') }}</p>
    <div class="home-uptime__entrance">
      <UButton to="/subscription" color="neutral" variant="link" icon="i-ph-link-bold" :label="t('subscription.title')" data-haptic="navigate" />
      <UButton v-if="error" color="neutral" variant="ghost" :label="t('common.tryAgain')" @click="refresh()" />
    </div>
  </section>
</template>

<style scoped>
.home-uptime { display: grid; gap: 0.5rem; }
.home-uptime__entrance { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 0.5rem; border-top: 1px solid var(--line); padding-top: 0.45rem; }
.home-uptime__entrance :deep(a) { padding-left: 0; min-height: 44px; }
.home-uptime__hint { margin: 0; color: var(--text-muted); font-size: 0.75rem; }
</style>
