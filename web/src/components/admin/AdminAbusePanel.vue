<script setup lang="ts">
import InlineNotice from '@/components/common/InlineNotice.vue'
import { abuseApi } from '@/api/abuse'
import { useAdminAbuse } from '@/composables/useAdminAbuse'
import { useClipboard } from '@/composables/useClipboard'
import { useI18n } from '@/i18n'
import AdminSectionState from './AdminSectionState.vue'
import AbuseNodesRecordsCard from './abuse/AbuseNodesRecordsCard.vue'
import AbusePolicyCard from './abuse/AbusePolicyCard.vue'
import AbuseRulesCard from './abuse/AbuseRulesCard.vue'

const { t } = useI18n(); const state = useAdminAbuse(); const clipboard = useClipboard()
function copy(id: string): void { void state.save(async () => { const value = await abuseApi.copyNodeKey(id); await clipboard.copy(value.key) }) }
function rotate(id: string): void { void state.save(async () => { const value = await abuseApi.rotateNodeKey(id); await clipboard.copy(value.key) }) }
</script>

<template><section class="admin-panel"><header class="admin-panel__heading"><div><p class="eyebrow">{{ t('adminAbuse.eyebrow') }}</p><h2>{{ t('adminAbuse.title') }}</h2><p>{{ t('adminAbuse.copy') }}</p></div></header><InlineNotice v-if="state.error.value" tone="warning">{{ state.error.value }}</InlineNotice><AdminSectionState :loading="state.loading.value" :error="state.error.value" @retry="state.load"><div v-if="state.policy.value" class="grid"><AbusePolicyCard :policy="state.policy.value" :punishments="state.punishments.value" :busy="state.busy.value" @save-policy="value => state.save(async () => { await abuseApi.savePolicy(value) })" @save-punishment="value => state.save(async () => { await abuseApi.savePunishment(value) })" /><AbuseRulesCard :rules="state.rules.value" :whitelist="state.whitelist.value" :busy="state.busy.value" @save-rule="value => state.save(async () => { await abuseApi.saveRule(value) })" @whitelist="(id, enabled) => state.save(async () => { await abuseApi.setWhitelist(id, enabled) })" /><AbuseNodesRecordsCard :nodes="state.nodes.value" :records="state.records.value" :statistics="state.statistics.value" :busy="state.busy.value" @copy="copy" @rotate="rotate" /></div></AdminSectionState></section></template>

<style scoped>.admin-panel,.grid { display:grid; gap:1rem; }.grid { grid-template-columns: repeat(auto-fit,minmax(280px,1fr)); align-items:start; padding-bottom:max(1rem,env(safe-area-inset-bottom)); }</style>
