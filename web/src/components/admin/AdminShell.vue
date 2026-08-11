<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

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
    { value: 'users', labelKey: 'adminNav.users' },
    { value: 'emby', labelKey: 'adminNav.emby' },
    { value: 'entitlements', labelKey: 'adminNav.entitlements' },
  ] },
  { labelKey: 'adminNav.system', sections: [
    { value: 'settings', labelKey: 'adminNav.settings' },
    { value: 'backups', labelKey: 'adminNav.backups' },
    { value: 'database', labelKey: 'adminNav.database' },
    { value: 'audit', labelKey: 'adminNav.audit' },
  ] },
] as const

const sectionItems = computed(() => groups.flatMap((group) => group.sections.map((section) => ({
  value: section.value,
  label: `${t(group.labelKey)} · ${t(section.labelKey)}`,
}))))
const selected = computed({
  get: () => String(route.params.section || 'settings'),
  set: (section: string) => void router.push(`/admin/${section}`),
})

function goToOnboarding(): void {
  void router.push('/onboarding')
}
</script>

<template>
  <div class="admin-shell">
    <header class="page-header admin-shell__header">
      <p class="eyebrow">{{ t('adminNav.eyebrow') }}</p>
      <h1>{{ t('adminNav.title') }}</h1>
      <p>{{ t('adminNav.copy') }}</p>
      <div v-if="!sessionStore.onboardingComplete" class="button-row">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-ph-user-plus-bold"
          :label="t('adminNav.setupAccount')"
          @click="goToOnboarding"
        />
      </div>
    </header>
    <nav class="admin-section-picker" :aria-label="t('adminNav.sections')">
      <UFormField :label="t('adminNav.section')">
        <USelect v-model="selected" class="admin-section-picker__select" :items="sectionItems" value-key="value" />
      </UFormField>
    </nav>
    <div class="admin-shell__content"><slot /></div>
  </div>
</template>

<style scoped>
.admin-section-picker { max-width: 420px; margin: 0 0 0.9rem; padding: 0.75rem; border: 1px solid var(--line); border-radius: var(--radius-card); background: var(--surface-raised); }
.admin-section-picker__select { width: 100%; }
</style>
