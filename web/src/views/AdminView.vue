<script setup lang="ts">
import type { Component } from 'vue'
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import AdminAuditPanel from '@/components/admin/AdminAuditPanel.vue'
import AdminBackupsPanel from '@/components/admin/AdminBackupsPanel.vue'
import AdminCatalogPanel from '@/components/admin/AdminCatalogPanel.vue'
import AdminEntitlementsPanel from '@/components/admin/AdminEntitlementsPanel.vue'
import AdminPaymentsPanel from '@/components/admin/AdminPaymentsPanel.vue'
import AdminSettingsPanel from '@/components/admin/AdminSettingsPanel.vue'
import AdminShell from '@/components/admin/AdminShell.vue'
import AdminUsersPanel from '@/components/admin/AdminUsersPanel.vue'

const route = useRoute()
const panels: Record<string, Component> = {
  settings: AdminSettingsPanel,
  catalog: AdminCatalogPanel,
  users: AdminUsersPanel,
  entitlements: AdminEntitlementsPanel,
  payments: AdminPaymentsPanel,
  backups: AdminBackupsPanel,
  audit: AdminAuditPanel,
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
