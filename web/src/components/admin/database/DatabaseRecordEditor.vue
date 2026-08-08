<script setup lang="ts">
import { computed, reactive, shallowRef, watch } from 'vue'
import { PhCheckCircle, PhFloppyDisk, PhWarning, PhX } from '@phosphor-icons/vue'

import type { DatabaseMutationReview, DatabaseRow, DatabaseTable, DatabaseValue } from '@/api/features'
import SwitchField from '@/components/common/SwitchField.vue'

type DeepReadonly<T> = T extends readonly (infer Item)[]
  ? readonly DeepReadonly<Item>[]
  : T extends object ? { readonly [Key in keyof T]: DeepReadonly<T[Key]> } : T

const props = defineProps<{ table: DatabaseTable; action: 'insert' | 'update' | 'delete'; row?: DatabaseRow; review: DeepReadonly<DatabaseMutationReview> | null; busy: boolean }>()
const emit = defineEmits<{
  cancel: []
  invalidate: []
  review: [payload: { action: 'insert' | 'update' | 'delete'; row?: DatabaseRow; values: Record<string, DatabaseValue>; reason: string }]
  apply: [payload: { action: 'insert' | 'update' | 'delete'; row?: DatabaseRow; values: Record<string, DatabaseValue>; reason: string; confirmation: string }]
}>()

const draft = reactive<Record<string, DatabaseValue>>({})
const textDraft = reactive<Record<string, string>>({})
const nullDraft = reactive<Record<string, boolean>>({})
const reason = shallowRef('')
const confirmation = shallowRef('')
const canReview = computed(() => reason.value.trim().length >= 4)
const canApply = computed(() => Boolean(props.review && confirmation.value === props.review.requiredConfirmation))
const title = computed(() => `${props.action === 'insert' ? 'Insert' : props.action === 'delete' ? 'Delete' : 'Edit'} ${props.table.name} row`)

watch([() => props.row, () => props.action, () => props.table], ([row]) => {
  for (const key of Object.keys(draft)) delete draft[key]
  for (const key of Object.keys(textDraft)) delete textDraft[key]
  for (const key of Object.keys(nullDraft)) delete nullDraft[key]
  for (const column of props.table.columns) {
    const value = row?.values[column.name] ?? (column.nullable ? null : /BOOL/i.test(column.declaredType) ? false : '')
    draft[column.name] = value
    nullDraft[column.name] = value === null
    textDraft[column.name] = typeof value === 'object' && value !== null
      ? value.blobBase64
      : value === null ? '' : String(value)
  }
  reason.value = ''
  confirmation.value = ''
}, { immediate: true })

watch(() => props.review, (review) => {
  if (!review) confirmation.value = ''
})

function isBoolean(columnType: string, value: DatabaseValue): boolean {
  return typeof value === 'boolean' || /BOOL/i.test(columnType)
}

function isBlob(columnType: string, value: DatabaseValue): boolean {
  return (typeof value === 'object' && value !== null) || /BLOB|BINARY/i.test(columnType)
}

function isNumeric(columnType: string): boolean {
  return /INT|REAL|FLOA|DOUB|NUMERIC|DECIMAL/i.test(columnType)
}

function setBoolean(columnName: string, value: boolean): void {
  draft[columnName] = value
  emit('invalidate')
}

function setNull(columnName: string, value: boolean): void {
  nullDraft[columnName] = value
  emit('invalidate')
}

function values(): Record<string, DatabaseValue> {
  const output: Record<string, DatabaseValue> = {}
  for (const column of props.table.columns) {
    const current = draft[column.name]
    if (!column.editable || column.sensitive) continue
    if (nullDraft[column.name]) {
      output[column.name] = null
      continue
    }
    output[column.name] = isBlob(column.declaredType, current)
      ? { blobBase64: textDraft[column.name] ?? '' }
      : typeof current === 'boolean'
      ? current
      : textDraft[column.name]
  }
  return output
}

