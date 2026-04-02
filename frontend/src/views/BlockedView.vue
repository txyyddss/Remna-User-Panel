<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const toast = useToast()

const isGroupBlock = computed(() => route.query.reason === 'group')
const actionError = ref('')
const loadingChannel = ref(false)
const loadingGroup = ref(false)

const access = computed(() => userStore.miniAppAccess)
const title = computed(() =>
  isGroupBlock.value ? 'Group Access Verification' : 'Telegram Mini App Only',
)
const message = computed(() =>
  isGroupBlock.value
    ? 'Finish the same channel -> one-time invite -> group verification path before the panel unlocks.'
    : 'Open this panel from the Telegram mini app to continue.',
)

const steps = computed(() => {
  const current = access.value
  return [
    {
      id: 'channel',
      title: 'Join The Channel',
      detail: 'Open the official notice channel first. Group verification stays locked until this passes.',
      done: !!current?.channel_joined,
    },
    {
      id: 'invite',
      title: 'Generate Invite Link',
      detail: current?.invite_link
        ? 'A one-time invite link is ready. Use it to enter the group.'
        : 'After channel verification, the backend will mint a single-use invite link for you.',
      done: !!current?.invite_link,
    },
    {
      id: 'group',
      title: 'Authorize Group Membership',
      detail: 'After joining the group, come back here and verify membership to unlock the mini app.',
      done: !!current?.group_joined,
    },
  ]
})

async function ensureAccessLoaded() {
  if (!isGroupBlock.value || access.value) {
    return
  }

  try {
    const data = await userStore.bootstrapMiniAppAccess()
    if (data.group_joined) {
      await userStore.refreshState()
      userStore.startAutoRefresh()
      router.replace('/')
    }
  } catch (e: any) {
    actionError.value = e.message || 'Failed to load access state.'
  }
}

async function handleChannelVerify() {
  if (!isGroupBlock.value) {
    return
  }

  loadingChannel.value = true
  actionError.value = ''
  try {
    const data = await userStore.verifyMiniAppChannel()
    if (data.invite_link) {
      toast.success('One-time invite link generated.')
    }
  } catch (e: any) {
    actionError.value = e.message || 'Failed to verify channel membership.'
    toast.error(actionError.value)
  } finally {
    loadingChannel.value = false
  }
}

async function handleGroupVerify() {
  if (!isGroupBlock.value) {
    return
  }

  loadingGroup.value = true
  actionError.value = ''
  try {
    const data = await userStore.verifyMiniAppGroup()
    if (!data.group_joined) {
      throw new Error('Group membership is still not verified.')
    }

    await userStore.refreshState()
    if (userStore.error) {
      throw new Error(userStore.error)
    }
    userStore.startAutoRefresh()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    toast.success('Group membership verified.')
    router.replace('/')
  } catch (e: any) {
    actionError.value = e.message || 'Failed to verify group membership.'
    toast.error(actionError.value)
  } finally {
    loadingGroup.value = false
  }
}

onMounted(() => {
  void ensureAccessLoaded()
})
</script>

<template>
  <div class="blocked-page">
    <div class="blocked-backdrop"></div>
    <div class="blocked-grid">
      <section class="blocked-hero card stagger-enter stagger-1">
        <div class="hero-kicker">{{ isGroupBlock ? 'Access Policy' : 'Launch Required' }}</div>
        <h1 class="hero-title">{{ title }}</h1>
        <p class="hero-text">{{ message }}</p>
        <p class="hero-subtext" v-if="isGroupBlock">
          This mini app now follows the full gate used in the reference flow. No shortcut path is available.
        </p>
      </section>

      <section v-if="isGroupBlock" class="flow-shell card stagger-enter stagger-2">
        <div class="flow-header">
          <div>
            <div class="flow-label">Guided Entry</div>
            <h2 class="flow-title">Join Sequence</h2>
          </div>
          <div class="flow-state" :class="{ ready: access?.group_joined }">
            {{ access?.group_joined ? 'Unlocked' : 'Locked' }}
          </div>
        </div>

        <div class="step-list">
          <article
            v-for="(step, index) in steps"
            :key="step.id"
            class="step-card"
            :class="{ done: step.done }"
          >
            <div class="step-index">0{{ index + 1 }}</div>
            <div class="step-copy">
              <h3 class="step-title">{{ step.title }}</h3>
              <p class="step-detail">{{ step.detail }}</p>
            </div>
            <div class="step-badge">{{ step.done ? 'Done' : 'Open' }}</div>
          </article>
        </div>

        <div class="action-stack">
          <a
            class="action-btn action-btn-secondary"
            v-if="access?.channel_url"
            :href="access.channel_url"
            target="_blank"
            rel="noreferrer"
          >
            Open Channel
          </a>

          <button
            class="action-btn action-btn-primary"
            :disabled="loadingChannel"
            @click="handleChannelVerify"
          >
            {{ loadingChannel ? 'Checking Channel...' : access?.invite_link ? 'Regenerate Invite Link' : 'I Joined The Channel' }}
          </button>

          <a
            v-if="access?.invite_link"
            class="action-btn action-btn-accent"
            :href="access.invite_link"
            target="_blank"
            rel="noreferrer"
          >
            Enter Group
          </a>

          <button
            class="action-btn action-btn-primary"
            :disabled="loadingGroup || !access?.invite_link"
            @click="handleGroupVerify"
          >
            {{ loadingGroup ? 'Verifying Group...' : 'I Joined The Group' }}
          </button>
        </div>

        <div v-if="actionError" class="error-box">
          {{ actionError }}
        </div>
      </section>

      <section v-else class="miniapp-card card stagger-enter stagger-2">
        <p class="miniapp-copy">
          Telegram init data is required for authentication, so browser-only visits stay blocked here.
        </p>
      </section>
    </div>
  </div>
