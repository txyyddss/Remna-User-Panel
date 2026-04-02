<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

const activeTab = ref('combos')
const config = ref<any>(null)
const internalSquads = ref<any[]>([])
const combos = ref<any[]>([])
const loading = ref(true)
const saving = ref(false)

// Combo form
const showComboForm = ref(false)
const editingCombo = ref<any>(null)
const comboForm = ref({
  name: '', description: '', squad_uuid: '', traffic_gb: 100,
  strategy: 'MONTH', cycle: 'monthly', price_rmb: 10, reset_price: 5
})

// User management
const users = ref<any[]>([])
const userTotal = ref(0)
const userSearch = ref('')
const userPage = ref(0)
const usersLoading = ref(false)
const editingUser = ref<any>(null)
const editUserForm = ref({ credit: 0, remnawave_uuid: '', jellyfin_user_id: '', is_admin: false })

// Config editing
const configEditing = ref(false)
const editableConfig = ref<any>(null)
const configSaving = ref(false)

onMounted(async () => {
  try {
    config.value = await api.getConfig()
    internalSquads.value = (await api.getInternalSquads()) || []
    await loadCombos()
  } catch (e) {}
  loading.value = false
})

async function loadCombos() {
  try { combos.value = (await api.adminListCombos()) || [] } catch (e) {}
}

function resetComboForm() {
  comboForm.value = { name: '', description: '', squad_uuid: '', traffic_gb: 100, strategy: 'MONTH', cycle: 'monthly', price_rmb: 10, reset_price: 5 }
  editingCombo.value = null
}

function startEditCombo(combo: any) {
  editingCombo.value = combo
  comboForm.value = {
    name: combo.name, description: combo.description, squad_uuid: combo.squad_uuid,
    traffic_gb: combo.traffic_gb, strategy: combo.strategy, cycle: combo.cycle,
    price_rmb: combo.price_rmb, reset_price: combo.reset_price
  }
  showComboForm.value = true
}

async function saveCombo() {
  saving.value = true
  try {
    if (editingCombo.value) {
      await api.updateCombo(editingCombo.value.uuid, comboForm.value)
    } else {
      await api.createCombo(comboForm.value)
    }
    showComboForm.value = false
    resetComboForm()
    await loadCombos()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    alert(e.message)
  }
  saving.value = false
}

async function deleteCombo(uuid: string) {
  if (!confirm('确定删除此套餐？')) return
  try {
    await api.deleteCombo(uuid)
    await loadCombos()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    alert(e.message)
  }
}

// Users
async function loadUsers() {
  usersLoading.value = true
  try {
    const resp = await api.adminListUsers({ search: userSearch.value, limit: 20, offset: userPage.value * 20 })
    users.value = resp.users || []
    userTotal.value = resp.total || 0
  } catch (e) {}
  usersLoading.value = false
}

function searchUsers() {
  userPage.value = 0
  loadUsers()
}

function startEditUser(u: any) {
  editingUser.value = u
  editUserForm.value = {
    credit: u.credit || 0,
    remnawave_uuid: u.remnawave_uuid || '',
    jellyfin_user_id: u.jellyfin_user_id || '',
    is_admin: !!u.is_admin,
  }
}

async function saveUser() {
  if (!editingUser.value) return
  saving.value = true
  try {
    await api.adminUpdateUser(editingUser.value.id, editUserForm.value)
    editingUser.value = null
    await loadUsers()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    alert(e.message)
  }
  saving.value = false
}

// Config
function startEditConfig() {
  editableConfig.value = JSON.parse(JSON.stringify(config.value))
  configEditing.value = true
}

async function saveConfig() {
  configSaving.value = true
  try {
    await api.updateConfig(editableConfig.value)
    config.value = await api.getConfig()
    configEditing.value = false
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    alert(e.message)
  }
  configSaving.value = false
}

