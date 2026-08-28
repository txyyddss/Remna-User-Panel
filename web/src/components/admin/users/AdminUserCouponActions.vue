<script setup lang="ts">
import { computed, shallowRef } from 'vue'

import { adminOperationsApi, type AdminUserDetail } from '@/api/adminOperations'
import { featuresApi, type CouponDefinition } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'
import { formatDateTime } from '@/utils/format'
import { notifyHaptic } from '@/utils/telegram'

const props = defineProps<{ userId: string; grants: AdminUserDetail['couponWallet'] }>()
const emit = defineEmits<{ changed: [] }>()
const { t } = useI18n()
const coupons = shallowRef<CouponDefinition[]>([])
const grantOpen = shallowRef(false)
const selectedCouponID = shallowRef('')
const reason = shallowRef('')
const discarding = shallowRef<AdminUserDetail['couponWallet'][number] | null>(null)
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)
const couponItems = computed(() => coupons.value
  .filter((coupon) => coupon.active && (coupon.kind === 'purchase_once' || coupon.kind === 'purchase_recurring'))
  .map((coupon) => ({ label: `${coupon.code} / ${coupon.name}`, value: coupon.id })))

async function openGrant(): Promise<void> {
  error.value = null; reason.value = ''; selectedCouponID.value = ''
  try { coupons.value = (await featuresApi.getAdminCoupons()).items } catch { error.value = t('adminUserProfile.actionFailed') }
  grantOpen.value = true
}

async function grant(): Promise<void> {
  if (!selectedCouponID.value || reason.value.length < 4 || busy.value) return
  busy.value = true
  try {
    await adminOperationsApi.grantUserCoupon(props.userId, selectedCouponID.value, reason.value, createUuid())
    grantOpen.value = false; emit('changed'); notifyHaptic('success')
  } catch { error.value = t('adminUserProfile.actionFailed'); notifyHaptic('error') } finally { busy.value = false }
}

async function discard(): Promise<void> {
  if (!discarding.value || busy.value) return
  busy.value = true
  try {
    await adminOperationsApi.discardUserCoupon(props.userId, discarding.value.id, createUuid())
    discarding.value = null; emit('changed'); notifyHaptic('success')
  } catch { error.value = t('adminUserProfile.actionFailed'); notifyHaptic('error') } finally { busy.value = false }
}
</script>

<template>
  <section class="admin-profile-section">
    <div class="admin-profile-section__heading"><div><h3>{{ t('adminUserProfile.couponWallet') }}</h3><p>{{ t('adminUserProfile.addCouponHint') }}</p></div><UButton color="neutral" variant="ghost" icon="i-ph-plus" :label="t('adminUserProfile.addCoupon')" data-haptic="open" @click="openGrant" /></div>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <div v-if="grants.length" class="admin-profile-list admin-profile-list--compact"><article v-for="grantItem in grants" :key="grantItem.id" class="admin-profile-row"><div class="admin-profile-row__main"><strong>{{ grantItem.coupon.name }}</strong><small>{{ grantItem.coupon.code }} / {{ formatDateTime(grantItem.createdAt) }}</small></div><StatusBadge tone="success" :label="t('adminUserProfile.activeCoupon')" /><UButton color="error" variant="ghost" square icon="i-ph-trash" :aria-label="t('adminUserProfile.removeCoupon')" data-haptic="destructive" @click="discarding = grantItem" /></article></div>
    <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noCoupons') }}</p>
    <UModal v-model:open="grantOpen" :title="t('adminUserProfile.addCoupon')" :description="t('adminUserProfile.addCouponHint')" :dismissible="!busy" :ui="{ footer: 'justify-end' }"><template #body><div class="form-stack"><UFormField name="coupon" :label="t('adminUserProfile.couponWallet')"><USelectMenu v-model="selectedCouponID" :items="couponItems" value-key="value" label-key="label" /></UFormField><UFormField name="reason" :label="t('adminReason.reason')" required><UTextarea v-model.trim="reason" :rows="3" :minlength="4" :maxlength="300" /></UFormField></div></template><template #footer="{ close }"><UButton color="neutral" variant="outline" :label="t('common.cancel')" @click="close" /><UButton color="primary" :label="t('common.confirm')" :loading="busy" :disabled="!selectedCouponID || reason.length < 4" data-haptic="confirm" @click="grant" /></template></UModal>
    <UModal :open="discarding !== null" :title="t('adminUserProfile.removeCoupon')" :description="t('adminUserProfile.removeCouponHint')" :dismissible="!busy" :ui="{ footer: 'justify-end' }" @update:open="!$event && (discarding = null)"><template #footer="{ close }"><UButton color="neutral" variant="outline" :label="t('common.cancel')" @click="close" /><UButton color="error" :label="t('adminUserProfile.removeCoupon')" :loading="busy" data-haptic="destructive" @click="discard" /></template></UModal>
  </section>
</template>
