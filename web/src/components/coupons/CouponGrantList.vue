<script setup lang="ts">
import { AnimatePresence, motion } from 'motion-v'

import type { CouponGrant } from '@/api/features'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { formatDate } from '@/utils/format'

const props = withDefaults(defineProps<{
  grants: readonly CouponGrant[]
  discardingId?: string | null
}>(), { discardingId: null })

const emit = defineEmits<{
  discard: [grant: CouponGrant]
}>()
const { reducedMotion, offset } = useMotionPreferences()
</script>

<template>
  <motion.div layout class="coupon-list">
    <AnimatePresence :initial="false" mode="popLayout">
      <motion.article v-for="grant in props.grants" :key="grant.id" layout class="coupon-grant" :initial="reducedMotion ? false : { opacity: 0, y: offset(6) }" :animate="{ opacity: 1, y: 0 }" :exit="{ opacity: 0 }" :transition="{ duration: reducedMotion ? 0.08 : 0.17, ease: 'easeOut' }">
        <span class="feature-icon feature-icon--small"><UIcon name="i-ph-ticket" /></span>
        <div class="coupon-grant__details">
          <strong>{{ grant.coupon.name }}</strong>
          <small>
            {{ $t('coupons.grantSummary', {
              code: grant.coupon.code,
              kind: $t(`coupons.kind.${grant.coupon.kind}`),
              expiry: grant.coupon.expiresAt ? $t('coupons.expires', { date: formatDate(grant.coupon.expiresAt) }) : $t('coupons.noExpiry'),
            }) }}
          </small>
        </div>
        <div class="coupon-grant__actions">
          <span>{{ $t('coupons.uses', { count: grant.coupon.perUserUseLimit === null ? $t('coupons.unlimited') : Math.max(0, grant.coupon.perUserUseLimit - grant.useCount) }) }}</span>
          <UButton
            size="xs"
            color="error"
            variant="ghost"
            icon="i-ph-trash"
            :loading="props.discardingId === grant.id"
            :disabled="props.discardingId !== null"
            :label="$t('coupons.discard')"
            data-haptic="open"
            @click="emit('discard', grant)"
          />
        </div>
      </motion.article>
    </AnimatePresence>
  </motion.div>
</template>

<style scoped>
.coupon-list { display: grid; gap: 0.5rem; }
.coupon-grant { min-height: 58px; display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 0.65rem; padding: 0.55rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.coupon-grant__details strong, .coupon-grant__details small { display: block; }
.coupon-grant__details strong { font-size: 0.78rem; }
.coupon-grant__details small { margin-top: 0.2rem; color: var(--text-faint); font-size: 0.62rem; }
.coupon-grant__actions { display: grid; justify-items: end; gap: 0.2rem; }
.coupon-grant__actions > span { color: var(--accent); font-family: var(--font-mono); font-size: 0.65rem; }
</style>
