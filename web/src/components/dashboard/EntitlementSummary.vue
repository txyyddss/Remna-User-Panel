<script setup lang="ts">
import { computed, toRefs } from 'vue'

import type { Purchase } from '@/api/types'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import { formatBytes, formatDate } from '@/utils/format'

const props = defineProps<{
  active?: Purchase | null
  queued?: Purchase | null
  squadNames?: readonly string[]
}>()

const { active, queued } = toRefs(props)
const { t } = useI18n()
const resetLabel = computed(() => props.active ? t(`home.reset.${props.active.resetStrategy}`) : '')
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
          <p>{{ squadNames?.length ? squadNames.join(' · ') : $t('dashboard.squadsIncluded', { count: active.squadUuids.length }) }}</p>
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
        data-haptic
      />
    </div>
  </section>
</template>
