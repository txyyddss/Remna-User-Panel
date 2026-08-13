import { computed, onMounted, reactive, readonly, shallowRef, watch } from 'vue'
import { useRouter } from 'vue-router'

import { api, ApiError } from '@/api/client'
import { featuresApi, type PublishedOnboarding } from '@/api/features'
import type { InviteLink, MembershipState, OnboardingStep } from '@/api/types'
import { useI18n } from '@/i18n'
import { useSessionStore } from '@/stores/session'
import { notifyHaptic, openExternalLink } from '@/utils/telegram'
import { isValid, usernameSchema } from '@/utils/validation'

function normalizeStep(step?: OnboardingStep): OnboardingStep {
  if (!step || step === 'complete') return 'membership'
  return step
}

type ErrorRecovery = (caught: unknown) => Promise<boolean>

export function useOnboarding() {
  const router = useRouter()
  const sessionStore = useSessionStore()
  const { locale, t } = useI18n()
  const step = shallowRef<OnboardingStep>(normalizeStep(sessionStore.user?.onboardingState))
  const loading = shallowRef(false)
  const error = shallowRef<string | null>(null)
  const invites = shallowRef<InviteLink[]>([])
  const membership = shallowRef<MembershipState | null>(null)
  const content = shallowRef<PublishedOnboarding | null>(null)
  const form = reactive({ username: '', agreementIds: [] as string[] })

  const progress = computed(() => {
    const order: OnboardingStep[] = ['intro', 'membership', 'username', 'agreement', 'complete']
    return Math.max(0, order.indexOf(step.value)) / (order.length - 1)
  })

  const usernameValid = computed(() => isValid(usernameSchema, form.username))
  const allAgreementsAccepted = computed(() => Boolean(content.value?.agreements.length)
    && content.value!.agreements.every((agreement) => form.agreementIds.includes(agreement.id)))
  const usernameHint = computed(() => {
    if (!form.username) return t('onboarding.useLowercase')
    return usernameValid.value ? t('onboarding.formatGood') : t('onboarding.formatInvalid')
  })

  async function run<T>(task: () => Promise<T>, recover?: ErrorRecovery): Promise<T | undefined> {
    loading.value = true
    error.value = null
    try {
      return await task()
    } catch (caught) {
      let recovered = false
      if (recover) {
        try {
          recovered = await recover(caught)
        } catch {
          recovered = false
        }
      }
      if (recovered) return undefined
      if (caught instanceof ApiError && caught.code === 'USERNAME_UNAVAILABLE') {
        error.value = t('onboarding.usernameUnavailable')
      } else {
        const key = caught instanceof ApiError ? `errors.api.${caught.code}` : ''
        const translated = key ? t(key) : ''
        error.value = translated && translated !== key ? translated : t('onboarding.somethingWrong')
      }
      notifyHaptic('error')
      return undefined
    } finally {
      loading.value = false
    }
  }

  async function loadInvites(): Promise<void> {
    const response = await run(() => api.createInvites())
    if (response) invites.value = response.invites.map((invite) => ({
      ...invite,
      label: t(invite.kind === 'group' ? 'onboarding.privateGroup' : 'onboarding.updatesChannel'),
    }))
  }

  async function loadContent(): Promise<void> {
    const response = await run(() => featuresApi.getPublishedOnboarding(locale.value))
    if (!response) return
    content.value = response
    form.agreementIds = form.agreementIds.filter((id) => response.agreements.some((agreement) => agreement.id === id))
  }

  function finishIntro(): void {
    step.value = 'membership'
  }

  function openInvite(invite: InviteLink): void {
    openExternalLink(invite.url)
  }

  async function checkMembership(): Promise<void> {
    const response = await run(() => api.checkMembership())
    if (!response) return
    membership.value = response
    sessionStore.updateSession(response.session)
    invites.value = invites.value.map((invite) => ({
      ...invite,
      joined: invite.kind === 'group' ? response.groupJoined : response.channelJoined,
    }))
    if (response.complete) {
      notifyHaptic('success')
      step.value = 'username'
    } else {
      error.value = t('onboarding.joinBoth')
    }
  }

  async function submitUsername(): Promise<void> {
    if (!usernameValid.value) return
    const response = await run(() => api.setUsername(form.username))
    if (!response) return
    sessionStore.updateSession(response)
    notifyHaptic('success')
    step.value = 'agreement'
  }

  async function refreshAgreementState(): Promise<boolean> {
    const session = await api.getMe()
    sessionStore.updateSession(session)
    if (session.user.onboardingState === 'complete') {
      notifyHaptic('success')
      step.value = 'complete'
      await router.replace('/home')
      return true
    }
    if (session.user.onboardingState !== 'agreement') {
      step.value = normalizeStep(session.user.onboardingState)
      return false
    }
    form.agreementIds = []
    const currentContent = await featuresApi.getPublishedOnboarding(locale.value)
    content.value = currentContent
    return false
  }

  async function recoverAgreementConflict(caught: unknown): Promise<boolean> {
    if (!(caught instanceof ApiError)) return false
    if (!['AGREEMENT_OUTDATED', 'ONBOARDING_CONFLICT', 'ONBOARDING_STATE_CONFLICT'].includes(caught.code)) return false
    return refreshAgreementState()
  }

  async function acceptAgreement(): Promise<void> {
    if (!content.value || !allAgreementsAccepted.value) return
    const response = await run(
      () => api.acceptAgreement(content.value!.agreementRevision, [...form.agreementIds]),
      recoverAgreementConflict,
    )
    if (!response) return
    sessionStore.updateSession(response)
    notifyHaptic('success')
    step.value = 'complete'
    await router.replace('/home')
  }

  function toggleAgreement(id: string): void {
    form.agreementIds = form.agreementIds.includes(id)
      ? form.agreementIds.filter((value) => value !== id)
      : [...form.agreementIds, id]
  }

  watch(step, (next) => {
    error.value = null
    if (next === 'membership' && invites.value.length === 0) void loadInvites()
  })

  watch(locale, () => void loadContent())

  onMounted(() => {
    void loadContent()
    if (step.value === 'membership') void loadInvites()
  })

  return {
    step: readonly(step),
    progress,
    loading: readonly(loading),
    error: readonly(error),
    invites: readonly(invites),
    membership: readonly(membership),
    content: readonly(content),
    form,
    usernameValid,
    usernameHint,
    allAgreementsAccepted,
    finishIntro,
    loadInvites,
    openInvite,
    checkMembership,
    submitUsername,
    acceptAgreement,
    toggleAgreement,
  }
}
