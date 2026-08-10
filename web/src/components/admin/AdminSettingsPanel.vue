<script setup lang="ts">
import { computed, reactive, watch } from 'vue'

import type { AdminSetting } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { useI18n } from '@/i18n'
import AdminSectionState from './AdminSectionState.vue'

const { items, loading, busy, error, load, perform } = useAdminSection<AdminSetting>('settings')
const draft = reactive<Record<string, string>>({})
const saved = reactive({ visible: false })
const { t } = useI18n()

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
  const key = `adminSettings.settingLabels.${setting.key.replaceAll('.', '_')}`
  const translated = t(key)
  if (translated !== key) return translated
  return setting.key
}

function categoryLabel(category: string): string {
  const key = `adminSettings.categories.${category.toLowerCase()}`
  const translated = t(key)
  return translated === key ? category : translated
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
      <div><h2>{{ t('adminSettings.title') }}</h2><p>{{ t('adminSettings.copy') }}</p></div>
      <UButton icon="i-ph-floppy-disk" :disabled="busy" :loading="busy" :label="busy ? t('common.saving') : t('adminSettings.save')" @click="save" />
    </div>
    <InlineNotice v-if="saved.visible" tone="success" :title="t('adminSettings.saved')">{{ t('adminSettings.savedHint') }}</InlineNotice>
    <AdminSectionState :loading="loading" :error="error" @retry="load()">
      <form class="settings-groups" @submit.prevent="save">
        <fieldset v-for="(settings, category) in grouped" :key="category" class="settings-group">
          <legend>{{ categoryLabel(String(category)) }}</legend>
          <div v-for="setting in settings" :key="setting.key" class="settings-field">
            <span class="settings-field__label">
              <span>{{ settingLabel(setting) }}</span>
              <span v-if="setting.configured" class="configured-label"><UIcon name="i-ph-check" /> {{ t('common.configured') }}</span>
            </span>
            <SwitchField
              v-if="isBoolean(setting)"
              :id="`setting-${setting.key}`"
              :model-value="draft[setting.key] === 'true'"
              :label="t('common.enabled')"
              :help="t('adminSettings.validated')"
              @update:model-value="setBoolean(setting.key, $event)"
            />
            <UInput
              v-if="!isBoolean(setting)"
              v-model="draft[setting.key]"
              :icon="isSensitive(setting) ? 'i-ph-key' : undefined"
              :type="isSensitive(setting) ? 'password' : 'text'"
              :placeholder="isSensitive(setting) && setting.configured ? t('adminSettings.keepSecret') : ''"
              :autocomplete="isSensitive(setting) ? 'new-password' : 'off'"
            />
            <small v-if="!isBoolean(setting)">{{ isSensitive(setting) ? t('adminSettings.secretHint') : t('adminSettings.validated') }}</small>
          </div>
        </fieldset>
      </form>
    </AdminSectionState>
  </section>
</template>
