<script setup lang="ts">
import { shallowRef, watch } from 'vue'
import { useRouter } from 'vue-router'

import type { Purchase } from '@/api/types'
import { api } from '@/api/client'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import InlineNotice from '@/components/common/InlineNotice.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import RolloverFlipCard from './RolloverFlipCard.vue'
import PurchaseActions from './PurchaseActions.vue'
import SquadAdditionDialog from '@/components/squad-addition/SquadAdditionDialog.vue'
import { formatDate } from '@/utils/format'
import { localizedError } from '@/i18n'
import { notifyHaptic } from '@/utils/telegram'

const props = defineProps<{
  active?: Purchase | null
  queued?: Purchase | null
  squadNames?: readonly string[]
  openSquadAddition?: boolean
}>()
const emit = defineEmits<{ queuedCancelled: []; autoRenewalChanged: []; squadsChanged: []; squadAdditionRequestConsumed: [] }>()

const router = useRouter()
const queuedCancelOpen = shallowRef(false)
const queuedCancelBusy = shallowRef(false)
const queuedCancelError = shallowRef<string | null>(null)
const squadAdditionOpen = shallowRef(false)

function openQueuedCancellation(): void {
  queuedCancelError.value = null
  queuedCancelOpen.value = true
}
async function cancelQueued(): Promise<void> {
  if (!props.queued || queuedCancelBusy.value) return
  queuedCancelBusy.value = true
  queuedCancelError.value = null
  try {
    await api.cancelQueuedPurchase(props.queued.id)
    queuedCancelOpen.value = false
    notifyHaptic('success')
    emit('queuedCancelled')
  } catch (caught) {
    queuedCancelError.value = localizedError(caught, 'errors.purchaseCancelFailed')
    notifyHaptic('error')
  } finally {
    queuedCancelBusy.value = false
  }
}

function goToCatalog(): void {
  void router.push('/catalog')
}

function openSquadAdditionDialog(): void {
  if (!props.active || props.queued) return
  squadAdditionOpen.value = true
}

watch(() => props.openSquadAddition, (requested) => {
  if (!requested) return
  if (!props.active || props.queued) {
    emit('squadAdditionRequestConsumed')
    return
  }
  openSquadAdditionDialog()
}, { immediate: true })
watch(squadAdditionOpen, (open, wasOpen) => {
  if (!open && wasOpen && props.openSquadAddition) emit('squadAdditionRequestConsumed')
})
</script>

<template>
  <section class="section-block home-ride">
    <div class="section-heading">
      <h2>{{ $t('dashboard.yourRide') }}</h2>
      <StatusBadge v-if="active" tone="success" :label="$t('common.active')" />
    </div>

    <div v-if="active || queued" class="home-ride__content">
      <RolloverFlipCard v-if="active" :active="active" :squad-names="squadNames" :add-squad-disabled="Boolean(queued)" @add-squad="openSquadAdditionDialog" />
      <div v-if="queued" class="home-ride__queued">
        <div class="home-ride__queued-content">
          <span>
            <StatusBadge tone="warning" :label="$t('common.queued')" />
            {{ $t('dashboard.queuedStarts', { name: queued.comboName, date: formatDate(queued.validFrom) }) }}
          </span>
          <UButton
            color="error"
            variant="ghost"
            size="sm"
            icon="i-ph-x-circle"
            :label="$t('home.cancelQueued')"
            data-haptic="open"
            @click="openQueuedCancellation"
          />
          <InlineNotice v-if="queuedCancelError" tone="warning">{{ queuedCancelError }}</InlineNotice>
        </div>
      </div>
      <PurchaseActions v-if="active" :purchase="active" @changed="emit('autoRenewalChanged')" />
    </div>

    <div v-else class="empty-inline">
      <div>
        <h3>{{ $t('dashboard.noActiveCombo') }}</h3>
        <p>{{ $t('dashboard.choosePlan') }}</p>
      </div>
      <UButton
        class="home-ride__catalog"
        color="neutral"
        variant="outline"
        trailing-icon="i-ph-arrow-right"
        :label="$t('catalog.viewCombos')"
        data-haptic="navigate"
        @click="goToCatalog"
      />
    </div>
  </section>
  <ConfirmDialog
    v-model:open="queuedCancelOpen"
    :title="$t('home.cancelQueuedTitle')"
    :description="$t('home.cancelQueuedDescription')"
    :confirm-label="$t('home.cancelQueuedConfirm')"
    :busy="queuedCancelBusy"
    danger
    @confirm="cancelQueued"
  />
  <SquadAdditionDialog v-if="active" v-model:open="squadAdditionOpen" :active="active" @changed="emit('squadsChanged')" />
</template>

<style scoped>
.home-ride__queued-content {
  display: grid;
  gap: 0.55rem;
  min-width: 0;
}

.home-ride__queued-content > span {
  min-width: 0;
}
</style>
