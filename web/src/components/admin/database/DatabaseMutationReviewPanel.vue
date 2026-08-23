<script setup lang="ts">
import type { DatabaseMutationReview, DatabaseValue } from '@/api/features'
import { useI18n } from '@/i18n'
import type { DeepReadonly } from './types'

defineProps<{ review: DeepReadonly<DatabaseMutationReview> }>()
const confirmation = defineModel<string>({ required: true })
const { t } = useI18n()

function displayValue(value: DatabaseValue | undefined): string {
  if (value === undefined) return t('databaseRecord.notPresent')
  if (value === null) return t('adminDatabase.nullValue')
  if (typeof value === 'object') return t('adminDatabase.blobValue')
  return String(value)
}
</script>

<template>
  <section class="database-review" :aria-label="t('databaseRecord.review')">
    <header>
      <UIcon name="i-ph-check-circle" />
      <h3>{{ t('databaseRecord.review') }}</h3>
    </header>
    <dl>
      <div v-for="column in review.changedColumns" :key="column">
        <dt>{{ column }}</dt>
        <dd>
          <span><small>{{ t('databaseRecord.before') }}</small><code>{{ displayValue(review.before?.[column]) }}</code></span>
          <span><small>{{ t('databaseRecord.after') }}</small><code>{{ displayValue(review.after?.[column]) }}</code></span>
        </dd>
      </div>
    </dl>
    <p>{{ t('databaseRecord.backupRequired') }}</p>
    <UFormField
      name="confirmation"
      :label="t('databaseRecord.typeConfirmation', { confirmation: review.requiredConfirmation })"
      required
    >
      <UInput v-model="confirmation" class="database-review__confirmation font-mono" autocomplete="off" />
    </UFormField>
  </section>
</template>

<style scoped>
.database-review { display: grid; gap: 0.8rem; padding: 0.8rem; border: 1px solid var(--accent); border-radius: var(--radius-control); background: var(--accent-soft); }
.database-review header { display: flex; align-items: center; gap: 0.5rem; color: var(--accent); }
.database-review h3,
.database-review p { margin: 0; }
.database-review dl { display: grid; gap: 0.7rem; margin: 0; }
.database-review dl > div { display: grid; gap: 0.35rem; }
.database-review dt { color: var(--text-muted); font-size: 0.7rem; font-weight: 800; overflow-wrap: anywhere; }
.database-review dd { display: grid; grid-template-columns: minmax(0, 1fr); gap: 0.45rem; margin: 0; }
.database-review dd span { min-width: 0; display: grid; gap: 0.2rem; padding: 0.5rem; border-radius: 8px; background: var(--surface); }
.database-review dd small { color: var(--text-faint); font-size: 0.62rem; }
.database-review dd code { font-family: var(--font-mono); font-size: 0.68rem; overflow-wrap: anywhere; }
.database-review p { color: var(--text-faint); font-size: 0.65rem; }
.database-review__confirmation :deep(input) { font-size: 1rem; }

@media (min-width: 640px) {
  .database-review dd { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
</style>
