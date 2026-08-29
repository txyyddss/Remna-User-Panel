<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/i18n'
import { indicatorIcon } from './catalogFlowProgress'

const step = defineModel<number>({ required: true })
const { t } = useI18n()

const stepperIndex = computed(() => Math.max(0, step.value - 1))

const items = computed(() => [
  { value: 1, title: t('catalog.steps.squads'), icon: 'i-ph-users-three' },
  { value: 2, title: t('catalog.steps.core'), icon: 'i-ph-cube' },
  { value: 3, title: t('catalog.steps.coupon'), icon: 'i-ph-ticket' },
  { value: 4, title: t('catalog.steps.review'), icon: 'i-ph-list-checks' },
])

</script>

<template>
  <section class="catalog-progress" :aria-label="$t('catalog.steps.progress')">
    <UStepper :model-value="stepperIndex" color="success" size="xs" :items="items" disabled aria-hidden="true">
      <template #indicator="{ item }">
        <UIcon
          :name="indicatorIcon(item.value, step, item.icon)"
          :class="{ 'catalog-progress__icon--completed': item.value < step }"
          data-slot="icon"
        />
      </template>
    </UStepper>
    <span class="sr-only" role="status" aria-live="polite">
      {{ $t('catalog.steps.current', { step, total: items.length }) }}
    </span>
  </section>
</template>

<style scoped>
.catalog-progress { overflow: hidden; margin: 0.3rem 0 1.15rem; padding: 0.15rem 0 0.35rem; }
.catalog-progress :deep([data-slot='root']) { min-width: 0; }
.catalog-progress :deep([data-slot='title']) { color: var(--text-muted); font-size: 0.64rem; line-height: 1.2; overflow-wrap: anywhere; }
.catalog-progress :deep([data-slot='icon']) { position: relative; z-index: 10; }
.catalog-progress :deep(.catalog-progress__icon--completed) { color: var(--accent-ink); }

@media (min-width: 480px) {
  .catalog-progress :deep([data-slot='title']) { font-size: 0.72rem; white-space: nowrap; }
}
</style>
