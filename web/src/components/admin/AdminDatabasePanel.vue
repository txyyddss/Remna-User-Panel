<script setup lang="ts">
import { computed, onBeforeUnmount, shallowRef, watch } from 'vue'

import type { DatabaseMutationInput, DatabaseQueryInput, DatabaseRow, DatabaseValue } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useAdminDatabase } from '@/composables/useAdminDatabase'
import { useI18n } from '@/i18n'
import AdminSectionState from './AdminSectionState.vue'
import DatabaseMobileRowCard from './database/DatabaseMobileRowCard.vue'
import DatabaseQueryControls from './database/DatabaseQueryControls.vue'
import DatabaseRecordEditor from './database/DatabaseRecordEditor.vue'
import DatabaseTablePicker from './database/DatabaseTablePicker.vue'
import type { DatabaseColumnOption, DatabaseOperatorOption, TextDatabaseFilter } from './database/types'

const {
  tables, selectedTableName, selectedTable, rows, nextCursor, loading, busy, error,
  review, lastRescueBackupId, loadTables, queryRows, selectTable,
  reviewMutation, applyMutation, clearReview,
} = useAdminDatabase()
type EditorState = { action: 'insert' | 'update' | 'delete'; row?: DatabaseRow }
type EditorPayload = { action: EditorState['action']; row?: DatabaseRow; values: Record<string, DatabaseValue>; reason: string }
type DisplayRow = Record<string, unknown> & { __row: DatabaseRow }

const editing = shallowRef<EditorState | null>(null)
const search = shallowRef('')
const filters = shallowRef<TextDatabaseFilter[]>([])
const { t } = useI18n()
const operators = computed<DatabaseOperatorOption[]>(() => [
  { value: 'eq', label: t('adminDatabase.operators.eq') }, { value: 'ne', label: t('adminDatabase.operators.ne') },
  { value: 'contains', label: t('adminDatabase.operators.contains') }, { value: 'starts_with', label: t('adminDatabase.operators.startsWith') },
  { value: 'gt', label: t('adminDatabase.operators.gt') }, { value: 'gte', label: t('adminDatabase.operators.gte') },
  { value: 'lt', label: t('adminDatabase.operators.lt') }, { value: 'lte', label: t('adminDatabase.operators.lte') },
  { value: 'is_null', label: t('adminDatabase.operators.isNull') }, { value: 'not_null', label: t('adminDatabase.operators.notNull') },
])
const columnItems = computed<DatabaseColumnOption[]>(() => (selectedTable.value?.columns ?? [])
  .filter((item) => !item.sensitive && item.declaredType.toUpperCase() !== 'BLOB')
  .map((column) => ({ value: column.name, label: column.name })))
const tableData = computed<DisplayRow[]>(() => rows.value.map((row) => ({
  __row: row,
  ...Object.fromEntries((selectedTable.value?.columns ?? []).map((column) => [
    column.name,
    column.sensitive ? t('adminDatabase.maskedValue') : displayValue(row.values[column.name]),
  ])),
})))
const tableColumns = computed(() => [
  ...(selectedTable.value?.columns ?? []).map((column) => ({ id: column.name, header: column.name, accessorFn: (row: DisplayRow) => row[column.name] })),
  { id: 'actions', header: t('adminDatabase.actions') },
])
let debounceTimer: ReturnType<typeof globalThis.setTimeout> | undefined

function queryInput(): DatabaseQueryInput {
  return { search: search.value.trim() || undefined, filters: filters.value.map((filter) => ({ ...filter })), limit: 50 }
}

function scheduleQuery(): void {
  if (debounceTimer) globalThis.clearTimeout(debounceTimer)
  debounceTimer = globalThis.setTimeout(() => { if (selectedTableName.value) void queryRows(selectedTableName.value, queryInput()) }, 280)
}

function updateSearch(value: string): void { search.value = value }
function updateFilters(value: TextDatabaseFilter[]): void { filters.value = value }

watch(search, scheduleQuery)
watch(filters, scheduleQuery, { deep: true })
onBeforeUnmount(() => { if (debounceTimer) globalThis.clearTimeout(debounceTimer) })

function displayValue(value: DatabaseValue): string {
  if (value === null) return t('adminDatabase.nullValue')
  if (typeof value === 'object') return t('adminDatabase.blobValue')
  return String(value)
}

function mutationInput(payload: EditorPayload): DatabaseMutationInput {
  return { action: payload.action, table: selectedTable.value!.name, key: payload.row?.key, values: payload.action === 'delete' ? undefined : payload.values, recordHash: payload.row?.recordHash, reason: payload.reason }
}

