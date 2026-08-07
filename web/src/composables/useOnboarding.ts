import { computed, onMounted, reactive, readonly, shallowRef, watch } from 'vue'
import { useRouter } from 'vue-router'

import { api, ApiError } from '@/api/client'
import type { InviteLink, MembershipState, OnboardingStep } from '@/api/types'
import { useSessionStore } from '@/stores/session'
import { notifyHaptic, openExternalLink } from '@/utils/telegram'

function normalizeStep(step?: OnboardingStep): OnboardingStep {
  if (!step || step === 'complete') return 'membership'
  return step
}

export function useOnboarding() {
  const router = useRouter()
  const sessionStore = useSessionStore()
  const step = shallowRef<OnboardingStep>(normalizeStep(sessionStore.user?.onboardingState))
  const loading = shallowRef(false)
  const error = shallowRef<string | null>(null)
  const invites = shallowRef<InviteLink[]>([])
  const membership = shallowRef<MembershipState | null>(null)
  const form = reactive({ username: '', agreement: false })

  const progress = computed(() => {
    const order: OnboardingStep[] = ['intro', 'membership', 'username', 'agreement', 'complete']
    return Math.max(0, order.indexOf(step.value)) / (order.length - 1)
  })

  const usernameValid = computed(() => /^[a-z]{3,9}$/.test(form.username))
  const usernameHint = computed(() => {
    if (!form.username) return 'Use 3-9 lowercase English letters.'
    return usernameValid.value ? 'This format looks good.' : 'Only 3-9 lowercase English letters are allowed.'
  })

  async function run<T>(task: () => Promise<T>): Promise<T | undefined> {
    loading.value = true
    error.value = null
    try {
      return await task()
    } catch (caught) {
      if (caught instanceof ApiError && caught.code === 'USERNAME_UNAVAILABLE') {
        error.value = 'That username is already in use. Try another one.'
      } else {
        error.value = caught instanceof Error ? caught.message : 'Something went wrong. Please try again.'
      }
      notifyHaptic('error')
      return undefined
    } finally {
      loading.value = false
    }
  }

  async function loadInvites(): Promise<void> {
    const response = await run(() => api.createInvites())
    if (response) invites.value = response.invites
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
      error.value = 'Join both spaces, then check again.'
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

  async function acceptAgreement(): Promise<void> {
    if (!form.agreement) return
    const response = await run(() => api.acceptAgreement())
    if (!response) return
    sessionStore.updateSession(response)
    notifyHaptic('success')
    step.value = 'complete'
    await router.replace('/home')
  }

  watch(step, (next) => {
    error.value = null
    if (next === 'membership' && invites.value.length === 0) void loadInvites()
  })

  onMounted(() => {
    if (step.value === 'membership') void loadInvites()
  })

  return {
    step: readonly(step),
    progress,
    loading: readonly(loading),
    error: readonly(error),
    invites: readonly(invites),
    membership: readonly(membership),
    form,
    usernameValid,
    usernameHint,
    finishIntro,
    loadInvites,
    openInvite,
    checkMembership,
    submitUsername,
    acceptAgreement,
  }
}
