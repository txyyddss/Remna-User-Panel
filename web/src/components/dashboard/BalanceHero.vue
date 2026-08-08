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
      <p class="eyebrow">{{ $t('dashboard.greeting', { name: firstName }) }}</p>
      <p class="balance-hero__label">{{ $t('billing.availableBalance') }}</p>
      <h1>{{ formatMoney(balance) }}</h1>
    </div>
    <div class="balance-hero__actions">
      <RouterLink class="button button--light" to="/balance">
        <PhPlus :size="19" weight="bold" />
        {{ $t('billing.addBalance') }}
      </RouterLink>
      <RouterLink class="text-button text-button--on-accent" to="/catalog">
        {{ $t('catalog.browsePlans') }}
        <PhArrowRight :size="18" />
      </RouterLink>
    </div>
  </section>
</template>
