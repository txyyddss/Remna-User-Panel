<script setup lang="ts">
import type { StatisticsNodesSnapshot } from '@/api/types'
import CountryFlag from '@/components/common/CountryFlag.vue'
import { formatBytes, formatDateTime } from '@/utils/format'
import { formatStatisticNumber } from './statisticsFormat'

defineProps<{
  snapshot: StatisticsNodesSnapshot | null
  loading: boolean
}>()
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
          <span class="statistics-node__state" :class="{ 'statistics-node__state--online': node.online }">
            {{ $t(node.online ? 'statistics.online' : 'statistics.offline') }}
          </span>
        </header>
        <dl>
          <div><dt>{{ $t('statistics.usersOnline') }}</dt><dd>{{ formatStatisticNumber(node.usersOnline) }}</dd></div>
          <div><dt>{{ $t('statistics.downloadRate') }}</dt><dd>{{ $t('statistics.perSecond', { value: formatBytes(node.rxBytesPerSec) }) }}</dd></div>
          <div><dt>{{ $t('statistics.uploadRate') }}</dt><dd>{{ $t('statistics.perSecond', { value: formatBytes(node.txBytesPerSec) }) }}</dd></div>
          <div><dt>{{ $t('statistics.xrayVersion') }}</dt><dd :title="node.xrayVersion">{{ node.xrayVersion || $t('common.notAvailable') }}</dd></div>
          <div><dt>{{ $t('statistics.multiplier') }}</dt><dd>{{ $t('statistics.multiplierValue', { value: formatStatisticNumber(node.multiplier, 2) }) }}</dd></div>
        </dl>
      </article>
    </div>
    <div v-else class="statistics-empty statistics-empty--panel">
      <UIcon :name="loading ? 'i-ph-spinner-gap' : 'i-ph-broadcast'" :class="{ 'icon-spin': loading }" aria-hidden="true" />
      <span>{{ $t(loading ? 'statistics.loadingNodes' : 'statistics.noNodes') }}</span>
    </div>
  </section>
</template>
