<script setup lang="ts">
import { computed, reactive, shallowRef, watch } from 'vue'

import type { CouponDefinition, LuckyDrawAdmin, LuckyDrawPrize, LuckyDrawWrite, Reward } from '@/api/features'
import SwitchField from '@/components/common/SwitchField.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { useI18n } from '@/i18n'
import { moneyFromTxbInput, signedMoneyFromTxbInput, txbInputFromMinor } from '@/utils/format'

type RewardKind = Reward['kind']
interface PrizeDraft {
  id: string
  name: string
  weight: string
  stockRemaining: string
  kind: RewardKind
  txbDelta: string
  couponId: string
  extensionDays: number
}

const props = defineProps<{ draw: LuckyDrawAdmin | null; coupons: readonly CouponDefinition[]; busy: boolean }>()
const emit = defineEmits<{ save: [value: LuckyDrawWrite]; cancel: [] }>()
const draft = reactive({ name: '', description: '', fee: '0.00', enabled: true, prizes: [] as PrizeDraft[] })
const validationError = shallowRef<string | null>(null)
const { t } = useI18n()
const rewardItems = computed(() => [
  { value: 'none', label: t('adminLuckyDraw.noPrize') },
  { value: 'txb_delta', label: t('adminLuckyDraw.txbChange') },
  { value: 'coupon_grant', label: t('adminLuckyDraw.couponGrant') },
  { value: 'subscription_extension', label: t('adminLuckyDraw.extension') },
])
const couponItems = computed(() => props.coupons.map((coupon) => ({ value: coupon.id, label: `${coupon.code} · ${coupon.name}` })))

function blankPrize(): PrizeDraft {
  return { id: '', name: t('adminLuckyDraw.noPrize'), weight: '1', stockRemaining: '', kind: 'none', txbDelta: '1.00', couponId: '', extensionDays: 1 }
}

function prizeDraft(prize: LuckyDrawPrize): PrizeDraft {
  return {
    id: prize.id,
    name: prize.name,
    weight: prize.weight,
    stockRemaining: prize.stockRemaining == null ? '' : String(prize.stockRemaining),
    kind: prize.reward.kind,
    txbDelta: prize.reward.kind === 'txb_delta' ? txbInputFromMinor(prize.reward.txbDeltaMinor) : '1.00',
    couponId: prize.reward.kind === 'coupon_grant' ? prize.reward.couponId : '',
    extensionDays: prize.reward.kind === 'subscription_extension' ? prize.reward.extensionDays : 1,
  }
}

watch(() => props.draw, (draw) => {
  draft.name = draw?.name ?? ''
  draft.description = draw?.description ?? ''
  draft.fee = txbInputFromMinor(draw?.feeTxbMinor ?? '0')
  draft.enabled = draw?.enabled ?? true
  draft.prizes = draw?.prizes.length ? draw.prizes.map(prizeDraft) : [blankPrize()]
}, { immediate: true })

function rewardFromDraft(prize: PrizeDraft): Reward | null {
  if (prize.kind === 'txb_delta') {
    const txbDeltaMinor = signedMoneyFromTxbInput(prize.txbDelta)
    return txbDeltaMinor && txbDeltaMinor !== '0' ? { kind: 'txb_delta', txbDeltaMinor } : null
  }
  if (prize.kind === 'coupon_grant') return prize.couponId ? { kind: 'coupon_grant', couponId: prize.couponId } : null
  if (prize.kind === 'subscription_extension') {
    return prize.extensionDays > 0 ? { kind: 'subscription_extension', extensionDays: Math.trunc(prize.extensionDays) } : null
  }
  return { kind: 'none' }
}

function save(): void {
  validationError.value = null
  if (!draft.name.trim()) {
    validationError.value = t('adminLuckyDraw.nameRequired')
    return
  }
  const feeTxbMinor = moneyFromTxbInput(draft.fee)
  if (!feeTxbMinor || !draft.prizes.length) {
    validationError.value = t('adminLuckyDraw.feePrizeRequired')
    return
  }
  const prizes: LuckyDrawPrize[] = []
  for (const prize of draft.prizes) {
    const reward = rewardFromDraft(prize)
    if (!reward || !/^\d+$/.test(prize.weight) || BigInt(prize.weight) <= 0n || !prize.name.trim()) {
      validationError.value = t('adminLuckyDraw.prizeInvalid', { index: prizes.length + 1 })
      return
    }
    const stockRemaining = prize.stockRemaining === '' ? null : Number.parseInt(prize.stockRemaining, 10)
    if (stockRemaining !== null && (!Number.isSafeInteger(stockRemaining) || stockRemaining < 0)) {
      validationError.value = t('adminLuckyDraw.stockInvalid', { index: prizes.length + 1 })
      return
    }
    prizes.push({ id: prize.id, name: prize.name.trim(), weight: prize.weight, stockRemaining, reward })
  }
  emit('save', { name: draft.name.trim(), description: draft.description.trim(), enabled: draft.enabled, feeTxbMinor, prizes })
}
</script>

