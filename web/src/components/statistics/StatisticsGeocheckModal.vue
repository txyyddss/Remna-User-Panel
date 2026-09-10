<script setup lang="ts">
import { computed, watch } from 'vue'
import { AnimatePresence, motion } from 'motion-v'

import type { NodeGeocheckTarget, StatisticsNodeGeocheck } from '@/api/types'
import { useImageZoom } from '@/composables/useImageZoom'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { t } from '@/i18n'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{
  node: NodeGeocheckTarget | null
  result: StatisticsNodeGeocheck | null
  loading: boolean
  error: string | null
}>()

const open = defineModel<boolean>('open', { required: true })
const zoom = useImageZoom()
const title = computed(() => t('statistics.geocheck.title', { node: props.node?.name ?? '' }))
const imageSource = computed(() => props.result ? `data:${props.result.image.mediaType};${props.result.image.encoding},${props.result.image.data}` : '')
const imageLabel = computed(() => t('statistics.geocheck.canvasLabel', { node: props.node?.name ?? '' }))
const stateKey = computed(() => props.node?.geocheckEnabled === false ? 'disabled' : props.loading ? 'loading' : props.error ? `error:${props.error}` : props.result ? `result:${props.result.checkedAt}` : 'unavailable')
const { reducedMotion } = useMotionPreferences()

watch(open, (visible) => { if (!visible) zoom.reset() })
watch(imageSource, () => zoom.reset())
useTelegramBackButton(computed(() => open.value), () => { open.value = false })
</script>

<template>
  <UModal
    v-model:open="open"
    :title="title"
    :description="result ? $t('statistics.geocheck.checkedAt', { date: formatDateTime(result.checkedAt) }) : undefined"
    :close="{ 'data-haptic': 'dismiss' }"
    :ui="{ content: 'statistics-geocheck-modal', body: 'statistics-geocheck-modal__body' }"
  >
    <template #body>
      <section class="statistics-geocheck">
        <AnimatePresence mode="wait" :initial="false">
          <motion.div
            :key="stateKey"
            :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: 4 }"
            :animate="{ opacity: 1, y: 0 }"
            :exit="{ opacity: 0 }"
            :transition="{ duration: reducedMotion ? 0.08 : 0.16, ease: 'easeOut' }"
          >
            <div v-if="node?.geocheckEnabled === false" class="statistics-geocheck__state" role="status">
              <UIcon name="i-ph-eye-slash" aria-hidden="true" />
              <span>{{ $t('statistics.geocheck.disabled') }}</span>
            </div>
            <div v-else-if="loading" class="statistics-geocheck__state" role="status" aria-live="polite">
              <UIcon class="icon-spin" name="i-ph-spinner-gap" aria-hidden="true" />
              <span>{{ $t('statistics.geocheck.loading') }}</span>
            </div>
            <div v-else-if="error" class="statistics-geocheck__state" role="alert">
              <UIcon name="i-ph-warning" aria-hidden="true" />
              <span>{{ error }}</span>
            </div>
            <template v-else-if="result">
              <div
                ref="canvas"
                class="statistics-geocheck__canvas"
                role="img"
                tabindex="0"
                :aria-label="imageLabel"
                @pointerdown="zoom.onPointerDown"
                @pointermove="zoom.onPointerMove"
                @pointerup="zoom.onPointerUp"
                @pointercancel="zoom.onPointerUp"
                @touchstart="zoom.onTouchStart"
                @touchmove="zoom.onTouchMove"
                @touchend="zoom.onTouchEnd"
                @touchcancel="zoom.onTouchEnd"
                @dblclick="zoom.onDoubleClick"
                @wheel.prevent="zoom.onWheel"
              >
                <img
                  class="statistics-geocheck__image"
                  :class="{ 'statistics-geocheck__image--interacting': zoom.isInteracting.value }"
                  :src="imageSource"
                  :alt="imageLabel"
                  :style="zoom.imageStyle.value"
                  draggable="false"
                />
              </div>
            </template>
            <div v-else class="statistics-geocheck__state" role="status">
              <UIcon name="i-ph-image-broken" aria-hidden="true" />
              <span>{{ $t('statistics.geocheck.unavailable') }}</span>
            </div>
          </motion.div>
        </AnimatePresence>
      </section>
    </template>
  </UModal>
</template>
