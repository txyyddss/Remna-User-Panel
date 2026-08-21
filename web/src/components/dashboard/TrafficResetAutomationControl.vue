<script setup lang="ts">
import InlineNotice from '@/components/common/InlineNotice.vue'

defineProps<{
  enabled: boolean | null
  loading: boolean
  saving: boolean
  error: string | null
}>()

defineEmits<{ update: [enabled: boolean] }>()
</script>

<template>
  <section class="reset-automation">
    <div class="reset-automation__copy">
      <strong>{{ $t('purchaseOperations.automation.label') }}</strong>
      <p>{{ $t('purchaseOperations.automation.description') }}</p>
    </div>
    <USwitch
      :model-value="enabled ?? false"
      :loading="loading || saving"
      :disabled="loading || saving || enabled === null"
      :aria-label="$t('purchaseOperations.automation.label')"
      @update:model-value="$emit('update', $event)"
    />
    <InlineNotice v-if="error" class="reset-automation__error" tone="warning">{{ error }}</InlineNotice>
  </section>
</template>

<style scoped>
.reset-automation { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: center; gap: 0.65rem; padding: 0.7rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.reset-automation__copy { min-width: 0; display: grid; gap: 0.2rem; }
.reset-automation__copy strong { font-size: 0.76rem; }
.reset-automation__copy p { margin: 0; color: var(--text-faint); font-size: 0.68rem; line-height: 1.45; }
.reset-automation__error { grid-column: 1 / -1; }
</style>
