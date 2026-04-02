<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const squads = ref<any[]>([])
const loading = ref(true)
const updating = ref(false)
const currentSquadUUID = computed(() => userStore.currentExternalSquadUUID)

async function loadSquads() {
  try {
    squads.value = (await api.getExternalSquads()) || []
  } catch (e) {
    squads.value = []
  } finally {
    loading.value = false
  }
}

async function selectSquad(uuid: string) {
  if (updating.value || currentSquadUUID.value === uuid) {
    return
  }

  updating.value = true
  try {
    await api.updateExternalSquad(uuid)
    await userStore.refreshState({ background: true })
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    alert(e.message)
  } finally {
    updating.value = false
  }
}

onMounted(loadSquads)
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">Switch Route</h1>
      <p class="page-subtitle">Choose the external route group to use for your subscription</p>
    </div>

    <div class="loading-page" v-if="loading">
      <div class="loading-spinner"></div>
    </div>

    <div class="stack" v-else>
      <button
        v-for="squad in squads"
        :key="squad.uuid"
        class="card squad-card"
        :class="{ active: currentSquadUUID === squad.uuid }"
        @click="selectSquad(squad.uuid)"
        :disabled="updating"
      >
        <div class="stack-xs">
          <span class="squad-name">{{ squad.name }}</span>
          <span class="text-xs text-muted">UUID: {{ squad.uuid }}</span>
        </div>
        <span class="text-xs text-accent">{{ currentSquadUUID === squad.uuid ? 'Selected' : 'Select →' }}</span>
      </button>
    </div>

    <div v-if="squads.length === 0 && !loading" class="empty-state">
      <span class="empty-state-icon">🌐</span>
      <p class="empty-state-text">No route groups are available</p>
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

.squad-card.active {
  border-color: var(--accent-primary);
  background: rgba(108, 92, 231, 0.08);
}

.squad-name {
  font-weight: 600;
}

.stack-xs {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
</style>
