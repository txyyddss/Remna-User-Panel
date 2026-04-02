<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const { toasts, dismiss } = useToast()
</script>

<template>
  <teleport to="body">
    <transition-group name="toast-slide" tag="div" class="toast-container">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="toast-item"
        :class="'toast-' + toast.type"
        @click="dismiss(toast.id)"
      >
        <span class="toast-icon">
          {{ toast.type === 'success' ? '✓' : toast.type === 'error' ? '✕' : 'ℹ' }}
        </span>
        <span class="toast-text">{{ toast.message }}</span>
      </div>
    </transition-group>
  </teleport>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: var(--space-lg);
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  width: calc(100% - 32px);
  max-width: 400px;
  pointer-events: none;
}

.toast-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background: rgba(22, 22, 35, 0.92);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  color: var(--text-primary);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  pointer-events: auto;
  cursor: pointer;
}

.toast-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  font-size: 0.75rem;
  font-weight: 700;
  flex-shrink: 0;
}

.toast-success {
  border-color: rgba(0, 184, 148, 0.4);
}

.toast-success .toast-icon {
  background: rgba(0, 184, 148, 0.2);
  color: var(--accent-success);
}

.toast-error {
  border-color: rgba(255, 107, 107, 0.4);
}

.toast-error .toast-icon {
  background: rgba(255, 107, 107, 0.2);
  color: var(--accent-danger);
}

.toast-info {
  border-color: rgba(108, 92, 231, 0.4);
}

.toast-info .toast-icon {
  background: rgba(108, 92, 231, 0.2);
  color: var(--accent-primary);
}

.toast-text {
  flex: 1;
  line-height: 1.4;
}

.toast-slide-enter-active {
  transition: all 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}

.toast-slide-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.toast-slide-enter-from {
  opacity: 0;
  transform: translateY(-16px) scale(0.95);
}

.toast-slide-leave-to {
  opacity: 0;
  transform: translateY(-8px) scale(0.95);
}
</style>
