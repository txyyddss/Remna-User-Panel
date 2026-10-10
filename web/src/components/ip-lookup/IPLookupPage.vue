<script setup lang="ts">
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import IPLookupForm from './IPLookupForm.vue'
import IPLookupReport from './IPLookupReport.vue'
import { useIPLookup } from './useIPLookup'
import { ipLookupErrorText } from './presentation'

const { ip, state, quote, refreshQuote, check, loading, busy, errorCode, validIP, submit, load } = useIPLookup()
</script>

<template>
  <main class="page ip-lookup-page">
    <header class="page-header">
      <div><p class="eyebrow">{{ $t('dashboard.aroundTx') }}</p><h1>{{ $t('ipLookup.title') }}</h1><p>{{ $t('ipLookup.subtitle') }}</p></div>
    </header>
    <SkeletonBlock v-if="loading" height="12rem" />
    <template v-else>
      <InlineNotice v-if="errorCode" tone="warning">{{ ipLookupErrorText(errorCode) }}</InlineNotice>
      <InlineNotice v-if="state && !state.enabled" tone="info">{{ $t('ipLookup.disabled') }}</InlineNotice>
      <IPLookupForm v-if="state?.enabled" v-model="ip" :state="state" :quote="quote" :refresh-quote="refreshQuote" :busy="busy" :valid-ip="validIP" @check="submit()" @refresh="submit(true)" />
      <p v-if="busy" role="status">{{ $t('ipLookup.processing') }}</p>
      <IPLookupReport v-if="check?.report" :check="check" />
      <p v-else-if="!busy && state?.enabled" class="ip-lookup-empty">{{ $t('ipLookup.empty') }}</p>
      <UButton v-if="!state" color="neutral" variant="outline" :label="$t('common.tryAgain')" @click="load" />
    </template>
  </main>
</template>

<style scoped>
.ip-lookup-page { max-width: 1080px; }
.ip-lookup-empty { padding: 2rem 0; color: var(--text-muted); }
</style>
