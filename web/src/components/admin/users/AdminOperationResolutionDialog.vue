<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

import type { OperationReceipt, OperationResolution } from '@/api/adminOperations'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useI18n } from '@/i18n'

const props = defineProps<{ open: boolean; operation: OperationReceipt | null; busy: boolean; error?: string | null }>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  resolve: [payload: { resolution: OperationResolution; reason: string }]
}>()
const { t } = useI18n()
const resolution = shallowRef<OperationResolution>('succeeded')
const reason = shallowRef('')
const resolutionItems = computed(() => {
  const values: OperationResolution[] = ['succeeded', 'failed', 'compensated', 'partial']
  return values.filter((value) => props.operation?.status !== 'partial' || value !== 'partial')
    .map((value) => ({ value, label: t(`adminUserProfile.operationStatus.${value}`) }))
})
const canResolve = computed(() => reason.value.trim().length >= 3 && resolutionItems.value.some((item) => item.value === resolution.value))

watch(() => props.open, (open) => {
  if (!open) return
  resolution.value = 'succeeded'
  reason.value = ''
})

function submit(): void {
  if (!canResolve.value) return
  emit('resolve', { resolution: resolution.value, reason: reason.value.trim() })
}
</script>

<template>
  <UModal :open="open" :title="t('adminUserProfile.resolveTitle')" :description="t('adminUserProfile.resolveHint')" :dismissible="!busy" :close="{ 'data-haptic': 'dismiss' }" :ui="{ footer: 'justify-end' }" @update:open="emit('update:open', $event)">
    <template #body>
      <UAlert color="warning" variant="soft" icon="i-ph-warning-circle" :title="t('adminUserProfile.resolveWarning')" :description="t('adminUserProfile.resolveWarningHint')" />
      <UForm id="operation-resolution" :state="{ resolution, reason }" class="form-stack" @submit="submit">
        <UFormField name="resolution" :label="t('adminUserProfile.resolution')" required><USelect v-model="resolution" class="w-full" :items="resolutionItems" value-key="value" /></UFormField>
        <UFormField name="reason" :label="t('adminReason.reason')" required><UTextarea v-model.trim="reason" :rows="3" :minlength="3" :maxlength="500" /></UFormField>
        <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      </UForm>
    </template>
    <template #footer>
      <UButton color="neutral" variant="outline" :label="t('common.cancel')" :disabled="busy" @click="emit('update:open', false)" />
      <UButton type="submit" form="operation-resolution" color="warning" icon="i-ph-gavel" :label="busy ? t('common.working') : t('adminUserProfile.recordResolution')" :loading="busy" :disabled="!canResolve || busy" />
    </template>
  </UModal>
</template>
