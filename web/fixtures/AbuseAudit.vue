<script setup lang="ts">
import { en } from '@nuxt/ui/locale'
import { shallowRef } from 'vue'
import type { OperationReceipt } from '@/api/types'
import AdminAbusePanel from '@/components/admin/AdminAbusePanel.vue'
import OperationStatusNotice from '@/components/common/OperationStatusNotice.vue'
import { t } from '@/i18n'
const operationMode = new URLSearchParams(location.search).has('operation')
const receipt = shallowRef<OperationReceipt>({ id: 'audit-operation', kind: 'subscription_revoke', status: 'queued', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() })
const error = shallowRef<string | null>(t('errors.adminAction'))
function refresh(): void { receipt.value = { ...receipt.value, status: 'succeeded' }; error.value = null }
Object.assign(window, { __operationAudit: { receipt, error, refresh } })
</script>
<template>
  <UApp :locale="en"><main id="main-content" style="max-width: 74rem; margin: 0 auto; padding: 1rem;"><OperationStatusNotice v-if="operationMode" :receipt="receipt" :error="error" @refresh="refresh" /><AdminAbusePanel v-else /></main></UApp>
</template>