function switchTab(tab: string) {
  activeTab.value = tab
  if (tab === 'users' && users.value.length === 0) loadUsers()
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">⚙️ 管理面板</h1>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <template v-else>
      <!-- Tab Navigation -->
      <div class="tab-bar">
        <button class="tab" :class="{ active: activeTab === 'combos' }" @click="switchTab('combos')">📦 套餐</button>
        <button class="tab" :class="{ active: activeTab === 'users' }" @click="switchTab('users')">👥 用户</button>
        <button class="tab" :class="{ active: activeTab === 'config' }" @click="switchTab('config')">⚙️ 配置</button>
      </div>

      <!-- ========== COMBOS TAB ========== -->
      <div v-if="activeTab === 'combos'">
        <div class="card">
          <div class="row-between">
            <h3>套餐管理</h3>
            <button class="btn btn-sm btn-primary" @click="showComboForm ? (showComboForm = false, resetComboForm()) : (showComboForm = true)">
              {{ showComboForm ? '取消' : '+ 新建' }}
            </button>
          </div>

          <!-- Combo Form -->
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
            <button class="btn btn-primary btn-block" @click="saveCombo" :disabled="saving">
              {{ saving ? '保存中...' : (editingCombo ? '更新套餐' : '创建套餐') }}
            </button>
          </div>
        </div>

        <!-- Combo List -->
        <div class="stack-sm mt-md">
          <div v-for="c in combos" :key="c.uuid" class="card combo-item" :class="{ inactive: !c.active }">
            <div class="row-between">
              <div>
                <div class="text-sm fw-semibold">{{ c.name }}</div>
                <div class="text-xs text-muted">¥{{ c.price_rmb }} · {{ c.traffic_gb }}GB · {{ c.strategy }}</div>
              </div>
              <div class="row" style="gap: var(--space-xs)">
                <button class="btn btn-xs btn-secondary" @click="startEditCombo(c)">✏️</button>
                <button class="btn btn-xs btn-danger" @click="deleteCombo(c.uuid)">🗑️</button>
              </div>
            </div>
          </div>
          <div v-if="combos.length === 0" class="text-sm text-muted text-center">暂无套餐</div>
        </div>
      </div>

      <!-- ========== USERS TAB ========== -->
      <div v-if="activeTab === 'users'">
        <div class="card">
          <h3 class="mb-sm">用户管理</h3>
          <div class="row" style="gap:var(--space-sm)">
            <input class="input" v-model="userSearch" placeholder="搜索用户名或ID" style="flex:1" @keyup.enter="searchUsers" />
            <button class="btn btn-sm btn-primary" @click="searchUsers">搜索</button>
          </div>
          <div class="text-xs text-muted mt-sm">共 {{ userTotal }} 个用户</div>
        </div>

        <div v-if="usersLoading" class="text-center mt-md"><div class="loading-spinner"></div></div>

        <div class="stack-sm mt-md" v-else>
          <div v-for="u in users" :key="u.id" class="card user-item" @click="startEditUser(u)">
            <div class="row-between">
              <div>
                <div class="text-sm fw-semibold">{{ u.telegram_name || 'Unknown' }}</div>
                <div class="text-xs text-muted">ID: {{ u.telegram_id }} · TXB: {{ (u.credit || 0).toFixed(2) }}</div>
              </div>
              <div class="text-right">
                <span v-if="u.is_admin" class="badge badge-warning" style="font-size:0.6rem">Admin</span>
                <span v-if="u.remnawave_uuid" class="badge badge-success" style="font-size:0.6rem">VPN</span>
              </div>
            </div>
          </div>

          <!-- Pagination -->
          <div class="row-between mt-sm" v-if="userTotal > 20">
            <button class="btn btn-sm btn-secondary" :disabled="userPage === 0" @click="userPage--; loadUsers()">上一页</button>
            <span class="text-xs text-muted">{{ userPage + 1 }} / {{ Math.ceil(userTotal / 20) }}</span>
            <button class="btn btn-sm btn-secondary" :disabled="(userPage + 1) * 20 >= userTotal" @click="userPage++; loadUsers()">下一页</button>
          </div>
        </div>

        <!-- Edit User Modal -->
        <teleport to="body">
          <transition name="fade">
            <div class="modal-overlay" v-if="editingUser" @click.self="editingUser = null">
              <div class="modal card">
                <h3 class="mb-md">编辑用户: {{ editingUser.telegram_name }}</h3>
                <div class="stack-sm">
                  <label class="text-xs text-muted">TXB 余额</label>
                  <input class="input" v-model.number="editUserForm.credit" type="number" step="0.01" />

                  <label class="text-xs text-muted">Remnawave UUID</label>
                  <input class="input" v-model="editUserForm.remnawave_uuid" placeholder="留空取消绑定" />

                  <label class="text-xs text-muted">Jellyfin User ID</label>
                  <input class="input" v-model="editUserForm.jellyfin_user_id" placeholder="留空取消绑定" />

                  <label class="checkbox">
                    <input type="checkbox" v-model="editUserForm.is_admin" />
                    <span class="text-sm">管理员权限</span>
                  </label>
                </div>
                <div class="row mt-lg" style="gap: var(--space-sm)">
                  <button class="btn btn-secondary" style="flex:1" @click="editingUser = null">取消</button>
                  <button class="btn btn-primary" style="flex:2" @click="saveUser" :disabled="saving">
                    {{ saving ? '保存中...' : '保存' }}
                  </button>
                </div>
              </div>
            </div>
          </transition>
        </teleport>
      </div>

      <!-- ========== CONFIG TAB ========== -->
      <div v-if="activeTab === 'config'">
        <div class="card">
          <div class="row-between mb-md">
            <h3>系统配置</h3>
            <button v-if="!configEditing" class="btn btn-sm btn-primary" @click="startEditConfig">编辑</button>
            <div v-else class="row" style="gap:var(--space-xs)">
              <button class="btn btn-sm btn-secondary" @click="configEditing = false">取消</button>
              <button class="btn btn-sm btn-primary" @click="saveConfig" :disabled="configSaving">
                {{ configSaving ? '...' : '保存' }}
              </button>
            </div>
          </div>

          <!-- Read-only view -->
          <div v-if="!configEditing && config" class="config-grid">
            <div class="config-item" v-for="(val, key) in config" :key="key">
              <span class="text-xs text-muted">{{ key }}</span>
              <pre class="config-value text-xs">{{ JSON.stringify(val, null, 2) }}</pre>
            </div>
          </div>

          <!-- Editable view -->
          <div v-if="configEditing && editableConfig" class="stack-sm">
            <template v-for="(section, sKey) in editableConfig" :key="sKey">
              <div class="config-section">
                <h4 class="text-sm mb-sm">{{ sKey }}</h4>
                <div class="stack-xs" v-if="typeof section === 'object' && section !== null">
                  <div v-for="(val, key) in section" :key="key" class="config-edit-row">
                    <label class="text-xs text-muted">{{ key }}</label>
                    <input v-if="typeof val === 'number'" class="input input-sm" type="number" v-model.number="editableConfig[sKey][key]" step="any" />
                    <input v-else-if="typeof val === 'boolean'" type="checkbox" v-model="editableConfig[sKey][key]" />
                    <input v-else class="input input-sm" v-model="editableConfig[sKey][key]" />
                  </div>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.tab-bar {
  display: flex;
  gap: var(--space-xs);
  margin-bottom: var(--space-md);
  background: var(--bg-glass);
  border-radius: var(--radius-md);
  padding: 3px;
}

.tab {
  flex: 1;
  padding: var(--space-sm);
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  cursor: pointer;
  font-family: var(--font-body);
  font-size: 0.8125rem;
  transition: all 0.2s;
}

.tab.active {
  background: var(--bg-card);
  color: var(--text-primary);
  font-weight: 600;
}

.combo-item.inactive {
  opacity: 0.5;
}

.btn-xs {
  padding: 4px 8px;
  font-size: 0.75rem;
}

.btn-danger {
  background: rgba(214, 48, 49, 0.15);
  color: #d63031;
  border: 1px solid rgba(214, 48, 49, 0.3);
}

.user-item {
  cursor: pointer;
  transition: all 0.2s;
}

.user-item:hover {
  border-color: var(--accent-primary);
}

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

.config-section {
  padding: var(--space-sm);
  background: var(--bg-glass);
  border-radius: var(--radius-sm);
}

.config-edit-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-sm);
  padding: 4px 0;
}

.config-edit-row label {
  min-width: 120px;
  flex-shrink: 0;
}

.config-edit-row .input-sm {
  padding: 4px 8px;
  font-size: 0.8125rem;
  flex: 1;
}

.stack-xs {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: flex-end;
  z-index: 200;
}

.modal {
  width: 100%;
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
  max-height: 80vh;
  overflow-y: auto;
}

.checkbox {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  cursor: pointer;
}

.checkbox input[type="checkbox"] {
  width: 18px;
  height: 18px;
  accent-color: var(--accent-primary);
}

.fw-semibold { font-weight: 600; }
.text-center { text-align: center; }
.text-right { text-align: right; }
</style>
