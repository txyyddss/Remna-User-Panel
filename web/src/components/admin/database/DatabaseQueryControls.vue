<script setup lang="ts">
import type { DatabaseFilterOperator } from '@/api/features'
import { useI18n } from '@/i18n'
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
  'add-filter': []
  'remove-filter': [index: number]
}>()
const { t } = useI18n()

function updateFilter(index: number, patch: Partial<TextDatabaseFilter>): void {
  emit('update:filters', props.filters.map((filter, filterIndex) => filterIndex === index ? { ...filter, ...patch } : { ...filter }))
}

function updateColumn(index: number, value: unknown): void {
  updateFilter(index, { column: String(value) })
}

function updateOperator(index: number, value: unknown): void {
  updateFilter(index, { operator: value as DatabaseFilterOperator })
}

function updateValue(index: number, value: unknown): void {
  updateFilter(index, { value: String(value) })
}

function isNullOperator(operator: DatabaseFilterOperator): boolean {
  return operator === 'is_null' || operator === 'not_null'
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
        size="sm"
        color="neutral"
        variant="outline"
        icon="i-ph-plus"
        :disabled="filters.length >= 5"
        :label="t('adminDatabase.addFilter')"
        @click="emit('add-filter')"
      />
    </div>
    <div v-for="(filter, index) in filters" :key="index" class="database-filter">
      <USelect
        :model-value="filter.column"
        :items="columnItems"
        value-key="value"
        :aria-label="t('adminDatabase.filterColumn')"
        @update:model-value="updateColumn(index, $event)"
      />
      <USelect
        :model-value="filter.operator"
        :items="operators"
        value-key="value"
        :aria-label="t('adminDatabase.filterOperator')"
        @update:model-value="updateOperator(index, $event)"
      />
      <UInput
        v-if="!isNullOperator(filter.operator)"
        class="database-filter__value"
        :model-value="filter.value ?? ''"
        :aria-label="t('adminDatabase.filterValue')"
        :placeholder="t('adminDatabase.value')"
        @update:model-value="updateValue(index, $event)"
      />
      <UButton
        class="database-filter__remove"
        color="neutral"
        variant="ghost"
        icon="i-ph-x"
        :aria-label="t('adminDatabase.removeFilter')"
        @click="emit('remove-filter', index)"
      />
    </div>
  </section>
</template>

<style scoped>
.database-query {
  display: grid;
  gap: 0.55rem;
  padding: 0.6rem;
  border-bottom: 1px solid var(--line);
}

.database-query__primary {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 0.45rem;
}

.database-query__primary > * {
  min-width: 0;
}

.database-filter {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 44px;
  gap: 0.4rem;
  align-items: start;
}

.database-filter__value {
  grid-column: 1 / -1;
}

.database-filter__remove {
  grid-column: 3;
  grid-row: 1;
  min-width: 44px;
  min-height: 44px;
}

@media (max-width: 639px) {
  .database-query__primary {
    grid-template-columns: 1fr;
  }

  .database-query__primary > * {
    width: 100%;
  }
}

@media (min-width: 900px) {
  .database-filter {
    grid-template-columns: minmax(120px, 1fr) minmax(110px, 1fr) minmax(120px, 1fr) 44px;
  }

  .database-filter__value {
    grid-column: 3;
    grid-row: 1;
  }

  .database-filter__remove {
    grid-column: 4;
  }
}
</style>
