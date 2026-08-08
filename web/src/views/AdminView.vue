<script setup lang="ts">
import type { Component } from 'vue'
import { computed, defineAsyncComponent } from 'vue'
import { useRoute } from 'vue-router'

import AdminSettingsPanel from '@/components/admin/AdminSettingsPanel.vue'
import AdminShell from '@/components/admin/AdminShell.vue'

const route = useRoute()
const panels: Record<string, Component> = {
  settings: AdminSettingsPanel,
  catalog: defineAsyncComponent(() => import('@/components/admin/AdminCatalogPanel.vue')),
  activity: defineAsyncComponent(() => import('@/components/admin/AdminActivityPanel.vue')),
  coupons: defineAsyncComponent(() => import('@/components/admin/AdminCouponsPanel.vue')),
  questionnaires: defineAsyncComponent(() => import('@/components/admin/AdminQuestionnairesPanel.vue')),
  emby: defineAsyncComponent(() => import('@/components/admin/AdminEmbyPanel.vue')),
  users: defineAsyncComponent(() => import('@/components/admin/AdminUsersPanel.vue')),
  entitlements: defineAsyncComponent(() => import('@/components/admin/AdminEntitlementsPanel.vue')),
  payments: defineAsyncComponent(() => import('@/components/admin/AdminPaymentsPanel.vue')),
  backups: defineAsyncComponent(() => import('@/components/admin/AdminBackupsPanel.vue')),
  database: defineAsyncComponent(() => import('@/components/admin/AdminDatabasePanel.vue')),
  audit: defineAsyncComponent(() => import('@/components/admin/AdminAuditPanel.vue')),
}
const panel = computed(() => panels[String(route.params.section)] ?? AdminSettingsPanel)
</script>

<template>
  <div class="page page--admin">
    <AdminShell>
      <component :is="panel" :key="String(route.params.section)" />
    </AdminShell>
  </div>
</template>
