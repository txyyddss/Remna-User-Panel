<script setup lang="ts">
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
    <USwitch
      :id="id"
      v-model="model"
      class="switch-control"
      :disabled="disabled"
      :aria-describedby="help ? `${id}-help` : undefined"
    />
  </div>
</template>

<style scoped>
.switch-field { min-height: 52px; display: flex; align-items: center; justify-content: space-between; gap: 1rem; }
.switch-field__copy label, .switch-field__copy small { display: block; }
.switch-field__copy label { color: var(--text-muted); font-size: 0.78rem; font-weight: 700; }
.switch-field__copy small { margin-top: 0.25rem; color: var(--text-faint); font-size: 0.68rem; line-height: 1.4; }
.switch-control { flex: 0 0 auto; }
</style>
