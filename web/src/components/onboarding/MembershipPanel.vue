<script setup lang="ts">
import { PhArrowClockwise, PhArrowUpRight, PhBroadcast, PhCheck, PhUsersThree } from '@phosphor-icons/vue'

import type { InviteLink } from '@/api/types'

defineProps<{
  invites: readonly InviteLink[]
  loading: boolean
}>()

defineEmits<{
  openInvite: [invite: InviteLink]
  check: []
  refresh: []
}>()

function iconFor(kind: InviteLink['kind']) {
  return kind === 'group' ? PhUsersThree : PhBroadcast
}
</script>

<template>
  <section class="onboarding-panel">
    <header class="onboarding-panel__header">
      <p class="eyebrow">Membership</p>
      <h1>Two quick joins.</h1>
      <p>Join the private group and the updates channel. Access is checked against your Telegram account.</p>
    </header>

    <div v-if="invites.length" class="invite-list">
      <button
        v-for="invite in invites"
        :key="invite.kind"
        class="invite-row"
        type="button"
        @click="$emit('openInvite', invite)"
      >
        <span class="invite-row__icon"><component :is="iconFor(invite.kind)" :size="22" /></span>
        <span class="invite-row__copy">
          <strong>{{ invite.label }}</strong>
          <small>{{ invite.joined ? 'Membership verified' : 'Invite expires in 30 minutes' }}</small>
        </span>
        <PhCheck v-if="invite.joined" :size="20" weight="bold" class="invite-row__check" />
        <PhArrowUpRight v-else :size="20" />
      </button>
    </div>
    <button v-else class="button button--secondary" type="button" :disabled="loading" @click="$emit('refresh')">
      <PhArrowClockwise :size="19" />
      Generate secure invites
    </button>

    <button class="button button--primary button--wide" type="button" :disabled="loading" @click="$emit('check')">
      <span>{{ loading ? 'Checking membership' : 'Already joined' }}</span>
      <PhArrowClockwise :size="19" :class="{ 'icon-spin': loading }" />
    </button>
  </section>
</template>
