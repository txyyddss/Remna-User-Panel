<script setup lang="ts">
import { PhCirclesFour } from '@phosphor-icons/vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import { useOnboarding } from '@/composables/useOnboarding'
import AgreementPanel from './AgreementPanel.vue'
import IntroSequence from './IntroSequence.vue'
import MembershipPanel from './MembershipPanel.vue'
import UsernamePanel from './UsernamePanel.vue'

const {
  step,
  progress,
  loading,
  error,
  invites,
  form,
  usernameValid,
  usernameHint,
  finishIntro,
  loadInvites,
  openInvite,
  checkMembership,
  submitUsername,
  acceptAgreement,
} = useOnboarding()
</script>

<template>
  <IntroSequence v-if="step === 'intro'" @complete="finishIntro" />
  <main v-else class="onboarding-shell">
    <header class="onboarding-shell__top">
      <span class="onboarding-shell__brand"><PhCirclesFour :size="20" weight="fill" /> TX Carpool</span>
      <span class="onboarding-shell__percent">{{ Math.round(progress * 100) }}%</span>
    </header>
    <div class="onboarding-shell__track" aria-label="Setup progress">
      <span :style="{ transform: `scaleX(${progress})` }" />
    </div>

    <div class="onboarding-shell__stage">
      <Transition name="onboarding-step" mode="out-in">
        <MembershipPanel
          v-if="step === 'membership'"
          key="membership"
          :invites="invites"
          :loading="loading"
          @open-invite="openInvite"
          @check="checkMembership"
          @refresh="loadInvites"
        />
        <UsernamePanel
          v-else-if="step === 'username'"
          key="username"
          v-model="form.username"
          :valid="usernameValid"
          :hint="usernameHint"
          :loading="loading"
          @submit="submitUsername"
        />
        <AgreementPanel
          v-else-if="step === 'agreement'"
          key="agreement"
          v-model="form.agreement"
          :loading="loading"
          @submit="acceptAgreement"
        />
      </Transition>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    </div>
  </main>
</template>
