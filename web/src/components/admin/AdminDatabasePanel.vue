<script setup lang="ts">
import { computed, onBeforeUnmount, shallowRef, watch } from 'vue'

import type { DatabaseFilter, DatabaseFilterOperator, DatabaseMutationInput, DatabaseQueryInput, DatabaseRow, DatabaseValue } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useAdminDatabase } from '@/composables/useAdminDatabase'
import { useI18n } from '@/i18n'
import AdminSectionState from './AdminSectionState.vue'
import DatabaseRecordEditor from './database/DatabaseRecordEditor.vue'

const {
  tables, selectedTableName, selectedTable, rows, nextCursor, loading, busy, error,
  review, lastRescueBackupId, loadTables, queryRows, selectTable,
  reviewMutation, applyMutation, clearReview,
} = useAdminDatabase()
type EditorState = { action: 'insert' | 'update' | 'delete'; row?: DatabaseRow }
type EditorPayload = { action: EditorState['action']; row?: DatabaseRow; values: Record<string, DatabaseValue>; reason: string }
type DisplayRow = Record<string, unknown> & { __row: DatabaseRow }
type TextDatabaseFilter = Omit<DatabaseFilter, 'value'> & { value?: string }

const editing = shallowRef<EditorState | null>(null)
const search = shallowRef('')
const filters = shallowRef<TextDatabaseFilter[]>([])
const { t } = useI18n()
const operators = computed<Array<{ value: DatabaseFilterOperator; label: string }>>(() => [
  { value: 'eq', label: t('adminDatabase.operators.eq') }, { value: 'ne', label: t('adminDatabase.operators.ne') },
  { value: 'contains', label: t('adminDatabase.operators.contains') }, { value: 'starts_with', label: t('adminDatabase.operators.startsWith') },
  { value: 'gt', label: t('adminDatabase.operators.gt') }, { value: 'gte', label: t('adminDatabase.operators.gte') },
  { value: 'lt', label: t('adminDatabase.operators.lt') }, { value: 'lte', label: t('adminDatabase.operators.lte') },
  { value: 'is_null', label: t('adminDatabase.operators.isNull') }, { value: 'not_null', label: t('adminDatabase.operators.notNull') },
])
const columnItems = computed(() => (selectedTable.value?.columns ?? [])
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
function addFilter(): void {
  const column = columnItems.value[0]
  if (!column || filters.value.length >= 5) return
  filters.value = [...filters.value, { column: column.value, operator: 'eq', value: '' }]
}
function removeFilter(index: number): void { filters.value = filters.value.filter((_, itemIndex) => itemIndex !== index) }

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
    <div class="admin-panel__heading"><div><h2>{{ t('adminDatabase.title') }}</h2><p>{{ t('adminDatabase.copy') }}</p></div><UButton icon="i-ph-plus" :disabled="!selectedTable" :label="t('adminDatabase.insertRow')" @click="openEditor('insert')" /></div>
    <InlineNotice tone="warning" :title="t('adminDatabase.bypassTitle')">{{ t('adminDatabase.bypassCopy') }}</InlineNotice>
    <InlineNotice v-if="lastRescueBackupId" tone="success" :title="t('adminDatabase.appliedTitle')">{{ t('adminDatabase.appliedCopy', { id: lastRescueBackupId }) }}</InlineNotice>
    <AdminSectionState :loading="loading" :error="error" @retry="loadTables">
      <div class="database-layout">
        <nav v-auto-animate class="database-tables" :aria-label="t('adminDatabase.tables')">
          <UButton v-for="table in tables" :key="table.name" class="database-table-button" :class="{ 'database-table-button--active': selectedTableName === table.name }" color="neutral" variant="ghost" icon="i-ph-database" @click="chooseTable(table.name)"><span>{{ table.name }}</span><small>{{ t('adminDatabase.highRisk') }}</small></UButton>
        </nav>
        <div class="database-rows" :aria-label="t('adminDatabase.rows')">
          <div class="database-query">
            <UInput v-model="search" type="search" icon="i-ph-magnifying-glass" :maxlength="200" :placeholder="t('adminDatabase.searchPlaceholder')" :aria-label="t('adminDatabase.search')" />
            <UButton size="sm" color="neutral" variant="outline" icon="i-ph-plus" :disabled="filters.length >= 5" :label="t('adminDatabase.addFilter')" @click="addFilter" />
            <div v-for="(filter, index) in filters" :key="index" class="database-filter">
              <USelect v-model="filter.column" :items="columnItems" value-key="value" :aria-label="t('adminDatabase.filterColumn')" />
              <USelect v-model="filter.operator" :items="operators" value-key="value" :aria-label="t('adminDatabase.filterOperator')" />
              <UInput v-if="!['is_null', 'not_null'].includes(filter.operator)" v-model="filter.value" :aria-label="t('adminDatabase.filterValue')" :placeholder="t('adminDatabase.value')" />
              <UButton color="neutral" variant="ghost" icon="i-ph-x" :aria-label="t('adminDatabase.removeFilter')" @click="removeFilter(index)" />
            </div>
          </div>
          <UTable v-if="selectedTable && rows.length" :data="tableData" :columns="tableColumns">
            <template #actions-cell="{ row }"><div class="row-actions"><UButton color="neutral" variant="ghost" icon="i-ph-pencil-simple" :aria-label="t('adminDatabase.editRow')" @click="openEditor('update', row.original.__row)" /><UButton color="error" variant="ghost" icon="i-ph-trash" :aria-label="t('adminDatabase.deleteRow')" @click="openEditor('delete', row.original.__row)" /></div></template>
          </UTable>
          <div v-else class="empty-inline"><div><h3>{{ t('adminDatabase.noRows') }}</h3><p>{{ t('adminDatabase.empty') }}</p></div></div>
          <UButton v-if="nextCursor && selectedTableName" class="m-3" color="neutral" variant="outline" :disabled="busy" :loading="busy" :label="busy ? t('common.loading') : t('adminDatabase.loadMore')" @click="queryRows(selectedTableName, queryInput(), { append: true })" />
        </div>
      </div>
    </AdminSectionState>
    <DatabaseRecordEditor v-if="editing && selectedTable" :table="selectedTable" :action="editing.action" :row="editing.row" :review="review" :busy="busy" @cancel="closeEditor" @invalidate="clearReview" @review="requestReview" @apply="apply" />
  </section>
</template>

<style scoped>
.database-layout { display: grid; gap: 0.8rem; padding: 1rem; }
.database-tables { display: flex; gap: 0.4rem; overflow-x: auto; padding-bottom: 0.35rem; }
.database-table-button { min-height: 44px; flex: 0 0 auto; border: 1px solid var(--line); border-radius: var(--radius-control); color: var(--text-muted); }
.database-table-button small { color: var(--warning); font-size: 0.58rem; }
.database-table-button--active { color: var(--accent); border-color: var(--accent); background: var(--accent-soft); }
.database-rows { overflow: auto; border: 1px solid var(--line); border-radius: var(--radius-control); }
.database-query { display: flex; flex-wrap: wrap; gap: 0.45rem; padding: 0.6rem; border-bottom: 1px solid var(--line); }
.database-query > :first-child { min-width: min(260px, 100%); flex: 1; }
.database-filter { width: 100%; display: grid; grid-template-columns: minmax(120px, 1fr) minmax(110px, 1fr) minmax(120px, 1fr) auto; gap: 0.35rem; }
@media (min-width: 900px) { .database-layout { grid-template-columns: 210px minmax(0, 1fr); } .database-tables { flex-direction: column; overflow: visible; } }
</style>
