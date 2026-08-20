<script setup lang="ts">
import type { AffiliateTier } from '@/api/features'
import { affiliateFormSchema } from './affiliateForm'
import AdminAffiliateTierEditor from './AdminAffiliateTierEditor.vue'
import { useAdminAffiliates } from './useAdminAffiliates'

const state = useAdminAffiliates()
function update(index: number, tier: AffiliateTier): void { state.replace(state.tiers.value.map((item, itemIndex) => itemIndex === index ? tier : item)) }
function move(index: number, offset: number): void {
  const next = [...state.tiers.value]
  const target = index + offset
  if (target < 0 || target >= next.length) return
  ;[next[index], next[target]] = [next[target]!, next[index]!]
  state.replace(next)
}
function remove(index: number): void { state.replace(state.tiers.value.filter((_, itemIndex) => itemIndex !== index)) }
function add(): void {
  const last = state.tiers.value.at(-1)
  state.replace([...state.tiers.value, { id: '', name: '', threshold: (last?.threshold ?? 0) + 1, enabled: true, commissionEnabled: false, commissionBps: 0, reward: { kind: 'none' } }])
}
</script>

<template>
  <section class="admin-panel admin-affiliates">
    <div class="admin-panel__heading"><div><h2>{{ $t('adminAffiliates.title') }}</h2><p>{{ $t('adminAffiliates.copy') }}</p></div></div>
    <USkeleton v-if="state.loading.value" class="h-64 w-full" />
    <template v-else-if="state.configuration.value">
      <div class="admin-affiliates__identity"><div><span>{{ $t('adminAffiliates.botStatus') }}</span><strong>{{ $t(`adminAffiliates.bot.${state.configuration.value.bot.status}`) }}</strong></div><code>{{ state.configuration.value.bot.username ? `@${state.configuration.value.bot.username}` : $t('adminAffiliates.notDiscovered') }}</code></div>
      <UForm :schema="affiliateFormSchema" :state="{ tiers: state.tiers.value }" class="admin-affiliates__form" @submit="state.save">
        <AdminAffiliateTierEditor v-for="(tier, index) in state.tiers.value" :key="tier.id || `new-${index}`" :tier="tier" :coupons="state.coupons.value" :index="index" :count="state.tiers.value.length" @update="update(index, $event)" @move="move(index, $event)" @remove="remove(index)" />
        <div class="button-row"><UButton color="neutral" variant="outline" icon="i-ph-plus" :label="$t('adminAffiliates.addTier')" @click="add" /><UButton type="submit" :loading="state.saving.value" :label="state.saving.value ? $t('common.saving') : $t('common.save')" /></div>
      </UForm>
    </template>
    <UAlert v-if="state.error.value" color="warning" variant="soft" icon="i-ph-warning" :description="state.error.value" />
  </section>
</template>
