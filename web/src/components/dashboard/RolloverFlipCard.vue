<script setup lang="ts">
import { shallowRef, watch } from 'vue'

import type { Purchase } from '@/api/types'
import RideSummaryFace from './RideSummaryFace.vue'
import RolloverDetailFace from './RolloverDetailFace.vue'
import { useRolloverDetail } from '@/composables/useRolloverDetail'
import { haptic } from '@/utils/telegram'

const props = defineProps<{
  active: Purchase
  squadNames?: readonly string[]
}>()

const flipped = shallowRef(false)
const { detail, loading, error, load, retry, reset } = useRolloverDetail()

function showDetail(): void {
  flipped.value = true
  haptic('light')
  void load(props.active.id)
}

function showSummary(): void {
  flipped.value = false
  haptic('light')
}

watch(() => props.active.id, () => {
  flipped.value = false
  reset()
})
</script>

<template>
  <div class="home-ride__summary home-ride__flip-card" :class="{ 'is-flipped': flipped }">
    <div class="home-ride__flip-inner">
      <div class="home-ride__flip-face home-ride__flip-face--front">
        <UButton
          id="your-ride-summary"
          type="button"
          class="home-ride__summary-trigger"
          block
          color="neutral"
          variant="ghost"
          aria-controls="your-ride-rollover"
          :aria-expanded="flipped"
          :aria-hidden="flipped"
          :inert="flipped"
          data-haptic
          @click="showDetail"
        >
          <RideSummaryFace :active="active" :squad-names="squadNames" />
        </UButton>
      </div>
      <div id="your-ride-rollover" class="home-ride__flip-face home-ride__flip-face--back" :aria-hidden="!flipped" :inert="!flipped">
        <RolloverDetailFace :detail="detail" :loading="loading" :error="error" @back="showSummary" @retry="retry" />
      </div>
    </div>
  </div>
</template>
