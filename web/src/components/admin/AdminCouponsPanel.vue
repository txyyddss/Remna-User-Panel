<script setup lang="ts">
import { onMounted, reactive, shallowRef } from 'vue'
import { PhPause, PhPencilSimple, PhPlus, PhTicket } from '@phosphor-icons/vue'

import type { CouponDefinition } from '@/api/features'
import { featuresApi } from '@/api/features'
import SwitchField from '@/components/common/SwitchField.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'

const items = shallowRef<CouponDefinition[]>([])
const loading = shallowRef(true)
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)
const editingId = shallowRef<string | null | undefined>(undefined)
const draft = reactive({
  code: '', name: '', kind: 'purchase_once' as CouponDefinition['kind'], discountMode: 'percent' as 'fixed' | 'percent',
  valueTxb: '5.00', percent: 10, factor: 2, capTxb: '', globalLimit: '', perUserLimit: '1',
  eligibleComboIds: '', eligibleSquadIds: '', expiresAt: '', active: true,
})

async function load(): Promise<void> {
  loading.value = true
  try { items.value = (await featuresApi.getAdminCoupons()).items } catch (caught) { error.value = caught instanceof Error ? caught.message : 'Coupons could not be loaded.' } finally { loading.value = false }
}

function edit(coupon?: CouponDefinition): void {
  editingId.value = coupon?.id ?? null
  Object.assign(draft, coupon ? {
    code: coupon.code,
    name: coupon.name,
    kind: coupon.kind,
    discountMode: coupon.discountMode ?? 'fixed',
    valueTxb: ['purchase_once', 'purchase_recurring'].includes(coupon.kind) && coupon.discountMode === 'percent' || coupon.kind === 'balance_multiply' ? '5.00' : txbInputFromMinor(coupon.valueMinorOrBps),
    percent: coupon.discountMode === 'percent' ? Number(coupon.valueMinorOrBps) / 100 : 10,
    factor: coupon.kind === 'balance_multiply' ? Number(coupon.valueMinorOrBps) / 10000 : 2,
    capTxb: coupon.percentCapMinor ? txbInputFromMinor(coupon.percentCapMinor) : '',
    globalLimit: coupon.globalUseLimit === null ? '' : String(coupon.globalUseLimit),
    perUserLimit: coupon.perUserUseLimit === null ? '' : String(coupon.perUserUseLimit),
    eligibleComboIds: coupon.eligibleComboIds.join(', '),
    eligibleSquadIds: coupon.eligibleSquadIds.join(', '),
    expiresAt: coupon.expiresAt?.slice(0, 10) ?? '',
    active: coupon.active,
  } : {
    code: '', name: '', kind: 'purchase_once', discountMode: 'percent', valueTxb: '5.00', percent: 10,
    factor: 2, capTxb: '', globalLimit: '', perUserLimit: '1', eligibleComboIds: '', eligibleSquadIds: '', expiresAt: '', active: true,
  })
}

async function save(): Promise<void> {
  busy.value = true
  try {
    const purchase = draft.kind === 'purchase_once' || draft.kind === 'purchase_recurring'
    const percentage = purchase && draft.discountMode === 'percent'
    const valueMinorOrBps = percentage
      ? String(Math.round(draft.percent * 100))
      : draft.kind === 'balance_multiply'
        ? String(Math.round(draft.factor * 10000))
        : moneyFromTxbInput(draft.valueTxb)
    if (!valueMinorOrBps) return
    await featuresApi.saveAdminCoupon(editingId.value ?? null, {
      code: draft.code.trim().toUpperCase(), name: draft.name, kind: draft.kind,
      discountMode: purchase ? draft.discountMode : undefined,
      valueMinorOrBps,
      percentCapMinor: percentage && draft.capTxb ? moneyFromTxbInput(draft.capTxb) : null,
      eligibleComboIds: purchase ? draft.eligibleComboIds.split(',').map((value) => value.trim()).filter(Boolean) : [],
      eligibleSquadIds: purchase ? draft.eligibleSquadIds.split(',').map((value) => value.trim()).filter(Boolean) : [],
      expiresAt: draft.expiresAt ? new Date(`${draft.expiresAt}T23:59:59Z`).toISOString() : null,
      active: draft.active, globalUseLimit: draft.globalLimit ? Number(draft.globalLimit) : null,
      perUserUseLimit: draft.perUserLimit ? Number(draft.perUserLimit) : null,
    })
    editingId.value = undefined
    await load()
  } catch (caught) { error.value = caught instanceof Error ? caught.message : 'The coupon could not be created.' } finally { busy.value = false }
}


async function deactivate(coupon: CouponDefinition): Promise<void> {
  if (!globalThis.confirm(`Deactivate coupon ${coupon.code}? Existing purchase grants will stop being eligible.`)) return
  busy.value = true
  error.value = null
  try {
    await featuresApi.deactivateAdminCoupon(coupon.id)
    await load()
  } catch (caught) {
    error.value = caught instanceof Error ? caught.message : 'The coupon could not be deactivated.'
  } finally {
    busy.value = false
  }
}

