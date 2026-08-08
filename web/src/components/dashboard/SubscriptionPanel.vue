<script setup lang="ts">
import { shallowRef, watch } from 'vue'
import { PhArrowSquareOut, PhCheck, PhCopy, PhKey, PhTrash } from '@phosphor-icons/vue'

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
      <span class="feature-icon feature-icon--small"><PhKey :size="19" /></span>
    </div>
    <template v-if="subscriptionUrl">
      <p class="subscription-section__lead">{{ $t('dashboard.subscriptionHint') }}</p>
      <div class="secret-field">
        <span>••••••••••••••••••••••••</span>
        <button class="icon-button" type="button" :aria-label="copied ? $t('common.copied') : $t('dashboard.copySubscription')" @click="copyLink">
          <PhCheck v-if="copied" :size="20" weight="bold" />
          <PhCopy v-else :size="20" />
        </button>
      </div>
      <div class="button-row">
        <button class="button button--secondary" type="button" @click="openLink">
          <PhArrowSquareOut :size="19" />
          {{ $t('common.open') }}
        </button>
        <button class="button button--ghost-danger" type="button" @click="confirmOpen = true">
          <PhTrash :size="19" />
          {{ $t('common.revoke') }}
        </button>
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
