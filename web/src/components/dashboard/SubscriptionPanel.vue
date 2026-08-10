<script setup lang="ts">
import { shallowRef, watch } from 'vue'

import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { useClipboard } from '@/composables/useClipboard'
import { openExternalLink } from '@/utils/telegram'

const props = defineProps<{
  subscriptionUrl?: string | null
  revoking: boolean
}>()

const emit = defineEmits<{ revoke: [] }>()
const confirmOpen = shallowRef(false)
const { copied, copy } = useClipboard()

watch(() => props.revoking, (next, previous) => {
  if (previous && !next) confirmOpen.value = false
})

function copyLink(): void {
  if (props.subscriptionUrl) void copy(props.subscriptionUrl)
}

function openLink(): void {
  if (props.subscriptionUrl) openExternalLink(props.subscriptionUrl)
}
</script>

<template>
  <section class="section-block subscription-section">
    <div class="section-heading">
      <h2>{{ $t('dashboard.subscriptionLink') }}</h2>
      <span class="feature-icon feature-icon--small"><UIcon name="i-ph-key" /></span>
    </div>
    <template v-if="subscriptionUrl">
      <p class="subscription-section__lead">{{ $t('dashboard.subscriptionHint') }}</p>
      <div class="secret-field">
        <span aria-hidden="true">••••••••••••••••••••••••</span>
        <UButton
          class="icon-button"
          color="neutral"
          variant="ghost"
          :icon="copied ? 'i-ph-check-bold' : 'i-ph-copy'"
          :aria-label="copied ? $t('common.copied') : $t('dashboard.copySubscription')"
          @click="copyLink"
        />
      </div>
      <div class="button-row">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-ph-arrow-square-out"
          :label="$t('common.open')"
          @click="openLink"
        />
        <UButton
          color="error"
          variant="ghost"
          icon="i-ph-trash"
          :label="$t('common.revoke')"
          @click="confirmOpen = true"
        />
      </div>
    </template>
    <div v-else class="empty-inline">
      <div>
        <h3>{{ $t('dashboard.noActiveLink') }}</h3>
        <p>{{ $t('dashboard.linkAfterActivation') }}</p>
      </div>
    </div>

    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="$t('dashboard.revokeTitle')"
      :description="$t('dashboard.revokeDescription')"
      :confirm-label="$t('dashboard.revokeLink')"
      :busy="revoking"
      danger
      @confirm="emit('revoke')"
    />
  </section>
</template>
