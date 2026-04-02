<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

const status = ref<any>(null)
const loading = ref(true)
const changing = ref(false)
const result = ref('')

onMounted(async () => {
  try { status.value = await api.getIPStatus() } catch (e) {}
  loading.value = false
})

async function changeIP() {
  changing.value = true
  result.value = ''
  try {
    const resp = await api.changeIP()
    result.value = '✅ ' + resp.message
    status.value = await api.getIPStatus()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    result.value = '❌ ' + e.message
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('error')
  }
  changing.value = false
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">🔄 Change IP</h1>
      <p class="page-subtitle">Disconnect current connection to get a new IP</p>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <template v-else>
      <div class="card">
        <div class="row-between mb-md">
          <span class="text-muted">Cooldown Time</span>
          <span class="mono text-sm">{{ status?.cooldown_hours || 6 }} hours</span>
        </div>

        <div class="row-between mb-md" v-if="status?.last_change">
          <span class="text-muted">Last Changed</span>
          <span class="text-sm">{{ new Date(status.last_change).toLocaleString('en-US') }}</span>
        </div>

        <button class="btn btn-primary btn-block btn-lg" @click="changeIP" :disabled="changing || !status?.can_change">
          {{ changing ? 'Processing...' : status?.can_change ? '🔄 Change IP' : '⏳ Cooldown Active' }}
        </button>

        <div v-if="!status?.can_change && status?.next_available" class="text-center text-sm text-muted mt-sm">
          Available Time: {{ new Date(status.next_available).toLocaleString('en-US') }}
        </div>
      </div>

      <div v-if="result" class="card mt-md text-sm">{{ result }}</div>
    </template>
  </div>
</template>
