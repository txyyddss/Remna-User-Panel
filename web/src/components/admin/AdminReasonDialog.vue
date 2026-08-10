<script setup lang="ts">
import { useI18n } from '@/i18n'

const open = defineModel<boolean>('open', { required: true })
const reason = defineModel<string>('reason', { required: true })

withDefaults(defineProps<{
  title: string
  description: string
  confirmLabel: string
  busy?: boolean
  danger?: boolean
}>(), { busy: false, danger: false })

defineEmits<{ confirm: [] }>()
const { t } = useI18n()
</script>

<template>
  <UModal v-model:open="open" :title="title" :description="description" :dismissible="!busy" :ui="{ footer: 'justify-end' }">
    <template #body>
      <UFormField name="reason" :label="t('adminReason.reason')" required>
        <UTextarea v-model.trim="reason" :rows="3" :minlength="4" :maxlength="300" :placeholder="t('adminReason.placeholder')" />
      </UFormField>
    </template>
    <template #footer="{ close }">
      <UButton color="neutral" variant="outline" :label="t('common.cancel')" :disabled="busy" @click="close" />
      <UButton :color="danger ? 'error' : 'primary'" trailing-icon="i-ph-arrow-right" :disabled="reason.length < 4 || busy" :loading="busy" :label="busy ? t('adminReason.working') : confirmLabel" @click="$emit('confirm')" />
    </template>
  </UModal>
</template>
