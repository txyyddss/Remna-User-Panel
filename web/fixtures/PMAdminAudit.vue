<script setup lang="ts">
import { computed } from 'vue'
import { en, zh_cn } from '@nuxt/ui/locale'
import { useI18n } from '@/i18n'
import AdminPMPanel from '@/components/admin/pm/AdminPMPanel.vue'
const { locale } = useI18n()
const uiLocale = computed(() => locale.value === 'zh-CN' ? zh_cn : en)
</script>
<template><UApp :locale="uiLocale"><AdminPMPanel /></UApp></template>