async function requestReview(payload: EditorPayload): Promise<void> { await reviewMutation(mutationInput(payload)) }

async function apply(payload: EditorPayload & { confirmation: string }): Promise<void> {
  if (await applyMutation(mutationInput(payload), payload.confirmation)) editing.value = null
}

function closeEditor(): void { editing.value = null; clearReview() }
function openEditor(action: EditorState['action'], row?: DatabaseRow): void { clearReview(); editing.value = { action, row } }
async function chooseTable(name: string): Promise<void> { closeEditor(); search.value = ''; filters.value = []; await selectTable(name) }
</script>

<template>
  <section class="admin-panel database-panel">
    <div class="admin-panel__heading">
      <div><h2>{{ t('adminDatabase.title') }}</h2><p>{{ t('adminDatabase.copy') }}</p></div>
      <UButton icon="i-ph-plus" :disabled="!selectedTable" :label="t('adminDatabase.insertRow')" @click="openEditor('insert')" />
    </div>
    <InlineNotice tone="warning" :title="t('adminDatabase.bypassTitle')">{{ t('adminDatabase.bypassCopy') }}</InlineNotice>
    <InlineNotice v-if="lastRescueBackupId" tone="success" :title="t('adminDatabase.appliedTitle')">{{ t('adminDatabase.appliedCopy', { id: lastRescueBackupId }) }}</InlineNotice>
    <AdminSectionState :loading="loading" :error="error" @retry="loadTables">
      <div class="database-layout">
        <DatabaseTablePicker :tables="tables" :selected="selectedTableName" :busy="busy" @select="chooseTable" />
        <div class="database-rows" :aria-label="t('adminDatabase.rows')">
          <DatabaseQueryControls
            :search="search"
            :filters="filters"
            :column-items="columnItems"
            :operators="operators"
            @update:search="updateSearch"
            @update:filters="updateFilters"
          />
          <div v-if="selectedTable && rows.length" class="database-results">
            <div class="database-table-view">
              <div class="database-table-scroll">
                <UTable :data="tableData" :columns="tableColumns">
                  <template #actions-cell="{ row }"><div class="row-actions"><UButton color="neutral" variant="ghost" icon="i-ph-pencil-simple" :aria-label="t('adminDatabase.editRow')" @click="openEditor('update', row.original.__row)" /><UButton color="error" variant="ghost" icon="i-ph-trash" :aria-label="t('adminDatabase.deleteRow')" data-haptic="destructive" @click="openEditor('delete', row.original.__row)" /></div></template>
                </UTable>
              </div>
            </div>
            <div class="database-mobile-list">
              <DatabaseMobileRowCard v-for="row in rows" :key="row.recordHash" :row="row" :columns="selectedTable.columns" @edit="openEditor('update', $event)" @delete="openEditor('delete', $event)" />
            </div>
          </div>
          <div v-else class="empty-inline"><div><h3>{{ t('adminDatabase.noRows') }}</h3><p>{{ t('adminDatabase.empty') }}</p></div></div>
          <UButton v-if="nextCursor && selectedTableName" class="database-load-more" color="neutral" variant="outline" :disabled="busy" :loading="busy" :label="busy ? t('common.loading') : t('adminDatabase.loadMore')" @click="queryRows(selectedTableName, queryInput(), { append: true })" />
        </div>
      </div>
    </AdminSectionState>
    <DatabaseRecordEditor v-if="editing && selectedTable" :table="selectedTable" :action="editing.action" :row="editing.row" :review="review" :busy="busy" @cancel="closeEditor" @invalidate="clearReview" @review="requestReview" @apply="apply" />
  </section>
</template>

<style scoped>
.database-panel,
.database-layout,
.database-rows,
.database-results {
  min-width: 0;
}

.database-layout {
  display: grid;
  gap: 0.8rem;
  padding: 1rem;
}

.database-rows { overflow: hidden; border: 1px solid var(--line); border-radius: var(--radius-control); }
.database-results { display: block; }
.database-table-view { display: none; min-width: 0; }
.database-table-scroll { min-width: 0; overflow-x: auto; overscroll-behavior-x: contain; }
.database-table-view :deep(table) { min-width: max-content; }
.database-mobile-list { display: grid; gap: 0.55rem; padding: 0.65rem; }
.database-load-more { margin: 0.65rem; }

@media (min-width: 900px) {
  .database-layout { grid-template-columns: 210px minmax(0, 1fr); }
  .database-table-view { display: block; }
  .database-mobile-list { display: none; }
}

@media (max-width: 639px) {
  .database-layout { padding: 0.75rem; }
}
</style>