onMounted(() => void load())
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading"><div><h2>Coupons</h2><p>One explicit coupon grant per purchase. Discounts never stack.</p></div><button class="button button--primary" type="button" @click="edit()"><PhPlus :size="18" />New coupon</button></div>
    <form v-if="editingId !== undefined" class="catalog-editor" @submit.prevent="save">
      <div class="catalog-editor__heading"><h3>{{ editingId ? 'Edit coupon' : 'Coupon definition' }}</h3></div>
      <label><span>Code</span><input v-model.trim="draft.code" required pattern="[A-Za-z0-9_-]+" maxlength="40" /></label>
      <label><span>Name</span><input v-model.trim="draft.name" required maxlength="80" /></label>
      <label><span>Kind</span><select v-model="draft.kind" :disabled="Boolean(editingId)"><option value="purchase_once">One-time discount</option><option value="purchase_recurring">Recurring discount</option><option value="balance_add">Balance adding</option><option value="balance_multiply">Balance multiplying</option></select><small v-if="editingId" class="field-hint">Create a new definition to change the financial effect.</small></label>
      <label v-if="draft.kind === 'purchase_once' || draft.kind === 'purchase_recurring'"><span>Discount mode</span><select v-model="draft.discountMode" :disabled="Boolean(editingId)"><option value="percent">Percent</option><option value="fixed">Fixed TXB</option></select></label>
      <label v-if="(draft.kind === 'purchase_once' || draft.kind === 'purchase_recurring') && draft.discountMode === 'percent'"><span>Discount, percent</span><input v-model.number="draft.percent" type="number" min="0.01" max="100" step="0.01" /></label>
      <label v-if="draft.kind === 'balance_multiply'"><span>Balance multiplier</span><input v-model.number="draft.factor" type="number" min="1.01" max="100" step="0.01" /></label>
      <TxbAmountField v-if="draft.kind === 'balance_add' || ((draft.kind === 'purchase_once' || draft.kind === 'purchase_recurring') && draft.discountMode === 'fixed')" id="coupon-value" v-model="draft.valueTxb" label="TXB value" min-minor="1" required />
      <TxbAmountField v-if="(draft.kind === 'purchase_once' || draft.kind === 'purchase_recurring') && draft.discountMode === 'percent'" id="coupon-cap" v-model="draft.capTxb" label="Optional percentage cap" hint="Leave blank for no TXB cap." />
      <label><span>Eligible combo IDs, comma separated</span><input v-model="draft.eligibleComboIds" :disabled="draft.kind === 'balance_add' || draft.kind === 'balance_multiply'" /></label>
      <label><span>Eligible squad IDs, comma separated</span><input v-model="draft.eligibleSquadIds" :disabled="draft.kind === 'balance_add' || draft.kind === 'balance_multiply'" /></label>
      <label><span>Expires after, optional</span><input v-model="draft.expiresAt" type="date" /></label>
      <label><span>Global use limit, blank for unlimited</span><input v-model="draft.globalLimit" inputmode="numeric" pattern="[0-9]*" /></label>
      <label><span>Per-user use limit, blank for unlimited</span><input v-model="draft.perUserLimit" inputmode="numeric" pattern="[0-9]*" /></label>
      <SwitchField id="coupon-active" v-model="draft.active" label="Active" />
      <div class="button-row"><button class="button button--secondary" type="button" @click="editingId = undefined">Cancel</button><button class="button button--primary" type="submit" :disabled="busy">{{ busy ? 'Saving' : editingId ? 'Save coupon' : 'Create coupon' }}</button></div>
    </form>
    <p v-if="error" class="field-error admin-error">{{ error }}</p>
    <div v-if="loading" class="admin-loading">Loading coupons</div>
    <div v-else class="admin-list"><article v-for="coupon in items" :key="coupon.id" class="admin-list-row"><span class="feature-icon feature-icon--small"><PhTicket :size="18" /></span><div><strong>{{ coupon.code }} · {{ coupon.name }}</strong><small>{{ coupon.kind.replaceAll('_', ' ') }} · {{ coupon.active ? 'active' : 'paused' }} · {{ coupon.globalUseLimit ?? 'unlimited' }} global limit</small></div><div class="row-actions"><button class="icon-button" type="button" :aria-label="`Edit ${coupon.code}`" @click="edit(coupon)"><PhPencilSimple :size="18" /></button><button v-if="coupon.active" class="button button--ghost-danger button--small" type="button" :disabled="busy" @click="deactivate(coupon)"><PhPause :size="17" />Deactivate</button></div></article><div v-if="!items.length" class="empty-inline"><div><h3>No coupons</h3><p>Create a definition or grant one as a draw prize.</p></div></div></div>
  </section>
</template>

<style scoped>.admin-error { margin: 1rem; }</style>
