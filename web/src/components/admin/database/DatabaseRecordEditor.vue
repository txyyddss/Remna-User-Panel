<script setup lang="ts">
import { computed, reactive, shallowRef, watch } from 'vue'

import type { DatabaseMutationReview, DatabaseRow, DatabaseTable, DatabaseValue } from '@/api/features'
import SwitchField from '@/components/common/SwitchField.vue'
import { useI18n } from '@/i18n'

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
const { t } = useI18n()
const canReview = computed(() => reason.value.trim().length >= 4)
const canApply = computed(() => Boolean(props.review && confirmation.value === props.review.requiredConfirmation))
const title = computed(() => t('databaseRecord.title', { action: t(`databaseRecord.actions.${props.action}`), table: props.table.name }))

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
      : typeof current === 'boolean' ? current : textDraft[column.name]
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
  if (value === undefined) return t('databaseRecord.notPresent')
  if (value === null) return t('adminDatabase.nullValue')
  if (typeof value === 'object') return t('adminDatabase.blobValue')
  return String(value)
}
</script>

<template>
  <UDrawer :open="true" :title="title" :description="t('databaseRecord.copy')" @update:open="!$event && emit('cancel')">
    <template #body>
      <UAlert color="warning" variant="soft" icon="i-ph-warning" :description="review?.warning || table.warning || t('databaseRecord.warning')" />
      <form class="database-form" @submit.prevent="submit">
        <div v-if="action === 'delete'" class="database-delete-summary">
          <strong>{{ t('databaseRecord.recordKey') }}</strong><code>{{ JSON.stringify(row?.key ?? {}) }}</code><p>{{ t('databaseRecord.deleteHint') }}</p>
        </div>
        <template v-for="column in action === 'delete' ? [] : table.columns" :key="column.name">
          <div class="database-field-wrap">
            <SwitchField v-if="column.editable && !column.sensitive && isBoolean(column.declaredType, draft[column.name]) && !nullDraft[column.name]" :id="`db-${column.name}`" :model-value="draft[column.name] === true" :disabled="review !== null" :label="column.name" :help="column.declaredType" @update:model-value="setBoolean(column.name, $event)" />
            <UFormField v-else :name="column.name" :label="column.name" :hint="column.declaredType">
              <UTextarea v-if="isBlob(column.declaredType, draft[column.name])" v-model.trim="textDraft[column.name]" class="font-mono" :rows="3" :disabled="review !== null || !column.editable || column.sensitive || nullDraft[column.name]" :placeholder="t('databaseRecord.base64')" @update:model-value="emit('invalidate')" />
              <UInput v-else v-model="textDraft[column.name]" class="font-mono" :inputmode="isNumeric(column.declaredType) ? 'decimal' : 'text'" :disabled="review !== null || !column.editable || column.sensitive || nullDraft[column.name]" :placeholder="column.sensitive ? t('databaseRecord.masked') : undefined" @update:model-value="emit('invalidate')" />
            </UFormField>
            <UCheckbox v-if="column.nullable && column.editable && !column.sensitive" :model-value="nullDraft[column.name]" :disabled="review !== null" :label="t('databaseRecord.storeNull')" @update:model-value="setNull(column.name, Boolean($event))" />
          </div>
        </template>
        <UFormField name="reason" :label="t('databaseRecord.reason')" required>
          <UTextarea v-model.trim="reason" :rows="3" :minlength="4" :disabled="review !== null" :placeholder="t('databaseRecord.reasonPlaceholder')" @update:model-value="emit('invalidate')" />
        </UFormField>
        <section v-if="review" class="database-diff" :aria-label="t('databaseRecord.review')">
          <div><UIcon name="i-ph-check-circle" /><h4>{{ t('databaseRecord.review') }}</h4></div>
          <dl><template v-for="column in review.changedColumns" :key="column"><dt>{{ column }}</dt><dd><span>{{ displayValue(review.before?.[column]) }}</span><strong>{{ t('databaseRecord.to') }}</strong><span>{{ displayValue(review.after?.[column]) }}</span></dd></template></dl>
          <p>{{ t('databaseRecord.backupRequired') }}</p>
        </section>
        <UFormField v-if="review" name="confirmation" :label="t('databaseRecord.typeConfirmation', { confirmation: review.requiredConfirmation })" required>
          <UInput v-model="confirmation" autocomplete="off" />
        </UFormField>
        <div class="button-row">
          <UButton v-if="review" color="neutral" variant="outline" :disabled="busy" :label="t('databaseRecord.changeDraft')" @click="emit('invalidate')" />
          <UButton type="submit" icon="i-ph-floppy-disk" :loading="busy" :disabled="busy || (review ? !canApply : !canReview)" :label="busy ? t('databaseRecord.working') : review ? t('databaseRecord.apply') : t('databaseRecord.reviewChange')" />
        </div>
      </form>
    </template>
  </UDrawer>
</template>

<style scoped>
.database-form { display: grid; gap: 0.8rem; }
.database-field-wrap { display: grid; gap: 0.4rem; padding-bottom: 0.7rem; border-bottom: 1px solid var(--line); }
.database-delete-summary { display: grid; gap: 0.45rem; padding: 0.8rem; border: 1px solid var(--danger); border-radius: var(--radius-control); background: var(--danger-soft); }
.database-delete-summary strong { color: var(--danger); font-size: 0.78rem; }
.database-delete-summary code { overflow-wrap: anywhere; font-family: var(--font-mono); font-size: 0.68rem; }
.database-delete-summary p { margin: 0; color: var(--text-muted); font-size: 0.68rem; }
.database-diff { padding: 0.8rem; border: 1px solid var(--accent); border-radius: var(--radius-control); background: var(--accent-soft); }
.database-diff > div { display: flex; align-items: center; gap: 0.5rem; color: var(--accent); }
.database-diff h4, .database-diff p { margin: 0; }
.database-diff dl { display: grid; gap: 0.55rem; margin: 0.8rem 0; }
.database-diff dt { color: var(--text-muted); font-size: 0.7rem; font-weight: 800; }
.database-diff dd { display: grid; grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr); align-items: center; gap: 0.45rem; margin: 0; }
.database-diff dd span { overflow: hidden; padding: 0.45rem; border-radius: 8px; background: var(--surface); font-family: var(--font-mono); font-size: 0.65rem; text-overflow: ellipsis; white-space: nowrap; }
.database-diff dd strong, .database-diff p { color: var(--text-faint); font-size: 0.62rem; }
</style>
