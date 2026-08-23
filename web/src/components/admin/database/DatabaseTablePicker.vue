<script setup lang="ts">
import { computed, shallowRef } from 'vue'

import type { DatabaseTable } from '@/api/features'
import { useI18n } from '@/i18n'
import { selectionHaptic } from '@/utils/telegramHaptics'
import type { DeepReadonly } from './types'

const props = defineProps<{
  tables: DeepReadonly<DatabaseTable[]>
  selected: string | null
  busy: boolean
}>()
const emit = defineEmits<{ select: [name: string] }>()
const { t } = useI18n()
const search = shallowRef('')

const items = computed(() => props.tables.map((table) => ({
  value: table.name,
  label: table.name,
  description: table.highRisk ? t('adminDatabase.highRisk') : undefined,
  icon: 'i-ph-database',
})))
const visibleTables = computed(() => {
  const needle = search.value.trim().toLowerCase()
  return needle ? props.tables.filter((table) => table.name.toLowerCase().includes(needle)) : props.tables
})

function choose(value: unknown): void {
  const name = String(value)
  if (!name || name === props.selected || props.busy) return
  selectionHaptic()
  emit('select', name)
}
</script>

<template>
  <nav class="database-table-picker" :aria-label="t('adminDatabase.tables')">
    <div class="database-table-picker__mobile">
      <UFormField :label="t('adminDatabase.tables')">
        <USelectMenu
          class="database-table-picker__select"
          :model-value="selected ?? undefined"
          :items="items"
          value-key="value"
          label-key="label"
          :search-input="{ placeholder: t('adminDatabase.tableSearch') }"
          :placeholder="t('adminDatabase.selectTable')"
          :disabled="busy"
          :aria-label="t('adminDatabase.selectTable')"
          @update:model-value="choose"
        />
      </UFormField>
    </div>
    <div class="database-table-picker__desktop">
      <UInput
        v-model="search"
        icon="i-ph-magnifying-glass"
        :placeholder="t('adminDatabase.tableSearch')"
        :aria-label="t('adminDatabase.tableSearch')"
      />
      <UButton
        v-for="table in visibleTables"
        :key="table.name"
        class="database-table-picker__button"
        :class="{ 'database-table-picker__button--active': selected === table.name }"
        color="neutral"
        variant="ghost"
        icon="i-ph-database"
        :disabled="busy"
        :aria-pressed="selected === table.name"
        @click="choose(table.name)"
      >
        <span>{{ table.name }}</span>
        <small v-if="table.highRisk">{{ t('adminDatabase.highRisk') }}</small>
      </UButton>
    </div>
  </nav>
</template>

<style scoped>
.database-table-picker,
.database-table-picker__select { min-width: 0; width: 100%; }
.database-table-picker__mobile { display: block; }
.database-table-picker__desktop { display: none; }

@media (min-width: 900px) {
  .database-table-picker__mobile { display: none; }
  .database-table-picker__desktop { display: flex; min-width: 0; flex-direction: column; gap: 0.4rem; }
  .database-table-picker__button { min-height: 44px; justify-content: flex-start; border: 1px solid var(--line); border-radius: var(--radius-control); color: var(--text-muted); }
  .database-table-picker__button small { margin-left: auto; color: var(--warning); font-size: 0.58rem; }
  .database-table-picker__button--active { border-color: var(--accent); color: var(--accent); background: var(--accent-soft); }
}
</style>
