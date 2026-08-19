<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useConnectionDrop } from '@/composables/useConnectionDrop'
import { useConnectionScan } from '@/composables/useConnectionScan'
import { t } from '@/i18n'
import ConnectionDropDialog from './ConnectionDropDialog.vue'
import ConnectionNodeList from './ConnectionNodeList.vue'
import type { ConnectionTarget } from './types'

const scan = useConnectionScan()
const drop = useConnectionDrop()
const selected = shallowRef<ConnectionTarget | null>(null)
const dialogOpen = shallowRef(false)
let refreshedOperationId = ''

const hasConnections = computed(() => scan.nodes.value.some((node) => node.ips.length > 0))
const operationTone = computed(() => drop.receipt.value?.status === 'succeeded' ? 'success' : 'warning')
const operationMessage = computed(() => t(`connections.operation.${drop.receipt.value?.status ?? 'queued'}`))

function selectTarget(target: ConnectionTarget): void {
  if (drop.blocksMutations.value) return
  drop.reset()
  selected.value = target
  dialogOpen.value = true
}

async function confirmDrop(): Promise<void> {
  if (!selected.value) return
  await drop.drop(selected.value.connection.handle)
}

watch(() => drop.receipt.value, (receipt) => {
  if (!receipt || receipt.status !== 'succeeded' || receipt.id === refreshedOperationId) return
  refreshedOperationId = receipt.id
  void scan.restart()
})
</script>

<template>
  <div class="page page--connections">
    <header class="connections-heading">
      <div>
        <span class="eyebrow">{{ $t('connections.eyebrow') }}</span>
        <h1>{{ $t('connections.title') }}</h1>
        <p>{{ $t('connections.subtitle') }}</p>
      </div>
      <UButton
        color="neutral"
        variant="outline"
        square
        icon="i-ph-arrow-clockwise"
        :loading="scan.loading.value || scan.polling.value"
        :disabled="drop.blocksMutations.value"
        :aria-label="$t('connections.scanAgain')"
        data-haptic
        @click="scan.restart()"
      />
    </header>

    <InlineNotice v-if="drop.receipt.value && !dialogOpen" :tone="operationTone" :title="$t(`operations.status.${drop.receipt.value.status}`)">
      {{ operationMessage }}
    </InlineNotice>

    <div v-if="scan.loading.value" class="connections-loading" aria-busy="true">
      <SkeletonBlock height="7rem" />
      <SkeletonBlock height="11rem" />
    </div>

    <section v-else-if="scan.polling.value" class="connections-progress" aria-live="polite">
      <div class="connections-progress__icon"><UIcon name="i-ph-radar" aria-hidden="true" /></div>
      <div><h2>{{ $t('connections.scanning') }}</h2><p>{{ $t('connections.progress', { percent: Math.round(scan.progressPercent.value) }) }}</p></div>
      <UProgress class="connections-progress__bar" :model-value="scan.progressPercent.value" :max="100" />
    </section>

    <section v-else-if="scan.error.value || scan.failed.value" class="error-state error-state--compact">
      <UIcon name="i-ph-warning-circle" class="connections-state-icon" aria-hidden="true" />
      <h2>{{ $t('connections.scanUnavailable') }}</h2>
      <p>{{ scan.error.value ?? $t('connections.scanFailed') }}</p>
      <UButton icon="i-ph-arrow-clockwise" :label="$t('common.tryAgain')" data-haptic @click="scan.retry()" />
    </section>

    <section v-else-if="scan.completed.value && !hasConnections" class="connections-empty">
      <UIcon name="i-ph-plugs" class="connections-state-icon" aria-hidden="true" />
      <h2>{{ $t('connections.emptyTitle') }}</h2>
      <p>{{ $t('connections.emptyDescription') }}</p>
      <UButton color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :label="$t('connections.scanAgain')" data-haptic @click="scan.restart()" />
    </section>

    <ConnectionNodeList
      v-else-if="scan.completed.value"
      :nodes="scan.nodes.value"
      :disabled="drop.busy.value || drop.blocksMutations.value"
      @drop="selectTarget"
    />

    <ConnectionDropDialog
      v-model:open="dialogOpen"
      :target="selected"
      :receipt="drop.receipt.value"
      :busy="drop.busy.value"
      :checking="drop.checking.value"
      :error="drop.error.value"
      @confirm="confirmDrop"
      @refresh="drop.refresh"
    />
  </div>
</template>