function submit(): void {
  if (!props.review) {
    if (canReview.value) emit('review', { action: props.action, row: props.row, values: values(), reason: reason.value.trim() })
    return
  }
  if (canApply.value) emit('apply', { action: props.action, row: props.row, values: values(), reason: reason.value.trim(), confirmation: confirmation.value })
}

function displayValue(value: DatabaseValue | undefined): string {
  if (value === undefined) return 'Not present'
  if (value === null) return 'NULL'
  if (typeof value === 'object') return '[BLOB]'
  return String(value)
}
</script>

<template>
  <aside class="database-drawer" aria-labelledby="database-editor-title">
    <header class="database-drawer__header"><div><h3 id="database-editor-title">{{ title }}</h3><p>Direct edits bypass domain synchronization hooks.</p></div><button class="icon-button" type="button" aria-label="Close record editor" @click="$emit('cancel')"><PhX :size="19" /></button></header>
    <div class="database-warning"><PhWarning :size="20" weight="fill" /><p>{{ review?.warning || table.warning || 'Every direct mutation is high risk and creates a rescue backup.' }}</p></div>
    <form class="database-form" @submit.prevent="submit">
      <div v-if="action === 'delete'" class="database-delete-summary"><strong>Record key</strong><code>{{ JSON.stringify(row?.key ?? {}) }}</code><p>The reviewed diff must show the full row removal before apply.</p></div>
      <template v-for="column in action === 'delete' ? [] : table.columns" :key="column.name">
        <div class="database-field-wrap">
          <SwitchField
            v-if="column.editable && !column.sensitive && isBoolean(column.declaredType, draft[column.name]) && !nullDraft[column.name]"
            :id="`db-${column.name}`"
            :model-value="draft[column.name] === true"
            :disabled="review !== null"
            :label="column.name"
            :help="column.declaredType"
            @update:model-value="setBoolean(column.name, $event)"
          />
          <label v-else class="database-field"><span>{{ column.name }} <small>{{ column.declaredType }}</small></span><textarea v-if="isBlob(column.declaredType, draft[column.name])" v-model.trim="textDraft[column.name]" rows="3" spellcheck="false" :disabled="review !== null || !column.editable || column.sensitive || nullDraft[column.name]" placeholder="Base64-encoded bytes" @input="$emit('invalidate')" /><input v-else v-model="textDraft[column.name]" :inputmode="isNumeric(column.declaredType) ? 'decimal' : 'text'" :disabled="review !== null || !column.editable || column.sensitive || nullDraft[column.name]" :placeholder="column.sensitive ? 'Masked, use settings to replace' : ''" @input="$emit('invalidate')" /></label>
          <label v-if="column.nullable && column.editable && !column.sensitive" class="database-null-toggle"><input :checked="nullDraft[column.name]" type="checkbox" :disabled="review !== null" @change="setNull(column.name, ($event.target as HTMLInputElement).checked)" />Store NULL</label>
        </div>
      </template>
      <label class="database-field"><span>Reason</span><textarea v-model.trim="reason" rows="3" required minlength="4" :disabled="review !== null" placeholder="Why is this direct edit necessary?" @input="$emit('invalidate')" /></label>

      <section v-if="review" class="database-diff" aria-labelledby="database-diff-title">
        <div><PhCheckCircle :size="20" /><h4 id="database-diff-title">Review exact changes</h4></div>
        <dl><template v-for="column in review.changedColumns" :key="column"><dt>{{ column }}</dt><dd><span>{{ displayValue(review.before?.[column]) }}</span><strong>to</strong><span>{{ displayValue(review.after?.[column]) }}</span></dd></template></dl>
        <p>Rescue backup required before apply.</p>
      </section>

      <label v-if="review" class="database-field"><span>Type {{ review.requiredConfirmation }}</span><input v-model="confirmation" required autocomplete="off" /></label>
      <div class="button-row">
        <button v-if="review" class="button button--secondary" type="button" :disabled="busy" @click="$emit('invalidate')">Change draft</button>
        <button class="button button--primary" type="submit" :disabled="busy || (review ? !canApply : !canReview)"><PhFloppyDisk :size="18" />{{ busy ? 'Working' : review ? 'Apply reviewed change' : 'Review change' }}</button>
      </div>
    </form>
  </aside>
