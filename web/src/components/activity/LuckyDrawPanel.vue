<script setup lang="ts">
import { AnimatePresence, motion } from 'motion-v'

import type { LuckyDraw } from '@/api/features'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { txbInputFromMinor } from '@/utils/format'

defineProps<{ draws: readonly LuckyDraw[]; busy: boolean }>()
defineEmits<{ draw: [id: string] }>()
const { reducedMotion, offset } = useMotionPreferences()
</script>

<template>
  <motion.section layout class="section-block draw-list">
    <div class="section-heading section-heading--stacked"><h2>{{ $t('activity.luckyDraws') }}</h2><p>{{ $t('activity.drawCopy') }}</p></div>
    <AnimatePresence :initial="false" mode="popLayout">
      <motion.article v-for="draw in draws" :key="draw.id" layout class="draw-panel" :initial="reducedMotion ? false : { opacity: 0, y: offset(8) }" :animate="{ opacity: 1, y: 0 }" :exit="{ opacity: 0, y: reducedMotion ? 0 : -6 }" :transition="{ duration: reducedMotion ? 0.08 : 0.18, ease: 'easeOut' }">
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
          data-haptic="confirm"
          @click="$emit('draw', draw.id)"
        />
      </motion.article>
      <motion.div v-if="!draws.length" key="empty" class="empty-inline" :initial="{ opacity: 0 }" :animate="{ opacity: 1 }" :exit="{ opacity: 0 }"><div><h3>{{ $t('activity.noDraws') }}</h3><p>{{ $t('activity.publishDraw') }}</p></div></motion.div>
    </AnimatePresence>
  </motion.section>
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
