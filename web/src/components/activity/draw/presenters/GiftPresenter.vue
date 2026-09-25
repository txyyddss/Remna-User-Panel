<script setup lang="ts">
import { onScopeDispose, shallowRef, watch } from 'vue'

import type { DrawPresenterProps } from '../selection'

const props = defineProps<DrawPresenterProps>()
const emit = defineEmits<{ finished: [] }>()
const opened = shallowRef(false)
let autoOpen: ReturnType<typeof globalThis.setTimeout> | undefined
let finish: ReturnType<typeof globalThis.setTimeout> | undefined

function open(): void {
  if (!props.result || opened.value) return
  opened.value = true
  if (autoOpen) globalThis.clearTimeout(autoOpen)
  finish = globalThis.setTimeout(() => emit('finished'), 1250)
}

watch(() => props.result, (value) => {
  if (value && !opened.value) autoOpen = globalThis.setTimeout(open, 2200)
}, { immediate: true })
onScopeDispose(() => {
  if (autoOpen) globalThis.clearTimeout(autoOpen)
  if (finish) globalThis.clearTimeout(finish)
})
</script>

<template>
  <div class="draw-gift" role="group" :aria-label="$t('activity.drawStyle.gift')">
    <UButton
      class="draw-gift__trigger"
      color="neutral"
      variant="ghost"
      :disabled="!result || opened"
      :aria-label="$t('activity.openGift')"
      @click="open"
    >
      <span class="draw-gift__scene" :class="{ 'draw-gift__scene--opened': opened, 'draw-gift__scene--waiting': !result }" aria-hidden="true">
        <strong v-if="result" class="draw-gift__prize">{{ result.prizeName }}</strong>
        <span class="draw-gift__lid"><span /></span>
        <span class="draw-gift__box"><span /></span>
      </span>
    </UButton>
    <p role="status">{{ $t(opened ? 'activity.giftOpeningHint' : result ? 'activity.giftReadyHint' : 'activity.drawing') }}</p>
  </div>
</template>

<style scoped>
.draw-gift { display: grid; justify-items: center; align-content: center; min-height: 15rem; gap: 0.4rem; }
.draw-gift__trigger { padding: 0; border-radius: var(--radius-panel); }
.draw-gift__trigger:disabled { opacity: 1; }
.draw-gift__scene { position: relative; display: block; width: 12.5rem; height: 11.5rem; }
.draw-gift__box, .draw-gift__lid { position: absolute; left: 50%; display: block; border: 1px solid var(--accent); background: var(--surface-raised); transform: translateX(-50%); }
.draw-gift__box { top: 5.4rem; z-index: 2; width: 7rem; height: 4.9rem; border-radius: 0 0 0.7rem 0.7rem; }
.draw-gift__lid { top: 4.3rem; z-index: 3; width: 7.8rem; height: 1.25rem; border-radius: 0.35rem; transform-origin: 20% 100%; transition: transform 700ms cubic-bezier(0.16, 1, 0.3, 1); }
.draw-gift__box span, .draw-gift__lid span { position: absolute; left: 50%; display: block; width: 1rem; height: 100%; background: var(--warning); opacity: 0.85; transform: translateX(-50%); }
.draw-gift__prize { position: absolute; top: 4rem; left: 5%; right: 5%; z-index: 4; display: grid; place-items: center; min-height: 2.7rem; padding: 0.35rem; border: 1px solid var(--line-strong); border-radius: 0.5rem; background: var(--surface-raised); color: var(--text); font-size: 0.9rem; line-height: 1.25; text-align: center; overflow-wrap: anywhere; opacity: 0; transform: translateY(0.8rem); transition: transform 780ms cubic-bezier(0.16, 1, 0.3, 1) 180ms, opacity 160ms ease 240ms; }
.draw-gift__scene--opened .draw-gift__lid { transform: translate(-75%, -4.5rem) rotate(-24deg); }
.draw-gift__scene--opened .draw-gift__prize { opacity: 1; transform: translateY(-3.5rem); }
.draw-gift__scene--waiting .draw-gift__lid { animation: gift-lid-breathe 1.4s ease-in-out infinite alternate; }
.draw-gift p { margin: 0; color: var(--text-muted); font-size: 0.78rem; text-align: center; }
@keyframes gift-lid-breathe { to { transform: translate(-50%, -0.3rem) rotate(-3deg); } }
@media (prefers-reduced-motion: reduce) { .draw-gift__lid, .draw-gift__prize { transition: none; } .draw-gift__scene--waiting .draw-gift__lid { animation: none; } }
</style>
