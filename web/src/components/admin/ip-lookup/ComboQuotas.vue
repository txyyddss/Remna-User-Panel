<script setup lang="ts">
import type { Combo } from '@/api/types'

const quotas = defineModel<Record<string, number | null>>({ required: true })
defineProps<{ combos: Combo[] }>()
</script>

<template>
  <section class="combo-quotas">
    <h2>{{ $t('ipLookup.quotaHeading') }}</h2><p>{{ $t('ipLookup.quotaHelp') }}</p>
    <UFormField v-for="combo in combos" :key="combo.id" :name="`quota-${combo.id}`" :label="combo.name" class="quota-row">
      <UInputNumber v-model="quotas[combo.id]" class="w-full" :min="0" :max="1000000" :step="1" :placeholder="$t('ipLookup.unconfigured')" />
    </UFormField>
  </section>
</template>

<style scoped>
.combo-quotas { display: grid; gap: 0.8rem; }
.combo-quotas h2 { font-size: 1rem; margin: 0; }
.combo-quotas p { color: var(--text-muted); font-size: 0.82rem; margin: 0; }
@media (min-width: 900px) { .quota-row { max-width: 28rem; } }
</style>
