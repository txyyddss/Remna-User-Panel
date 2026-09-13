<script setup lang="ts">
import { AnimatePresence, motion } from 'motion-v'

import type { AdminUserDetail } from '@/api/adminOperations'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { motionSpring } from '@/composables/motionPresets'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { t } from '@/i18n'
import { formatDateTime } from '@/utils/format'

type IPBlock = AdminUserDetail['ipBlocks'][number]

defineProps<{ items: readonly IPBlock[]; busy: boolean }>()
const emit = defineEmits<{ unblock: [block: IPBlock] }>()
const { reducedMotion } = useMotionPreferences()
</script>

<template>
  <section class="admin-profile-section admin-ip-blocks">
    <div class="admin-profile-section__heading">
      <div><h3>{{ t('adminUserProfile.ipBlocks') }}</h3><p>{{ t('adminUserProfile.ipBlocksHint') }}</p></div>
    </div>
    <div v-if="items.length" class="admin-profile-list admin-profile-list--compact">
      <AnimatePresence :initial="false" mode="popLayout">
        <motion.div
          v-for="block in items"
          :key="block.id"
          class="admin-profile-row"
          layout
          :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: 5 }"
          :animate="{ opacity: 1, y: 0 }"
          :exit="{ opacity: 0, y: reducedMotion ? 0 : -4 }"
          :transition="reducedMotion ? { duration: 0.08 } : motionSpring"
        >
          <div class="admin-profile-row__main">
            <strong><UIcon name="i-ph-shield-warning" />{{ block.ip }}</strong>
            <small>{{ t('adminUserProfile.ipBlockExpiry', { date: formatDateTime(block.expiresAt) }) }}</small>
          </div>
          <StatusBadge :tone="block.status === 'active' ? 'success' : 'warning'" :label="t(`adminUserProfile.ipBlockStatus.${block.status}`)" />
          <div class="admin-profile-row__actions">
            <UButton color="neutral" variant="outline" icon="i-ph-shield-check" :label="t('adminUserProfile.unblockIP')" :disabled="busy || block.status === 'unblocking'" @click="emit('unblock', block)" />
          </div>
        </motion.div>
      </AnimatePresence>
    </div>
    <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noIPBlocks') }}</p>
  </section>
</template>
