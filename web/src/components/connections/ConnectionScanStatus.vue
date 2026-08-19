<script setup lang="ts">
import { computed } from 'vue'

import { t } from '@/i18n'

interface Props {
  progressPercent: number
  starting: boolean
}

const props = defineProps<Props>()

const displayPercent = computed(() => Math.min(100, Math.max(0, Math.round(props.progressPercent))))
const progressValue = computed(() => props.starting ? null : displayPercent.value)
const descriptionKey = computed(() => props.starting ? 'connections.startingDescription' : 'connections.scanDescription')

function progressValueText(value: number | null | undefined): string {
  return t('connections.progress', { percent: Math.round(value ?? 0) })
}

function progressValueLabel(): string {
  return t('connections.scanProgress')
}
</script>

<template>
  <section class="connection-scan" aria-live="polite" aria-busy="true">
    <div class="connection-scan__visual" aria-hidden="true">
      <div class="connection-radar">
        <span class="connection-radar__ring connection-radar__ring--inner" />
        <span class="connection-radar__ring connection-radar__ring--outer" />
        <span class="connection-radar__sweep" />
        <span class="connection-radar__device connection-radar__device--phone">
          <UIcon name="i-ph-device-mobile" />
        </span>
        <span class="connection-radar__device connection-radar__device--laptop">
          <UIcon name="i-ph-laptop" />
        </span>
        <span class="connection-radar__device connection-radar__device--desktop">
          <UIcon name="i-ph-desktop" />
        </span>
        <span class="connection-radar__core"><UIcon name="i-ph-broadcast" /></span>
      </div>
    </div>

    <div class="connection-scan__content">
      <p class="connection-scan__status">
        <span />{{ $t('connections.scanStatus') }}
      </p>
      <h2>{{ $t('connections.scanning') }}</h2>
      <p class="connection-scan__description">{{ $t(descriptionKey) }}</p>

      <div class="connection-scan__progress">
        <div class="connection-scan__progress-meta">
          <span>{{ $t('connections.scanProgress') }}</span>
          <strong v-if="!starting">{{ displayPercent }}%</strong>
          <UIcon v-else name="i-ph-spinner-gap" class="connection-scan__spinner" aria-hidden="true" />
        </div>
        <UProgress
          :model-value="progressValue"
          :max="100"
          size="xs"
          animation="swing"
          :get-value-label="progressValueLabel"
          :get-value-text="progressValueText"
        />
      </div>
    </div>
  </section>
</template>
