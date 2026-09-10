<script setup lang="ts">
import { motion } from 'motion-v'

import CountryFlag from '@/components/common/CountryFlag.vue'
import { motionDurations } from '@/composables/motionPresets'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import type { TrafficUsageSegment } from './TrafficUsageBar.vue'

defineProps<{ segment: TrafficUsageSegment }>()

const { reducedMotion } = useMotionPreferences()
</script>

<template>
  <motion.article
    class="traffic-usage-bar__details"
    role="status"
    aria-live="polite"
    :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: -4 }"
    :animate="{ opacity: 1, y: 0 }"
    :transition="{ duration: reducedMotion ? 0.08 : motionDurations.fast, ease: 'easeOut' }"
  >
    <header class="traffic-usage-bar__header">
      <span class="traffic-usage-bar__swatch" :style="{ backgroundColor: segment.color }" aria-hidden="true" />
      <div>
        <strong>{{ segment.name }}</strong>
        <span v-if="!segment.isOther" class="traffic-usage-bar__country"><CountryFlag :code="segment.countryCode" />{{ segment.countryCode }}</span>
      </div>
    </header>
    <p v-if="segment.isOther" class="traffic-usage-bar__description">{{ $t('home.trafficOtherDescription') }}</p>
    <dl class="traffic-usage-bar__stats">
      <div><dt>{{ $t('home.trafficNodeUsage') }}</dt><dd>{{ segment.bytesLabel }}</dd></div>
      <div><dt>{{ $t('home.trafficNodeShare') }}</dt><dd>{{ segment.percentageLabel }}%</dd></div>
      <div><dt>{{ $t('home.trafficNodeMultiplier') }}</dt><dd>{{ segment.multiplierLabel }}</dd></div>
    </dl>
  </motion.article>
</template>

<style scoped>
.traffic-usage-bar__details { margin-top: 0.65rem; padding: 0.8rem 0.9rem; border: 1px solid var(--line); border-radius: 12px; background: var(--surface-raised); }
.traffic-usage-bar__header { display: flex; align-items: center; gap: 0.65rem; }
.traffic-usage-bar__header strong { display: block; color: var(--text); font-size: 0.88rem; }
.traffic-usage-bar__country { display: inline-flex; align-items: center; gap: 0.2rem; color: var(--text-muted); font-size: 0.74rem; }
.traffic-usage-bar__swatch { width: 0.65rem; height: 0.65rem; flex: 0 0 auto; border-radius: 3px; }
.traffic-usage-bar__description { margin: 0.55rem 0 0; color: var(--text-muted); font-size: 0.76rem; }
.traffic-usage-bar__stats { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.65rem; margin: 0.7rem 0 0; }
.traffic-usage-bar__stats div { min-width: 0; }
.traffic-usage-bar__stats dt { color: var(--text-muted); font-size: 0.68rem; }
.traffic-usage-bar__stats dd { margin: 0.15rem 0 0; color: var(--text); font-size: 0.78rem; font-weight: 700; overflow-wrap: anywhere; }
@media (max-width: 360px) { .traffic-usage-bar__stats { gap: 0.35rem; } .traffic-usage-bar__stats dt, .traffic-usage-bar__stats dd { font-size: 0.66rem; } }
</style>
