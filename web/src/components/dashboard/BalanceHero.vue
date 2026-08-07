<script setup lang="ts">
import { PhArrowRight, PhPlus } from '@phosphor-icons/vue'
import { RouterLink } from 'vue-router'

import type { Money } from '@/api/types'
import { formatMoney } from '@/utils/format'

defineProps<{
  balance: Money
  firstName: string
}>()
</script>

<template>
  <section class="balance-hero">
    <div class="balance-hero__copy">
      <p class="eyebrow">Good to see you, {{ firstName }}</p>
      <p class="balance-hero__label">Available balance</p>
      <h1>{{ formatMoney(balance) }}</h1>
    </div>
    <div class="balance-hero__actions">
      <RouterLink class="button button--light" to="/balance">
        <PhPlus :size="19" weight="bold" />
        Add balance
      </RouterLink>
      <RouterLink class="text-button text-button--on-accent" to="/catalog">
        Browse plans
        <PhArrowRight :size="18" />
      </RouterLink>
    </div>
  </section>
</template>
