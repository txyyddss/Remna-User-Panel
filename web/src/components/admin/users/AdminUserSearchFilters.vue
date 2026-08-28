<script setup lang="ts">
import { computed, reactive, shallowRef } from 'vue'

import type { Combo, SquadProduct } from '@/api/adminOperations'
import { useI18n } from '@/i18n'

export interface AdminUserSearchFiltersValue {
  state: '' | 'active' | 'non_active'
  comboIds: string[]
  squadUuids: string[]
  match: 'and' | 'or'
}

const props = defineProps<{ combos: readonly Combo[]; squads: readonly SquadProduct[] }>()
const emit = defineEmits<{ apply: [value: AdminUserSearchFiltersValue] }>()
const { t } = useI18n()
const drawerOpen = shallowRef(false)
const value = reactive<AdminUserSearchFiltersValue>({ state: '', comboIds: [], squadUuids: [], match: 'and' })
const stateItems = computed(() => [
  { label: t('adminUsers.allStates'), value: '' },
  { label: t('adminUsers.activeEntitlement'), value: 'active' },
  { label: t('adminUsers.nonActiveEntitlement'), value: 'non_active' },
])
const matchItems = computed(() => [
  { label: t('adminUsers.matchAnd'), value: 'and' },
  { label: t('adminUsers.matchOr'), value: 'or' },
])
const comboItems = computed(() => props.combos.map((item) => ({ label: item.name, value: item.id })))
const squadItems = computed(() => props.squads.map((item) => ({ label: item.name, value: item.remnaSquadUuid })))

function apply(): void {
  emit('apply', { state: value.state, comboIds: [...value.comboIds], squadUuids: [...value.squadUuids], match: value.match })
  drawerOpen.value = false
}

function clear(): void {
  value.state = ''
  value.comboIds = []
  value.squadUuids = []
  value.match = 'and'
  apply()
}
</script>

<template>
  <div class="admin-user-filters">
    <div class="admin-user-filters__desktop">
      <USelect v-model="value.state" :items="stateItems" value-key="value" label-key="label" :aria-label="t('adminUsers.entitlementState')" />
      <USelectMenu v-model="value.comboIds" :items="comboItems" value-key="value" label-key="label" multiple :placeholder="t('adminUsers.combos')" />
      <USelectMenu v-model="value.squadUuids" :items="squadItems" value-key="value" label-key="label" multiple :placeholder="t('adminUsers.squads')" />
      <URadioGroup v-model="value.match" :items="matchItems" value-key="value" />
      <UButton color="primary" icon="i-ph-funnel" :label="t('adminUsers.applyFilters')" data-haptic="confirm" @click="apply" />
      <UButton color="neutral" variant="ghost" :label="t('common.clear')" data-haptic="dismiss" @click="clear" />
    </div>
    <UButton class="admin-user-filters__mobile" color="neutral" variant="outline" icon="i-ph-funnel" :label="t('adminUsers.filters')" data-haptic="open" @click="drawerOpen = true" />
    <UDrawer v-model:open="drawerOpen" :title="t('adminUsers.filters')" :description="t('adminUsers.filtersHint')">
      <template #body>
        <div class="form-stack admin-user-filters__drawer">
          <UFormField name="state" :label="t('adminUsers.entitlementState')"><USelect v-model="value.state" :items="stateItems" value-key="value" label-key="label" /></UFormField>
          <UFormField name="combo" :label="t('adminUsers.combos')"><USelectMenu v-model="value.comboIds" :items="comboItems" value-key="value" label-key="label" multiple /></UFormField>
          <UFormField name="squad" :label="t('adminUsers.squads')"><USelectMenu v-model="value.squadUuids" :items="squadItems" value-key="value" label-key="label" multiple /></UFormField>
          <UFormField name="match" :label="t('adminUsers.match')"><URadioGroup v-model="value.match" :items="matchItems" value-key="value" /></UFormField>
        </div>
      </template>
      <template #footer><UButton color="neutral" variant="outline" :label="t('common.clear')" @click="clear" /><UButton color="primary" :label="t('adminUsers.applyFilters')" data-haptic="confirm" @click="apply" /></template>
    </UDrawer>
  </div>
</template>

<style scoped>
.admin-user-filters__desktop { display: flex; flex-wrap: wrap; gap: 0.5rem; align-items: center; }
.admin-user-filters__desktop > * { min-height: 44px; min-width: 8rem; }
.admin-user-filters__mobile { display: none; min-height: 44px; }
.admin-user-filters__drawer :deep(.u-select-menu) { width: 100%; }
@media (max-width: 639px) { .admin-user-filters__desktop { display: none; } .admin-user-filters__mobile { display: inline-flex; } }
</style>
