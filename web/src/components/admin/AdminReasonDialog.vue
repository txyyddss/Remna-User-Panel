<script setup lang="ts">
import {
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogOverlay,
  DialogPortal,
  DialogRoot,
  DialogTitle,
} from 'reka-ui'
import { PhArrowRight, PhX } from '@phosphor-icons/vue'
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
  <DialogRoot v-model:open="open">
    <DialogPortal to="#overlays">
      <DialogOverlay class="dialog-overlay" />
      <DialogContent class="dialog-content dialog-content--compact">
        <header class="dialog-header">
          <div><DialogTitle class="dialog-title">{{ title }}</DialogTitle><DialogDescription class="dialog-description">{{ description }}</DialogDescription></div>
          <DialogClose class="icon-button" :aria-label="t('common.close')"><PhX :size="19" /></DialogClose>
        </header>
        <label class="form-stack">
          <span class="field-label">{{ t('adminReason.reason') }}</span>
          <textarea v-model.trim="reason" rows="3" minlength="4" maxlength="300" :placeholder="t('adminReason.placeholder')" />
        </label>
        <button class="button button--wide" :class="danger ? 'button--danger' : 'button--primary'" type="button" :disabled="reason.length < 4 || busy" @click="$emit('confirm')">
          {{ busy ? t('adminReason.working') : confirmLabel }} <PhArrowRight :size="18" />
        </button>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>
