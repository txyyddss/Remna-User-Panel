<script setup lang="ts">
import type { Component } from 'vue'
import {
  PhArchive,
  PhBooks,
  PhCreditCard,
  PhGear,
  PhListMagnifyingGlass,
  PhStack,
  PhUsers,
} from '@phosphor-icons/vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()

const sections: Array<{ value: string; label: string; icon: Component }> = [
  { value: 'settings', label: 'Settings', icon: PhGear },
  { value: 'catalog', label: 'Catalog', icon: PhBooks },
  { value: 'users', label: 'Users', icon: PhUsers },
  { value: 'entitlements', label: 'Entitlements', icon: PhStack },
  { value: 'payments', label: 'Payments', icon: PhCreditCard },
  { value: 'backups', label: 'Backups', icon: PhArchive },
  { value: 'audit', label: 'Audit', icon: PhListMagnifyingGlass },
]
</script>

<template>
  <div class="admin-shell">
    <header class="page-header admin-shell__header">
      <p class="eyebrow">Control room</p>
      <h1>TX administration.</h1>
      <p>Configuration and account changes are validated, logged, and reversible where possible.</p>
    </header>
    <nav class="admin-tabs" aria-label="Admin sections">
      <RouterLink
        v-for="section in sections"
        :key="section.value"
        :to="`/admin/${section.value}`"
        class="admin-tab"
        :class="{ 'admin-tab--active': route.params.section === section.value }"
      >
        <component :is="section.icon" :size="18" />
        {{ section.label }}
      </RouterLink>
    </nav>
    <div class="admin-shell__content">
      <slot />
    </div>
  </div>
</template>
