<script setup lang="ts">
import { computed } from 'vue'

import { useTelegramProtection } from '@/composables/useTelegramProtection'
import { useI18n } from '@/i18n'

const open = defineModel<boolean>('open', { required: true })

const props = withDefaults(defineProps<{
  title: string
  description: string
  confirmLabel: string
  busy?: boolean
  danger?: boolean
  showClose?: boolean
  centered?: boolean
}>(), {
  busy: false,
  danger: false,
  showClose: true,
  centered: false,
})

defineEmits<{ confirm: [] }>()

const { t } = useI18n()
const modalUi = computed(() => ({
  footer: 'justify-end',
  ...(props.centered ? { header: 'tg-overlay-header--centered', wrapper: 'tg-overlay-copy--centered' } : {}),
}))
useTelegramProtection(computed(() => open.value && (props.danger || props.busy)))
</script>

<template>
  <UModal
    v-model:open="open"
    :description="description"
    :dismissible="!busy"
    :close="showClose"
    :ui="modalUi"
  >
    <template #title>
      <span class="dialog-title">
        <UIcon
          name="i-ph-warning-fill"
          class="dialog-icon"
          :class="{ 'dialog-icon--danger': danger }"
          aria-hidden="true"
        />
        <span>{{ title }}</span>
      </span>
    </template>
    <template #footer="{ close }">
      <UButton
        :label="t('common.cancel')"
        color="neutral"
        variant="outline"
        :disabled="busy"
        data-haptic="dismiss"
        @click="close"
      />
      <UButton
        :label="busy ? t('common.working') : confirmLabel"
        :color="danger ? 'error' : 'primary'"
        :loading="busy"
        :data-haptic="danger ? 'destructive' : 'confirm'"
        @click="$emit('confirm')"
      />
    </template>
  </UModal>
</template>

<style scoped>
.dialog-title { display: inline-flex; align-items: center; gap: 0.55rem; }
.dialog-title .dialog-icon { width: 1.35rem; height: 1.35rem; margin: 0; font-size: 1.35rem; }
</style>
