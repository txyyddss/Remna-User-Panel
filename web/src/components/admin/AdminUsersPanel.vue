<script setup lang="ts">
import { computed, reactive, shallowRef } from 'vue'

import type { AdminUserSummary } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { useI18n } from '@/i18n'
import { formatDate, formatMoney, moneyFromTxbInput } from '@/utils/format'
import AdminSectionState from './AdminSectionState.vue'

const { items, loading, busy, error, load, perform } = useAdminSection<AdminUserSummary>('users')
const { t } = useI18n()
const query = shallowRef('')
const selected = shallowRef<AdminUserSummary | null>(null)
const form = reactive({ amountTxb: '', reason: '' })
const success = shallowRef<string | null>(null)

const filteredUsers = computed(() => {
  const needle = query.value.trim().toLowerCase()
  if (!needle) return items.value
  return items.value.filter(({ user }) =>
    [user.firstName, user.lastName, user.username, String(user.telegramId)].some((value) => value?.toLowerCase().includes(needle)),
  )
})

function displayName(summary: AdminUserSummary): string {
  return [summary.user.firstName, summary.user.lastName].filter(Boolean).join(' ') || t('nav.member')
}

async function adjust(sign: 1 | -1): Promise<void> {
  const amountMinor = moneyFromTxbInput(form.amountTxb)
  if (!selected.value || !amountMinor || form.reason.length < 4) return
  const current = selected.value
  const amount = sign === -1 ? `-${amountMinor}` : amountMinor
  const ok = await perform(() => import('@/api/client').then(({ api }) => api.adjustBalance(current.user.id, amount, form.reason)))
  if (ok) {
    success.value = t('adminUsers.adjusted', { name: displayName(current) })
    selected.value = null
    form.amountTxb = ''
    form.reason = ''
  }
}
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>{{ t('adminUsers.title') }}</h2><p>{{ t('adminUsers.copy') }}</p></div>
      <UInput v-model="query" type="search" icon="i-ph-magnifying-glass" :placeholder="t('adminUsers.search')" :aria-label="t('adminUsers.search')" />
    </div>
    <InlineNotice v-if="success" tone="success">{{ success }}</InlineNotice>
    <AdminSectionState :loading="loading" :error="error" @retry="load()">
      <div v-auto-animate class="admin-list">
        <article v-for="summary in filteredUsers" :key="summary.user.id" class="admin-list-row admin-list-row--user">
          <UAvatar :text="displayName(summary).slice(0, 2).toUpperCase()" size="sm" />
          <div>
            <strong>{{ displayName(summary) }}</strong>
            <small>{{ t('adminUsers.meta', { identity: summary.user.username ? `@${summary.user.username}` : summary.user.telegramId, date: formatDate(summary.createdAt) }) }}</small>
          </div>
          <strong>{{ formatMoney(summary.balance) }}</strong>
          <StatusBadge :tone="summary.synchronization.status === 'synchronized' ? 'success' : 'warning'" :label="summary.synchronization.status === 'synchronized' ? t('adminUsers.synchronized') : t('adminUsers.notProvisioned')" />
          <UButton size="sm" color="neutral" variant="outline" icon="i-ph-user-focus" :label="t('adminUsers.adjust')" @click="selected = summary" />
        </article>
        <div v-if="!filteredUsers.length" class="empty-inline"><div><h3>{{ t('adminUsers.none') }}</h3><p>{{ t('adminUsers.noneHint') }}</p></div></div>
      </div>
    </AdminSectionState>

    <UDrawer :open="Boolean(selected)" :title="selected ? t('adminUsers.adjustNamed', { name: displayName(selected) }) : t('adminUsers.adjust')" :description="t('adminUsers.adjustHint')" @update:open="!$event && (selected = null)">
      <template #body>
        <form v-if="selected" class="form-stack" @submit.prevent>
          <TxbAmountField id="balance-adjustment" v-model="form.amountTxb" :label="t('adminUsers.amount')" min-minor="1" required />
          <UFormField name="reason" :label="t('adminReason.reason')" required>
            <UInput v-model.trim="form.reason" :minlength="4" :maxlength="300" />
          </UFormField>
          <div class="button-row">
            <UButton color="neutral" variant="outline" icon="i-ph-plus" :disabled="busy" :label="t('adminUsers.add')" @click="adjust(1)" />
            <UButton color="error" variant="ghost" icon="i-ph-minus" :disabled="busy" :label="t('adminUsers.deduct')" @click="adjust(-1)" />
          </div>
        </form>
      </template>
    </UDrawer>
  </section>
</template>
