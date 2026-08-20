<script setup lang="ts">
import type { AffiliateReferralPage } from '@/api/features'
import { formatDate, formatDateTime } from '@/utils/format'

defineProps<{ page: AffiliateReferralPage; loading: boolean }>()
defineEmits<{ 'update:page': [page: number] }>()
</script>

<template>
  <section class="affiliate-section affiliate-referrals">
    <div class="affiliate-section__heading"><div><span>{{ $t('affiliates.history') }}</span><h2>{{ $t('affiliates.referrals') }}</h2></div><strong>{{ page.total }}</strong></div>
    <div v-if="loading" class="affiliate-referral-list"><USkeleton v-for="index in 5" :key="index" class="h-16 w-full" /></div>
    <div v-else-if="page.items.length" class="affiliate-referral-list">
      <article v-for="item in page.items" :key="`${item.username}-${item.registeredAt}`" class="affiliate-referral-row">
        <div><strong>{{ item.username || $t('affiliates.memberFallback') }}</strong><small>{{ $t('affiliates.joinedOn', { date: formatDate(item.registeredAt) }) }}</small></div>
        <div><UBadge :color="item.status === 'successful' ? 'success' : 'neutral'" variant="soft">{{ $t(`affiliates.status.${item.status}`) }}</UBadge><small>{{ item.paybackAt ? formatDateTime(item.paybackAt) : $t('affiliates.pending') }}</small><strong>{{ item.commissionAmount?.display ?? $t('affiliates.pending') }}</strong></div>
      </article>
    </div>
    <p v-else class="affiliate-empty">{{ $t('affiliates.noReferrals') }}</p>
    <UPagination v-if="page.totalPages > 1" :page="page.page" :items-per-page="page.pageSize" :total="page.total" @update:page="$emit('update:page', $event)" />
  </section>
</template>
