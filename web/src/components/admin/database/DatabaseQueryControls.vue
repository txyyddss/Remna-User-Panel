<script setup lang="ts">
import { shallowRef } from 'vue'

import { useI18n } from '@/i18n'
import DatabaseFilterFields from './DatabaseFilterFields.vue'
import type { DatabaseColumnOption, DatabaseOperatorOption, TextDatabaseFilter } from './types'

const props = defineProps<{
  search: string
  filters: readonly TextDatabaseFilter[]
  columnItems: readonly DatabaseColumnOption[]
  operators: readonly DatabaseOperatorOption[]
}>()
const emit = defineEmits<{
  'update:search': [value: string]
  'update:filters': [value: TextDatabaseFilter[]]
}>()
const { t } = useI18n()
const filterDrawerOpen = shallowRef(false)
const mobileDraft = shallowRef<TextDatabaseFilter[]>([])
const drawerUi = {
  container: 'database-filter-drawer__container',
  header: 'database-filter-drawer__header',
  body: 'database-filter-drawer__body',
  footer: 'database-filter-drawer__footer',
} as const

function cloneFilters(filters: readonly TextDatabaseFilter[]): TextDatabaseFilter[] {
  return filters.map((filter) => ({ ...filter }))
}

function withAddedFilter(filters: readonly TextDatabaseFilter[]): TextDatabaseFilter[] {
  const column = props.columnItems[0]
  if (!column || filters.length >= 5) return cloneFilters(filters)
  return [...cloneFilters(filters), { column: column.value, operator: 'eq', value: '' }]
}

function addDesktopFilter(): void {
  emit('update:filters', withAddedFilter(props.filters))
}

function openMobileFilters(): void {
  mobileDraft.value = cloneFilters(props.filters)
  filterDrawerOpen.value = true
}

function applyMobileFilters(): void {
  emit('update:filters', cloneFilters(mobileDraft.value))
  filterDrawerOpen.value = false
}
</script>

<template>
  <section class="database-query" :aria-label="t('adminDatabase.search')">
    <div class="database-query__primary">
      <UInput
        :model-value="search"
        type="search"
        icon="i-ph-magnifying-glass"
        :maxlength="200"
        :placeholder="t('adminDatabase.searchPlaceholder')"
        :aria-label="t('adminDatabase.search')"
        @update:model-value="emit('update:search', String($event))"
      />
      <UButton
        class="database-query__mobile-filter"
        color="neutral"
        variant="outline"
        icon="i-ph-funnel"
        :label="t('adminDatabase.filterCount', { count: filters.length })"
        data-haptic="open"
        @click="openMobileFilters"
      />
      <UButton
        class="database-query__desktop-add"
        size="sm"
        color="neutral"
        variant="outline"
        icon="i-ph-plus"
        :disabled="filters.length >= 5 || !columnItems.length"
        :label="t('adminDatabase.addFilter')"
        @click="addDesktopFilter"
      />
    </div>
    <DatabaseFilterFields
      class="database-query__desktop-filters"
      :filters="filters"
      :column-items="columnItems"
      :operators="operators"
      @update:filters="emit('update:filters', $event)"
    />
    <UDrawer
      v-model:open="filterDrawerOpen"
      class="database-filter-drawer"
      :title="t('adminDatabase.filters')"
      :description="t('adminDatabase.filterCopy')"
      :ui="drawerUi"
    >
      <template #close>
        <UButton icon="i-ph-x" color="neutral" variant="ghost" :aria-label="t('common.cancel')" data-haptic="dismiss" />
      </template>
      <template #body>
        <DatabaseFilterFields
          :filters="mobileDraft"
          :column-items="columnItems"
          :operators="operators"
          @update:filters="mobileDraft = $event"
        />
        <UButton
          class="database-query__mobile-add"
          block
          color="neutral"
          variant="outline"
          icon="i-ph-plus"
          :disabled="mobileDraft.length >= 5 || !columnItems.length"
          :label="t('adminDatabase.addFilter')"
          @click="mobileDraft = withAddedFilter(mobileDraft)"
        />
      </template>
      <template #footer>
        <div class="database-query__drawer-actions">
          <UButton
            color="neutral"
            variant="ghost"
            :disabled="!mobileDraft.length"
            :label="t('adminDatabase.clearFilters')"
            data-haptic="dismiss"
            @click="mobileDraft = []"
          />
          <UButton color="neutral" variant="outline" :label="t('common.cancel')" data-haptic="dismiss" @click="filterDrawerOpen = false" />
          <UButton :label="t('adminDatabase.applyFilters')" data-haptic="confirm" @click="applyMobileFilters" />
        </div>
      </template>
    </UDrawer>
  </section>
</template>

<style scoped>
.database-query { display: grid; gap: 0.55rem; padding: 0.6rem; border-bottom: 1px solid var(--line); }
.database-query__primary { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 0.45rem; }
.database-query__primary > * { min-width: 0; }
.database-query__desktop-add,
.database-query__desktop-filters { display: none; }
.database-query__mobile-add { margin-top: 0.75rem; }
.database-query__drawer-actions { display: grid; gap: 0.55rem; }
.database-query__drawer-actions :deep(button) { width: 100%; }
:global(.database-filter-drawer__container) { height: 100%; gap: 0; overflow: hidden; }
:global(.database-filter-drawer__header) { flex: 0 0 auto; padding-bottom: 0.75rem; border-bottom: 1px solid var(--line); }
:global(.database-filter-drawer__body) { min-height: 0; overflow-y: auto; overscroll-behavior: contain; padding-block: 1rem; }
:global(.database-filter-drawer__footer) { flex: 0 0 auto; padding-top: 0.75rem; border-top: 1px solid var(--line); }

@media (max-width: 899px) {
  :global(.database-filter-drawer) { height: min(80dvh, calc(var(--tg-viewport-height, 100dvh) - var(--app-safe-top) - 0.5rem)); }
}

@media (min-width: 900px) {
  .database-query__mobile-filter { display: none; }
  .database-query__desktop-add { display: inline-flex; }
  .database-query__desktop-filters { display: grid; }
}
</style>
