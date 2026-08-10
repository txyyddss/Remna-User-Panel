<script setup lang="ts">
import { computed, onBeforeUnmount, shallowRef, watch } from 'vue'
import { PhDatabase, PhMagnifyingGlass, PhPencilSimple, PhPlus, PhTrash, PhX } from '@phosphor-icons/vue'

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

const editing = shallowRef<EditorState | null>(null)
const search = shallowRef('')
const filters = shallowRef<DatabaseFilter[]>([])
const { t } = useI18n()
const operators = computed<Array<{ value: DatabaseFilterOperator; label: string }>>(() => [
  { value: 'eq', label: t('adminDatabase.operators.eq') }, { value: 'ne', label: t('adminDatabase.operators.ne') }, { value: 'contains', label: t('adminDatabase.operators.contains') },
  { value: 'starts_with', label: t('adminDatabase.operators.startsWith') }, { value: 'gt', label: t('adminDatabase.operators.gt') }, { value: 'gte', label: t('adminDatabase.operators.gte') },
  { value: 'lt', label: t('adminDatabase.operators.lt') }, { value: 'lte', label: t('adminDatabase.operators.lte') }, { value: 'is_null', label: t('adminDatabase.operators.isNull') }, { value: 'not_null', label: t('adminDatabase.operators.notNull') },
])
let debounceTimer: ReturnType<typeof setTimeout> | undefined

function queryInput(): DatabaseQueryInput {
  return { search: search.value.trim() || undefined, filters: filters.value.map((filter) => ({ ...filter })), limit: 50 }
}

function scheduleQuery(): void {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    if (selectedTableName.value) void queryRows(selectedTableName.value, queryInput())
  }, 280)
}

function addFilter(): void {
  const column = selectedTable.value?.columns.find((item) => !item.sensitive && item.declaredType.toUpperCase() !== 'BLOB')
  if (!column || filters.value.length >= 5) return
  filters.value = [...filters.value, { column: column.name, operator: 'eq', value: '' }]
}

function removeFilter(index: number): void {
  filters.value = filters.value.filter((_, itemIndex) => itemIndex !== index)
}

watch(search, scheduleQuery)
watch(filters, scheduleQuery, { deep: true })
onBeforeUnmount(() => { if (debounceTimer) clearTimeout(debounceTimer) })

function displayValue(value: DatabaseValue): string {
  if (value === null) return t('adminDatabase.nullValue')
  if (typeof value === 'object') return t('adminDatabase.blobValue')
  return String(value)
}

function mutationInput(payload: EditorPayload): DatabaseMutationInput {
  return {
    action: payload.action,
    table: selectedTable.value!.name,
    key: payload.row?.key,
    values: payload.action === 'delete' ? undefined : payload.values,
    recordHash: payload.row?.recordHash,
    reason: payload.reason,
  }
}

async function requestReview(payload: EditorPayload): Promise<void> {
  await reviewMutation(mutationInput(payload))
}

async function apply(payload: EditorPayload & { confirmation: string }): Promise<void> {
  if (await applyMutation(mutationInput(payload), payload.confirmation)) editing.value = null
}

function closeEditor(): void {
  editing.value = null
  clearReview()
}

function openEditor(action: EditorState['action'], row?: DatabaseRow): void {
  clearReview()
  editing.value = { action, row }
}

async function chooseTable(name: string): Promise<void> {
  closeEditor()
  search.value = ''
  filters.value = []
  await selectTable(name)
}
</script>