</template>

<style scoped>
.blocked-page {
  min-height: 100dvh;
  padding: clamp(20px, 4vw, 40px);
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(circle at top left, rgba(232, 88, 37, 0.16), transparent 34%),
    radial-gradient(circle at bottom right, rgba(34, 190, 165, 0.2), transparent 32%),
    linear-gradient(160deg, #0d1116 0%, #121920 48%, #0b0f13 100%);
}

.blocked-backdrop {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.04) 1px, transparent 1px);
  background-size: 44px 44px;
  mask-image: linear-gradient(to bottom, rgba(0, 0, 0, 0.85), transparent);
  pointer-events: none;
}

.blocked-grid {
  position: relative;
  z-index: 1;
  max-width: 980px;
  margin: 0 auto;
  display: grid;
  gap: 20px;
}

.blocked-hero,
.flow-shell,
.miniapp-card {
  border-color: rgba(255, 255, 255, 0.08);
  background: rgba(7, 10, 14, 0.78);
  backdrop-filter: blur(16px);
}

.hero-kicker,
.flow-label {
  font-size: 0.72rem;
  letter-spacing: 0.24em;
  text-transform: uppercase;
  color: #f2a65a;
  margin-bottom: 12px;
}

.hero-title,
.flow-title,
.step-title {
  font-family: 'JetBrains Mono', monospace;
}

.hero-title {
  font-size: clamp(1.8rem, 5vw, 3.6rem);
  line-height: 0.92;
  margin-bottom: 16px;
  text-transform: uppercase;
}

.hero-text,
.hero-subtext,
.step-detail,
.miniapp-copy {
  color: var(--text-secondary);
}

.hero-subtext {
  margin-top: 12px;
  max-width: 48ch;
}

.flow-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.flow-state {
  min-width: 96px;
  text-align: center;
  padding: 10px 14px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.12em;
  font-size: 0.72rem;
}

.flow-state.ready {
  color: #1fd1a6;
  border-color: rgba(31, 209, 166, 0.4);
}

.step-list {
  display: grid;
  gap: 12px;
}

.step-card {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 14px;
  align-items: flex-start;
  padding: 16px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.step-card.done {
  border-color: rgba(31, 209, 166, 0.35);
  background: rgba(31, 209, 166, 0.08);
}

.step-index {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  font-family: 'JetBrains Mono', monospace;
  font-weight: 700;
  background: rgba(242, 166, 90, 0.12);
  color: #f2a65a;
}

.step-title {
  margin-bottom: 6px;
  font-size: 1rem;
}

.step-detail {
  font-size: 0.88rem;
  line-height: 1.6;
}

.step-badge {
  padding: 8px 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.06);
  font-size: 0.74rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--text-secondary);
}

.action-stack {
  display: grid;
  gap: 12px;
  margin-top: 20px;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 52px;
  border-radius: 16px;
  padding: 0 18px;
  text-decoration: none;
  border: 1px solid transparent;
  transition: transform 0.2s ease, border-color 0.2s ease, opacity 0.2s ease;
  font-weight: 700;
}

.action-btn:hover {
  transform: translateY(-1px);
}

.action-btn:disabled {
  opacity: 0.6;
  transform: none;
}

.action-btn-primary {
  background: linear-gradient(135deg, #e85825 0%, #f2a65a 100%);
  color: #120b08;
}

.action-btn-secondary {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
  border-color: rgba(255, 255, 255, 0.1);
}

.action-btn-accent {
  background: linear-gradient(135deg, #1fd1a6 0%, #67d4ff 100%);
  color: #08161a;
}

.error-box {
  margin-top: 16px;
  padding: 14px 16px;
  border-radius: 14px;
  background: rgba(255, 107, 122, 0.08);
  border: 1px solid rgba(255, 107, 122, 0.24);
  color: #ff8d98;
  font-size: 0.88rem;
}

@media (max-width: 700px) {
  .step-card {
    grid-template-columns: 1fr;
  }

  .step-badge {
    justify-self: start;
  }

  .flow-header {
    flex-direction: column;
  }

  .flow-state {
    min-width: 0;
  }
}
</style>
