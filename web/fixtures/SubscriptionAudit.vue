<script setup lang="ts">
import { computed } from 'vue'
import { en, zh_cn } from '@nuxt/ui/locale'
import UserHome from '@/components/dashboard/UserHome.vue'
import SubscriptionPage from '@/components/subscription/SubscriptionPage.vue'
import AdminConnectivitySettings from '@/components/admin/connectivity/AdminConnectivitySettings.vue'
import { useI18n } from '@/i18n'
defineProps<{ mode: string }>()
const { locale } = useI18n()
const uiLocale = computed(() => locale.value === 'zh-CN' ? zh_cn : en)
</script>
<template><UApp :locale="uiLocale"><UserHome v-if="mode === 'home'" /><AdminConnectivitySettings v-else-if="mode === 'admin'" /><SubscriptionPage v-else /></UApp></template>
