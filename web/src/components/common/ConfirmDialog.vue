<script setup lang="ts">
import { useI18n } from '@/i18n'

const open = defineModel<boolean>('open', { required: true })

withDefaults(defineProps<{
  title: string
  description: string
  confirmLabel: string
  busy?: boolean
  danger?: boolean
}>(), {
  busy: false,
  danger: false,
})

defineEmits<{ confirm: [] }>()

const { t } = useI18n()
</script>

<template>
  <UModal
    v-model:open="open"
    :description="description"
    :dismissible="!busy"
    :ui="{ footer: 'justify-end' }"
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
        @click="close"
      />
      <UButton
        :label="busy ? t('common.working') : confirmLabel"
        :color="danger ? 'error' : 'primary'"
        :loading="busy"
        :data-haptic="danger ? 'heavy' : 'light'"
        @click="$emit('confirm')"
      />
    </template>
  </UModal>
</template>

<style scoped>
.dialog-title { display: inline-flex; align-items: center; gap: 0.55rem; }
.dialog-title .dialog-icon { width: 1.35rem; height: 1.35rem; margin: 0; font-size: 1.35rem; }
</style>
