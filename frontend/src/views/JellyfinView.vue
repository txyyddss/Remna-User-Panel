<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { api } from '@/api'

const userStore = useUserStore()
const loading = ref(true)
const qcCode = ref('')
const qcLoading = ref(false)
const qcMessage = ref('')
const parentalRating = ref(0)
const devices = ref<any[]>([])
const showPwdForm = ref(false)
const currentPwd = ref('')
const newPwd = ref('')

onMounted(async () => {
  try {
    if (userStore.hasJellyfin) {
      const devResp = await api.jellyfinGetDevices()
      devices.value = devResp?.Items || []
      parentalRating.value = userStore.jellyfin?.parental_rating || 0
    }
  } catch (e) {}
  loading.value = false
})

async function authorizeQC() {
  if (!qcCode.value) return
  qcLoading.value = true
  qcMessage.value = ''
  try {
    await api.jellyfinQuickConnect(qcCode.value)
    qcMessage.value = '✅ 授权成功！'
    qcCode.value = ''
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    qcMessage.value = '❌ ' + e.message
  }
  qcLoading.value = false
}

async function updateRating() {
  try {
    await api.jellyfinUpdateParentalRating(parentalRating.value)
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e) {}
}

async function changePassword() {
  try {
    await api.jellyfinUpdatePassword(currentPwd.value, newPwd.value)
    showPwdForm.value = false
    currentPwd.value = ''
    newPwd.value = ''
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    alert(e.message)
  }
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">🎬 Jellyfin</h1>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <template v-else-if="userStore.hasJellyfin">
      <div class="card">
        <div class="row-between">
          <h3>账户信息</h3>
          <span class="badge badge-success">活跃</span>
        </div>
        <div class="text-sm text-muted mt-sm">
          到期: {{ new Date(userStore.jellyfin?.expires_at).toLocaleDateString('zh-CN') }}
        </div>
      </div>

      <!-- Parental Rating -->
      <div class="card mt-md">
        <h3 class="mb-sm">🔒 内容分级</h3>
        <p class="text-xs text-muted mb-md">调整可观看内容的最高分级 (0=全部限制, 22=无限制)</p>
        <div class="row-between mb-sm">
          <span class="text-sm">分级: {{ parentalRating }}</span>
          <span class="text-xs text-muted">0 ~ 22</span>
        </div>
        <input type="range" min="0" max="22" step="1" v-model.number="parentalRating" @change="updateRating" />
      </div>

      <!-- Quick Connect -->
      <div class="card mt-md">
        <h3 class="mb-sm">⚡ Quick Connect</h3>
        <div class="row" style="gap:var(--space-sm)">
          <input class="input" v-model="qcCode" placeholder="输入授权码" style="flex:1" />
          <button class="btn btn-primary btn-sm" @click="authorizeQC" :disabled="qcLoading">授权</button>
        </div>
        <div v-if="qcMessage" class="text-sm mt-sm">{{ qcMessage }}</div>
      </div>

      <!-- Password -->
      <div class="card mt-md">
        <div class="row-between">
          <h3>🔑 密码管理</h3>
          <button class="btn btn-sm btn-secondary" @click="showPwdForm = !showPwdForm">{{ showPwdForm ? '取消' : '修改' }}</button>
        </div>
        <div v-if="showPwdForm" class="stack-sm mt-md">
          <input class="input" v-model="currentPwd" type="password" placeholder="当前密码" />
          <input class="input" v-model="newPwd" type="password" placeholder="新密码" />
          <button class="btn btn-primary btn-sm" @click="changePassword">确认修改</button>
        </div>
      </div>

      <!-- Devices -->
      <div class="card mt-md" v-if="devices.length > 0">
        <h3 class="mb-md">📱 设备列表</h3>
        <div class="stack-sm">
          <div v-for="dev in devices" :key="dev.Id" class="device-item">
            <div>
              <div class="text-sm">{{ dev.AppName || '未知应用' }}</div>
              <div class="text-xs text-muted">{{ dev.Name }}</div>
            </div>
            <span class="text-xs text-muted">{{ new Date(dev.DateLastActivity).toLocaleDateString('zh-CN') }}</span>
          </div>
        </div>
      </div>
    </template>

    <div class="empty-state" v-else>
      <span class="empty-state-icon">🎬</span>
      <p class="empty-state-text">还没有 Jellyfin 账户</p>
      <p class="text-xs text-muted mt-sm">¥2/月 · 支持多设备</p>
      <button class="btn btn-primary mt-md">开通影视服务</button>
    </div>
  </div>
</template>

<style scoped>
.device-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-sm) 0;
  border-bottom: 1px solid var(--border-subtle);
}

.device-item:last-child {
  border-bottom: none;
}
</style>
