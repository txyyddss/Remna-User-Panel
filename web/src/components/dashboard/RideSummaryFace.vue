<script setup lang="ts">
import { computed } from 'vue'

import type { Purchase } from '@/api/types'
import { useI18n } from '@/i18n'
import { formatBytes, formatDate } from '@/utils/format'

const props = defineProps<{
  active: Purchase
  squadNames?: readonly string[]
  addSquadDisabled?: boolean
}>()
const emit = defineEmits<{ addSquad: []; openRollover: [] }>()

const { t } = useI18n()
const resetLabel = computed(() => t(`home.reset.${props.active.resetStrategy}`))
</script>

<template>
  <div class="home-ride__summary-face">
    <div class="home-ride__primary">
      <span class="feature-icon"><UIcon name="i-ph-stack" /></span>
      <div class="home-ride__name-block">
        <div class="home-ride__name-row">
          <h3>{{ active.comboName }}</h3>
          <UTooltip v-if="!addSquadDisabled" :text="$t('home.squadAddition.open')">
            <UButton
              class="home-ride__add-squad"
              color="success"
              variant="solid"
              size="sm"
              icon="i-ph-plus"
              :aria-label="$t('home.squadAddition.open')"
              data-haptic="open"
              @click.stop="emit('addSquad')"
            />
          </UTooltip>
        </div>
        <p>{{ squadNames?.length ? squadNames.join(t('home.squadSeparator')) : $t('dashboard.squadsIncluded', { count: active.squadUuids.length }) }}</p>
      </div>
      <UButton
        class="home-ride__flip-icon"
        color="neutral"
        variant="ghost"
        icon="i-ph-arrows-clockwise"
        :aria-label="$t('home.rolloverOpen')"
        data-haptic="open"
        @click="emit('openRollover')"
      />
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
    <span class="home-ride__flip-hint">{{ $t('home.rolloverOpen') }}</span>
  </div>
</template>

<style scoped>
.home-ride__name-block { min-width: 0; }
.home-ride__name-row { display: flex; align-items: center; gap: 0.55rem; min-width: 0; }
.home-ride__name-row h3 { min-width: 0; margin: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.home-ride__add-squad {
  position: relative;
  z-index: 2;
  flex: 0 0 auto;
  min-width: 2.75rem;
  min-height: 2.75rem;
  padding-inline: 0.75rem;
  pointer-events: auto;
  transition: background-color 180ms var(--ease-out), transform 120ms var(--ease-out);
}
.home-ride__add-squad:active { transform: translateY(1px); }
@media (min-width: 900px) { .home-ride__add-squad { display: none; } }
</style>
