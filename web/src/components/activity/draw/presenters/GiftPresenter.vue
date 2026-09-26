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
        <span class="draw-gift__lid"><span class="draw-gift__bow"><i /><i /></span></span>
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
.draw-gift__box { top: 5.8rem; z-index: 3; width: 8rem; height: 4.9rem; border-radius: 0.12rem 0.12rem 0.55rem 0.55rem; }
.draw-gift__box span { position: absolute; top: 0; left: 50%; width: 1.05rem; height: 100%; background: var(--warning); transform: translateX(-50%); }
.draw-gift__box::after { position: absolute; top: 42%; left: 0; right: 0; height: 0.9rem; background: var(--warning); content: ''; }
.draw-gift__lid { top: 4.9rem; z-index: 4; width: 8.8rem; height: 1.25rem; border-radius: 0.25rem; transform-origin: 12% 90%; transition: transform 700ms cubic-bezier(0.16, 1, 0.3, 1); }
.draw-gift__lid::after { position: absolute; top: 0; left: 50%; width: 1.05rem; height: 100%; background: var(--warning); content: ''; transform: translateX(-50%); }
.draw-gift__bow { position: absolute; bottom: 100%; left: 50%; display: flex; transform: translateX(-50%); }
.draw-gift__bow i { display: block; width: 1.2rem; height: 0.9rem; border: 0.27rem solid var(--warning); border-radius: 75% 20% 75% 20%; transform: rotate(20deg); }
.draw-gift__bow i + i { transform: scaleX(-1) rotate(20deg); }
.draw-gift__prize { position: absolute; top: 5.8rem; left: 5%; right: 5%; z-index: 2; display: grid; place-items: center; min-height: 2.6rem; padding: 0.35rem; border: 1px solid var(--line-strong); border-radius: 0.45rem; background: var(--canvas); color: var(--text); font-size: 0.9rem; line-height: 1.25; text-align: center; overflow-wrap: anywhere; opacity: 0; transform: translateY(0); transition: transform 900ms cubic-bezier(0.16, 1, 0.3, 1) 180ms, opacity 180ms ease 400ms; }
.draw-gift__scene--opened .draw-gift__lid { z-index: 1; transform: translate(-100%, -2.7rem) rotate(-20deg); }
.draw-gift__scene--opened .draw-gift__prize { opacity: 1; transform: translateY(-5rem); }
.draw-gift__scene--waiting .draw-gift__lid { animation: gift-lid-breathe 1.4s ease-in-out infinite alternate; }
.draw-gift p { margin: 0; color: var(--text-muted); font-size: 0.78rem; text-align: center; }
@keyframes gift-lid-breathe { to { transform: translate(-50%, -0.3rem) rotate(-3deg); } }
@media (prefers-reduced-motion: reduce) { .draw-gift__lid, .draw-gift__prize { transition: none; } .draw-gift__scene--waiting .draw-gift__lid { animation: none; } }
</style>
