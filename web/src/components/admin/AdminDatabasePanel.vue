<script setup lang="ts">
import { shallowRef } from 'vue'
import { PhDatabase, PhPencilSimple, PhPlus, PhTrash } from '@phosphor-icons/vue'

import type { DatabaseMutationInput, DatabaseRow, DatabaseValue } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useAdminDatabase } from '@/composables/useAdminDatabase'
import AdminSectionState from './AdminSectionState.vue'
import DatabaseRecordEditor from './database/DatabaseRecordEditor.vue'

const {
  tables, selectedTableName, selectedTable, rows, nextCursor, loading, busy, error,
  review, lastRescueBackupId, loadTables, loadRows, selectTable,
  reviewMutation, applyMutation, clearReview,
} = useAdminDatabase()
type EditorState = { action: 'insert' | 'update' | 'delete'; row?: DatabaseRow }
type EditorPayload = { action: EditorState['action']; row?: DatabaseRow; values: Record<string, DatabaseValue>; reason: string }

const editing = shallowRef<EditorState | null>(null)

function displayValue(value: DatabaseValue): string {
  if (value === null) return 'NULL'
  if (typeof value === 'object') return '[BLOB]'
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
  await selectTable(name)
}
</script>

<template>
  <section class="admin-panel database-panel">
    <div class="admin-panel__heading"><div><h2>Database editor</h2><p>Typed records only. Raw SQL and internal SQLite tables are unavailable.</p></div><button class="button button--primary" type="button" :disabled="!selectedTable" @click="openEditor('insert')"><PhPlus :size="18" />Insert row</button></div>
    <InlineNotice tone="warning" title="Domain hooks are bypassed">Use normal admin actions whenever possible. Every direct change requires a reason, diff review, typed confirmation, and rescue backup.</InlineNotice>
    <InlineNotice v-if="lastRescueBackupId" tone="success" title="Reviewed change applied">Rescue backup {{ lastRescueBackupId }} was created before the mutation.</InlineNotice>
    <AdminSectionState :loading="loading" :error="error" @retry="loadTables">
      <div class="database-layout">
        <nav class="database-tables" aria-label="Application tables">
          <button v-for="table in tables" :key="table.name" class="database-table-button" :class="{ 'database-table-button--active': selectedTableName === table.name }" type="button" @click="chooseTable(table.name)"><PhDatabase :size="17" /><span>{{ table.name }}</span><small>High risk</small></button>
        </nav>
        <div class="database-rows" tabindex="0" aria-label="Table rows">
          <table v-if="selectedTable && rows.length">
            <thead><tr><th v-for="column in selectedTable.columns" :key="column.name" scope="col">{{ column.name }}</th><th scope="col">Actions</th></tr></thead>
            <tbody><tr v-for="row in rows" :key="row.recordHash"><td v-for="column in selectedTable.columns" :key="column.name"><span>{{ column.sensitive ? 'MASKED' : displayValue(row.values[column.name]) }}</span></td><td><div class="row-actions"><button class="icon-button" type="button" aria-label="Edit row" @click="openEditor('update', row)"><PhPencilSimple :size="18" /></button><button class="icon-button icon-button--danger" type="button" aria-label="Delete row" @click="openEditor('delete', row)"><PhTrash :size="18" /></button></div></td></tr></tbody>
          </table>
          <div v-else class="empty-inline"><div><h3>No rows</h3><p>This table is empty.</p></div></div>
          <button v-if="nextCursor && selectedTableName" class="button button--secondary" type="button" :disabled="busy" @click="loadRows(selectedTableName, { append: true })">{{ busy ? 'Loading' : 'Load more' }}</button>
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
