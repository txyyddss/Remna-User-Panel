<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import InlineNotice from '@/components/common/InlineNotice.vue'
import OperationStatusNotice from '@/components/common/OperationStatusNotice.vue'
import SettingsSwitchRow from './SettingsSwitchRow.vue'
import EarlyActivationDialog from './EarlyActivationDialog.vue'
import { useComboControls } from './useComboControls'
import { useI18n } from '@/i18n'

const controls = useComboControls()
const { t } = useI18n()
const open = shallowRef(false)
const message = computed(() => {
  const status = controls.command.receipt.value?.status
  return t(status === 'succeeded' ? 'settings.combo.updated' : status === 'failed' || status === 'pending_review' || status === 'partial' ? 'settings.combo.syncFailed' : 'settings.combo.syncPending')
})
async function activate(confirmation: string): Promise<void> { if (await controls.activate(confirmation)) open.value = false }
</script>

<template>
  <div class="combo-controls">
    <USkeleton v-if="controls.loading.value" class="h-24" />
    <template v-else-if="controls.value.value?.activePurchase">
      <h3>{{ $t('settings.combo.squads') }}</h3>
      <SettingsSwitchRow v-for="squad in controls.value.value.squads" :key="squad.uuid" :label="squad.name" :enabled="squad.enabled" :description="squad.enabled && !squad.canDisable && !controls.blocked.value ? $t('settings.combo.keepOne') : undefined" :disabled="controls.blocked.value || (squad.enabled && !squad.canDisable)" @update="controls.switchSquad(squad.uuid, $event)" />
      <div v-if="controls.quote.value?.eligible" class="combo-controls__action">
        <div><strong>{{ $t('settings.combo.activateTitle') }}</strong><p>{{ $t('settings.combo.activateHint') }}</p></div>
        <UButton color="neutral" variant="ghost" trailing-icon="i-ph-arrow-right" :label="$t('settings.combo.activate')" :disabled="controls.blocked.value" @click="open = true" />
      </div>
    </template>
    <OperationStatusNotice v-if="controls.command.receipt.value" :receipt="controls.command.receipt.value" :error="controls.error.value" :checking="controls.command.checking.value" :message="message" @refresh="controls.command.refresh" />
    <InlineNotice v-else-if="controls.error.value" tone="warning">{{ controls.error.value }}<UButton variant="link" :label="$t('common.tryAgain')" @click="controls.load" /></InlineNotice>
    <EarlyActivationDialog v-model:open="open" :quote="controls.quote.value" :busy="controls.command.busy.value" @confirm="activate" />
  </div>
</template>

<style scoped>
.combo-controls { min-width: 0; padding-top: 0.75rem; }
.combo-controls h3 { margin: 0; padding-top: 0.75rem; color: var(--text-muted); font-size: 0.875rem; font-weight: 600; }
.combo-controls__action { display: flex; align-items: center; justify-content: space-between; gap: 1rem; min-height: 60px; padding-block: 0.75rem; }
.combo-controls__action > div { min-width: 0; flex: 1; }
.combo-controls__action strong { font-size: 0.875rem; }
.combo-controls__action p { margin: 0.25rem 0 0; color: var(--text-muted); font-size: 0.8rem; line-height: 1.45; }
.combo-controls__action :deep(button) { flex-shrink: 0; }
</style>
