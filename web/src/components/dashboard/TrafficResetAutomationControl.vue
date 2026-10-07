<script setup lang="ts">
import { motion } from 'motion-v'
import { useId } from 'vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { selectionHaptic } from '@/utils/telegram'

defineProps<{
  enabled: boolean | null
  loading: boolean
  saving: boolean
  error: string | null
}>()

const emit = defineEmits<{ update: [enabled: boolean] }>()
const { reducedMotion } = useMotionPreferences()
const id = useId()

function update(enabled: boolean): void {
  selectionHaptic()
  emit('update', enabled)
}
</script>

<template>
  <motion.section class="reset-automation" layout :transition="{ duration: reducedMotion ? 0.08 : 0.16, ease: 'easeOut' }">
    <div class="reset-automation__copy">
      <label :for="id"><strong>{{ $t('purchaseOperations.automation.label') }}</strong></label>
      <p>{{ $t('purchaseOperations.automation.description') }}</p>
    </div>
    <USwitch
      :id="id"
      :model-value="enabled ?? false"
      :loading="loading || saving"
      :disabled="loading || saving || enabled === null"
      :aria-label="$t('purchaseOperations.automation.label')"
      @update:model-value="update"
    />
    <div v-if="error" class="reset-automation__error"><InlineNotice tone="warning">{{ error }}</InlineNotice></div>
  </motion.section>
</template>

<style scoped>
.reset-automation { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: center; gap: 0.65rem; padding: 0.7rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.reset-automation :deep([role="switch"]) { position: relative; }
.reset-automation :deep([role="switch"])::before { content: ''; position: absolute; width: 44px; height: 44px; top: 50%; left: 50%; transform: translate(-50%, -50%); }
.reset-automation__copy { min-width: 0; display: grid; gap: 0.2rem; }
.reset-automation__copy strong { font-size: 0.76rem; }
.reset-automation__copy p { margin: 0; color: var(--text-faint); font-size: 0.68rem; line-height: 1.45; }
.reset-automation__error { grid-column: 1 / -1; }
</style>
