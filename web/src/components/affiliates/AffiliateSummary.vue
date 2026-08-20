<script setup lang="ts">
import type { AffiliateOverview } from '@/api/features'
import { useClipboard } from '@/composables/useClipboard'

defineProps<{ overview: AffiliateOverview }>()
const clipboard = useClipboard()
</script>

<template>
  <section class="affiliate-section affiliate-summary">
    <div class="affiliate-link-row">
      <div><span>{{ $t('affiliates.inviteLink') }}</span><strong>{{ overview.inviteLink ?? $t('affiliates.linkUnavailable') }}</strong></div>
      <UButton
        color="neutral" variant="outline" icon="i-ph-copy" :disabled="!overview.inviteLink"
        :label="clipboard.copied.value ? $t('common.copied') : $t('affiliates.copy')"
        @click="overview.inviteLink && clipboard.copy(overview.inviteLink)"
      />
    </div>
    <dl class="affiliate-metrics">
      <div><dt>{{ $t('affiliates.totalCommission') }}</dt><dd>{{ overview.totalCommission.display }}</dd></div>
      <div><dt>{{ $t('affiliates.registered') }}</dt><dd>{{ overview.registeredCount }}</dd></div>
      <div><dt>{{ $t('affiliates.successful') }}</dt><dd>{{ overview.successfulCount }}</dd></div>
      <div><dt>{{ $t('affiliates.conversion') }}</dt><dd>{{ (overview.conversionBps / 100).toFixed(2) }}%</dd></div>
    </dl>
  </section>
</template>
