<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

const config = ref<any>(null)
const internalSquads = ref<any[]>([])
const loading = ref(true)
const showComboForm = ref(false)
const comboForm = ref({
  name: '', description: '', squad_uuid: '', traffic_gb: 100,
  strategy: 'MONTH', cycle: 'monthly', price_rmb: 10, reset_price: 5
})

onMounted(async () => {
  try {
    config.value = await api.getConfig()
    internalSquads.value = (await api.getInternalSquads()) || []
  } catch (e) {}
  loading.value = false
})

async function createCombo() {
  try {
    await api.createCombo(comboForm.value)
    showComboForm.value = false
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    alert(e.message)
  }
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">⚙️ 管理面板</h1>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <template v-else>
      <!-- Config Overview -->
      <div class="card">
        <h3 class="mb-md">配置概览</h3>
        <div v-if="config" class="config-grid">
          <div class="config-item" v-for="(val, key) in config" :key="key">
            <span class="text-xs text-muted">{{ key }}</span>
            <pre class="config-value text-xs">{{ JSON.stringify(val, null, 2) }}</pre>
          </div>
        </div>
      </div>

      <!-- Create Combo -->
      <div class="card mt-md">
        <div class="row-between">
          <h3>套餐管理</h3>
          <button class="btn btn-sm btn-primary" @click="showComboForm = !showComboForm">
            {{ showComboForm ? '取消' : '+ 新建' }}
          </button>
        </div>

        <div v-if="showComboForm" class="stack-sm mt-md">
          <input class="input" v-model="comboForm.name" placeholder="套餐名称" />
          <input class="input" v-model="comboForm.description" placeholder="描述" />
          <select class="input" v-model="comboForm.squad_uuid">
            <option value="">选择线路组</option>
            <option v-for="s in internalSquads" :key="s.uuid" :value="s.uuid">{{ s.name }}</option>
          </select>
          <div class="grid-2">
            <input class="input" v-model.number="comboForm.traffic_gb" type="number" placeholder="流量(GB)" />
            <select class="input" v-model="comboForm.strategy">
              <option value="NO_RESET">不重置</option>
              <option value="DAY">日</option>
              <option value="WEEK">周</option>
              <option value="MONTH">月</option>
            </select>
          </div>
          <div class="grid-2">
            <input class="input" v-model.number="comboForm.price_rmb" type="number" placeholder="价格(¥)" step="0.01" />
            <input class="input" v-model.number="comboForm.reset_price" type="number" placeholder="重置价格(¥)" step="0.01" />
          </div>
          <select class="input" v-model="comboForm.cycle">
            <option value="monthly">月付</option>
            <option value="quarterly">季付</option>
            <option value="semiannual">半年付</option>
            <option value="annual">年付</option>
          </select>
          <button class="btn btn-primary btn-block" @click="createCombo">创建套餐</button>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.config-grid {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.config-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: var(--space-sm);
  background: var(--bg-glass);
  border-radius: var(--radius-sm);
}

.config-value {
  font-family: var(--font-display);
  font-size: 0.6875rem;
  color: var(--text-secondary);
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}
</style>
