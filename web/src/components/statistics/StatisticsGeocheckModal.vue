<script setup lang="ts">
import { computed, useTemplateRef, watch } from 'vue'

import type { StatisticsNode, StatisticsNodeGeocheck } from '@/api/types'
import { useImageZoom } from '@/composables/useImageZoom'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { t } from '@/i18n'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{
  node: StatisticsNode | null
  result: StatisticsNodeGeocheck | null
  loading: boolean
  error: string | null
}>()

const open = defineModel<boolean>('open', { required: true })
const zoom = useImageZoom()
const canvas = useTemplateRef<globalThis.HTMLElement>('canvas')
const title = computed(() => t('statistics.geocheck.title', { node: props.node?.name ?? '' }))
const imageSource = computed(() => props.result ? `data:${props.result.image.mediaType};${props.result.image.encoding},${props.result.image.data}` : '')
const imageLabel = computed(() => t('statistics.geocheck.canvasLabel', { node: props.node?.name ?? '' }))

function zoomIn(): void { zoom.zoomIn(canvas.value ?? undefined) }
function zoomOut(): void { zoom.zoomOut(canvas.value ?? undefined) }

watch(open, (visible) => { if (!visible) zoom.reset() })
watch(imageSource, () => zoom.reset())
useTelegramBackButton(computed(() => open.value), () => { open.value = false })
</script>

<template>
  <UModal
    v-model:open="open"
    :title="title"
    :description="result ? $t('statistics.geocheck.checkedAt', { date: formatDateTime(result.checkedAt) }) : undefined"
    :close="{ 'data-haptic': '' }"
    :ui="{ content: 'statistics-geocheck-modal', body: 'statistics-geocheck-modal__body' }"
  >
    <template #body>
      <section class="statistics-geocheck">
        <div v-if="loading" class="statistics-geocheck__state" role="status" aria-live="polite">
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
          <div class="statistics-geocheck__controls" role="group" :aria-label="$t('statistics.geocheck.controls')">
            <UTooltip :text="$t('statistics.geocheck.zoomOut')"><UButton color="neutral" variant="ghost" square icon="i-ph-magnifying-glass-minus" :disabled="zoom.scale.value <= 1" :aria-label="$t('statistics.geocheck.zoomOut')" data-haptic @click="zoomOut" /></UTooltip>
            <UTooltip :text="$t('statistics.geocheck.resetZoom')"><UButton color="neutral" variant="ghost" square icon="i-ph-arrow-counter-clockwise" :disabled="!zoom.isZoomed.value" :aria-label="$t('statistics.geocheck.resetZoom')" data-haptic @click="zoom.reset" /></UTooltip>
            <UTooltip :text="$t('statistics.geocheck.zoomIn')"><UButton color="neutral" variant="ghost" square icon="i-ph-magnifying-glass-plus" :disabled="zoom.scale.value >= 4" :aria-label="$t('statistics.geocheck.zoomIn')" data-haptic @click="zoomIn" /></UTooltip>
          </div>
        </template>
        <div v-else class="statistics-geocheck__state" role="status">
          <UIcon name="i-ph-image-broken" aria-hidden="true" />
          <span>{{ $t('statistics.geocheck.unavailable') }}</span>
        </div>
      </section>
    </template>
  </UModal>
</template>
