<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const status = ref<any>(null)
const loading = ref(true)
const changing = ref(false)
const result = ref('')
const subscriptionInput = ref('')

const suggestedSubscription = computed(() => userStore.subKeys?.subscription_url || '')

async function loadStatus() {
  try {
    status.value = await api.getIPStatus()
  } catch {
    status.value = null
  } finally {
    loading.value = false
  }
}

async function changeIP() {
  changing.value = true
  result.value = ''
  try {
    const subscription = subscriptionInput.value.trim() || suggestedSubscription.value
    const resp = await api.changeIP({ subscription })
    result.value = resp.message
    await loadStatus()
  } catch (e: any) {
    result.value = e.message
  } finally {
    changing.value = false
  }
}

onMounted(async () => {
  await userStore.refreshSubInfo()
  subscriptionInput.value = suggestedSubscription.value
  await loadStatus()
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">Change IP</h1>
      <p class="page-subtitle">Standalone reconnect tool. Use your subscription link or short UUID if needed.</p>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <div v-else class="stack">
      <div class="card">
        <div class="row-between mb-md">
          <span class="text-muted">Cooldown</span>
          <span class="mono">{{ status?.cooldown_hours || 6 }} hours</span>
        </div>

        <div class="row-between mb-md" v-if="status?.last_change">
          <span class="text-muted">Last change</span>
          <span class="text-sm">{{ new Date(status.last_change).toLocaleString('en-US') }}</span>
        </div>

        <label class="field-label">Subscription link or short UUID</label>
        <input class="input" v-model="subscriptionInput" placeholder="Leave filled to use your current subscription" />

        <button class="btn btn-primary btn-block btn-lg mt-md" @click="changeIP" :disabled="changing || !status?.can_change">
          {{ changing ? 'Disconnecting...' : status?.can_change ? 'Disconnect Current Sessions' : 'Cooldown Active' }}
        </button>

        <p class="text-xs text-muted mt-sm">This action drops current Remnawave connections so the next reconnect gets a fresh exit IP.</p>

        <div v-if="!status?.can_change && status?.next_available" class="text-sm text-muted mt-md">
          Next available time: {{ new Date(status.next_available).toLocaleString('en-US') }}
        </div>
      </div>

      <div v-if="result" class="card text-sm">{{ result }}</div>
    </div>
  </div>
</template>

<style scoped>
.field-label {
  display: block;
  margin-bottom: var(--space-sm);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--text-muted);
}
</style>
