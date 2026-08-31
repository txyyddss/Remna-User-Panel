<script setup lang="ts">
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useCommunityMembership } from '@/composables/useCommunityMembership'
import CommunityMembershipRows from './CommunityMembershipRows.vue'

const { membership, loading, refreshing, joining, error, load, join } = useCommunityMembership()
</script>

<template>
  <main class="page page--community">
    <header class="page-header community-page__header">
      <p class="eyebrow">{{ $t('dashboard.aroundTx') }}</p>
      <h1>{{ $t('community.title') }}</h1>
      <p>{{ $t('community.subtitle') }}</p>
    </header>

    <InlineNotice tone="info">{{ $t('community.accessNote') }}</InlineNotice>

    <section v-if="loading" class="community-page__loading" aria-live="polite">
      <h2>{{ $t('community.loadingTitle') }}</h2>
      <p>{{ $t('community.loadingDescription') }}</p>
      <SkeletonBlock height="9.5rem" />
    </section>
    <template v-else>
      <InlineNotice v-if="error" tone="warning" :title="$t('community.errorTitle')">{{ error }}</InlineNotice>
      <CommunityMembershipRows
        v-if="membership"
        :active-combo="membership.activeCombo"
        :group-joined="membership.groupJoined"
        :channel-joined="membership.channelJoined"
        :joining="joining"
        @join="join"
      />
      <UButton
        color="neutral"
        variant="outline"
        size="lg"
        class="community-page__refresh"
        :loading="refreshing"
        :label="$t('common.refresh')"
        data-haptic="refresh"
        @click="load()"
      />
    </template>
  </main>
</template>

<style scoped>
.page--community { display: grid; gap: 1rem; padding-bottom: max(1rem, var(--tg-content-safe-area-inset-bottom, env(safe-area-inset-bottom))); }
.community-page__header { margin-bottom: 0.15rem; }
.community-page__header p:last-child, .community-page__loading p { margin: 0.35rem 0 0; color: var(--text-muted); }
.community-page__loading { display: grid; gap: 0.75rem; }
.community-page__loading h2 { margin: 0; font-size: 1rem; }
.community-page__refresh { justify-self: start; min-height: 44px; }
</style>
