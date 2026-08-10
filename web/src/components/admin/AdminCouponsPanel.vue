<script setup lang="ts">
import { computed, onMounted, reactive, shallowRef } from 'vue'

import { featuresApi, type CouponDefinition } from '@/api/features'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { localizedError, useI18n } from '@/i18n'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'

const items = shallowRef<CouponDefinition[]>([])
const loading = shallowRef(true)
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)
const editingId = shallowRef<string | null | undefined>(undefined)
const deactivating = shallowRef<CouponDefinition | null>(null)
const draft = reactive({
  code: '', name: '', kind: 'purchase_once' as CouponDefinition['kind'], discountMode: 'percent' as 'fixed' | 'percent',
  valueTxb: '5.00', percent: 10, factor: 2, capTxb: '', globalLimit: '', perUserLimit: '1',
  eligibleComboIds: '', eligibleSquadIds: '', expiresAt: '', active: true,
})
const { t } = useI18n()
const kindItems = computed(() => [
  { value: 'purchase_once', label: t('adminCoupons.oneTime') },
  { value: 'purchase_recurring', label: t('adminCoupons.recurring') },
  { value: 'balance_add', label: t('adminCoupons.balanceAdd') },
  { value: 'balance_multiply', label: t('adminCoupons.balanceMultiply') },
])
const discountItems = computed(() => [
  { value: 'percent', label: t('adminCoupons.percent') },
  { value: 'fixed', label: t('adminCoupons.fixed') },
])

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try { items.value = (await featuresApi.getAdminCoupons()).items }
  catch (caught) { error.value = localizedError(caught, 'adminCoupons.loadFailed') }
  finally { loading.value = false }
}

function edit(coupon?: CouponDefinition): void {
  editingId.value = coupon?.id ?? null
  Object.assign(draft, coupon ? {
    code: coupon.code, name: coupon.name, kind: coupon.kind, discountMode: coupon.discountMode ?? 'fixed',
    valueTxb: ['purchase_once', 'purchase_recurring'].includes(coupon.kind) && coupon.discountMode === 'percent' || coupon.kind === 'balance_multiply' ? '5.00' : txbInputFromMinor(coupon.valueMinorOrBps),
    percent: coupon.discountMode === 'percent' ? Number(coupon.valueMinorOrBps) / 100 : 10,
    factor: coupon.kind === 'balance_multiply' ? Number(coupon.valueMinorOrBps) / 10000 : 2,
    capTxb: coupon.percentCapMinor ? txbInputFromMinor(coupon.percentCapMinor) : '',
    globalLimit: coupon.globalUseLimit === null ? '' : String(coupon.globalUseLimit),
    perUserLimit: coupon.perUserUseLimit === null ? '' : String(coupon.perUserUseLimit),
    eligibleComboIds: coupon.eligibleComboIds.join(', '), eligibleSquadIds: coupon.eligibleSquadIds.join(', '),
    expiresAt: coupon.expiresAt?.slice(0, 10) ?? '', active: coupon.active,
  } : {
    code: '', name: '', kind: 'purchase_once', discountMode: 'percent', valueTxb: '5.00', percent: 10,
    factor: 2, capTxb: '', globalLimit: '', perUserLimit: '1', eligibleComboIds: '', eligibleSquadIds: '', expiresAt: '', active: true,
  })
}

