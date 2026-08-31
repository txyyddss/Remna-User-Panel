<script setup lang="ts">
import { computed } from 'vue'

import InlineNotice from '@/components/common/InlineNotice.vue'
import LanguageControl from '@/components/layout/LanguageControl.vue'
import { useOnboarding } from '@/composables/useOnboarding'
import { useI18n } from '@/i18n'
import AgreementPanel from './AgreementPanel.vue'
import IntroSequence from './IntroSequence.vue'
import { useOnboardingMainButton } from './useOnboardingMainButton'
import UsernamePanel from './UsernamePanel.vue'

const {
  step,
  progress,
  loading,
  error,
  content,
  form,
  usernameValid,
  usernameHint,
  allAgreementsAccepted,
  finishIntro,
  submitUsername,
  acceptAgreement,
  toggleAgreement,
} = useOnboarding()

const { t } = useI18n()
const mainAction = computed(() => {
  if (step.value === 'username') return {
    text: loading.value ? t('onboarding.checkingAvailability') : t('onboarding.reserveUsername'),
    disabled: !usernameValid.value || loading.value,
    loading: loading.value,
    run: submitUsername,
  }
  if (step.value === 'agreement') return {
    text: loading.value ? t('onboarding.finishing') : t('onboarding.finish'),
    disabled: !allAgreementsAccepted.value || loading.value,
    loading: loading.value,
    run: acceptAgreement,
  }
  return null
})
const { available: mainButtonAvailable } = useOnboardingMainButton(mainAction)
</script>

<template>
  <IntroSequence v-if="step === 'intro' && content" :messages="content.welcome" @complete="finishIntro" />
  <main v-else class="onboarding-shell">
    <UProgress
      class="onboarding-shell__track"
      :aria-label="$t('onboarding.setupProgress')"
      :model-value="progress * 100"
      :max="100"
    />

    <div class="onboarding-shell__stage">
      <Transition name="onboarding-step" mode="out-in">
        <UsernamePanel
          v-if="step === 'username'"
          key="username"
          v-model="form.username"
          :valid="usernameValid"
          :hint="usernameHint"
          :loading="loading"
          :show-action="!mainButtonAvailable"
          @submit="submitUsername"
        />
        <AgreementPanel
          v-else-if="step === 'agreement'"
          key="agreement"
          :agreements="content?.agreements ?? []"
          :selected-ids="form.agreementIds"
          :all-accepted="allAgreementsAccepted"
          :loading="loading"
          :show-action="!mainButtonAvailable"
          @toggle="toggleAgreement"
          @submit="acceptAgreement"
        />
      </Transition>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    </div>
    <footer class="onboarding-shell__locale"><LanguageControl /></footer>
  </main>
</template>