<template>
  <section class="admin-panel database-panel">
    <div class="admin-panel__heading"><div><h2>{{ t('adminDatabase.title') }}</h2><p>{{ t('adminDatabase.copy') }}</p></div><button class="button button--primary" type="button" :disabled="!selectedTable" @click="openEditor('insert')"><PhPlus :size="18" />{{ t('adminDatabase.insertRow') }}</button></div>
    <InlineNotice tone="warning" :title="t('adminDatabase.bypassTitle')">{{ t('adminDatabase.bypassCopy') }}</InlineNotice>
    <InlineNotice v-if="lastRescueBackupId" tone="success" :title="t('adminDatabase.appliedTitle')">{{ t('adminDatabase.appliedCopy', { id: lastRescueBackupId }) }}</InlineNotice>
    <AdminSectionState :loading="loading" :error="error" @retry="loadTables">
      <div class="database-layout">
        <nav class="database-tables" :aria-label="t('adminDatabase.tables')">
          <button v-for="table in tables" :key="table.name" class="database-table-button" :class="{ 'database-table-button--active': selectedTableName === table.name }" type="button" @click="chooseTable(table.name)"><PhDatabase :size="17" /><span>{{ table.name }}</span><small>{{ t('adminDatabase.highRisk') }}</small></button>
        </nav>
        <div class="database-rows" tabindex="0" :aria-label="t('adminDatabase.rows')">
          <div class="database-query">
            <label class="database-search"><PhMagnifyingGlass :size="17" /><input v-model="search" type="search" maxlength="200" :placeholder="t('adminDatabase.searchPlaceholder')" :aria-label="t('adminDatabase.search')" /></label>
            <button class="button button--secondary button--small" type="button" :disabled="filters.length >= 5" @click="addFilter"><PhPlus :size="16" />{{ t('adminDatabase.addFilter') }}</button>
            <div v-for="(filter, index) in filters" :key="index" class="database-filter">
              <select v-model="filter.column" :aria-label="t('adminDatabase.filterColumn')"><option v-for="column in selectedTable?.columns.filter((item) => !item.sensitive && item.declaredType.toUpperCase() !== 'BLOB')" :key="column.name" :value="column.name">{{ column.name }}</option></select>
              <select v-model="filter.operator" :aria-label="t('adminDatabase.filterOperator')"><option v-for="operator in operators" :key="operator.value" :value="operator.value">{{ operator.label }}</option></select>
              <input v-if="!['is_null', 'not_null'].includes(filter.operator)" v-model="filter.value" :aria-label="t('adminDatabase.filterValue')" :placeholder="t('adminDatabase.value')" />
              <button class="icon-button" type="button" :aria-label="t('adminDatabase.removeFilter')" @click="removeFilter(index)"><PhX :size="16" /></button>
            </div>
          </div>
          <table v-if="selectedTable && rows.length">
            <thead><tr><th v-for="column in selectedTable.columns" :key="column.name" scope="col">{{ column.name }}</th><th scope="col">{{ t('adminDatabase.actions') }}</th></tr></thead>
            <tbody><tr v-for="row in rows" :key="row.recordHash"><td v-for="column in selectedTable.columns" :key="column.name"><span>{{ column.sensitive ? t('adminDatabase.maskedValue') : displayValue(row.values[column.name]) }}</span></td><td><div class="row-actions"><button class="icon-button" type="button" :aria-label="t('adminDatabase.editRow')" @click="openEditor('update', row)"><PhPencilSimple :size="18" /></button><button class="icon-button icon-button--danger" type="button" :aria-label="t('adminDatabase.deleteRow')" @click="openEditor('delete', row)"><PhTrash :size="18" /></button></div></td></tr></tbody>
          </table>
          <div v-else class="empty-inline"><div><h3>{{ t('adminDatabase.noRows') }}</h3><p>{{ t('adminDatabase.empty') }}</p></div></div>
          <button v-if="nextCursor && selectedTableName" class="button button--secondary" type="button" :disabled="busy" @click="queryRows(selectedTableName, queryInput(), { append: true })">{{ busy ? t('common.loading') : t('adminDatabase.loadMore') }}</button>
        </div>
      </div>
    </AdminSectionState>
    <DatabaseRecordEditor v-if="editing && selectedTable" :table="selectedTable" :action="editing.action" :row="editing.row" :review="review" :busy="busy" @cancel="closeEditor" @invalidate="clearReview" @review="requestReview" @apply="apply" />
  </section>
</template>

<style scoped>
.database-layout { display: grid; gap: 0.8rem; padding: 1rem; }
.database-tables { display: flex; gap: 0.4rem; overflow-x: auto; padding-bottom: 0.35rem; }
.database-table-button { min-height: 44px; display: inline-flex; align-items: center; gap: 0.45rem; flex: 0 0 auto; padding: 0.55rem 0.7rem; border: 1px solid var(--line); border-radius: var(--radius-control); color: var(--text-muted); background: var(--surface-raised); font-size: 0.72rem; cursor: pointer; }
.database-table-button small { color: var(--warning); font-size: 0.58rem; }
.database-table-button--active { color: var(--accent); border-color: #557763; background: var(--accent-soft); }
.database-rows { overflow: auto; border: 1px solid var(--line); border-radius: var(--radius-control); }
.database-query { position: sticky; left: 0; display: flex; flex-wrap: wrap; align-items: center; gap: 0.45rem; padding: 0.6rem; border-bottom: 1px solid var(--line); background: var(--surface-raised); }
.database-search { min-width: min(260px, 100%); display: flex; align-items: center; gap: 0.4rem; flex: 1; padding: 0 0.55rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface); }
.database-search input { min-width: 0; flex: 1; border: 0; background: transparent; }
.database-filter { width: 100%; display: grid; grid-template-columns: minmax(120px, 1fr) minmax(110px, 1fr) minmax(120px, 1fr) auto; gap: 0.35rem; }
.database-rows table { min-width: 100%; border-collapse: collapse; font-size: 0.7rem; }
.database-rows th, .database-rows td { max-width: 220px; padding: 0.55rem 0.7rem; border-bottom: 1px solid var(--line); text-align: left; white-space: nowrap; }
.database-rows th { position: sticky; top: 0; z-index: 1; color: var(--text-muted); background: var(--surface-raised); }
.database-rows td span { display: block; overflow: hidden; text-overflow: ellipsis; font-family: var(--font-mono); }
.database-rows > .button { margin: 0.7rem; }

@media (min-width: 900px) {
  .database-layout { grid-template-columns: 210px minmax(0, 1fr); }
  .database-tables { flex-direction: column; overflow: visible; }
}
</style>
