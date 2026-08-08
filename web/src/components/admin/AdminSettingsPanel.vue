<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { PhCheck, PhFloppyDisk, PhKey } from '@phosphor-icons/vue'

import type { AdminSetting } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import AdminSectionState from './AdminSectionState.vue'

const { items, loading, busy, error, load, perform } = useAdminSection<AdminSetting>('settings')
const draft = reactive<Record<string, string>>({})
const saved = reactive({ visible: false })

const grouped = computed(() => items.value.reduce<Record<string, AdminSetting[]>>((groups, item) => {
  const category = item.category
  groups[category] ??= []
  groups[category].push(item)
  return groups
}, {}))

watch(items, (next) => {
  for (const item of next) draft[item.key] = item.encrypted ? '' : item.value
}, { immediate: true })

function settingLabel(setting: AdminSetting): string {
  return setting.key.replace(/[._]/g, ' ').replace(/\b\w/g, (character) => character.toUpperCase())
}

function isSensitive(setting: AdminSetting): boolean {
  return setting.encrypted
}

function isBoolean(setting: AdminSetting): boolean {
  return !setting.encrypted && (
    setting.value === 'true'
    || setting.value === 'false'
    || /(^|[._])(enabled|active|visible|allow|require|disabled)$/.test(setting.key)
  )
}

function setBoolean(key: string, value: boolean): void {
  draft[key] = value ? 'true' : 'false'
}

async function save(): Promise<void> {
  const values = Object.entries(draft).filter(([, value]) => value !== '')
  const { api } = await import('@/api/client')
  saved.visible = await perform(() => Promise.all(values.map(([key, value]) => api.updateAdminSetting(key, value))))
}
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>System settings</h2><p>Secrets are write-only and stay masked after save.</p></div>
      <button class="button button--primary" type="button" :disabled="busy" @click="save">
        <PhFloppyDisk :size="18" /> {{ busy ? 'Saving' : 'Save settings' }}
      </button>
    </div>
    <InlineNotice v-if="saved.visible" tone="success" title="Settings saved">New values will be used by the next integration request.</InlineNotice>
    <AdminSectionState :loading="loading" :error="error" @retry="load()">
      <form class="settings-groups" @submit.prevent="save">
        <fieldset v-for="(settings, category) in grouped" :key="category" class="settings-group">
          <legend>{{ category }}</legend>
          <div v-for="setting in settings" :key="setting.key" class="settings-field">
            <span class="settings-field__label">
              <span>{{ settingLabel(setting) }}</span>
              <span v-if="setting.configured" class="configured-label"><PhCheck :size="14" /> Configured</span>
            </span>
            <SwitchField
              v-if="isBoolean(setting)"
              :id="`setting-${setting.key}`"
              :model-value="draft[setting.key] === 'true'"
              label="Enabled"
              help="Applied after validation."
              @update:model-value="setBoolean(setting.key, $event)"
            />
            <span v-if="!isBoolean(setting)" class="input-shell">
              <PhKey v-if="isSensitive(setting)" :size="18" />
              <input
                v-model="draft[setting.key]"
                :type="isSensitive(setting) ? 'password' : 'text'"
                :placeholder="isSensitive(setting) && setting.configured ? 'Leave blank to keep current value' : ''"
                :autocomplete="isSensitive(setting) ? 'new-password' : 'off'"
              />
            </span>
            <small v-if="!isBoolean(setting)">{{ isSensitive(setting) ? 'Stored encrypted and never returned in plaintext.' : 'Applied after validation.' }}</small>
          </div>
        </fieldset>
      </form>
    </AdminSectionState>
  </section>
</template>
