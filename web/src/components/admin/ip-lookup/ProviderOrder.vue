<script setup lang="ts">
import { providerName } from '@/components/ip-lookup/presentation'

const order = defineModel<string[]>({ required: true })
defineProps<{ title: string; help: string; disabled?: boolean }>()

function move(index: number, delta: number): void {
  const target = index + delta
  if (target < 0 || target >= order.value.length) return
  const next = [...order.value]
  const [provider] = next.splice(index, 1)
  if (!provider) return
  next.splice(target, 0, provider)
  order.value = next
}
</script>

<template>
  <section class="provider-order">
    <div><h2>{{ title }}</h2><p>{{ help }}</p></div>
    <ol>
      <li v-for="(id, index) in order" :key="id">
        <span class="order-number">{{ index + 1 }}</span><span class="order-name">{{ providerName(id) }}</span>
        <UButton type="button" color="neutral" variant="ghost" icon="i-ph-arrow-up" :disabled="disabled || index === 0" :aria-label="$t('ipLookup.moveUp', { provider: providerName(id) })" @click="move(index, -1)" />
        <UButton type="button" color="neutral" variant="ghost" icon="i-ph-arrow-down" :disabled="disabled || index === order.length - 1" :aria-label="$t('ipLookup.moveDown', { provider: providerName(id) })" @click="move(index, 1)" />
      </li>
    </ol>
  </section>
</template>

<style scoped>
.provider-order { display: grid; gap: 0.6rem; min-width: 0; }
.provider-order h2 { font-size: 1rem; margin: 0; }
.provider-order p { color: var(--text-muted); font-size: 0.8rem; margin: 0.2rem 0 0; }
.provider-order ol { list-style: none; padding: 0; margin: 0; }
.provider-order li { display: flex; align-items: center; gap: 0.5rem; border-bottom: 1px solid var(--line); }
.order-number { width: 1.2rem; color: var(--text-faint); font: 0.75rem var(--font-mono); }
.order-name { flex: 1; min-width: 0; overflow-wrap: anywhere; font-size: 0.85rem; }
</style>
