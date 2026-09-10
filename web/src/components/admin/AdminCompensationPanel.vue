<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import { AnimatePresence, motion } from 'motion-v'

import type { NodeCompensationEvent } from '@/api/contracts/compensation'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useNodeCompensation } from '@/composables/useNodeCompensation'
import { motionSpring } from '@/composables/motionPresets'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { useI18n } from '@/i18n'
import AdminSectionState from './AdminSectionState.vue'
import CompensationConfigCard from './compensation/CompensationConfigCard.vue'
import CompensationEventCard from './compensation/CompensationEventCard.vue'
import CompensationReviewModal from './compensation/CompensationReviewModal.vue'
import {
  compensationStatusChoices,
  fromCompensationStatus,
  toCompensationStatus,
  type CompensationStatusChoice,
} from './compensation/statusFilter'

const { t } = useI18n()
const state = useNodeCompensation()
const { reducedMotion } = useMotionPreferences()
const reviewOpen = shallowRef(false)
const selected = shallowRef<NodeCompensationEvent | null>(null)
const statusChoice = computed<CompensationStatusChoice>({
  get: () => fromCompensationStatus(state.status.value),
  set: value => void state.changeStatus(toCompensationStatus(value)),
})
const statusItems = computed(() => compensationStatusChoices.map(value => ({
  value, label: t(value === 'all' ? 'adminCompensation.allStatuses' : `adminCompensation.status.${value}`),
})))

function openReview(event: NodeCompensationEvent): void {
  selected.value = event
  reviewOpen.value = true
}

async function review(action: 'approve' | 'dismiss', minutes: number, reason: string): Promise<void> {
  if (!selected.value) return
  if (await state.review(selected.value, action, minutes, reason)) reviewOpen.value = false
}
</script>

<template>
  <section class="admin-panel compensation-panel">
    <div class="admin-panel__heading compensation-panel__heading">
      <div><p class="eyebrow">{{ t('adminCompensation.eyebrow') }}</p><h2>{{ t('adminCompensation.title') }}</h2><p>{{ t('adminCompensation.copy') }}</p></div>
    </div>
    <AdminSectionState :loading="state.loading.value" :error="state.error.value" @retry="state.load">
      <CompensationConfigCard v-if="state.config.value" :config="state.config.value" :busy="state.busy.value" @save="state.saveConfig" />
      <div class="compensation-panel__toolbar">
        <UFormField :label="t('adminCompensation.filter')">
          <USelect v-model="statusChoice" class="w-full" :items="statusItems" value-key="value" />
        </UFormField>
      </div>
      <InlineNotice v-if="state.error.value" tone="warning">{{ state.error.value }}</InlineNotice>
      <div class="compensation-panel__events">
        <AnimatePresence :initial="false" mode="popLayout">
          <motion.div v-for="event in state.events.value" :key="event.id" layout :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: 5 }" :animate="{ opacity: 1, y: 0 }" :exit="{ opacity: 0, y: reducedMotion ? 0 : -4 }" :transition="reducedMotion ? { duration: 0.08 } : motionSpring"><CompensationEventCard :event="event" @review="openReview" /></motion.div>
        </AnimatePresence>
        <div v-if="!state.events.value.length" class="empty-inline"><div><h3>{{ t('adminCompensation.none') }}</h3><p>{{ t('adminCompensation.noneHint') }}</p></div></div>
      </div>
      <UButton v-if="state.nextCursor.value" block color="neutral" variant="outline" icon="i-ph-caret-down" :label="t('adminCompensation.loadMore')" :loading="state.busy.value" @click="state.loadMore" />
    </AdminSectionState>
    <CompensationReviewModal v-model:open="reviewOpen" :event="selected" :busy="state.busy.value" :error="state.error.value" @review="review" />
  </section>
</template>

<style scoped>
.compensation-panel { display: grid; grid-template-columns: minmax(0, 1fr); gap: 1rem; padding-bottom: max(1rem, env(safe-area-inset-bottom)); }
.compensation-panel__heading { align-items: start; justify-items: start; text-align: left; }
.compensation-panel__heading > :only-child { justify-self: start; }
.compensation-panel__toolbar { display: grid; margin-top: 1rem; }
.compensation-panel__events { display: grid; gap: 0.75rem; margin: 0.75rem 0; }
@media (min-width: 700px) { .compensation-panel__toolbar { grid-template-columns: minmax(220px, 0.45fr); } }
</style>
