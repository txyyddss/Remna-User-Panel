<script setup lang="ts">
import { AnimatePresence, motion } from 'motion-v'

import type { AffiliateReferral, AffiliateReferralPage } from '@/api/features'
import { motionSpring } from '@/composables/motionPresets'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { formatDate, formatDateTime } from '@/utils/format'
import { selectionHaptic } from '@/utils/telegram'

const props = defineProps<{ page: AffiliateReferralPage; loading: boolean }>()
const emit = defineEmits<{ 'update:page': [page: number] }>()
const { reducedMotion } = useMotionPreferences()

function referralName(referral: AffiliateReferral): string {
  return [referral.firstName, referral.lastName].map((name) => name.trim()).filter(Boolean).join(' ')
}

function updatePage(page: number): void {
  if (page === props.page.page) return
  selectionHaptic()
  emit('update:page', page)
}
</script>

<template>
  <section class="affiliate-section affiliate-referrals">
    <div class="affiliate-section__heading"><div><span>{{ $t('affiliates.history') }}</span><h2>{{ $t('affiliates.referrals') }}</h2></div><strong>{{ page.total }}</strong></div>
    <div v-if="loading" class="affiliate-referral-list"><USkeleton v-for="index in 5" :key="index" class="h-16 w-full" /></div>
    <div v-else-if="page.items.length" class="affiliate-referral-list">
      <AnimatePresence :initial="false" mode="popLayout">
        <motion.article
          v-for="item in page.items"
          :key="`${item.firstName}-${item.lastName}-${item.registeredAt}`"
          class="affiliate-referral-row"
          layout
          :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: 5 }"
          :animate="{ opacity: 1, y: 0 }"
          :exit="{ opacity: 0, y: reducedMotion ? 0 : -4 }"
          :transition="reducedMotion ? { duration: 0.08 } : motionSpring"
        >
          <div><strong>{{ referralName(item) || $t('affiliates.memberFallback') }}</strong><small>{{ $t('affiliates.joinedOn', { date: formatDate(item.registeredAt) }) }}</small></div>
          <div>
            <UBadge :color="item.status === 'successful' ? 'success' : 'neutral'" variant="soft">{{ $t(`affiliates.status.${item.status}`) }}</UBadge>
            <template v-if="item.status === 'successful'">
              <small v-if="item.paybackAt">{{ formatDateTime(item.paybackAt) }}</small>
              <strong v-if="item.commissionAmount">{{ item.commissionAmount.display }}</strong>
            </template>
          </div>
        </motion.article>
      </AnimatePresence>
    </div>
    <p v-else class="affiliate-empty">{{ $t('affiliates.noReferrals') }}</p>
    <UPagination v-if="page.totalPages > 1" class="affiliate-pagination" :page="page.page" :items-per-page="page.pageSize" :total="page.total" data-haptic-skip @update:page="updatePage" />
  </section>
</template>
