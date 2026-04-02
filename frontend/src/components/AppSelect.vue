<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const props = defineProps<{
  modelValue: string | number
  options: { value: string | number; label: string }[]
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string | number): void
}>()

const open = ref(false)
const root = ref<HTMLElement | null>(null)

const selectedLabel = computed(() => {
  const match = props.options.find((o) => o.value === props.modelValue)
  return match?.label || props.placeholder || 'Select...'
})

const isPlaceholder = computed(() => {
  return !props.options.some((o) => o.value === props.modelValue)
})

function toggle() {
  open.value = !open.value
}

function select(value: string | number) {
  emit('update:modelValue', value)
  open.value = false
}

function onClickOutside(e: MouseEvent) {
  if (root.value && !root.value.contains(e.target as Node)) {
    open.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', onClickOutside, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onClickOutside, true)
})
</script>

<template>
  <div ref="root" class="app-select" :class="{ open }">
    <button type="button" class="select-trigger" @click="toggle">
      <span class="select-label" :class="{ placeholder: isPlaceholder }">{{ selectedLabel }}</span>
      <svg class="select-chevron" width="12" height="12" viewBox="0 0 12 12" fill="none">
        <path d="M3 4.5L6 7.5L9 4.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
    </button>
    <transition name="select-drop">
      <div v-if="open" class="select-dropdown">
        <button
          v-for="option in options"
          :key="option.value"
          type="button"
          class="select-option"
          :class="{ selected: option.value === modelValue }"
          @click="select(option.value)"
        >
          {{ option.label }}
          <svg v-if="option.value === modelValue" class="select-check" width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path d="M3 7L6 10L11 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.app-select {
  position: relative;
  width: 100%;
}

.select-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: var(--space-sm) var(--space-md);
  font-family: var(--font-body);
  font-size: 0.875rem;
  background: linear-gradient(180deg, rgba(13, 22, 39, 0.96), rgba(10, 18, 32, 0.96));
  color: var(--text-primary);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
  text-align: left;
  min-height: 38px;
}

.select-trigger:hover {
  border-color: rgba(255, 255, 255, 0.12);
}

.app-select.open .select-trigger {
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px var(--accent-primary-glow);
}

.select-label.placeholder {
  color: var(--text-muted);
}

.select-chevron {
  color: var(--text-secondary);
  transition: transform 0.2s ease;
  flex-shrink: 0;
}

.app-select.open .select-chevron {
  transform: rotate(180deg);
}

.select-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  right: 0;
  background: rgba(18, 18, 30, 0.98);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.5);
  z-index: 500;
  max-height: 220px;
  overflow-y: auto;
  padding: 4px;
}

.select-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: var(--space-sm) var(--space-md);
  font-family: var(--font-body);
  font-size: 0.875rem;
  color: var(--text-secondary);
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.15s;
  text-align: left;
}

.select-option:hover {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-primary);
}

.select-option.selected {
  color: var(--accent-primary);
  background: rgba(108, 92, 231, 0.08);
}

.select-check {
  color: var(--accent-primary);
  flex-shrink: 0;
}

.select-drop-enter-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.select-drop-leave-active {
  transition: all 0.15s ease-in;
}

.select-drop-enter-from {
  opacity: 0;
  transform: translateY(-6px) scaleY(0.95);
}

.select-drop-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.select-dropdown::-webkit-scrollbar {
  width: 4px;
}

.select-dropdown::-webkit-scrollbar-thumb {
  background: var(--bg-tertiary);
  border-radius: 100px;
}
</style>
