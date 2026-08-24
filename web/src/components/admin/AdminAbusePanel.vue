<script setup lang="ts">
import { abuseApi } from '@/api/abuse'
import AbuseNodesRecordsCard from '@/components/admin/abuse/AbuseNodesRecordsCard.vue'
import AbusePolicyCard from '@/components/admin/abuse/AbusePolicyCard.vue'
import AbuseRulesCard from '@/components/admin/abuse/AbuseRulesCard.vue'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useAdminAbuse } from '@/composables/useAdminAbuse'
import { useClipboard } from '@/composables/useClipboard'
import { useI18n } from '@/i18n'
import AdminSectionState from './AdminSectionState.vue'

const { t } = useI18n()
const state = useAdminAbuse()
const clipboard = useClipboard()

function copy(id: string): void {
  void state.execute(async () => {
    const value = await abuseApi.copyNodeKey(id)
    await clipboard.copy(value.key)
  }, { errorKey: 'adminAbuse.copyFailed', reload: false })
}

function rotate(id: string): void {
  void state.execute(async () => {
    const value = await abuseApi.rotateNodeKey(id)
    await clipboard.copy(value.key)
  }, { errorKey: 'adminAbuse.rotateFailed' })
}
</script>

<template>
  <section class="admin-panel">
    <header class="admin-panel__heading">
      <div>
        <p class="eyebrow">{{ t('adminAbuse.eyebrow') }}</p>
        <h2>{{ t('adminAbuse.title') }}</h2>
        <p>{{ t('adminAbuse.copy') }}</p>
      </div>
    </header>

    <InlineNotice v-if="state.error.value" tone="warning">
      {{ state.error.value }}
    </InlineNotice>

    <AdminSectionState :loading="state.loading.value" :error="state.error.value" @retry="state.load">
      <div v-if="state.policy.value" class="abuse-admin__grid">
        <AbusePolicyCard
          :policy="state.policy.value"
          :punishments="state.punishments.value"
          :busy="state.busy.value"
          @save-policy="value => state.execute(async () => { await abuseApi.savePolicy(value) })"
          @save-punishment="value => state.execute(async () => { await abuseApi.savePunishment(value) })"
        />
        <AbuseRulesCard
          :rules="state.rules.value"
          :whitelist="state.whitelist.value"
          :busy="state.busy.value"
          @save-rule="value => state.execute(async () => { await abuseApi.saveRule(value) })"
          @delete-rule="(id, revision) => state.execute(async () => { await abuseApi.deleteRule(id, revision) })"
          @whitelist="(id, enabled) => state.execute(async () => { await abuseApi.setWhitelist(id, enabled) })"
        />
        <AbuseNodesRecordsCard
          :nodes="state.nodes.value"
          :records="state.records.value"
          :statistics="state.statistics.value"
          :busy="state.busy.value"
          @copy="copy"
          @rotate="rotate"
        />
      </div>
    </AdminSectionState>
  </section>
</template>

<style scoped>
.admin-panel,
.abuse-admin__grid {
  display: grid;
  gap: 1rem;
}

.abuse-admin__grid {
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 22rem), 1fr));
  align-items: start;
  padding-bottom: max(1rem, env(safe-area-inset-bottom));
}
</style>
