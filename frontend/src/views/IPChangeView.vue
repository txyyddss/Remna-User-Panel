<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ApiError, api } from '@/api'
import { useUserStore } from '@/stores/user'
import { useToast } from '@/composables/useToast'
import type { IPChangeStatus } from '@/types'

const userStore = useUserStore()
const toast = useToast()

const loading = ref(true)
const submitting = ref(false)
const subscription = ref('')
const reason = ref('')
const error = ref('')
const activeMessageLink = ref('')
const lookup = ref<IPChangeStatus | null>(null)
let refreshTimer: number | null = null

const statusLabel = computed(() => {
  switch (lookup.value?.status) {
    case 'PENDING':
      return 'Pending Votes'
    case 'CHANGING':
      return 'Changing'
    default:
      return 'Waiting'
  }
})

const statusDescription = computed(() => {
  switch (lookup.value?.status) {
    case 'PENDING':
      return `${lookup.value?.count || 0}/5 approvals recorded. The request is waiting in the Telegram group.`
    case 'CHANGING':
      return 'Voting passed. The upstream swap process is now in progress.'
    default:
      return 'No active IP replacement request is being processed right now.'
  }
})

const statusClass = computed(() => {
  switch (lookup.value?.status) {
    case 'PENDING':
      return 'pending'
    case 'CHANGING':
      return 'changing'
    default:
      return 'waiting'
  }
})

async function loadLookup() {
  try {
    lookup.value = await api.getIPLookup()
  } catch {
    lookup.value = { count: 0, status: 'WAITING' }
  } finally {
    loading.value = false
  }
}

async function submitRequest() {
  if (!subscription.value.trim() || !reason.value.trim()) {
    error.value = 'Subscription link and reason are required.'
    toast.error(error.value)
    return
  }

  submitting.value = true
  error.value = ''
  activeMessageLink.value = ''

  try {
    await api.changeIP({
      subscription: subscription.value.trim(),
      reason: reason.value.trim(),
    })
    subscription.value = ''
    reason.value = ''
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    toast.success('Request submitted. Wait for the Telegram group vote.')
    await loadLookup()
  } catch (e: unknown) {
    if (e instanceof ApiError) {
      error.value = e.message
      const data = e.data as { message_link?: string } | undefined
      activeMessageLink.value = data?.message_link || ''
    } else if (e instanceof Error) {
      error.value = e.message
    } else {
      error.value = 'Failed to submit IP change request.'
    }
    toast.error(error.value)
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  if (userStore.subKeys?.subscription_url) {
    subscription.value = userStore.subKeys.subscription_url
  }

  await loadLookup()
  refreshTimer = window.setInterval(() => {
    void loadLookup()
  }, 5000)
})

onUnmounted(() => {
  if (refreshTimer !== null) {
    window.clearInterval(refreshTimer)
    refreshTimer = null
  }
})
</script>

<template>
  <div class="page">
    <div class="page-header stagger-enter stagger-1">
      <div class="status-pill" :class="statusClass">{{ statusLabel }}</div>
      <h1 class="page-title">IP Change Queue</h1>
      <p class="page-subtitle">
        This follows the reference flow exactly: submit a subscription link and reason, wait for the Telegram vote, then wait for the swap callback.
      </p>
    </div>

    <div v-if="loading" class="loading-page">
      <div class="loading-spinner"></div>
    </div>

    <template v-else>
      <div class="ip-grid">
        <section class="card input-card stagger-enter stagger-2">
          <div class="section-label">Submit Request</div>
          <label class="field-label" for="subscription">Subscription Link</label>
          <input
            id="subscription"
            v-model="subscription"
            class="field-input"
            type="text"
            placeholder="https://sub.1391399.xyz/..."
            autocomplete="off"
          >

          <label class="field-label" for="reason">Reason</label>
          <textarea
            id="reason"
            v-model="reason"
            class="field-input field-textarea"
            rows="4"
            placeholder="Why do you want to replace the IP?"
          />

          <button
            class="btn btn-primary btn-block btn-lg submit-btn"
            :disabled="submitting"
            @click="submitRequest"
          >
            {{ submitting ? 'Submitting...' : 'Submit IP Change Request' }}
          </button>

          <p v-if="error" class="error-text">{{ error }}</p>
          <a
            v-if="activeMessageLink"
            class="thread-link"
            :href="activeMessageLink"
            target="_blank"
            rel="noreferrer"
          >
            Open Current Telegram Thread
          </a>
        </section>

        <section class="card status-card stagger-enter stagger-3">
          <div class="section-label">Live Queue State</div>
          <div class="status-orb" :class="statusClass">
            <span class="status-count">{{ lookup?.status === 'WAITING' ? '0' : lookup?.count || 0 }}</span>
            <span class="status-caption">{{ lookup?.status === 'CHANGING' ? 'in progress' : 'votes' }}</span>
          </div>

          <h2 class="status-title">{{ statusLabel }}</h2>
          <p class="status-text">{{ statusDescription }}</p>

          <div class="status-metrics">
            <div class="metric-box">
              <span class="metric-label">Votes</span>
              <strong class="metric-value">{{ lookup?.count || 0 }}/5</strong>
            </div>
            <div class="metric-box">
              <span class="metric-label">State</span>
              <strong class="metric-value">{{ lookup?.status || 'WAITING' }}</strong>
            </div>
          </div>
        </section>
      </div>

      <section class="card info-card stagger-enter stagger-4">
        <div class="section-label">Rules</div>
        <ul class="rule-list">
          <li>Only one active request can exist globally while status is `PENDING` or `CHANGING`.</li>
          <li>The request must pass the Telegram group vote before the swap phase starts.</li>
          <li>After a completed swap, the same subscription enters a 6-hour cooldown.</li>
          <li>The reference-only service squads are enforced before a request is accepted.</li>
        </ul>
      </section>
    </template>
  </div>
