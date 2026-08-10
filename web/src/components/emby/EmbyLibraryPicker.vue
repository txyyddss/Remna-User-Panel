<script setup lang="ts">
import type { EmbyLibrary } from '@/api/features'

const props = defineProps<{ libraries: readonly EmbyLibrary[]; selectedIds: readonly string[]; disabled?: boolean }>()
const emit = defineEmits<{ toggle: [id: string] }>()
</script>

<template>
  <fieldset class="emby-library-picker">
    <legend>{{ $t('emby.disabledLibraries') }}</legend>
    <label v-for="library in libraries" :key="library.id" class="emby-library">
      <span class="feature-icon feature-icon--small"><UIcon name="i-ph-film-slate" /></span>
      <span>{{ library.name }}</span>
      <UCheckbox
        :model-value="props.selectedIds.includes(library.id)"
        :disabled="disabled"
        :aria-label="library.name"
        @update:model-value="emit('toggle', library.id)"
      />
    </label>
    <small class="field-hint">{{ $t('emby.disabledLibrariesHint') }}</small>
  </fieldset>
</template>

<style scoped>
.emby-library-picker { display: grid; gap: 0.5rem; margin: 0; padding: 0; border: 0; }
.emby-library-picker legend { margin-bottom: 0.4rem; color: var(--text-muted); font-size: 0.78rem; font-weight: 700; }
.emby-library { min-height: 54px; display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 0.65rem; padding: 0.5rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); font-size: 0.8rem; font-weight: 700; }
</style>
