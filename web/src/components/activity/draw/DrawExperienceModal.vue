<script setup lang="ts">
import { computed, defineAsyncComponent, onErrorCaptured, shallowRef, watch } from 'vue'

import type { ActivityResult, LuckyDraw } from '@/api/features'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { useTelegramProtection } from '@/composables/useTelegramProtection'
import ActivityResultContent from '../ActivityResultContent.vue'
import DrawCelebration from './DrawCelebration.vue'
import SimplePresenter from './presenters/SimplePresenter.vue'
import type { DrawPhase, DrawStyle } from './selection'

const props = defineProps<{
  draw: LuckyDraw | null
  style: DrawStyle
  phase: DrawPhase
  result: ActivityResult | null
  error: string | null
}>()
const emit = defineEmits<{ retry: []; reveal: []; close: [] }>()

const presenters = {
  simple: SimplePresenter,
  wheel: defineAsyncComponent(() => import('./presenters/WheelPresenter.vue')),
  grid: defineAsyncComponent(() => import('./presenters/GridPresenter.vue')),
  cards: defineAsyncComponent(() => import('./presenters/CardsPresenter.vue')),
  gift: defineAsyncComponent(() => import('./presenters/GiftPresenter.vue')),
  slot: defineAsyncComponent(() => import('./presenters/SlotPresenter.vue')),
  scratch: defineAsyncComponent(() => import('./presenters/ScratchPresenter.vue')),
}
const presenter = computed(() => presenters[props.style])
const presenterFailed = shallowRef(false)
useTelegramProtection(computed(() => Boolean(props.draw) && props.phase === 'running'))
useTelegramBackButton(computed(() => Boolean(props.draw)), () => {
  if (props.phase === 'settling') emit('reveal')
  else dismiss()
})

watch(() => props.draw?.id, () => { presenterFailed.value = false })
watch(() => props.result, (value) => {
  if (value && presenterFailed.value) emit('reveal')
})
onErrorCaptured(() => {
  presenterFailed.value = true
  if (props.result) emit('reveal')
  return false
})

function dismiss(): void {
  if (props.phase === 'receipt' || props.phase === 'error') emit('close')
}
</script>

<template>
  <UModal
    :open="Boolean(draw)"
    :title="phase === 'receipt' ? $t('activity.drawRecorded') : draw?.name"
    :description="phase === 'receipt' ? $t('activity.resultDescription.drawComplete') : $t('activity.drawPresentationHint')"
    :close="false"
    :dismissible="false"
    scrollable
    :ui="{ content: 'tg-modal--windowed draw-experience', header: 'tg-overlay-header--centered', wrapper: 'tg-overlay-copy--centered' }"
    @update:open="!$event && dismiss()"
  >
    <template v-if="draw" #body>
      <div v-if="phase === 'receipt' && result" class="draw-experience__receipt" role="status" aria-live="polite">
        <DrawCelebration :result="result" />
        <ActivityResultContent :result="result" />
      </div>
      <div v-else-if="phase === 'error'" class="draw-experience__message" role="alert">
        <UIcon name="i-ph-warning-circle" aria-hidden="true" />
        <p>{{ error || $t('errors.activityFailed') }}</p>
      </div>
      <div v-else class="draw-experience__stage">
        <component
          :is="presenter"
          v-if="!presenterFailed"
          :key="draw.id + ':' + style"
          :prizes="draw.prizes ?? []"
          :result="result"
          @finished="$emit('reveal')"
        />
        <div v-else class="draw-experience__message" role="status">
          <UIcon name="i-ph-gift" aria-hidden="true" />
          <p>{{ $t('activity.drawing') }}</p>
        </div>
        <p class="draw-experience__note">{{ $t('activity.previewNote') }}</p>
      </div>
    </template>
    <template #footer>
      <UButton v-if="phase === 'running'" block disabled :label="$t('activity.drawing')" />
      <UButton v-else-if="phase === 'settling'" block :label="$t(style === 'scratch' ? 'activity.reveal' : 'activity.skipAnimation')" @click="$emit('reveal')" />
      <div v-else-if="phase === 'error'" class="draw-experience__actions">
        <UButton :label="$t('common.tryAgain')" @click="$emit('retry')" />
        <UButton color="neutral" variant="outline" :label="$t('common.close')" @click="$emit('close')" />
      </div>
      <UButton v-else block :label="$t('common.close')" data-haptic="dismiss" @click="$emit('close')" />
    </template>
  </UModal>
</template>

<style scoped>
.draw-experience__stage { display: grid; gap: 0.6rem; align-items: center; min-height: 16rem; }
.draw-experience__receipt { position: relative; }
.draw-experience__note { margin: 0; color: var(--text-faint); font-size: 0.72rem; line-height: 1.45; text-align: center; }
.draw-experience__message { display: grid; place-items: center; align-content: center; min-height: 13rem; gap: 0.65rem; color: var(--warning); text-align: center; }
.draw-experience__message :deep(svg) { width: 2.5rem; height: 2.5rem; }
.draw-experience__message p { margin: 0; color: var(--text-muted); }
.draw-experience__actions { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem; width: 100%; }
</style>
