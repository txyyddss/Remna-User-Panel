<script setup lang="ts">
import type { InviteLink } from '@/api/types'

defineProps<{
  invites: readonly InviteLink[]
  loading: boolean
  showAction?: boolean
}>()

defineEmits<{
  openInvite: [invite: InviteLink]
  check: []
  refresh: []
}>()

function iconFor(kind: InviteLink['kind']): string {
  return kind === 'group' ? 'i-ph-users-three' : 'i-ph-broadcast'
}
</script>

<template>
  <section class="onboarding-panel">
    <header class="onboarding-panel__header">
      <p class="eyebrow">{{ $t('onboarding.membership') }}</p>
      <h1>{{ $t('onboarding.twoQuickJoins') }}</h1>
      <p>{{ $t('onboarding.membershipCopy') }}</p>
    </header>

    <div v-if="invites.length" class="invite-list">
      <UButton
        v-for="invite in invites"
        :key="invite.kind"
        class="invite-row"
        color="neutral"
        variant="ghost"
        @click="$emit('openInvite', invite)"
      >
        <span class="invite-row__icon"><UIcon :name="iconFor(invite.kind)" /></span>
        <span class="invite-row__copy">
          <strong>{{ invite.label }}</strong>
          <small>{{ invite.joined ? $t('onboarding.membershipVerified') : $t('onboarding.inviteExpiry') }}</small>
        </span>
        <UIcon :name="invite.joined ? 'i-ph-check-bold' : 'i-ph-arrow-up-right'" class="invite-row__check" />
      </UButton>
    </div>
    <UButton
      v-else
      color="neutral"
      variant="outline"
      icon="i-ph-arrow-clockwise"
      :disabled="loading"
      :label="$t('onboarding.generateInvites')"
      @click="$emit('refresh')"
    />

    <UButton
      v-if="showAction !== false"
      block
      icon="i-ph-arrow-clockwise"
      :loading="loading"
      :label="loading ? $t('onboarding.checkingMembership') : $t('onboarding.alreadyJoined')"
      @click="$emit('check')"
    />
  </section>
</template>
