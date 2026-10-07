<script setup lang="ts">
import { useId } from 'vue'
import { selectionHaptic } from '@/utils/telegram'
defineProps<{ label: string; description?: string; enabled: boolean; disabled?: boolean; saving?: boolean }>()
const emit = defineEmits<{ update: [enabled: boolean] }>()
const id = useId()
function update(enabled: boolean): void { selectionHaptic(); emit('update', enabled) }
</script>

<template>
  <div class="settings-row">
    <label :for="id" class="settings-row__copy"><strong>{{ label }}</strong><span v-if="description">{{ description }}</span></label>
    <USwitch :id="id" :model-value="enabled" :disabled="disabled || saving" :loading="saving" :aria-label="label" @update:model-value="update(Boolean($event))" />
  </div>
</template>

<style scoped>
.settings-row { display: flex; justify-content: space-between; align-items: center; gap: 1rem; min-height: 60px; padding: 0.75rem 0; }
.settings-row__copy { min-width: 0; flex: 1; min-height: 44px; display: flex; flex-direction: column; justify-content: center; cursor: pointer; }
.settings-row__copy strong { font-size: 0.875rem; font-weight: 600; }
.settings-row__copy span { margin: 0.25rem 0 0; color: var(--text-muted); font-size: 0.8rem; line-height: 1.45; }
.settings-row :deep([role="switch"]) { position: relative; }
.settings-row :deep([role="switch"])::before { content: ''; position: absolute; width: 44px; height: 44px; top: 50%; left: 50%; transform: translate(-50%, -50%); }
</style>
