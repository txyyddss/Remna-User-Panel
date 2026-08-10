<script setup lang="ts">
import { computed } from 'vue'
import { PhCaretDown, PhUserPlus } from '@phosphor-icons/vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { useI18n } from '@/i18n'
import { useSessionStore } from '@/stores/session'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const sessionStore = useSessionStore()

const groups = [
  { labelKey: 'adminNav.commerce', sections: [
    { value: 'catalog', labelKey: 'adminNav.catalog' },
    { value: 'coupons', labelKey: 'adminNav.coupons' },
    { value: 'payments', labelKey: 'adminNav.payments' },
  ] },
  { labelKey: 'adminNav.community', sections: [
    { value: 'activity', labelKey: 'adminNav.activity' },
    { value: 'questionnaires', labelKey: 'adminNav.questionnaires' },
    { value: 'onboarding', labelKey: 'adminNav.onboarding' },
  ] },
  { labelKey: 'adminNav.accounts', sections: [
    { value: 'entitlements', labelKey: 'adminNav.entitlements' },
  ] },
  { labelKey: 'adminNav.system', sections: [
    { value: 'settings', labelKey: 'adminNav.settings' },
    { value: 'backups', labelKey: 'adminNav.backups' },
    { value: 'database', labelKey: 'adminNav.database' },
    { value: 'audit', labelKey: 'adminNav.audit' },
  ] },
] as const

const selected = computed(() => String(route.params.section || 'settings'))

function navigate(event: Event): void {
  const section = (event.target as HTMLSelectElement).value
  if (section) void router.push(`/admin/${section}`)
}
</script>

<template>
  <div class="admin-shell">
    <header class="page-header admin-shell__header">
      <p class="eyebrow">{{ t('adminNav.eyebrow') }}</p>
      <h1>{{ t('adminNav.title') }}</h1>
      <p>{{ t('adminNav.copy') }}</p>
      <div v-if="!sessionStore.onboardingComplete" class="button-row">
        <RouterLink class="button button--secondary" to="/onboarding">
          <PhUserPlus :size="19" weight="bold" />
          {{ t('adminNav.setupAccount') }}
        </RouterLink>
      </div>
    </header>
    <nav class="admin-section-picker" :aria-label="t('adminNav.sections')">
      <label>
        <span>{{ t('adminNav.section') }}</span>
        <span class="admin-section-picker__control">
          <select :value="selected" @change="navigate">
            <optgroup v-for="group in groups" :key="group.labelKey" :label="t(group.labelKey)">
              <option v-for="section in group.sections" :key="section.value" :value="section.value">{{ t(section.labelKey) }}</option>
            </optgroup>
          </select>
          <PhCaretDown :size="16" aria-hidden="true" />
        </span>
      </label>
    </nav>
    <div class="admin-shell__content"><slot /></div>
  </div>
</template>

<style scoped>
.admin-section-picker { margin: 0 0 0.9rem; padding: 0.75rem; border: 1px solid var(--line); border-radius: var(--radius-card); background: var(--surface-raised); }
.admin-section-picker label { display: grid; gap: 0.35rem; }
.admin-section-picker label > span:first-child { color: var(--text-faint); font-size: 0.66rem; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase; }
.admin-section-picker__control { position: relative; display: block; }
.admin-section-picker select { width: 100%; min-height: 44px; appearance: none; padding: 0.65rem 2.4rem 0.65rem 0.75rem; border: 1px solid var(--line); border-radius: var(--radius-control); color: var(--text); background: var(--surface); font: inherit; font-size: 0.82rem; }
.admin-section-picker__control > svg { position: absolute; top: 50%; right: 0.8rem; pointer-events: none; transform: translateY(-50%); color: var(--text-muted); }
@media (min-width: 760px) { .admin-section-picker { max-width: 390px; } }
</style>
