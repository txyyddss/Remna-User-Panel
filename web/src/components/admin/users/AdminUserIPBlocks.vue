<script setup lang="ts">
import type { AdminUserDetail } from '@/api/adminOperations'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { t } from '@/i18n'
import { formatDateTime } from '@/utils/format'

type IPBlock = AdminUserDetail['ipBlocks'][number]

defineProps<{ items: readonly IPBlock[]; busy: boolean }>()
const emit = defineEmits<{ unblock: [block: IPBlock] }>()
</script>

<template>
  <section class="admin-profile-section admin-ip-blocks">
    <div class="admin-profile-section__heading">
      <div><h3>{{ t('adminUserProfile.ipBlocks') }}</h3><p>{{ t('adminUserProfile.ipBlocksHint') }}</p></div>
    </div>
    <div v-if="items.length" class="admin-profile-list admin-profile-list--compact">
      <div v-for="block in items" :key="block.id" class="admin-profile-row">
        <div class="admin-profile-row__main">
          <strong><UIcon name="i-ph-shield-warning" />{{ block.ip }}</strong>
          <small>{{ t('adminUserProfile.ipBlockExpiry', { date: formatDateTime(block.expiresAt) }) }}</small>
        </div>
        <StatusBadge :tone="block.status === 'active' ? 'success' : 'warning'" :label="t(`adminUserProfile.ipBlockStatus.${block.status}`)" />
        <div class="admin-profile-row__actions">
          <UButton color="neutral" variant="outline" icon="i-ph-shield-check" :label="t('adminUserProfile.unblockIP')" :disabled="busy || block.status === 'unblocking'" @click="emit('unblock', block)" />
        </div>
      </div>
    </div>
    <p v-else class="admin-profile-empty">{{ t('adminUserProfile.noIPBlocks') }}</p>
  </section>
</template>
