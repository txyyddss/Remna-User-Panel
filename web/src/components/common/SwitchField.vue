<script setup lang="ts">
import { SwitchRoot, SwitchThumb } from 'reka-ui'

withDefaults(defineProps<{
  id: string
  label: string
  help?: string
  disabled?: boolean
}>(), {
  help: '',
  disabled: false,
})

const model = defineModel<boolean>({ required: true })
</script>

<template>
  <div class="switch-field">
    <span class="switch-field__copy">
      <label :for="id">{{ label }}</label>
      <small v-if="help" :id="`${id}-help`">{{ help }}</small>
    </span>
    <SwitchRoot
      :id="id"
      v-model="model"
      class="switch-control"
      :disabled="disabled"
      :aria-describedby="help ? `${id}-help` : undefined"
    >
      <SwitchThumb class="switch-thumb" />
    </SwitchRoot>
  </div>
</template>

<style scoped>
.switch-field {
  min-height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.switch-field__copy label,
.switch-field__copy small {
  display: block;
}

.switch-field__copy label {
  color: var(--text-muted);
  font-size: 0.78rem;
  font-weight: 700;
}

.switch-field__copy small {
  margin-top: 0.25rem;
  color: var(--text-faint);
  font-size: 0.68rem;
  line-height: 1.4;
}

.switch-control {
  width: 46px;
  height: 26px;
  flex: 0 0 auto;
  position: relative;
  padding: 2px;
  border: 1px solid var(--line-strong);
  border-radius: 999px;
  background: var(--surface-pressed);
  cursor: pointer;
  transition: border-color 180ms var(--ease-out), background-color 180ms var(--ease-out);
}

.switch-control[data-state='checked'] {
  border-color: var(--accent);
  background: var(--accent);
}

.switch-thumb {
  width: 20px;
  height: 20px;
  display: block;
  border-radius: 50%;
  background: var(--text-muted);
  transform: translateX(0);
  transition: transform 180ms var(--ease-out), background-color 180ms var(--ease-out);
}

.switch-control[data-state='checked'] .switch-thumb {
  background: var(--accent-ink);
  transform: translateX(20px);
}
</style>
