<script setup lang="ts">
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
      <UButton
        class="button button--light"
        to="/balance"
        icon="i-ph-plus-bold"
        :label="$t('billing.addBalance')"
      />
      <UButton
        class="text-button text-button--on-accent"
        to="/catalog"
        color="neutral"
        variant="link"
        trailing-icon="i-ph-arrow-right"
        :label="$t('catalog.browsePlans')"
      />
    </div>
  </section>
</template>
