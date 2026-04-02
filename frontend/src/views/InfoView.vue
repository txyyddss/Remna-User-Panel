<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import { api } from '@/api'

const userStore = useUserStore()
const bandwidth = ref<any[]>([])
const devices = ref<any[]>([])
const history = ref<any[]>([])
const loading = ref(true)
const activeTab = ref('bandwidth')

function formatBytes(b: number): string {
  if (b < 1073741824) return `${(b / 1048576).toFixed(2)} MB`
  return `${(b / 1073741824).toFixed(2)} GB`
}

const nodeAggregated = computed(() => {
  const map: Record<string, { name: string; country: string; total: number }> = {}
  for (const item of bandwidth.value) {
    if (!map[item.nodeUuid]) {
      map[item.nodeUuid] = { name: item.nodeName, country: item.countryCode, total: 0 }
    }
    map[item.nodeUuid].total += item.total
  }
  return Object.values(map).sort((a, b) => b.total - a.total)
})

const totalBandwidth = computed(() => nodeAggregated.value.reduce((sum, n) => sum + n.total, 0))

onMounted(async () => {
  try {
    const [bw, dev, hist] = await Promise.allSettled([
      api.getBandwidth(),
      api.getDevices(),
      api.getSubHistory(),
    ])
    if (bw.status === 'fulfilled') bandwidth.value = bw.value || []
    if (dev.status === 'fulfilled') devices.value = dev.value || []
    if (hist.status === 'fulfilled') history.value = hist.value || []
  } catch (e) {}
  loading.value = false
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">📊 使用信息</h1>
    </div>

    <!-- Tabs -->
    <div class="tabs">
      <button class="tab" :class="{ active: activeTab === 'bandwidth' }" @click="activeTab = 'bandwidth'">流量</button>
      <button class="tab" :class="{ active: activeTab === 'devices' }" @click="activeTab = 'devices'">设备</button>
      <button class="tab" :class="{ active: activeTab === 'history' }" @click="activeTab = 'history'">历史</button>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <!-- Bandwidth Tab -->
    <div v-if="!loading && activeTab === 'bandwidth'" class="stack mt-md">
      <div class="card">
        <h3 class="mb-sm">总流量使用</h3>
        <div class="stat-value">{{ formatBytes(totalBandwidth) }}</div>
        <span class="text-xs text-muted">过去30天</span>
      </div>

      <div class="card" v-if="nodeAggregated.length > 0">
        <h4 class="mb-md">节点使用分布</h4>
        <div v-for="(node, i) in nodeAggregated" :key="i" class="node-item">
          <div class="node-header">
            <span class="text-sm">{{ node.country?.toUpperCase() }} · {{ node.name }}</span>
            <span class="mono text-sm">{{ formatBytes(node.total) }}</span>
          </div>
          <div class="progress" style="height:4px">
            <div class="progress-bar" :style="{ width: (node.total / totalBandwidth * 100) + '%' }"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- Devices Tab -->
    <div v-if="!loading && activeTab === 'devices'" class="stack mt-md">
      <div v-for="dev in devices" :key="dev.hwid" class="card">
        <div class="row-between">
          <div>
            <div class="text-sm">{{ dev.platform || '未知' }} {{ dev.deviceModel || '' }}</div>
            <div class="text-xs text-muted">{{ dev.osVersion || '' }}</div>
          </div>
          <span class="badge badge-success">在线</span>
        </div>
      </div>
      <div v-if="devices.length === 0" class="empty-state">
        <span class="empty-state-icon">📱</span>
        <p class="empty-state-text">暂无设备记录</p>
      </div>
    </div>

    <!-- History Tab -->
    <div v-if="!loading && activeTab === 'history'" class="stack mt-md">
      <div v-for="h in history" :key="h.id" class="card">
        <div class="row-between">
          <div>
            <div class="text-sm truncate" style="max-width:200px">{{ h.userAgent }}</div>
            <div class="text-xs text-muted">{{ h.ip }}</div>
          </div>
          <span class="text-xs text-muted">{{ new Date(h.createdAt).toLocaleString('zh-CN') }}</span>
        </div>
      </div>
      <div v-if="history.length === 0" class="empty-state">
        <span class="empty-state-icon">📜</span>
        <p class="empty-state-text">暂无订阅请求记录</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tabs {
  display: flex;
  background: var(--bg-glass);
  border-radius: var(--radius-md);
  padding: 4px;
  gap: 4px;
}

.tab {
  flex: 1;
  padding: var(--space-sm);
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-family: var(--font-body);
  font-size: 0.875rem;
  font-weight: 500;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s;
}

.tab.active {
  background: var(--bg-card);
  color: var(--text-primary);
}

.node-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  padding: var(--space-sm) 0;
  border-bottom: 1px solid var(--border-subtle);
}

.node-item:last-child {
  border-bottom: none;
}

.node-header {
  display: flex;
  justify-content: space-between;
}
</style>
