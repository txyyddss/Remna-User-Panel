<script setup lang="ts">
import type { Purchase } from '@/api/types'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { formatBytes, formatDate } from '@/utils/format'

defineProps<{
  active?: Purchase | null
  queued?: Purchase | null
}>()
</script>

<template>
  <section class="section-block">
    <div class="section-heading">
      <h2>{{ $t('dashboard.yourRide') }}</h2>
      <StatusBadge v-if="active" tone="success" :label="$t('common.active')" />
    </div>

    <div v-if="active" class="entitlement-panel">
      <div class="entitlement-panel__primary">
        <span class="feature-icon"><UIcon name="i-ph-stack" /></span>
        <div>
          <h3>{{ active.comboName }}</h3>
          <p>{{ $t('dashboard.squadsIncluded', { count: active.squadUuids.length }) }}</p>
        </div>
      </div>
      <dl class="metric-pair">
        <div>
          <dt><UIcon name="i-ph-gauge" /> {{ $t('dashboard.traffic') }}</dt>
          <dd>{{ formatBytes(active.trafficLimitBytes) }}</dd>
        </div>
        <div>
          <dt><UIcon name="i-ph-calendar-blank" /> {{ $t('dashboard.renews') }}</dt>
          <dd>{{ formatDate(active.validUntil) }}</dd>
        </div>
      </dl>
      <div v-if="queued" class="queued-plan">
        <span>
          <StatusBadge tone="warning" :label="$t('common.queued')" />
          {{ $t('dashboard.queuedStarts', { name: queued.comboName, date: formatDate(queued.validFrom) }) }}
        </span>
      </div>
    </div>

    <div v-else class="empty-inline">
      <div>
        <h3>{{ $t('dashboard.noActiveCombo') }}</h3>
        <p>{{ $t('dashboard.choosePlan') }}</p>
      </div>
      <UButton
        to="/catalog"
        color="neutral"
        variant="outline"
        trailing-icon="i-ph-arrow-right"
        :label="$t('catalog.viewCombos')"
      />
    </div>
  </section>
</template>
