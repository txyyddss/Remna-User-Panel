<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

const squads = ref<any[]>([])
const loading = ref(true)
const updating = ref(false)

async function selectSquad(uuid: string) {
  updating.value = true
  try {
    await api.updateExternalSquad(uuid)
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    alert(e.message)
  }
  updating.value = false
}

onMounted(async () => {
  try { squads.value = (await api.getExternalSquads()) || [] } catch (e) {}
  loading.value = false
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">🌐 线路选择</h1>
      <p class="page-subtitle">切换外部线路组</p>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <div class="stack" v-else>
      <button v-for="squad in squads" :key="squad.uuid" class="card squad-card" @click="selectSquad(squad.uuid)" :disabled="updating">
        <span class="squad-name">{{ squad.name }}</span>
        <span class="text-xs text-accent">选择 →</span>
      </button>
    </div>

    <div v-if="squads.length === 0 && !loading" class="empty-state">
      <span class="empty-state-icon">🌐</span>
      <p class="empty-state-text">暂无可用线路</p>
    </div>
  </div>
</template>

<style scoped>
.squad-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  border: none;
  background: var(--bg-card);
  text-align: left;
  font-family: var(--font-body);
  color: var(--text-primary);
}

.squad-name {
  font-weight: 500;
}
</style>
