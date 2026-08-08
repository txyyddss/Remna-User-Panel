<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import { DialogContent, DialogDescription, DialogOverlay, DialogPortal, DialogRoot, DialogTitle } from 'reka-ui'
import { PhWarning } from '@phosphor-icons/vue'

const props = defineProps<{ open: boolean; backupName: string; busy: boolean }>()
const emit = defineEmits<{ 'update:open': [value: boolean]; restore: [payload: { reason: string; confirmation: string }] }>()
const reason = shallowRef('')
const confirmation = shallowRef('')
const requiredConfirmation = computed(() => `RESTORE ${props.backupName}`)
const canRestore = computed(() => reason.value.trim().length >= 4 && confirmation.value === requiredConfirmation.value)

watch(() => props.open, (open) => {
  if (open) { reason.value = ''; confirmation.value = '' }
})
</script>

<template>
  <DialogRoot :open="open" @update:open="emit('update:open', $event)">
    <DialogPortal to="#overlays">
      <DialogOverlay class="dialog-overlay" />
      <DialogContent class="dialog-content dialog-content--compact">
        <span class="dialog-icon dialog-icon--danger"><PhWarning :size="24" weight="fill" /></span>
        <DialogTitle class="dialog-title">Stage database restore?</DialogTitle>
        <DialogDescription class="dialog-description">A rescue backup is created first. The service then restarts and reconnects on the selected verified snapshot.</DialogDescription>
        <label class="form-stack"><span class="field-label">Reason</span><textarea v-model.trim="reason" rows="3" minlength="4" maxlength="300" required /></label>
        <label class="form-stack"><span class="field-label">Type {{ requiredConfirmation }}</span><input v-model="confirmation" class="compact-select" autocomplete="off" required /></label>
        <div class="dialog-actions"><button class="button button--secondary" type="button" :disabled="busy" @click="emit('update:open', false)">Cancel</button><button class="button button--danger" type="button" :disabled="busy || !canRestore" @click="emit('restore', { reason: reason.trim(), confirmation })">{{ busy ? 'Staging restore' : 'Create rescue backup and restore' }}</button></div>
      </DialogContent>
    </DialogPortal>
  </DialogRoot>
</template>
