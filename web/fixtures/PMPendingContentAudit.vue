<script setup lang="ts">
import { computed, onMounted, shallowRef } from 'vue'
import { en, zh_cn } from '@nuxt/ui/locale'
import { useI18n } from '@/i18n'
import type { PMConversation } from '@/api/types'
import PMDeliveryDialog from '@/components/admin/pm/PMDeliveryDialog.vue'
const { locale } = useI18n()
const uiLocale = computed(() => locale.value === 'zh-CN' ? zh_cn : en)
const open = shallowRef(false)
const conversation: PMConversation = { id: 'pending-content-review', userId: 'member', telegramId: '42', name: 'Casey', username: 'casey', chatId: '-100123', topicId: '50', profileMessageId: '51', topicState: 'ready', profileState: 'ready', blocked: false, muted: false, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() }
onMounted(() => { open.value = true })
</script>
<template><UApp :locale="uiLocale"><PMDeliveryDialog v-model:open="open" :conversation="conversation" /></UApp></template>
