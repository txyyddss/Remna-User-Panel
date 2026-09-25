<script setup lang="ts">
import { Vue3Lottie } from 'vue3-lottie'

import type { DrawPresenterProps } from '../selection'
import giftBox from '../assets/gift-box.json'

defineProps<DrawPresenterProps>()
defineEmits<{ finished: [] }>()
</script>

<template>
  <div class="draw-gift" role="status">
    <Vue3Lottie v-if="result" :key="result.id" :animation-data="giftBox" :loop="1" :height="200" :width="200" @on-complete="$emit('finished')" />
    <div v-else class="draw-gift__waiting"><UIcon name="i-ph-gift" aria-hidden="true" /></div>
    <p>{{ $t(result ? 'activity.drawRecorded' : 'activity.drawing') }}</p>
  </div>
</template>

<style scoped>
.draw-gift { display: grid; place-items: center; align-content: center; min-height: 15rem; }
.draw-gift__waiting { display: grid; place-items: center; width: 12.5rem; height: 12.5rem; color: var(--accent); font-size: 5rem; }
.draw-gift p { margin: 0; color: var(--text-muted); font-size: 0.82rem; }
</style>
