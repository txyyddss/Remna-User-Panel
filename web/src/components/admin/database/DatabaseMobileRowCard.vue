<script setup lang="ts">
import { computed } from 'vue'

import type { DatabaseRow } from '@/api/features'
import { useI18n } from '@/i18n'

const props = defineProps<{ row: DatabaseRow }>()
const emit = defineEmits<{
  edit: [row: DatabaseRow]
  delete: [row: DatabaseRow]
}>()
const { t } = useI18n()
const keyDisplay = computed(() => JSON.stringify(props.row.key))
</script>

<template>
  <article class="database-row-card">
    <div class="database-row-card__identity">
      <span>{{ t('databaseRecord.recordKey') }}</span>
      <code>{{ keyDisplay }}</code>
    </div>
    <div class="database-row-card__actions">
      <UButton
        color="neutral"
        variant="ghost"
        icon="i-ph-pencil-simple"
        :aria-label="t('adminDatabase.editRow')"
        @click="emit('edit', row)"
      />
      <UButton
        color="error"
        variant="ghost"
        icon="i-ph-trash"
        :aria-label="t('adminDatabase.deleteRow')"
        @click="emit('delete', row)"
      />
    </div>
  </article>
</template>

<style scoped>
.database-row-card {
  min-width: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.7rem;
  border: 1px solid var(--line);
  border-radius: var(--radius-control);
  background: var(--surface-raised);
}

.database-row-card__identity {
  min-width: 0;
  display: grid;
  gap: 0.25rem;
}

.database-row-card__identity span {
  color: var(--text-faint);
  font-size: 0.66rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.database-row-card__identity code {
  min-width: 0;
  overflow-wrap: anywhere;
  color: var(--text);
  font-family: var(--font-mono);
  font-size: 0.72rem;
  line-height: 1.4;
}

.database-row-card__actions {
  display: flex;
  flex: 0 0 auto;
  gap: 0.25rem;
}

.database-row-card__actions :deep(button) {
  min-width: 44px;
  min-height: 44px;
}
</style>
