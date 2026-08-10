<script setup lang="ts">
import type { LuckyDraw } from '@/api/features'
import { txbInputFromMinor } from '@/utils/format'

defineProps<{ draws: readonly LuckyDraw[]; busy: boolean }>()
defineEmits<{ draw: [id: string] }>()
</script>

<template>
  <section v-auto-animate class="section-block draw-list">
    <div class="section-heading section-heading--stacked"><h2>{{ $t('activity.luckyDraws') }}</h2><p>{{ $t('activity.drawCopy') }}</p></div>
    <article v-for="draw in draws" :key="draw.id" class="draw-panel">
      <span class="feature-icon"><UIcon name="i-ph-gift" /></span>
      <div class="draw-panel__copy">
        <h3>{{ draw.name }}</h3>
        <p>{{ draw.description || $t('activity.weightedPrize') }}</p>
        <span class="draw-panel__safety"><UIcon name="i-ph-shield-check" /> {{ $t('activity.drawSafety') }}</span>
      </div>
      <UButton
        :disabled="!draw.enabled || busy"
        :loading="busy"
        :label="busy ? $t('activity.drawing') : $t('activity.drawFor', { amount: txbInputFromMinor(draw.feeTxbMinor) })"
        @click="$emit('draw', draw.id)"
      />
    </article>
    <div v-if="!draws.length" class="empty-inline"><div><h3>{{ $t('activity.noDraws') }}</h3><p>{{ $t('activity.publishDraw') }}</p></div></div>
  </section>
</template>

<style scoped>
.draw-panel { display: grid; grid-template-columns: auto minmax(0, 1fr); align-items: start; gap: 0.8rem; }
.draw-list { display: grid; gap: 0.7rem; }
.draw-panel { padding: 0.7rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.draw-panel__copy h3, .draw-panel__copy p { margin: 0; }
.draw-panel__copy h3 { font-size: 1.05rem; }
.draw-panel__copy p { margin-top: 0.35rem; color: var(--text-muted); font-size: 0.82rem; line-height: 1.5; }
.draw-panel__safety { display: flex; gap: 0.4rem; margin-top: 0.65rem; color: var(--text-faint); font-size: 0.66rem; line-height: 1.4; }
.draw-panel :deep(button) { grid-column: 1 / -1; width: 100%; }
@media (min-width: 640px) { .draw-panel { grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; } .draw-panel :deep(button) { grid-column: auto; width: auto; } }
</style>
