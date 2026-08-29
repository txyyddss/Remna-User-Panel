<script setup lang="ts">
import { shallowRef, watch } from 'vue'

import type { SquadProduct } from '@/api/types'

const open = defineModel<boolean>('open', { required: true })
const props = defineProps<{ squad: SquadProduct | null }>()
const emit = defineEmits<{ submit: [code: string]; cancel: [] }>()
const code = shallowRef('')

watch(open, (visible) => {
  if (visible) code.value = ''
})

function submit(): void {
  const value = code.value.trim()
  if (!value) return
  emit('submit', value)
}

function cancel(close: () => void): void {
  emit('cancel')
  close()
}
</script>

<template>
  <UModal v-model:open="open" :title="$t('catalog.activationTitle')" :description="$t('catalog.activationDescription')" :dismissible="false" :close="false" :ui="{ header: 'tg-overlay-header--centered', wrapper: 'tg-overlay-copy--centered' }">
    <template #body>
      <div class="activation-dialog__heading">
        <UIcon name="i-ph-lock-key" aria-hidden="true" />
        <strong>{{ props.squad?.name }}</strong>
      </div>
      <UFormField :label="$t('catalog.activationCodeLabel')" :description="$t('catalog.activationCodeHint')">
        <UInput v-model="code" type="password" autocomplete="one-time-code" inputmode="text" autofocus />
      </UFormField>
    </template>
    <template #footer="{ close }">
      <UButton color="neutral" variant="outline" :label="$t('common.cancel')" data-haptic="dismiss" @click="cancel(close)" />
      <UButton :label="$t('catalog.activationContinue')" :disabled="!code.trim()" data-haptic="confirm" @click="submit" />
    </template>
  </UModal>
</template>

<style scoped>
.activation-dialog__heading { display: flex; align-items: center; gap: 0.55rem; margin-bottom: 1rem; color: var(--text-strong); }
.activation-dialog__heading > :first-child { color: var(--accent); font-size: 1.2rem; }
</style>
