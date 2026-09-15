<script setup lang="ts">
import { computed } from 'vue'

import type { Money } from '@/api/types'
import { formatCatalogMoneyLines } from '@/utils/displayCurrency'

const props = defineProps<{
  money: Money
}>()

const lines = computed(() => formatCatalogMoneyLines(props.money))
</script>

<template>
  <span class="dual-currency-money">
    <span>{{ lines.primary }}</span>
    <small v-if="lines.approximation" class="dual-currency-money__approximation">{{ lines.approximation }}</small>
  </span>
</template>

<style scoped>
.dual-currency-money { display: grid; gap: 0.12rem; }
.dual-currency-money__approximation { color: var(--text-muted); font-family: var(--font-mono); font-size: 0.56em; font-weight: 600; line-height: 1.25; }
</style>
