<script setup lang="ts">
import type { Component } from 'vue'
import {
  PhArchive,
  PhBooks,
  PhDatabase,
  PhGameController,
  PhGift,
  PhCreditCard,
  PhGear,
  PhListMagnifyingGlass,
  PhListChecks,
  PhMonitorPlay,
  PhStack,
  PhUserPlus,
  PhUsers,
} from '@phosphor-icons/vue'
import { RouterLink, useRoute } from 'vue-router'

import { useSessionStore } from '@/stores/session'

const route = useRoute()
const sessionStore = useSessionStore()

const groups: Array<{ label: string; sections: Array<{ value: string; label: string; icon: Component }> }> = [
  { label: 'Commerce', sections: [
    { value: 'catalog', label: 'Catalog', icon: PhBooks },
    { value: 'coupons', label: 'Coupons', icon: PhGift },
    { value: 'payments', label: 'Payments', icon: PhCreditCard },
  ] },
  { label: 'Community', sections: [
    { value: 'activity', label: 'Activity', icon: PhGameController },
    { value: 'questionnaires', label: 'Questionnaires', icon: PhListChecks },
    { value: 'emby', label: 'Emby', icon: PhMonitorPlay },
  ] },
  { label: 'Accounts', sections: [
    { value: 'users', label: 'Users', icon: PhUsers },
    { value: 'entitlements', label: 'Entitlements', icon: PhStack },
  ] },
  { label: 'System', sections: [
    { value: 'settings', label: 'Settings', icon: PhGear },
    { value: 'backups', label: 'Backups', icon: PhArchive },
    { value: 'database', label: 'Database', icon: PhDatabase },
    { value: 'audit', label: 'Audit', icon: PhListMagnifyingGlass },
  ] },
]
</script>

<template>
  <div class="admin-shell">
    <header class="page-header admin-shell__header">
      <p class="eyebrow">Control room</p>
      <h1>TX administration.</h1>
      <p>Configuration and account changes are validated, logged, and reversible where possible.</p>
      <div v-if="!sessionStore.onboardingComplete" class="button-row">
        <RouterLink class="button button--secondary" to="/onboarding">
          <PhUserPlus :size="19" weight="bold" />
          Set up user account
        </RouterLink>
      </div>
    </header>
    <nav class="admin-tabs" aria-label="Admin sections">
      <div v-for="group in groups" :key="group.label" class="admin-tabs__group">
        <span class="admin-tabs__label">{{ group.label }}</span>
        <RouterLink
          v-for="section in group.sections"
          :key="section.value"
          :to="`/admin/${section.value}`"
          class="admin-tab"
          :class="{ 'admin-tab--active': route.params.section === section.value }"
        >
          <component :is="section.icon" :size="18" />
          {{ section.label }}
        </RouterLink>
      </div>
    </nav>
    <div class="admin-shell__content">
      <slot />
    </div>
  </div>
</template>
