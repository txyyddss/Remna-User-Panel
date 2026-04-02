<script setup lang="ts">
import { useConfirm } from '@/composables/useConfirm'

const { visible, title, message, accept, dismiss } = useConfirm()
</script>

<template>
  <teleport to="body">
    <transition name="confirm-fade">
      <div v-if="visible" class="confirm-overlay" @click.self="dismiss">
        <transition name="confirm-slide">
          <div v-if="visible" class="confirm-dialog card">
            <h3 class="confirm-title">{{ title }}</h3>
            <p class="confirm-message">{{ message }}</p>
            <div class="confirm-actions">
              <button class="btn btn-secondary" style="flex: 1" @click="dismiss">Cancel</button>
              <button class="btn btn-danger" style="flex: 1" @click="accept">Confirm</button>
            </div>
          </div>
        </transition>
      </div>
    </transition>
  </teleport>
</template>

<style scoped>
.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(3, 10, 21, 0.72);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9000;
  padding: var(--space-lg);
}

.confirm-dialog {
  width: 100%;
  max-width: 340px;
  text-align: center;
}

.confirm-title {
  margin-bottom: var(--space-sm);
}

.confirm-message {
  color: var(--text-secondary);
  font-size: 0.875rem;
  margin-bottom: var(--space-lg);
  line-height: 1.5;
}

.confirm-actions {
  display: flex;
  gap: var(--space-sm);
}

.confirm-fade-enter-active,
.confirm-fade-leave-active {
  transition: opacity 0.25s ease;
}

.confirm-fade-enter-from,
.confirm-fade-leave-to {
  opacity: 0;
}

.confirm-slide-enter-active {
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.confirm-slide-leave-active {
  transition: all 0.2s ease-in;
}

.confirm-slide-enter-from {
  opacity: 0;
  transform: scale(0.9) translateY(10px);
}

.confirm-slide-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>
