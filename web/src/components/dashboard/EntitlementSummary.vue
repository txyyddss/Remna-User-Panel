<script setup lang="ts">
import { PhArrowRight, PhCalendarBlank, PhGauge, PhStack } from '@phosphor-icons/vue'
import { RouterLink } from 'vue-router'

import type { Purchase } from '@/api/types'
import { formatBytes, formatDate } from '@/utils/format'
import StatusBadge from '@/components/common/StatusBadge.vue'

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
        <span class="feature-icon"><PhStack :size="24" /></span>
        <div>
          <h3>{{ active.comboName }}</h3>
          <p>{{ $t('dashboard.squadsIncluded', { count: active.squadUuids.length }) }}</p>
        </div>
      </div>
      <dl class="metric-pair">
        <div>
          <dt><PhGauge :size="17" /> {{ $t('dashboard.traffic') }}</dt>
          <dd>{{ formatBytes(active.trafficLimitBytes) }}</dd>
        </div>
        <div>
          <dt><PhCalendarBlank :size="17" /> {{ $t('dashboard.renews') }}</dt>
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
      <RouterLink class="button button--secondary" to="/catalog">
        {{ $t('catalog.viewCombos') }}
        <PhArrowRight :size="18" />
      </RouterLink>
    </div>
  </section>
</template>
