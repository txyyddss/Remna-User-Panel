<script setup lang="ts">
import type { DatabaseFilterOperator } from '@/api/features'
import { useI18n } from '@/i18n'
import type { DatabaseColumnOption, DatabaseOperatorOption, TextDatabaseFilter } from './types'

const props = defineProps<{
  filters: readonly TextDatabaseFilter[]
  columnItems: readonly DatabaseColumnOption[]
  operators: readonly DatabaseOperatorOption[]
}>()
const emit = defineEmits<{ 'update:filters': [value: TextDatabaseFilter[]] }>()
const { t } = useI18n()

function updateFilter(index: number, patch: Partial<TextDatabaseFilter>): void {
  emit('update:filters', props.filters.map((filter, itemIndex) => itemIndex === index ? { ...filter, ...patch } : { ...filter }))
}

function removeFilter(index: number): void {
  emit('update:filters', props.filters.filter((_, itemIndex) => itemIndex !== index).map((filter) => ({ ...filter })))
}

function updateOperator(index: number, value: unknown): void {
  updateFilter(index, { operator: String(value) as DatabaseFilterOperator })
}

function isNullOperator(operator: DatabaseFilterOperator): boolean {
  return operator === 'is_null' || operator === 'not_null'
}
</script>

<template>
  <div class="database-filter-fields">
    <div v-for="(filter, index) in filters" :key="index" class="database-filter-fields__row">
      <USelect
        :model-value="filter.column"
        :items="columnItems"
        value-key="value"
        :aria-label="t('adminDatabase.filterColumn')"
        @update:model-value="updateFilter(index, { column: String($event) })"
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
        class="database-filter-fields__value"
        :model-value="filter.value ?? ''"
        :aria-label="t('adminDatabase.filterValue')"
        :placeholder="t('adminDatabase.value')"
        @update:model-value="updateFilter(index, { value: String($event) })"
      />
      <UButton
        class="database-filter-fields__remove"
        color="neutral"
        variant="ghost"
        icon="i-ph-x"
        :aria-label="t('adminDatabase.removeFilter')"
        @click="removeFilter(index)"
      />
    </div>
  </div>
</template>

<style scoped>
.database-filter-fields { display: grid; gap: 0.55rem; }
.database-filter-fields__row { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 44px; gap: 0.4rem; align-items: start; }
.database-filter-fields__value { grid-column: 1 / -1; }
.database-filter-fields__remove { grid-column: 3; grid-row: 1; min-width: 44px; min-height: 44px; }

@media (min-width: 900px) {
  .database-filter-fields__row { grid-template-columns: minmax(120px, 1fr) minmax(110px, 1fr) minmax(120px, 1fr) 44px; }
  .database-filter-fields__value { grid-column: 3; grid-row: 1; }
  .database-filter-fields__remove { grid-column: 4; }
}
</style>
