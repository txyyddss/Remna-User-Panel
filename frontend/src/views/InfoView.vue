<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api'

const bandwidth = ref<any[]>([])
const devices = ref<any[]>([])
const history = ref<any[]>([])
const loading = ref(true)
const activeTab = ref<'bandwidth' | 'devices' | 'history'>('bandwidth')

function formatBytes(bytes: number): string {
  if (!bytes) {
    return '0 B'
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return `${(bytes / 1024 ** index).toFixed(index === 0 ? 0 : 2)} ${units[index]}`
}

const nodeAggregated = computed(() => {
  const map: Record<string, { name: string; country: string; total: number }> = {}
  for (const item of bandwidth.value) {
    const key = item.nodeUuid || item.nodeUUID || item.nodeName
    if (!key) {
      continue
    }
    if (!map[key]) {
      map[key] = {
        name: item.nodeName || 'Unknown node',
        country: item.countryCode || '--',
        total: 0,
      }
    }
    map[key].total += Number(item.total || 0)
  }
  return Object.values(map).sort((a, b) => b.total - a.total)
})

const totalBandwidth = computed(() => nodeAggregated.value.reduce((sum, node) => sum + node.total, 0))

onMounted(async () => {
  try {
    const [bw, dev, hist] = await Promise.allSettled([
      api.getBandwidth(),
      api.getDevices(),
      api.getSubHistory(),
    ])

    if (bw.status === 'fulfilled') {
      bandwidth.value = bw.value || []
    }
    if (dev.status === 'fulfilled') {
      devices.value = dev.value || []
    }
    if (hist.status === 'fulfilled') {
      history.value = hist.value || []
    }
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">Usage</h1>
      <p class="page-subtitle">Traffic by node, active devices, and subscription fetch history.</p>
    </div>

    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'bandwidth' }" @click="activeTab = 'bandwidth'">Traffic</button>
      <button class="tab" :class="{ active: activeTab === 'devices' }" @click="activeTab = 'devices'">Devices</button>
      <button class="tab" :class="{ active: activeTab === 'history' }" @click="activeTab = 'history'">History</button>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <div v-else-if="activeTab === 'bandwidth'" class="stack mt-md">
      <div class="card">
        <h3 class="mb-sm">30-Day Traffic</h3>
        <div class="stat-value">{{ formatBytes(totalBandwidth) }}</div>
        <p class="text-sm text-muted mt-sm">Aggregated from Remnawave bandwidth stats.</p>
      </div>

      <div class="card" v-if="nodeAggregated.length">
        <h3 class="mb-md">Node Distribution</h3>
        <div v-for="node in nodeAggregated" :key="`${node.country}-${node.name}`" class="node-item">
          <div class="row-between text-sm">
            <span>{{ node.country.toUpperCase() }} · {{ node.name }}</span>
            <span class="mono">{{ formatBytes(node.total) }}</span>
          </div>
          <div class="progress mt-sm">
            <div class="progress-bar" :style="{ width: `${totalBandwidth ? (node.total / totalBandwidth) * 100 : 0}%` }"></div>
          </div>
        </div>
      </div>

      <div v-else class="card text-sm text-muted">No traffic usage has been recorded recently.</div>
    </div>

    <div v-else-if="activeTab === 'devices'" class="stack mt-md">
      <div v-for="device in devices" :key="device.hwid" class="card">
        <div class="row-between">
          <div>
            <div class="text-sm">{{ device.platform || 'Unknown platform' }} {{ device.deviceModel || '' }}</div>
            <div class="text-xs text-muted">{{ device.osVersion || device.userAgent || 'No extra details' }}</div>
          </div>
          <span class="badge badge-success">{{ device.hwid ? 'Bound' : 'Active' }}</span>
        </div>
      </div>

      <div v-if="devices.length === 0" class="card text-sm text-muted">No hardware device records found.</div>
    </div>

    <div v-else class="stack mt-md">
      <div v-for="item in history" :key="item.id" class="card">
        <div class="row-between">
          <div>
            <div class="text-sm">{{ item.userAgent || 'Unknown user agent' }}</div>
            <div class="text-xs text-muted">{{ item.ip || 'Unknown IP' }}</div>
          </div>
          <span class="text-xs text-muted">{{ new Date(item.createdAt).toLocaleString('en-US') }}</span>
        </div>
      </div>

      <div v-if="history.length === 0" class="card text-sm text-muted">No subscription request history yet.</div>
    </div>
  </div>
</template>

<style scoped>
.tabs {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 4px;
  padding: 4px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.03);
}

.tab {
  min-height: 42px;
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
}

.tab.active {
  background: rgba(91, 141, 239, 0.16);
  color: var(--text-primary);
}

.node-item + .node-item {
  margin-top: var(--space-md);
}
</style>