</template>

<style scoped>
.database-drawer { width: min(100%, 560px); max-height: calc(100dvh - 1rem); overflow: auto; position: fixed; right: 0; bottom: 0; z-index: 60; padding: 1rem; border: 1px solid var(--line-strong); border-radius: var(--radius-panel) var(--radius-panel) 0 0; background: var(--surface); box-shadow: var(--shadow-panel); }
.database-drawer__header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.database-drawer__header h3, .database-drawer__header p { margin: 0; }
.database-drawer__header p { margin-top: 0.3rem; color: var(--text-muted); font-size: 0.7rem; }
.database-warning { display: flex; gap: 0.6rem; margin: 1rem 0; padding: 0.8rem; border-radius: var(--radius-control); color: var(--warning); background: var(--warning-soft); }
.database-warning p { margin: 0; font-size: 0.72rem; line-height: 1.45; }
.database-form { display: grid; gap: 0.8rem; margin-top: 1rem; }
.database-field-wrap { display: grid; gap: 0.35rem; padding-bottom: 0.65rem; border-bottom: 1px solid var(--line); }
.database-field { display: grid; gap: 0.35rem; }
.database-field > span { color: var(--text-muted); font-size: 0.75rem; font-weight: 700; }
.database-field > span small { color: var(--text-faint); font-family: var(--font-mono); font-weight: 400; }
.database-field input, .database-field textarea { width: 100%; min-height: 44px; padding: 0.65rem 0.75rem; border: 1px solid var(--line-strong); border-radius: var(--radius-control); outline: 0; color: var(--text); background: var(--surface-raised); font-family: var(--font-mono); font-size: 0.72rem; }
.database-field textarea { resize: vertical; font-family: var(--font-sans); }
.database-field textarea[spellcheck="false"] { font-family: var(--font-mono); }
.database-null-toggle { min-height: 32px; display: inline-flex; align-items: center; gap: 0.45rem; justify-self: start; color: var(--text-muted); font-size: 0.7rem; }
.database-null-toggle input { width: 18px; height: 18px; accent-color: var(--accent); }
.database-delete-summary { display: grid; gap: 0.45rem; padding: 0.8rem; border: 1px solid #57312f; border-radius: var(--radius-control); background: var(--danger-soft); }
.database-delete-summary strong { color: var(--danger); font-size: 0.78rem; }
.database-delete-summary code { overflow-wrap: anywhere; font-family: var(--font-mono); font-size: 0.68rem; }
.database-delete-summary p { margin: 0; color: var(--text-muted); font-size: 0.68rem; line-height: 1.4; }
.database-diff { padding: 0.8rem; border: 1px solid #4d735a; border-radius: var(--radius-control); background: var(--accent-soft); }
.database-diff > div { display: flex; align-items: center; gap: 0.5rem; color: var(--accent); }
.database-diff h4, .database-diff p { margin: 0; }
.database-diff dl { display: grid; gap: 0.55rem; margin: 0.8rem 0; }
.database-diff dt { color: var(--text-muted); font-size: 0.7rem; font-weight: 800; }
.database-diff dd { display: grid; grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr); align-items: center; gap: 0.45rem; margin: 0; }
.database-diff dd span { overflow: hidden; padding: 0.45rem; border-radius: 8px; background: var(--surface); font-family: var(--font-mono); font-size: 0.65rem; text-overflow: ellipsis; white-space: nowrap; }
.database-diff dd strong, .database-diff p { color: var(--text-faint); font-size: 0.62rem; }

@media (min-width: 900px) { .database-drawer { top: 0; bottom: 0; border-radius: var(--radius-panel) 0 0 var(--radius-panel); } }
</style>
