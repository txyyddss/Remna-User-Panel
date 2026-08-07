<script setup lang="ts">
import { computed, reactive, shallowRef } from 'vue'
import { PhMagnifyingGlass, PhMinus, PhPlus, PhUserFocus } from '@phosphor-icons/vue'

import type { AdminUserSummary } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { formatDate, formatMoney } from '@/utils/format'
import AdminSectionState from './AdminSectionState.vue'

const { items, loading, busy, error, load, perform } = useAdminSection<AdminUserSummary>('users')
const query = shallowRef('')
const selected = shallowRef<AdminUserSummary | null>(null)
const form = reactive({ amountMinor: '', reason: '' })
const success = shallowRef<string | null>(null)

const filteredUsers = computed(() => {
  const needle = query.value.trim().toLowerCase()
  if (!needle) return items.value
  return items.value.filter(({ user }) =>
    [user.firstName, user.lastName, user.username, String(user.telegramId)].some((value) => value?.toLowerCase().includes(needle)),
  )
})

function displayName(summary: AdminUserSummary): string {
  return [summary.user.firstName, summary.user.lastName].filter(Boolean).join(' ') || 'Telegram member'
}

async function adjust(sign: 1 | -1): Promise<void> {
  if (!selected.value || !/^\d+$/.test(form.amountMinor) || form.reason.length < 4) return
  const amount = sign === -1 ? `-${form.amountMinor}` : form.amountMinor
  const ok = await perform(() => import('@/api/client').then(({ api }) => api.adjustBalance(selected.value!.user.id, amount, form.reason)))
  if (ok) {
    success.value = `Balance adjusted for ${displayName(selected.value)}.`
    selected.value = null
    form.amountMinor = ''
    form.reason = ''
  }
}
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>Users</h2><p>Search accounts, control purchasing, and append audited balance entries.</p></div>
      <label class="admin-search"><PhMagnifyingGlass :size="18" /><input v-model="query" type="search" placeholder="Search users" aria-label="Search users" /></label>
    </div>
    <InlineNotice v-if="success" tone="success">{{ success }}</InlineNotice>
    <AdminSectionState :loading="loading" :error="error" @retry="load()">
      <div class="admin-list">
        <article v-for="summary in filteredUsers" :key="summary.user.id" class="admin-list-row admin-list-row--user">
          <span class="avatar avatar--small">{{ displayName(summary).slice(0, 2).toUpperCase() }}</span>
          <div><strong>{{ displayName(summary) }}</strong><small>{{ summary.user.username ? `@${summary.user.username}` : summary.user.telegramId }} / Joined {{ formatDate(summary.createdAt) }}</small></div>
          <strong>{{ formatMoney(summary.balance) }}</strong>
          <StatusBadge :tone="summary.synchronization?.status === 'synchronized' ? 'success' : 'warning'" :label="summary.synchronization?.status ?? summary.user.onboardingState" />
          <div class="row-actions">
            <button class="button button--secondary button--small" type="button" @click="selected = summary"><PhUserFocus :size="17" /> Adjust</button>
          </div>
        </article>
        <div v-if="!filteredUsers.length" class="empty-inline"><div><h3>No matching users</h3><p>Try a Telegram name, username, or numeric ID.</p></div></div>
      </div>
    </AdminSectionState>

    <form v-if="selected" class="admin-drawer" @submit.prevent>
      <div class="admin-drawer__heading"><div><h3>Adjust {{ displayName(selected) }}</h3><p>Enter TXB minor units. For example, 1250 is 12.50 TXB.</p></div><button class="text-button" type="button" @click="selected = null">Close</button></div>
      <label><span>Amount, minor units</span><input v-model="form.amountMinor" required inputmode="numeric" pattern="[0-9]+" /></label>
      <label><span>Reason</span><input v-model.trim="form.reason" required minlength="4" maxlength="300" /></label>
      <div class="button-row">
        <button class="button button--secondary" type="button" :disabled="busy" @click="adjust(1)"><PhPlus :size="18" /> Add TXB</button>
        <button class="button button--ghost-danger" type="button" :disabled="busy" @click="adjust(-1)"><PhMinus :size="18" /> Deduct TXB</button>
      </div>
    </form>
  </section>
</template>
