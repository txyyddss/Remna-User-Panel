<script setup lang="ts">
import type { CommunitySpace } from '@/api/types'

const props = defineProps<{
  activeCombo: boolean
  groupJoined: boolean
  channelJoined: boolean
  joining: readonly CommunitySpace[]
}>()

const emit = defineEmits<{ join: [space: CommunitySpace] }>()

const spaces: ReadonlyArray<{ kind: CommunitySpace; icon: string; titleKey: string; descriptionKey: string }> = [
  { kind: 'group', icon: 'i-ph-users-three', titleKey: 'community.groupTitle', descriptionKey: 'community.groupDescription' },
  { kind: 'channel', icon: 'i-ph-broadcast', titleKey: 'community.channelTitle', descriptionKey: 'community.channelDescription' },
]

function joined(space: CommunitySpace): boolean {
  return space === 'group' ? props.groupJoined : props.channelJoined
}

function state(space: CommunitySpace): 'joined' | 'unavailable' | 'ready' {
  if (joined(space)) return 'joined'
  return props.activeCombo ? 'ready' : 'unavailable'
}
</script>

<template>
  <section class="community-rows" :aria-label="$t('community.title')">
    <article v-for="space in spaces" :key="space.kind" class="community-row">
      <span class="community-row__icon" aria-hidden="true"><UIcon :name="space.icon" /></span>
      <div class="community-row__copy">
        <h2>{{ $t(space.titleKey) }}</h2>
        <p>{{ $t(space.descriptionKey) }}</p>
        <small>{{ $t(`community.${state(space.kind)}Description`) }}</small>
      </div>
      <div class="community-row__action">
        <UBadge v-if="state(space.kind) !== 'ready'" :color="state(space.kind) === 'joined' ? 'success' : 'neutral'" variant="subtle">
          {{ $t(`community.${state(space.kind)}`) }}
        </UBadge>
        <UButton
          v-else
          color="primary"
          size="lg"
          class="community-row__join"
          :loading="joining.includes(space.kind)"
          :disabled="joining.includes(space.kind)"
          :label="joining.includes(space.kind) ? $t('community.joining') : $t('community.join')"
          data-haptic="open"
          @click="emit('join', space.kind)"
        />
      </div>
    </article>
  </section>
</template>

<style scoped>
.community-rows { overflow: hidden; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }
.community-row { display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 0.8rem; padding: 1rem; }
.community-row + .community-row { border-top: 1px solid var(--line); }
.community-row__icon { display: grid; width: 2.5rem; height: 2.5rem; place-items: center; border: 1px solid var(--line); border-radius: 50%; color: var(--accent); background: var(--accent-soft); font-size: 1.25rem; }
.community-row__copy { min-width: 0; }
.community-row__copy h2, .community-row__copy p, .community-row__copy small { margin: 0; }
.community-row__copy h2 { font-size: 0.98rem; }
.community-row__copy p { margin-top: 0.2rem; color: var(--text-muted); font-size: 0.78rem; }
.community-row__copy small { display: block; margin-top: 0.45rem; color: var(--text-faint); font-size: 0.69rem; line-height: 1.35; }
.community-row__action { justify-self: end; }
.community-row__join { min-width: 4.8rem; min-height: 44px; }
@media (min-width: 640px) { .community-row { padding: 1.15rem 1.25rem; } }
</style>