<template>
  <form class="catalog-editor lucky-draw-editor" @submit.prevent="save">
    <div class="catalog-editor__heading">
      <div><h3>{{ draw ? t('adminLuckyDraw.edit') : t('adminLuckyDraw.new') }}</h3><p>{{ t('adminLuckyDraw.copy') }}</p></div>
    </div>
    <UFormField name="draw-name" :label="t('adminLuckyDraw.name')" required><UInput v-model.trim="draft.name" class="w-full" :maxlength="80" /></UFormField>
    <TxbAmountField id="lucky-draw-fee" v-model="draft.fee" :label="t('adminLuckyDraw.fee')" min-minor="0" required />
    <UFormField class="catalog-editor__wide" name="draw-description" :label="t('adminLuckyDraw.description')"><UTextarea v-model.trim="draft.description" class="w-full" :rows="2" :maxlength="300" /></UFormField>
    <SwitchField id="lucky-draw-enabled" v-model="draft.enabled" :label="t('adminLuckyDraw.available')" :help="t('adminLuckyDraw.availableHint')" />

    <section class="prize-list catalog-editor__wide" aria-labelledby="prize-list-title">
      <div class="prize-list__heading"><div><h4 id="prize-list-title">{{ t('adminLuckyDraw.weightedPrizes') }}</h4><p>{{ t('adminLuckyDraw.prizeHint') }}</p></div><UButton color="neutral" variant="outline" icon="i-ph-plus" :label="t('adminLuckyDraw.addPrize')" @click="draft.prizes.push(blankPrize())" /></div>
      <div v-auto-animate class="prize-list__items">
        <article v-for="(prize, index) in draft.prizes" :key="`${prize.id}-${index}`" class="prize-row">
          <UFormField :name="`prize-name-${index}`" :label="t('adminLuckyDraw.prizeName')" required><UInput v-model.trim="prize.name" class="w-full" :maxlength="80" /></UFormField>
          <UFormField :name="`prize-weight-${index}`" :label="t('adminLuckyDraw.weight')" required><UInput v-model.trim="prize.weight" class="w-full" inputmode="numeric" pattern="[0-9]+" /></UFormField>
          <UFormField :name="`prize-stock-${index}`" :label="t('adminLuckyDraw.stock')"><UInput v-model.trim="prize.stockRemaining" class="w-full" inputmode="numeric" pattern="[0-9]*" :placeholder="t('adminLuckyDraw.unlimited')" /></UFormField>
          <UFormField :name="`prize-reward-${index}`" :label="t('adminLuckyDraw.reward')"><USelect v-model="prize.kind" class="w-full" :items="rewardItems" /></UFormField>
          <UFormField v-if="prize.kind === 'txb_delta'" :name="`prize-change-${index}`" :label="t('adminLuckyDraw.signedChange')" required><UInput v-model.trim="prize.txbDelta" class="w-full" inputmode="decimal" :placeholder="t('adminLuckyDraw.signedPlaceholder')" /></UFormField>
          <UFormField v-else-if="prize.kind === 'coupon_grant'" :name="`prize-coupon-${index}`" :label="t('adminLuckyDraw.coupon')" required><USelect v-model="prize.couponId" class="w-full" :items="couponItems" :placeholder="t('adminLuckyDraw.selectCoupon')" /></UFormField>
          <UFormField v-else-if="prize.kind === 'subscription_extension'" :name="`prize-days-${index}`" :label="t('adminLuckyDraw.extensionDays')" required><UInput v-model.number="prize.extensionDays" class="w-full" type="number" :min="1" :max="3650" :step="1" /></UFormField>
          <UButton class="prize-row__remove" color="error" variant="ghost" square icon="i-ph-trash" :disabled="draft.prizes.length === 1" :aria-label="t('adminLuckyDraw.removePrize', { name: prize.name || t('adminLuckyDraw.prizeNumber', { index: index + 1 }) })" @click="draft.prizes.splice(index, 1)" />
        </article>
      </div>
    </section>

    <p v-if="validationError" class="field-error catalog-editor__wide" role="alert">{{ validationError }}</p>

    <div class="button-row catalog-editor__wide"><UButton color="neutral" variant="outline" :label="t('common.cancel')" @click="emit('cancel')" /><UButton type="submit" :loading="busy" :disabled="busy" :label="busy ? t('common.saving') : t('adminLuckyDraw.save')" /></div>
  </form>
</template>

<style scoped>
.catalog-editor__heading p, .prize-list p { margin: 0.25rem 0 0; color: var(--text-muted); font-size: 0.76rem; }
.prize-list { display: grid; gap: 0.7rem; }
.prize-list__items { display: grid; gap: 0.7rem; }
.prize-list__heading { display: flex; align-items: center; justify-content: space-between; gap: 0.8rem; }
.prize-list h4 { margin: 0; }
.prize-row { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.55rem; padding: 0.7rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface); }
.prize-row__remove { align-self: end; justify-self: end; }
@media (min-width: 820px) { .prize-row { grid-template-columns: repeat(3, minmax(0, 1fr)) auto; } }
</style>
