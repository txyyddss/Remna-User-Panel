<script setup lang="ts">
import { computed } from 'vue'
import { AnimatePresence, motion } from 'motion-v'

import InlineNotice from '@/components/common/InlineNotice.vue'
import LanguageControl from '@/components/layout/LanguageControl.vue'
import { useOnboarding } from '@/composables/useOnboarding'
import { useI18n } from '@/i18n'
import AgreementPanel from './AgreementPanel.vue'
import IntroSequence from './IntroSequence.vue'
import { useOnboardingMainButton } from './useOnboardingMainButton'
import UsernamePanel from './UsernamePanel.vue'
import { motionDurations } from '@/composables/motionPresets'
import { useMotionPreferences } from '@/composables/useMotionPreferences'

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
const { reducedMotion, offset } = useMotionPreferences()
const stepDirection = computed(() => step.value === 'agreement' ? 1 : -1)
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
      <AnimatePresence mode="wait" :initial="false">
        <motion.div
          v-if="step === 'username'"
          key="username"
          :initial="{ opacity: 0, y: offset(10) * stepDirection }"
          :animate="{ opacity: 1, y: 0 }"
          :exit="{ opacity: 0, y: reducedMotion ? 0 : -8 * stepDirection }"
          :transition="{ duration: reducedMotion ? 0.08 : motionDurations.step, ease: 'easeOut' }"
        >
          <UsernamePanel
            v-model="form.username"
            :valid="usernameValid"
            :hint="usernameHint"
            :loading="loading"
            :show-action="!mainButtonAvailable"
            @submit="submitUsername"
          />
        </motion.div>
        <motion.div
          v-else-if="step === 'agreement'"
          key="agreement"
          :initial="{ opacity: 0, y: offset(10) * stepDirection }"
          :animate="{ opacity: 1, y: 0 }"
          :exit="{ opacity: 0, y: reducedMotion ? 0 : -8 * stepDirection }"
          :transition="{ duration: reducedMotion ? 0.08 : motionDurations.step, ease: 'easeOut' }"
        >
          <AgreementPanel
            :agreements="content?.agreements ?? []"
            :selected-ids="form.agreementIds"
            :all-accepted="allAgreementsAccepted"
            :loading="loading"
            :show-action="!mainButtonAvailable"
            @toggle="toggleAgreement"
            @submit="acceptAgreement"
          />
        </motion.div>
      </AnimatePresence>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    </div>
    <footer class="onboarding-shell__locale"><LanguageControl /></footer>
  </main>
</template>
