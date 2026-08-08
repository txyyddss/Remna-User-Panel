<script setup lang="ts">
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useEmby } from '@/composables/useEmby'
import EmbyPreferencesPanel from './EmbyPreferencesPanel.vue'
import EmbySetupPanel from './EmbySetupPanel.vue'

const { overview, loading, busy, error, message, load, setup, updatePreferences, changePassword } = useEmby()
</script>

<template>
  <div class="page page--emby">
    <header class="page-header"><p class="eyebrow">Shared media</p><h1>Your Emby access.</h1><p>Choose libraries and age rating while account safety controls remain locked.</p></header>
    <SkeletonBlock v-if="loading" height="28rem" />
    <template v-else-if="overview">
      <InlineNotice v-if="message" tone="success">{{ message }}</InlineNotice>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      <InlineNotice v-if="!overview.configured" tone="warning" title="Emby is not configured">An administrator must connect Emby before setup can start.</InlineNotice>
      <EmbyPreferencesPanel
        v-if="overview.account && !(overview.account.status === 'failed' && !overview.account.retryable)"
        :account="overview.account"
        :ratings="overview.ratings"
        :libraries="overview.libraries"
        :busy="busy"
        @save="updatePreferences"
        @change-password="changePassword"
      />
      <EmbySetupPanel
        v-else-if="overview.configured"
        :price="overview.setupPrice"
        :ratings="overview.ratings"
        :libraries="overview.libraries"
        :busy="busy === 'setup'"
        @setup="setup"
      />
    </template>
    <div v-else class="error-state"><h1>Emby is unavailable.</h1><p>{{ error }}</p><button class="button button--primary" type="button" @click="load">Try again</button></div>
  </div>
</template>