async function save(): Promise<void> {
  busy.value = true
  error.value = null
  try {
    const purchase = draft.kind === 'purchase_once' || draft.kind === 'purchase_recurring'
    const percentage = purchase && draft.discountMode === 'percent'
    const valueMinorOrBps = percentage ? String(Math.round(draft.percent * 100))
      : draft.kind === 'balance_multiply' ? String(Math.round(draft.factor * 10000)) : moneyFromTxbInput(draft.valueTxb)
    if (!valueMinorOrBps) return
    await featuresApi.saveAdminCoupon(editingId.value ?? null, {
      code: draft.code.trim().toUpperCase(), name: draft.name, kind: draft.kind,
      discountMode: purchase ? draft.discountMode : undefined, valueMinorOrBps,
      percentCapMinor: percentage && draft.capTxb ? moneyFromTxbInput(draft.capTxb) : null,
      eligibleComboIds: purchase ? draft.eligibleComboIds.split(',').map((value) => value.trim()).filter(Boolean) : [],
      eligibleSquadIds: purchase ? draft.eligibleSquadIds.split(',').map((value) => value.trim()).filter(Boolean) : [],
      expiresAt: draft.expiresAt ? new Date(`${draft.expiresAt}T23:59:59Z`).toISOString() : null,
      active: draft.active, globalUseLimit: draft.globalLimit ? Number(draft.globalLimit) : null,
      perUserUseLimit: draft.perUserLimit ? Number(draft.perUserLimit) : null,
    })
    editingId.value = undefined
    await load()
  } catch (caught) { error.value = localizedError(caught, 'adminCoupons.saveFailed') }
  finally { busy.value = false }
}

async function deactivate(): Promise<void> {
  if (!deactivating.value) return
  busy.value = true
  error.value = null
  try {
    await featuresApi.deactivateAdminCoupon(deactivating.value.id)
    deactivating.value = null
    await load()
  } catch (caught) { error.value = localizedError(caught, 'adminCoupons.deactivateFailed') }
  finally { busy.value = false }
}

function kindLabel(kind: CouponDefinition['kind']): string {
  return t(`adminCoupons.${kind === 'purchase_once' ? 'oneTime' : kind === 'purchase_recurring' ? 'recurring' : kind === 'balance_add' ? 'balanceAdd' : 'balanceMultiply'}`)
}

