<script setup lang="ts">
import { computed } from 'vue'
import { en, zh_cn } from '@nuxt/ui/locale'
import SettingsPage from '@/components/settings/SettingsPage.vue'
import PurchaseActions from '@/components/dashboard/PurchaseActions.vue'
import type { Purchase } from '@/api/types'
import { useI18n } from '@/i18n'
const { locale } = useI18n()
const uiLocale = computed(() => locale.value === 'zh-CN' ? zh_cn : en)
defineProps<{ purchase: Purchase }>()
const refundMode = new globalThis.URLSearchParams(globalThis.location.search).has('refund')
</script>
<template><UApp :locale="uiLocale"><PurchaseActions v-if="refundMode" :purchase="purchase" /><SettingsPage v-else /></UApp></template>
