<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/i18n'

const step = defineModel<number>({ required: true })
const { t } = useI18n()

const items = computed(() => [
  { value: 1, title: t('catalog.steps.core'), icon: 'i-ph-cube' },
  { value: 2, title: t('catalog.steps.squads'), icon: 'i-ph-users-three' },
  { value: 3, title: t('catalog.steps.nodes'), icon: 'i-ph-network' },
  { value: 4, title: t('catalog.steps.coupon'), icon: 'i-ph-ticket' },
  { value: 5, title: t('catalog.steps.review'), icon: 'i-ph-list-checks' },
  { value: 6, title: t('catalog.steps.payment'), icon: 'i-ph-credit-card' },
])

function indicatorIcon(value: number, icon: string): string {
  return value < step.value ? 'i-ph-check-bold' : icon
}
</script>

<template>
  <section class="catalog-progress" :aria-label="$t('catalog.steps.progress')">
    <UStepper v-model="step" color="neutral" size="xs" :items="items" disabled>
      <template #indicator="{ item }">
        <UIcon :name="indicatorIcon(item.value, item.icon)" data-slot="icon" />
      </template>
    </UStepper>
  </section>
</template>

<style scoped>
.catalog-progress { overflow-x: auto; margin: 0.3rem 0 1.15rem; padding: 0.2rem 0; }
.catalog-progress :deep([data-slot='title']) { font-size: 0.58rem; white-space: nowrap; }
</style>