onMounted(() => void load())
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading"><div><h2>{{ t('adminCoupons.title') }}</h2><p>{{ t('adminCoupons.copy') }}</p></div><UButton icon="i-ph-plus" :label="t('adminCoupons.new')" @click="edit()" /></div>
    <form v-if="editingId !== undefined" class="catalog-editor" @submit.prevent="save">
      <div class="catalog-editor__heading"><h3>{{ editingId ? t('adminCoupons.edit') : t('adminCoupons.definition') }}</h3></div>
      <UFormField name="coupon-code" :label="t('adminCoupons.code')" required><UInput v-model.trim="draft.code" class="w-full" pattern="[A-Za-z0-9_-]+" :maxlength="40" /></UFormField>
      <UFormField name="coupon-name" :label="t('adminCoupons.name')" required><UInput v-model.trim="draft.name" class="w-full" :maxlength="80" /></UFormField>
      <UFormField name="coupon-kind" :label="t('adminCoupons.kind')" :hint="editingId ? t('adminCoupons.financialHint') : undefined"><USelect v-model="draft.kind" class="w-full" :items="kindItems" :disabled="Boolean(editingId)" /></UFormField>
      <UFormField v-if="draft.kind === 'purchase_once' || draft.kind === 'purchase_recurring'" name="discount-mode" :label="t('adminCoupons.discountMode')"><USelect v-model="draft.discountMode" class="w-full" :items="discountItems" :disabled="Boolean(editingId)" /></UFormField>
      <UFormField v-if="(draft.kind === 'purchase_once' || draft.kind === 'purchase_recurring') && draft.discountMode === 'percent'" name="discount-percent" :label="t('adminCoupons.discountPercent')"><UInput v-model.number="draft.percent" class="w-full" type="number" :min="0.01" :max="100" :step="0.01" /></UFormField>
      <UFormField v-if="draft.kind === 'balance_multiply'" name="balance-multiplier" :label="t('adminCoupons.balanceMultiplier')"><UInput v-model.number="draft.factor" class="w-full" type="number" :min="1.01" :max="100" :step="0.01" /></UFormField>
      <TxbAmountField v-if="draft.kind === 'balance_add' || ((draft.kind === 'purchase_once' || draft.kind === 'purchase_recurring') && draft.discountMode === 'fixed')" id="coupon-value" v-model="draft.valueTxb" :label="t('adminCoupons.txbValue')" min-minor="1" required />
      <TxbAmountField v-if="(draft.kind === 'purchase_once' || draft.kind === 'purchase_recurring') && draft.discountMode === 'percent'" id="coupon-cap" v-model="draft.capTxb" :label="t('adminCoupons.cap')" :hint="t('adminCoupons.capHint')" />
      <UFormField name="combo-ids" :label="t('adminCoupons.comboIds')"><UInput v-model="draft.eligibleComboIds" class="w-full" :disabled="draft.kind === 'balance_add' || draft.kind === 'balance_multiply'" /></UFormField>
      <UFormField name="squad-ids" :label="t('adminCoupons.squadIds')"><UInput v-model="draft.eligibleSquadIds" class="w-full" :disabled="draft.kind === 'balance_add' || draft.kind === 'balance_multiply'" /></UFormField>
      <UFormField name="coupon-expiry" :label="t('adminCoupons.expires')"><UInput v-model="draft.expiresAt" class="w-full" type="date" /></UFormField>
      <UFormField name="global-limit" :label="t('adminCoupons.globalLimit')"><UInput v-model="draft.globalLimit" class="w-full" inputmode="numeric" pattern="[0-9]*" /></UFormField>
      <UFormField name="user-limit" :label="t('adminCoupons.userLimit')"><UInput v-model="draft.perUserLimit" class="w-full" inputmode="numeric" pattern="[0-9]*" /></UFormField>
      <SwitchField id="coupon-active" v-model="draft.active" :label="t('common.active')" />
      <div class="button-row"><UButton color="neutral" variant="outline" :label="t('common.cancel')" @click="editingId = undefined" /><UButton type="submit" :loading="busy" :disabled="busy" :label="busy ? t('common.saving') : editingId ? t('adminCoupons.save') : t('adminCoupons.create')" /></div>
    </form>
    <UAlert v-if="error" class="admin-error" color="warning" variant="soft" icon="i-ph-warning" :description="error" />
    <USkeleton v-if="loading" class="m-4 h-24" />
    <div v-else v-auto-animate class="admin-list">
      <article v-for="coupon in items" :key="coupon.id" class="admin-list-row"><span class="feature-icon feature-icon--small"><UIcon name="i-ph-ticket" /></span><div><strong>{{ coupon.code }} · {{ coupon.name }}</strong><small>{{ t('adminCoupons.summary', { kind: kindLabel(coupon.kind), status: coupon.active ? t('common.active') : t('adminCatalog.paused'), uses: coupon.usageCount, limit: coupon.globalUseLimit ?? t('adminCoupons.unlimited') }) }}</small></div><div class="row-actions"><UButton color="neutral" variant="ghost" square icon="i-ph-pencil-simple" :aria-label="t('adminCoupons.editNamed', { code: coupon.code })" @click="edit(coupon)" /><UButton v-if="coupon.active" size="sm" color="error" variant="ghost" icon="i-ph-pause" :disabled="busy" :label="t('adminCoupons.deactivate')" @click="deactivating = coupon" /></div></article>
      <div v-if="!items.length" class="empty-inline"><div><h3>{{ t('adminCoupons.none') }}</h3><p>{{ t('adminCoupons.noneHint') }}</p></div></div>
    </div>
    <ConfirmDialog :open="Boolean(deactivating)" :title="t('adminCoupons.deactivate')" :description="t('adminCoupons.deactivateConfirm', { code: deactivating?.code ?? '' })" :confirm-label="t('adminCoupons.deactivate')" :busy="busy" danger @update:open="!$event && (deactivating = null)" @confirm="deactivate" />
  </section>
</template>

<style scoped>.admin-error { margin: 1rem; }</style>
