<script setup lang="ts">
import { computed } from 'vue'
import { en, zh_cn } from '@nuxt/ui/locale'
import SettingsPage from '@/components/settings/SettingsPage.vue'
import { useI18n } from '@/i18n'
const { locale } = useI18n()
const uiLocale = computed(() => locale.value === 'zh-CN' ? zh_cn : en)
</script>
<template><UApp :locale="uiLocale"><SettingsPage /></UApp></template>