</template>

<style scoped>
.ip-grid {
  display: grid;
  gap: 20px;
}

.section-label {
  margin-bottom: 14px;
  font-size: 0.72rem;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--accent-secondary);
}

.status-pill {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  padding: 0 14px;
  border-radius: 999px;
  margin-bottom: 14px;
  font-size: 0.72rem;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  background: rgba(255, 255, 255, 0.06);
}

.status-pill.waiting {
  color: var(--text-secondary);
}

.status-pill.pending {
  color: #f0b45c;
}

.status-pill.changing {
  color: #32d2ae;
}

.input-card,
.status-card,
.info-card {
  border-color: rgba(255, 255, 255, 0.08);
}

.field-label {
  display: block;
  margin-bottom: 8px;
  color: var(--text-secondary);
  font-size: 0.82rem;
}

.field-input {
  width: 100%;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.04);
  border-radius: 16px;
  padding: 14px 16px;
  color: var(--text-primary);
  margin-bottom: 16px;
}

.field-input:focus {
  outline: none;
  border-color: rgba(var(--accent-primary-rgb), 0.45);
  box-shadow: 0 0 0 4px rgba(var(--accent-primary-rgb), 0.12);
}

.field-textarea {
  min-height: 120px;
  resize: vertical;
}

.submit-btn {
  margin-top: 4px;
}

.error-text {
  margin-top: 14px;
  color: #ff8d98;
  font-size: 0.86rem;
}

.thread-link {
  display: inline-flex;
  margin-top: 10px;
  color: var(--accent-secondary);
  text-decoration: none;
  font-size: 0.88rem;
}

.status-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.status-orb {
  width: 156px;
  height: 156px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  margin: 10px 0 18px;
  position: relative;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.status-orb::before {
  content: '';
  position: absolute;
  inset: 10px;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.status-orb.waiting {
  background: radial-gradient(circle at top, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0.03));
}

.status-orb.pending {
  background: radial-gradient(circle at top, rgba(240, 180, 92, 0.24), rgba(240, 180, 92, 0.06));
}

.status-orb.changing {
  background: radial-gradient(circle at top, rgba(50, 210, 174, 0.26), rgba(50, 210, 174, 0.08));
}

.status-count {
  font-family: var(--font-display);
  font-size: 2rem;
  font-weight: 700;
}

.status-caption {
  display: block;
  font-size: 0.76rem;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--text-secondary);
}

.status-title {
  font-size: 1.2rem;
  margin-bottom: 10px;
}

.status-text {
  color: var(--text-secondary);
  max-width: 34ch;
}

.status-metrics {
  width: 100%;
  display: grid;
  gap: 12px;
  margin-top: 18px;
}

.metric-box {
  padding: 14px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.04);
}

.metric-label {
  display: block;
  color: var(--text-muted);
  font-size: 0.72rem;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  margin-bottom: 6px;
}

.metric-value {
  font-size: 1rem;
}

.rule-list {
  list-style: none;
  padding: 0;
}

.rule-list li {
  position: relative;
  padding-left: 20px;
  color: var(--text-secondary);
  line-height: 1.7;
}

.rule-list li + li {
  margin-top: 10px;
}

.rule-list li::before {
  content: '•';
  position: absolute;
  left: 0;
  color: var(--accent-primary);
}

@media (min-width: 900px) {
  .ip-grid {
    grid-template-columns: minmax(0, 1.15fr) minmax(320px, 0.85fr);
    align-items: start;
  }
}
</style>
