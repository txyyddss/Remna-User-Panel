<script setup lang="ts">
import type { ConnectivityAttempt } from '@/api/connectivity'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'
import { connectivityError, connectivityTone } from './presentation'

defineProps<{ attempt: ConnectivityAttempt | null }>()
const { t } = useI18n()
</script>

<template>
  <div class="attempt-summary">
    <div class="attempt-summary__status">
      <StatusBadge :tone="connectivityTone(attempt?.status)" :label="t(`hostConnectivity.status.${attempt?.status ?? 'unchecked'}`)" />
      <span v-if="attempt?.latencyMs != null">{{ t('hostConnectivity.latency', { value: Math.round(attempt.latencyMs) }) }}</span>
      <span v-if="attempt?.httpStatus != null">{{ t('hostConnectivity.httpStatus', { value: attempt.httpStatus }) }}</span>
    </div>
    <small v-if="attempt">{{ formatDateTime(attempt.startedAt) }}</small>
    <small v-if="attempt?.errorCode" class="attempt-summary__error">{{ connectivityError(attempt.errorCode, 'hostConnectivity.checkFailed') }}</small>
  </div>
</template>

<style scoped>
.attempt-summary { display: grid; gap: 0.25rem; color: var(--text-muted); font-size: 0.75rem; min-width: 0; }
.attempt-summary__status { display: flex; flex-wrap: wrap; align-items: center; gap: 0.45rem; }
.attempt-summary__error { overflow-wrap: anywhere; }
</style>
