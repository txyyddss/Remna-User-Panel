<script setup lang="ts">
import { abuseApi } from '@/api/abuse'
import AbuseNodesCard from '@/components/admin/abuse/AbuseNodesCard.vue'
import AbusePolicyCard from '@/components/admin/abuse/AbusePolicyCard.vue'
import AbusePunishmentLadder from '@/components/admin/abuse/AbusePunishmentLadder.vue'
import AbuseRecordsCard from '@/components/admin/abuse/AbuseRecordsCard.vue'
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
  }, { errorKey: 'adminAbuse.copyFailed', reload: false, successHaptic: 'copy' })
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
      <UButton
        color="neutral"
        variant="outline"
        icon="i-ph-arrow-clockwise"
        :label="t('common.refresh')"
        :loading="state.loading.value"
        @click="state.load"
      />
    </header>

    <AdminSectionState :loading="state.loading.value" :error="state.loadError.value" @retry="state.load">
      <div v-if="state.policy.value" class="abuse-admin__grid">
        <InlineNotice v-if="state.operationError.value" class="abuse-admin__notice" tone="warning">
          {{ state.operationError.value }}
        </InlineNotice>
        <AbusePolicyCard
          :policy="state.policy.value"
          :busy="state.busy.value"
          @save="value => state.execute(async () => { await abuseApi.savePolicy(value) })"
        />
        <AbusePunishmentLadder
          :punishments="state.punishments.value"
          :busy="state.busy.value"
          @save="value => state.execute(async () => { await abuseApi.savePunishment(value) })"
        />
        <AbuseRulesCard
          :rules="state.rules.value"
          :whitelist="state.whitelist.value"
          :busy="state.busy.value"
          @save-rule="value => state.execute(async () => { await abuseApi.saveRule(value) })"
          @delete-rule="(id, revision) => state.execute(async () => { await abuseApi.deleteRule(id, revision) })"
          @whitelist="(id, enabled) => state.execute(async () => { await abuseApi.setWhitelist(id, enabled) })"
        />
        <AbuseNodesCard :nodes="state.nodes.value" :statistics="state.statistics.value" :busy="state.busy.value" @copy="copy" @rotate="rotate" />
        <AbuseRecordsCard
          :records="state.records.value"
          :has-more="Boolean(state.nextRecordsCursor.value)"
          :loading-more="state.loadingMoreRecords.value"
          @load-more="state.loadMoreRecords"
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

.abuse-admin__notice { grid-column: 1 / -1; }

.admin-panel__heading {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 1rem;
}

.abuse-admin__grid {
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 22rem), 1fr));
  align-items: start;
  padding-bottom: max(1rem, env(safe-area-inset-bottom));
}

@media (max-width: 620px) {
  .admin-panel__heading {
    align-items: stretch;
    flex-direction: column;
  }
}
</style>
