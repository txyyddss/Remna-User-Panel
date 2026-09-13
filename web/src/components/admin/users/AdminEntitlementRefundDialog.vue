<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import { AnimatePresence, motion } from 'motion-v'

import type { AdminEntitlement } from '@/api/adminOperations'
import InlineNotice from '@/components/common/InlineNotice.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { useTelegramProtection } from '@/composables/useTelegramProtection'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { useI18n } from '@/i18n'
import { moneyFromTxbInput } from '@/utils/format'

const open = defineModel<boolean>('open', { required: true })
const props = withDefaults(defineProps<{ item: AdminEntitlement | null; busy?: boolean; error?: string | null }>(), { busy: false, error: null })
const emit = defineEmits<{ refund: [body: { reason: string; amountTxbMinor: string }] }>()
const { t } = useI18n()
const reason = shallowRef('')
const amount = shallowRef('')
const maximum = computed(() => props.item?.price.minor ?? '0')
const amountMinor = computed(() => moneyFromTxbInput(amount.value))
const valid = computed(() => reason.value.length >= 4 && amountMinor.value !== '' && BigInt(amountMinor.value || '0') <= BigInt(maximum.value))
const workflowState = computed(() => props.busy ? 'processing' : props.error ? `error:${props.error}` : 'review')
const { reducedMotion } = useMotionPreferences()

watch(open, (value) => {
  if (!value || !props.item) return
  reason.value = ''
  amount.value = (Number(BigInt(props.item.price.minor)) / 100).toFixed(2)
})
useTelegramProtection(computed(() => open.value && (props.busy || reason.value.trim() !== '' || amount.value !== '')))
</script>

<template>
  <UModal v-model:open="open" :title="t('adminUserProfile.refundTitle')" :description="t('adminUserProfile.refundHint')" :dismissible="!busy" :close="false" :ui="{ header: 'tg-overlay-header--centered', wrapper: 'tg-overlay-copy--centered', footer: 'justify-end' }">
    <template #body>
      <AnimatePresence mode="wait" :initial="false">
        <motion.div :key="workflowState" class="form-stack" :initial="reducedMotion ? { opacity: 0 } : { opacity: 0, y: 4 }" :animate="{ opacity: 1, y: 0 }" :exit="{ opacity: 0 }" :transition="{ duration: reducedMotion ? 0.08 : 0.18, ease: 'easeOut' }">
          <TxbAmountField id="admin-entitlement-refund" v-model="amount" :label="t('adminUserProfile.refundAmount')" min-minor="1" :max-minor="maximum" required />
          <UFormField name="reason" :label="t('adminReason.reason')" required><UTextarea v-model.trim="reason" :rows="3" :minlength="4" :maxlength="300" :placeholder="t('adminReason.placeholder')" /></UFormField>
          <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
        </motion.div>
      </AnimatePresence>
    </template>
    <template #footer="{ close }">
      <UButton color="neutral" variant="outline" :label="t('common.cancel')" :disabled="busy" data-haptic="dismiss" @click="close" />
      <UButton color="warning" :label="t('adminUserProfile.issueRefund')" :disabled="!valid || busy" :loading="busy" data-haptic="destructive" @click="amountMinor && emit('refund', { reason, amountTxbMinor: amountMinor })" />
    </template>
  </UModal>
</template>
