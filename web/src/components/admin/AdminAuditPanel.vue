<script setup lang="ts">
import { shallowRef } from 'vue'
import { AnimatePresence, motion } from 'motion-v'

import type { AuditEvent } from '@/api/types'
import { useAdminSection } from '@/composables/useAdminSection'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'
import AdminSectionState from './AdminSectionState.vue'

const { items, loading, error, load } = useAdminSection<AuditEvent>('audit-events')
const action = shallowRef('')
const { t } = useI18n()
const { reducedMotion, offset } = useMotionPreferences()

function detailText(detail: AuditEvent['detail']): string {
  return typeof detail === 'string' ? detail : JSON.stringify(detail)
}

function applyFilter(): void {
  void load({ action: action.value || undefined })
}
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>{{ t('adminAudit.title') }}</h2><p>{{ t('adminAudit.copy') }}</p></div>
      <form class="admin-search" @submit.prevent="applyFilter">
        <UInput v-model.trim="action" icon="i-ph-funnel" :placeholder="t('adminAudit.filter')" :aria-label="t('adminAudit.filterLabel')" />
      </form>
    </div>
    <AdminSectionState :loading="loading" :error="error" @retry="load()">
      <motion.div layout class="timeline-list">
        <AnimatePresence :initial="false" mode="popLayout">
          <motion.article v-for="event in items" :key="event.id" layout class="timeline-row" :initial="reducedMotion ? false : { opacity: 0, y: offset(6) }" :animate="{ opacity: 1, y: 0 }" :exit="{ opacity: 0 }" :transition="{ duration: reducedMotion ? 0.08 : 0.16, ease: 'easeOut' }">
            <span class="timeline-row__icon"><UIcon name="i-ph-shield-check" /></span>
            <div>
              <strong>{{ event.action }}</strong>
              <p>{{ detailText(event.detail) }}</p>
              <small>{{ t('adminAudit.meta', { actor: event.actorUserId ?? t('adminAudit.system'), target: `${event.targetType} ${event.targetId}`, date: formatDateTime(event.createdAt) }) }}</small>
            </div>
          </motion.article>
          <motion.div v-if="!items.length" key="empty" class="empty-inline" :initial="{ opacity: 0 }" :animate="{ opacity: 1 }"><div><h3>{{ t('adminAudit.none') }}</h3><p>{{ t('adminAudit.noneHint') }}</p></div></motion.div>
        </AnimatePresence>
      </motion.div>
    </AdminSectionState>
  </section>
</template>
