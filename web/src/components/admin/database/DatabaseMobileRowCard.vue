<script setup lang="ts">
import { computed, shallowRef } from 'vue'

import type { DatabaseColumn, DatabaseRow, DatabaseValue } from '@/api/features'
import { useI18n } from '@/i18n'

const props = defineProps<{
  row: DatabaseRow
  columns: readonly DatabaseColumn[]
}>()
const emit = defineEmits<{
  edit: [row: DatabaseRow]
  delete: [row: DatabaseRow]
}>()
const { t } = useI18n()
const expanded = shallowRef(false)
const keyEntries = computed(() => Object.entries(props.row.key))
const safeFields = computed(() => props.columns
  .filter((column) => !column.sensitive && column.primaryKeyPosition === 0 && !(column.name in props.row.key))
  .map((column) => ({ name: column.name, value: props.row.values[column.name] })))
const hiddenCount = computed(() => Math.max(0, safeFields.value.length - 3))
const visibleFields = computed(() => expanded.value ? safeFields.value : safeFields.value.slice(0, 3))

function displayValue(value: DatabaseValue | undefined): string {
  if (value === undefined) return t('databaseRecord.notPresent')
  if (value === null) return t('adminDatabase.nullValue')
  if (typeof value === 'object') return t('adminDatabase.blobValue')
  return String(value)
}
</script>

<template>
  <article class="database-row-card">
    <section class="database-row-card__section" :aria-label="t('databaseRecord.recordKey')">
      <span class="database-row-card__label">{{ t('databaseRecord.recordKey') }}</span>
      <dl class="database-row-card__key-list">
        <div v-for="([name, value]) in keyEntries" :key="name">
          <dt>{{ name }}</dt>
          <dd><code>{{ value }}</code></dd>
        </div>
      </dl>
    </section>
    <dl v-if="visibleFields.length" v-auto-animate class="database-row-card__preview">
      <div v-for="field in visibleFields" :key="field.name">
        <dt>{{ field.name }}</dt>
        <dd><code>{{ displayValue(field.value) }}</code></dd>
      </div>
    </dl>
    <UButton
      v-if="hiddenCount"
      class="database-row-card__expand"
      color="neutral"
      variant="ghost"
      :icon="expanded ? 'i-ph-caret-up' : 'i-ph-caret-down'"
      :label="expanded ? t('adminDatabase.showLess') : t('adminDatabase.showMore', { count: hiddenCount })"
      :aria-expanded="expanded"
      data-haptic="open"
      @click="expanded = !expanded"
    />
    <div class="database-row-card__actions">
      <UButton
        block
        color="neutral"
        variant="outline"
        icon="i-ph-pencil-simple"
        :label="t('adminDatabase.editRow')"
        data-haptic="open"
        @click="emit('edit', row)"
      />
      <UButton
        block
        color="error"
        variant="soft"
        icon="i-ph-trash"
        :label="t('adminDatabase.deleteRow')"
        data-haptic="destructive"
        @click="emit('delete', row)"
      />
    </div>
  </article>
</template>

<style scoped>
.database-row-card { min-width: 0; padding: 0.8rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.database-row-card__section { min-width: 0; }
.database-row-card__label { display: block; margin-bottom: 0.35rem; color: var(--text-faint); font-size: 0.66rem; font-weight: 800; letter-spacing: 0.05em; text-transform: uppercase; }
.database-row-card__key-list,
.database-row-card__preview { display: grid; gap: 0.4rem; margin: 0; }
.database-row-card__key-list > div,
.database-row-card__preview > div { min-width: 0; display: grid; grid-template-columns: minmax(5rem, 0.4fr) minmax(0, 1fr); gap: 0.55rem; align-items: baseline; }
.database-row-card dt { min-width: 0; color: var(--text-faint); font-size: 0.67rem; overflow-wrap: anywhere; }
.database-row-card dd { min-width: 0; margin: 0; color: var(--text); font-size: 0.72rem; text-align: right; }
.database-row-card code { font-family: var(--font-mono); overflow-wrap: anywhere; }
.database-row-card__preview { margin-top: 0.7rem; padding-top: 0.7rem; border-top: 1px solid var(--line); }
.database-row-card__preview dd { display: -webkit-box; overflow: hidden; -webkit-box-orient: vertical; -webkit-line-clamp: 2; }
.database-row-card__expand { width: 100%; min-height: 44px; margin-top: 0.35rem; }
.database-row-card__actions { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.5rem; margin-top: 0.7rem; padding-top: 0.7rem; border-top: 1px solid var(--line); }
.database-row-card__actions :deep(button) { min-width: 0; min-height: 44px; }
</style>
