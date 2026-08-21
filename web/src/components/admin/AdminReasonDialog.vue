<script setup lang="ts">
import { computed } from 'vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import { useTelegramProtection } from '@/composables/useTelegramProtection'
import { useI18n } from '@/i18n'

const open = defineModel<boolean>('open', { required: true })
const reason = defineModel<string>('reason', { required: true })

const props = withDefaults(defineProps<{
  title: string
  description: string
  confirmLabel: string
  busy?: boolean
  danger?: boolean
  error?: string | null
}>(), { busy: false, danger: false, error: null })

defineEmits<{ confirm: [] }>()
const { t } = useI18n()
useTelegramProtection(computed(() => open.value && (props.danger || props.busy || reason.value.trim() !== '')))
</script>

<template>
  <UModal v-model:open="open" :title="title" :description="description" :dismissible="!busy" :ui="{ footer: 'justify-end' }">
    <template #body>
      <UFormField name="reason" :label="t('adminReason.reason')" required>
        <UTextarea v-model.trim="reason" :rows="3" :minlength="4" :maxlength="300" :placeholder="t('adminReason.placeholder')" />
      </UFormField>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    </template>
    <template #footer="{ close }">
      <UButton color="neutral" variant="outline" :label="t('common.cancel')" :disabled="busy" data-haptic="dismiss" @click="close" />
      <UButton :color="danger ? 'error' : 'primary'" trailing-icon="i-ph-arrow-right" :disabled="reason.length < 4 || busy" :loading="busy" :label="busy ? t('adminReason.working') : confirmLabel" :data-haptic="danger ? 'destructive' : 'confirm'" @click="$emit('confirm')" />
    </template>
  </UModal>
</template>
