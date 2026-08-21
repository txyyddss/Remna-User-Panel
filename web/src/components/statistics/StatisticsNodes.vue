<script setup lang="ts">
import type { StatisticsNode, StatisticsNodesSnapshot } from '@/api/types'
import CountryFlag from '@/components/common/CountryFlag.vue'
import { formatBytes, formatDateTime } from '@/utils/format'
import { formatStatisticNumber } from './statisticsFormat'

defineProps<{
  snapshot: StatisticsNodesSnapshot | null
  loading: boolean
}>()

const emit = defineEmits<{ openGeocheck: [node: StatisticsNode] }>()
</script>

<template>
  <section class="statistics-section">
    <div class="statistics-section__heading">
      <h2>{{ $t('statistics.nodes') }}</h2>
      <span v-if="snapshot" class="statistics-section__meta">{{ formatDateTime(snapshot.generatedAt) }}</span>
    </div>
    <div v-if="snapshot?.stale" class="statistics-stale-label"><UIcon name="i-ph-clock-counter-clockwise" aria-hidden="true" />{{ $t('statistics.lastKnown') }}</div>
    <div v-if="snapshot?.nodes.length" class="statistics-node-grid">
      <article v-for="node in snapshot.nodes" :key="node.uuid" class="statistics-node">
        <header>
          <CountryFlag :code="node.countryCode" />
          <UTooltip :text="node.name"><h3>{{ node.name }}</h3></UTooltip>
          <UTooltip :text="$t(node.online ? 'statistics.online' : 'statistics.offline')">
            <span
              class="statistics-node__state"
              :class="{ 'statistics-node__state--online': node.online }"
              role="img"
              :aria-label="$t(node.online ? 'statistics.online' : 'statistics.offline')"
            >
              <UIcon :name="node.online ? 'i-ph-wifi-high' : 'i-ph-wifi-slash'" aria-hidden="true" />
            </span>
          </UTooltip>
          <UTooltip :text="$t('statistics.geocheck.open')">
            <UButton
              type="button"
              color="neutral"
              variant="ghost"
              square
              class="statistics-node__geocheck"
              icon="i-ph-globe-hemisphere-west"
              :aria-label="$t('statistics.geocheck.open')"
              data-haptic
              @click="emit('openGeocheck', node)"
            />
          </UTooltip>
        </header>
        <dl>
          <div>
            <dt><UTooltip :text="$t('statistics.usersOnline')"><span class="statistics-node__metric-icon" role="img" :aria-label="$t('statistics.usersOnline')"><UIcon name="i-ph-users" aria-hidden="true" /></span></UTooltip></dt>
            <dd>{{ formatStatisticNumber(node.usersOnline) }}</dd>
          </div>
          <div>
            <dt><UTooltip :text="$t('statistics.downloadRate')"><span class="statistics-node__metric-icon" role="img" :aria-label="$t('statistics.downloadRate')"><UIcon name="i-ph-arrow-down" aria-hidden="true" /></span></UTooltip></dt>
            <dd :title="$t('statistics.perSecond', { value: formatBytes(node.rxBytesPerSec) })">{{ $t('statistics.perSecond', { value: formatBytes(node.rxBytesPerSec) }) }}</dd>
          </div>
          <div>
            <dt><UTooltip :text="$t('statistics.uploadRate')"><span class="statistics-node__metric-icon" role="img" :aria-label="$t('statistics.uploadRate')"><UIcon name="i-ph-arrow-up" aria-hidden="true" /></span></UTooltip></dt>
            <dd :title="$t('statistics.perSecond', { value: formatBytes(node.txBytesPerSec) })">{{ $t('statistics.perSecond', { value: formatBytes(node.txBytesPerSec) }) }}</dd>
          </div>
        </dl>
      </article>
    </div>
    <div v-else class="statistics-empty statistics-empty--panel">
      <UIcon :name="loading ? 'i-ph-spinner-gap' : 'i-ph-broadcast'" :class="{ 'icon-spin': loading }" aria-hidden="true" />
      <span>{{ $t(loading ? 'statistics.loadingNodes' : 'statistics.noNodes') }}</span>
    </div>
  </section>
</template>
