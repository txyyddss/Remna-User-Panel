<script setup lang="ts">
import { computed, shallowRef, toRefs, watch } from 'vue'
import { useRouter } from 'vue-router'

import type { Purchase } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import { formatBytes, formatDate, formatMoney } from '@/utils/format'
import { api } from '@/api/client'
import { createUuid } from '@/utils/browserCompatibility'
import { localizedError } from '@/i18n'
import { notifyHaptic } from '@/utils/telegram'

const props = defineProps<{
  active?: Purchase | null
  queued?: Purchase | null
  squadNames?: readonly string[]
}>()

const { active, queued } = toRefs(props)
const { t } = useI18n()
const router = useRouter()
const resetLabel = computed(() => props.active ? t(`home.reset.${props.active.resetStrategy}`) : '')
const renewalOpen = shallowRef(false)
const renewalTerms = shallowRef(1)
const renewalQuote = shallowRef<import('@/api/types').RenewalQuote | null>(null)
const renewalBusy = shallowRef(false)
const renewalError = shallowRef<string | null>(null)
const renewalSuccess = shallowRef(false)

async function loadRenewalQuote(): Promise<void> {
  if (!props.active) return
  renewalError.value = null
  try { renewalQuote.value = await api.quoteRenewal(props.active.id, renewalTerms.value) } catch (caught) { renewalQuote.value = null; renewalError.value = localizedError(caught, 'errors.renewalFailed') }
}
async function renew(): Promise<void> {
  if (!props.active || !renewalQuote.value || renewalBusy.value) return
  renewalBusy.value = true
  renewalError.value = null
  try { await api.renewPurchase(props.active.id, renewalTerms.value, createUuid()); renewalSuccess.value = true; notifyHaptic('success') } catch (caught) { renewalError.value = localizedError(caught, 'errors.renewalFailed'); notifyHaptic('error') } finally { renewalBusy.value = false }
}
watch(renewalOpen, (open) => { if (open) { renewalSuccess.value = false; void loadRenewalQuote() } })
watch(renewalTerms, () => { if (renewalOpen.value) void loadRenewalQuote() })

function goToCatalog(): void {
  void router.push('/catalog')
}
</script>

<template>
  <section class="section-block home-ride">
    <div class="section-heading">
      <h2>{{ $t('dashboard.yourRide') }}</h2>
      <StatusBadge v-if="active" tone="success" :label="$t('common.active')" />
    </div>

    <div v-if="active" class="home-ride__summary">
      <div class="home-ride__primary">
        <span class="feature-icon"><UIcon name="i-ph-stack" /></span>
        <div>
          <h3>{{ active.comboName }}</h3>
          <p>{{ squadNames?.length ? squadNames.join(t('home.squadSeparator')) : $t('dashboard.squadsIncluded', { count: active.squadUuids.length }) }}</p>
        </div>
      </div>
      <dl class="home-ride__metrics">
        <div>
          <dt><UIcon name="i-ph-gauge" /> {{ $t('dashboard.traffic') }}</dt>
          <dd>{{ formatBytes(active.trafficLimitBytes) }}</dd>
        </div>
        <div>
          <dt><UIcon name="i-ph-arrow-clockwise" /> {{ $t('home.resetCadence') }}</dt>
          <dd>{{ resetLabel }}</dd>
        </div>
        <div>
          <dt><UIcon name="i-ph-calendar-blank" /> {{ $t('dashboard.renews') }}</dt>
          <dd>{{ formatDate(active.validUntil) }}</dd>
        </div>
      </dl>
      <div v-if="queued" class="home-ride__queued">
        <span>
          <StatusBadge tone="warning" :label="$t('common.queued')" />
          {{ $t('dashboard.queuedStarts', { name: queued.comboName, date: formatDate(queued.validFrom) }) }}
        </span>
      </div>
      <UButton block color="neutral" variant="outline" trailing-icon="i-ph-arrow-clockwise" :label="$t('home.renew')" data-haptic @click="renewalOpen = true" />
    </div>

    <div v-else class="empty-inline">
      <div>
        <h3>{{ $t('dashboard.noActiveCombo') }}</h3>
        <p>{{ $t('dashboard.choosePlan') }}</p>
      </div>
      <UButton
        color="neutral"
        variant="outline"
        trailing-icon="i-ph-arrow-right"
        :label="$t('catalog.viewCombos')"
        data-haptic
        @click="goToCatalog"
      />
    </div>
  </section>
  <UModal v-model:open="renewalOpen" :title="$t('home.renewTitle')" :description="$t('home.renewHint')">
    <template #body>
      <div class="renewal-dialog">
        <div class="renewal-dialog__terms">
          <div class="renewal-dialog__terms-heading"><span>{{ $t('home.renewTerms') }}</span><strong>{{ renewalTerms }}</strong></div>
          <USlider v-model="renewalTerms" :min="1" :max="6" :step="1" :aria-label="$t('home.renewTerms')" />
        </div>
        <div v-if="renewalQuote" class="renewal-dialog__quote">
          <div class="renewal-dialog__total"><span>{{ $t('home.renewPrice') }}</span><strong>{{ formatMoney(renewalQuote.totalPrice) }}</strong></div>
          <small>{{ formatDate(renewalQuote.effectiveAt) }} {{ t('common.rangeSeparator') }} {{ formatDate(renewalQuote.expiresAt) }}</small>
        </div>
        <InlineNotice v-if="renewalSuccess" tone="success">{{ $t('home.renewSuccess') }}</InlineNotice>
        <InlineNotice v-if="renewalError" tone="warning">{{ renewalError }}</InlineNotice>
        <UButton v-if="!renewalSuccess" block :loading="renewalBusy" :disabled="!renewalQuote" :label="$t('home.renewAction')" data-haptic @click="renew" />
      </div>
    </template>
  </UModal>
</template>
