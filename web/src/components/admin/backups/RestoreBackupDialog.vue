<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

import { useI18n } from '@/i18n'
import { useTelegramProtection } from '@/composables/useTelegramProtection'

const props = defineProps<{ open: boolean; backupName: string; busy: boolean }>()
const emit = defineEmits<{ 'update:open': [value: boolean]; restore: [payload: { reason: string; confirmation: string }] }>()
const { t } = useI18n()
const reason = shallowRef('')
const confirmation = shallowRef('')
const requiredConfirmation = computed(() => `${t('restoreBackup.confirmationPrefix')} ${props.backupName}`)
const canRestore = computed(() => reason.value.trim().length >= 4 && confirmation.value === requiredConfirmation.value)
useTelegramProtection(computed(() => props.open))

watch(() => props.open, (open) => {
  if (open) { reason.value = ''; confirmation.value = '' }
})
</script>

<template>
  <UModal :open="open" :title="t('restoreBackup.title')" :description="t('restoreBackup.copy')" :dismissible="!busy" :close="false" :ui="{ header: 'tg-overlay-header--centered', wrapper: 'tg-overlay-copy--centered', footer: 'justify-end' }" @update:open="emit('update:open', $event)">
    <template #body>
      <UIcon name="i-ph-warning-fill" class="dialog-icon dialog-icon--danger" aria-hidden="true" />
      <UFormField name="reason" :label="t('adminReason.reason')" required>
        <UTextarea v-model.trim="reason" :rows="3" :minlength="4" :maxlength="300" />
      </UFormField>
      <UFormField name="confirmation" :label="t('databaseRecord.typeConfirmation', { confirmation: requiredConfirmation })" required>
        <UInput v-model="confirmation" autocomplete="off" />
      </UFormField>
    </template>
    <template #footer>
      <UButton color="neutral" variant="outline" :label="t('common.cancel')" :disabled="busy" @click="emit('update:open', false)" />
      <UButton color="error" :disabled="busy || !canRestore" :loading="busy" :label="busy ? t('restoreBackup.staging') : t('restoreBackup.confirm')" data-haptic="destructive" @click="emit('restore', { reason: reason.trim(), confirmation })" />
    </template>
  </UModal>
</template>
