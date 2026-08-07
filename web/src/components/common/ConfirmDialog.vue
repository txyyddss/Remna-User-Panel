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
import { PhWarning } from '@phosphor-icons/vue'

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
</script>

<template>
  <DialogRoot v-model:open="open">
    <DialogPortal to="#overlays">
      <DialogOverlay class="dialog-overlay" />
      <DialogContent class="dialog-content dialog-content--compact">
        <span class="dialog-icon" :class="{ 'dialog-icon--danger': danger }"><PhWarning :size="24" weight="fill" /></span>
        <DialogTitle class="dialog-title">{{ title }}</DialogTitle>
        <DialogDescription class="dialog-description">{{ description }}</DialogDescription>
        <div class="dialog-actions">
          <DialogClose class="button button--secondary" :disabled="busy">Cancel</DialogClose>
          <button
            class="button"
            :class="danger ? 'button--danger' : 'button--primary'"
            type="button"
            :disabled="busy"
            @click="$emit('confirm')"
          >
            {{ busy ? 'Working' : confirmLabel }}
          </button>
        </div>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>
