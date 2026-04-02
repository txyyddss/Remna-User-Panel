<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '@/api'

const activeTab = ref('combos')
const loading = ref(true)
const saving = ref(false)

const config = ref<any>(null)
const editableConfig = ref<any>(null)
const configEditing = ref(false)
const configSaving = ref(false)

const internalSquads = ref<any[]>([])
const combos = ref<any[]>([])
const showComboForm = ref(false)
const editingCombo = ref<any>(null)
const comboForm = ref({
  name: '',
  description: '',
  squad_uuid: '',
  traffic_gb: 100,
  strategy: 'MONTH',
  cycle: 'monthly',
  price_rmb: 10,
  reset_price: 5,
})

const users = ref<any[]>([])
const userTotal = ref(0)
const userSearch = ref('')
const userPage = ref(0)
const usersLoading = ref(false)
const editingUser = ref<any>(null)
const userDetailLoading = ref(false)
const userForm = ref<any>({
  credit: 0,
  is_admin: false,
  remnawave_uuid: '',
  jellyfin_user_id: '',
  subscription: {
    remnawave_uuid: '',
    combo_uuid: '',
    status: 'active',
    expires_at: '',
  },
  jellyfin: {
    jellyfin_user_id: '',
    username: '',
    parental_rating: 0,
    expires_at: '',
  },
})

const orders = ref<any[]>([])
const orderTotal = ref(0)
const ordersLoading = ref(false)
const orderFilters = ref({
  search: '',
  status: '',
  service_status: '',
  order_type: '',
  page: 0,
})
const selectedOrder = ref<any>(null)
const orderDetailLoading = ref(false)
const orderEdit = ref<any>({
  amount: 0,
  final_amount: 0,
  status: '',
  service_status: '',
  payment_method: '',
  payment_type: '',
  upstream_id: '',
  admin_note: '',
})

function toDateTimeLocal(value?: string | null) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const offset = date.getTimezoneOffset()
  const localDate = new Date(date.getTime() - offset * 60000)
  return localDate.toISOString().slice(0, 16)
}

function toRFC3339(value: string) {
  if (!value) return ''
  return new Date(value).toISOString()
}

async function loadConfig() {
  config.value = await api.getConfig()
}

async function loadInternalSquads() {
  try {
    internalSquads.value = (await api.getInternalSquads()) || []
  } catch (e) {
    internalSquads.value = []
  }
}

async function loadCombos() {
  try {
    combos.value = (await api.adminListCombos()) || []
  } catch (e) {
    combos.value = []
  }
}

function resetComboForm() {
  comboForm.value = {
    name: '',
    description: '',
    squad_uuid: '',
    traffic_gb: 100,
    strategy: 'MONTH',
    cycle: 'monthly',
    price_rmb: 10,
    reset_price: 5,
  }
  editingCombo.value = null
}

function startEditCombo(combo: any) {
  editingCombo.value = combo
  comboForm.value = {
    name: combo.name,
    description: combo.description,
    squad_uuid: combo.squad_uuid,
    traffic_gb: combo.traffic_gb,
    strategy: combo.strategy,
    cycle: combo.cycle,
    price_rmb: combo.price_rmb,
    reset_price: combo.reset_price,
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
    await loadCombos()
    showComboForm.value = false
    resetComboForm()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    alert(e.message)
  } finally {
    saving.value = false
  }
}

async function deleteCombo(uuid: string) {
  if (!confirm('Delete this plan?')) return
  try {
    await api.deleteCombo(uuid)
    await loadCombos()
  } catch (e: any) {
    alert(e.message)
  }
}

async function loadUsers() {
  usersLoading.value = true
  try {
    const resp = await api.adminListUsers({ search: userSearch.value, limit: 20, offset: userPage.value * 20 })
    users.value = resp.users || []
    userTotal.value = resp.total || 0
  } catch (e) {
    users.value = []
    userTotal.value = 0
  } finally {
    usersLoading.value = false
  }
}

async function openUser(user: any) {
  editingUser.value = user
  userDetailLoading.value = true
  try {
    const detail = await api.adminGetUser(user.id)
    userForm.value = {
      credit: detail.user.credit || 0,
      is_admin: !!detail.user.is_admin,
      remnawave_uuid: detail.user.remnawave_uuid || '',
      jellyfin_user_id: detail.user.jellyfin_user_id || '',
      subscription: {
        remnawave_uuid: detail.user.remnawave_uuid || '',
        combo_uuid: detail.subscription?.combo_uuid || '',
        status: detail.subscription?.status || 'active',
        expires_at: toDateTimeLocal(detail.subscription?.expires_at),
      },
      jellyfin: {
        jellyfin_user_id: detail.jellyfin?.jellyfin_user_id || detail.user.jellyfin_user_id || '',
        username: detail.jellyfin?.username || '',
        parental_rating: detail.jellyfin?.parental_rating || 0,
        expires_at: toDateTimeLocal(detail.jellyfin?.expires_at),
      },
    }
  } catch (e: any) {
    alert(e.message)
  } finally {
    userDetailLoading.value = false
  }
}

async function saveUser() {
  if (!editingUser.value) return
  saving.value = true
  try {
    await api.adminUpdateUser(editingUser.value.id, {
      credit: userForm.value.credit,
      is_admin: userForm.value.is_admin,
      remnawave_uuid: userForm.value.remnawave_uuid,
      jellyfin_user_id: userForm.value.jellyfin_user_id,
      subscription: {
        remnawave_uuid: userForm.value.subscription.remnawave_uuid,
        combo_uuid: userForm.value.subscription.combo_uuid,
        status: userForm.value.subscription.status,
        expires_at: toRFC3339(userForm.value.subscription.expires_at),
      },
      jellyfin: {
        jellyfin_user_id: userForm.value.jellyfin.jellyfin_user_id,
        username: userForm.value.jellyfin.username,
        parental_rating: userForm.value.jellyfin.parental_rating,
        expires_at: toRFC3339(userForm.value.jellyfin.expires_at),
      },
    })
    editingUser.value = null
    await loadUsers()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    alert(e.message)
  } finally {
    saving.value = false
  }
}

async function loadOrders() {
  ordersLoading.value = true
  try {
    const resp = await api.adminListOrders({
      search: orderFilters.value.search,
      status: orderFilters.value.status,
      service_status: orderFilters.value.service_status,
      order_type: orderFilters.value.order_type,
      limit: 20,
      offset: orderFilters.value.page * 20,
    })
    orders.value = resp.orders || []
    orderTotal.value = resp.total || 0
  } catch (e) {
    orders.value = []
    orderTotal.value = 0
  } finally {
    ordersLoading.value = false
  }
}

async function openOrder(order: any) {
  selectedOrder.value = null
  orderDetailLoading.value = true
  try {
    selectedOrder.value = await api.getOrder(order.uuid)
    orderEdit.value = {
      amount: selectedOrder.value.amount,
      final_amount: selectedOrder.value.final_amount,
      status: selectedOrder.value.status,
      service_status: selectedOrder.value.service_status,
      payment_method: selectedOrder.value.payment_method || '',
      payment_type: selectedOrder.value.payment_type || '',
      upstream_id: selectedOrder.value.upstream_id || '',
      admin_note: selectedOrder.value.admin_note || '',
    }
  } catch (e: any) {
    alert(e.message)
  } finally {
    orderDetailLoading.value = false
  }
}

async function saveOrder() {
  if (!selectedOrder.value) return
  saving.value = true
  try {
    selectedOrder.value = await api.adminUpdateOrder(selectedOrder.value.uuid, orderEdit.value)
    await loadOrders()
  } catch (e: any) {
    alert(e.message)
  } finally {
    saving.value = false
  }
}

async function runOrderAction(action: string) {
  if (!selectedOrder.value) return
  saving.value = true
  try {
    selectedOrder.value = await api.adminOrderAction(selectedOrder.value.uuid, action)
    await loadOrders()
  } catch (e: any) {
    alert(e.message)
  } finally {
    saving.value = false
  }
}

function startEditConfig() {
  editableConfig.value = JSON.parse(JSON.stringify(config.value))
  configEditing.value = true
}

async function saveConfig() {
  configSaving.value = true
  try {
    await api.updateConfig(editableConfig.value)
    await loadConfig()
    configEditing.value = false
  } catch (e: any) {
    alert(e.message)
  } finally {
    configSaving.value = false
  }
}

async function initialize() {
  loading.value = true
  await Promise.allSettled([
    loadConfig(),
    loadInternalSquads(),
    loadCombos(),
  ])
  loading.value = false
}

onMounted(initialize)
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">Admin Panel</h1>
    </div>

    <div class="loading-page" v-if="loading">
      <div class="loading-spinner"></div>
    </div>

    <template v-else>
      <div class="tab-bar">
        <button class="tab" :class="{ active: activeTab === 'combos' }" @click="activeTab = 'combos'">Plans</button>
        <button class="tab" :class="{ active: activeTab === 'users' }" @click="activeTab = 'users'; if (!users.length) loadUsers()">Users</button>
        <button class="tab" :class="{ active: activeTab === 'billing' }" @click="activeTab = 'billing'; if (!orders.length) loadOrders()">Billing</button>
        <button class="tab" :class="{ active: activeTab === 'config' }" @click="activeTab = 'config'">Config</button>
      </div>

      <div v-if="activeTab === 'combos'">
        <div class="card">
          <div class="row-between">
            <h3>Plan Management</h3>
            <button class="btn btn-sm btn-primary" @click="showComboForm ? (showComboForm = false, resetComboForm()) : (showComboForm = true)">
              {{ showComboForm ? 'Cancel' : '+ New' }}
            </button>
          </div>

          <div v-if="showComboForm" class="stack-sm mt-md">
            <input class="input" v-model="comboForm.name" placeholder="Plan name" />
            <input class="input" v-model="comboForm.description" placeholder="Description" />
            <select class="input" v-model="comboForm.squad_uuid">
              <option value="">Select internal squad</option>
              <option v-for="squad in internalSquads" :key="squad.uuid" :value="squad.uuid">{{ squad.name }}</option>
            </select>
            <div class="grid-2">
              <input class="input" v-model.number="comboForm.traffic_gb" type="number" placeholder="Traffic (GB)" />
              <select class="input" v-model="comboForm.strategy">
                <option value="NO_RESET">No Reset</option>
                <option value="DAY">Daily</option>
                <option value="WEEK">Weekly</option>
                <option value="MONTH">Monthly</option>
              </select>
            </div>
            <div class="grid-2">
              <input class="input" v-model.number="comboForm.price_rmb" type="number" step="0.01" placeholder="Price (RMB)" />
              <input class="input" v-model.number="comboForm.reset_price" type="number" step="0.01" placeholder="Reset price (RMB)" />
            </div>
            <select class="input" v-model="comboForm.cycle">
              <option value="monthly">Monthly</option>
              <option value="quarterly">Quarterly</option>
              <option value="semiannual">Semi-Annual</option>
              <option value="annual">Annual</option>
            </select>
            <button class="btn btn-primary btn-block" @click="saveCombo" :disabled="saving">{{ saving ? 'Saving...' : (editingCombo ? 'Update Plan' : 'Create Plan') }}</button>
          </div>
        </div>

        <div class="stack-sm mt-md">
          <div v-for="combo in combos" :key="combo.uuid" class="card combo-item" :class="{ inactive: !combo.active }">
            <div class="row-between">
              <div>
                <div class="text-sm fw-semibold">{{ combo.name }}</div>
                <div class="text-xs text-muted">¥{{ combo.price_rmb }} · {{ combo.traffic_gb }}GB · {{ combo.strategy }}</div>
              </div>
              <div class="row" style="gap: var(--space-xs)">
                <button class="btn btn-xs btn-secondary" @click="startEditCombo(combo)">Edit</button>
                <button class="btn btn-xs btn-danger" @click="deleteCombo(combo.uuid)">Delete</button>
              </div>
            </div>
          </div>
          <div v-if="combos.length === 0" class="text-sm text-muted text-center">No plans found</div>
        </div>
      </div>

      <div v-if="activeTab === 'users'">
        <div class="card">
          <div class="row" style="gap:var(--space-sm)">
            <input class="input" v-model="userSearch" placeholder="Search by name or Telegram ID" style="flex:1" @keyup.enter="userPage = 0; loadUsers()" />
            <button class="btn btn-sm btn-primary" @click="userPage = 0; loadUsers()">Search</button>
          </div>
          <div class="text-xs text-muted mt-sm">Total {{ userTotal }} users</div>
        </div>

        <div v-if="usersLoading" class="text-center mt-md"><div class="loading-spinner"></div></div>
        <div v-else class="stack-sm mt-md">
          <div v-for="user in users" :key="user.id" class="card user-item" @click="openUser(user)">
            <div class="row-between">
              <div>
                <div class="text-sm fw-semibold">{{ user.telegram_name || 'Unknown' }}</div>
                <div class="text-xs text-muted">ID: {{ user.telegram_id }} · Credit: {{ Number(user.credit || 0).toFixed(2) }}</div>
              </div>
              <div class="text-right">
                <span v-if="user.is_admin" class="badge badge-warning small-badge">Admin</span>
                <span v-if="user.remnawave_uuid" class="badge badge-success small-badge">VPN</span>
                <span v-if="user.jellyfin_user_id" class="badge badge-success small-badge">Jellyfin</span>
              </div>
            </div>
          </div>

          <div class="row-between mt-sm" v-if="userTotal > 20">
            <button class="btn btn-sm btn-secondary" :disabled="userPage === 0" @click="userPage--; loadUsers()">Prev</button>
            <span class="text-xs text-muted">{{ userPage + 1 }} / {{ Math.ceil(userTotal / 20) }}</span>
            <button class="btn btn-sm btn-secondary" :disabled="(userPage + 1) * 20 >= userTotal" @click="userPage++; loadUsers()">Next</button>
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'billing'">
        <div class="card">
          <div class="stack-sm">
            <input class="input" v-model="orderFilters.search" placeholder="Search order, upstream ID, or user" @keyup.enter="orderFilters.page = 0; loadOrders()" />
            <div class="grid-2">
              <select class="input" v-model="orderFilters.status">
                <option value="">All payment statuses</option>
                <option value="pending">pending</option>
                <option value="processing">processing</option>
                <option value="paid">paid</option>
                <option value="cancelled">cancelled</option>
                <option value="refunded">refunded</option>
              </select>
              <select class="input" v-model="orderFilters.service_status">
                <option value="">All service statuses</option>
                <option value="pending">pending</option>
                <option value="fulfilled">fulfilled</option>
                <option value="waiting_admin">waiting_admin</option>
                <option value="cancelled">cancelled</option>
                <option value="refunded">refunded</option>
                <option value="failed">failed</option>
              </select>
            </div>
            <div class="grid-2">
              <select class="input" v-model="orderFilters.order_type">
                <option value="">All bill types</option>
                <option value="combo">combo</option>
                <option value="jellyfin">jellyfin</option>
                <option value="custom">custom</option>
              </select>
              <button class="btn btn-primary" @click="orderFilters.page = 0; loadOrders()">Apply Filters</button>
            </div>
          </div>
        </div>

        <div v-if="ordersLoading" class="text-center mt-md"><div class="loading-spinner"></div></div>
        <div v-else class="stack-sm mt-md">
          <div v-for="order in orders" :key="order.uuid" class="card user-item" @click="openOrder(order)">
            <div class="row-between">
              <div>
                <div class="text-sm fw-semibold">{{ order.order_type }} · {{ order.user_telegram_name }}</div>
                <div class="text-xs text-muted">{{ order.uuid }} · ¥{{ Number(order.final_amount || 0).toFixed(2) }}</div>
              </div>
              <div class="text-right">
                <div class="text-xs">{{ order.status }}</div>
                <div class="text-xs text-muted">{{ order.service_status }}</div>
              </div>
            </div>
          </div>

          <div class="row-between mt-sm" v-if="orderTotal > 20">
            <button class="btn btn-sm btn-secondary" :disabled="orderFilters.page === 0" @click="orderFilters.page--; loadOrders()">Prev</button>
            <span class="text-xs text-muted">{{ orderFilters.page + 1 }} / {{ Math.ceil(orderTotal / 20) }}</span>
            <button class="btn btn-sm btn-secondary" :disabled="(orderFilters.page + 1) * 20 >= orderTotal" @click="orderFilters.page++; loadOrders()">Next</button>
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'config'">
        <div class="card">
          <div class="row-between mb-md">
            <h3>System Config</h3>
            <button v-if="!configEditing" class="btn btn-sm btn-primary" @click="startEditConfig">Edit</button>
            <div v-else class="row" style="gap:var(--space-xs)">
              <button class="btn btn-sm btn-secondary" @click="configEditing = false">Cancel</button>
              <button class="btn btn-sm btn-primary" @click="saveConfig" :disabled="configSaving">{{ configSaving ? 'Saving...' : 'Save' }}</button>
            </div>
          </div>

          <div v-if="!configEditing && config" class="config-grid">
            <div class="config-item" v-for="(val, key) in config" :key="key">
              <span class="text-xs text-muted">{{ key }}</span>
              <pre class="config-value text-xs">{{ JSON.stringify(val, null, 2) }}</pre>
            </div>
          </div>

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

    <teleport to="body">
      <transition name="fade">
        <div v-if="editingUser" class="modal-overlay" @click.self="editingUser = null">
          <div class="modal card">
            <h3 class="mb-md">Edit User: {{ editingUser.telegram_name }}</h3>
            <div v-if="userDetailLoading" class="loading-page"><div class="loading-spinner"></div></div>
            <div v-else class="stack-sm">
              <label class="text-xs text-muted">Credit</label>
              <input class="input" v-model.number="userForm.credit" type="number" step="0.01" />

              <label class="text-xs text-muted">Remnawave UUID</label>
              <input class="input" v-model="userForm.remnawave_uuid" />

              <label class="text-xs text-muted">Jellyfin User ID</label>
              <input class="input" v-model="userForm.jellyfin_user_id" />

              <label class="checkbox">
                <input type="checkbox" v-model="userForm.is_admin" />
                <span class="text-sm">Admin privileges</span>
              </label>

              <div class="card muted-card">
                <h4 class="mb-sm">Subscription</h4>
                <div class="stack-sm">
                  <input class="input" v-model="userForm.subscription.remnawave_uuid" placeholder="Remnawave UUID override" />
                  <select class="input" v-model="userForm.subscription.combo_uuid">
                    <option value="">Select combo</option>
                    <option v-for="combo in combos" :key="combo.uuid" :value="combo.uuid">{{ combo.name }}</option>
                  </select>
                  <select class="input" v-model="userForm.subscription.status">
                    <option value="active">active</option>
                    <option value="disabled">disabled</option>
                    <option value="expired">expired</option>
                  </select>
                  <input class="input" type="datetime-local" v-model="userForm.subscription.expires_at" />
                </div>
              </div>

              <div class="card muted-card">
                <h4 class="mb-sm">Jellyfin</h4>
                <div class="stack-sm">
                  <input class="input" v-model="userForm.jellyfin.jellyfin_user_id" placeholder="Jellyfin User ID" />
                  <input class="input" v-model="userForm.jellyfin.username" placeholder="Display username" />
                  <input class="input" v-model.number="userForm.jellyfin.parental_rating" type="number" min="0" max="22" />
                  <input class="input" type="datetime-local" v-model="userForm.jellyfin.expires_at" />
                </div>
              </div>
            </div>

            <div class="row mt-lg" style="gap: var(--space-sm)">
              <button class="btn btn-secondary" style="flex:1" @click="editingUser = null">Cancel</button>
              <button class="btn btn-primary" style="flex:2" @click="saveUser" :disabled="saving">{{ saving ? 'Saving...' : 'Save User' }}</button>
            </div>
          </div>
        </div>
      </transition>
    </teleport>

    <teleport to="body">
      <transition name="fade">
        <div v-if="selectedOrder || orderDetailLoading" class="modal-overlay" @click.self="selectedOrder = null">
          <div class="modal card">
            <h3 class="mb-md">Bill Detail</h3>
            <div v-if="orderDetailLoading" class="loading-page"><div class="loading-spinner"></div></div>
            <div v-else-if="selectedOrder" class="stack-sm">
              <div class="text-xs text-muted">{{ selectedOrder.uuid }}</div>
              <div class="text-sm">User: {{ selectedOrder.user_telegram_name }} ({{ selectedOrder.user_telegram_id }})</div>
              <input class="input" v-model.number="orderEdit.amount" type="number" step="0.01" placeholder="Amount" />
              <input class="input" v-model.number="orderEdit.final_amount" type="number" step="0.01" placeholder="Final amount" />
              <select class="input" v-model="orderEdit.status">
                <option value="pending">pending</option>
                <option value="processing">processing</option>
                <option value="paid">paid</option>
                <option value="cancelled">cancelled</option>
                <option value="refunded">refunded</option>
              </select>
              <select class="input" v-model="orderEdit.service_status">
                <option value="pending">pending</option>
                <option value="fulfilled">fulfilled</option>
                <option value="waiting_admin">waiting_admin</option>
                <option value="cancelled">cancelled</option>
                <option value="refunded">refunded</option>
                <option value="failed">failed</option>
              </select>
              <div class="grid-2">
                <input class="input" v-model="orderEdit.payment_method" placeholder="Payment method" />
                <input class="input" v-model="orderEdit.payment_type" placeholder="Payment type" />
              </div>
              <input class="input" v-model="orderEdit.upstream_id" placeholder="Upstream ID" />
              <textarea class="input order-note" v-model="orderEdit.admin_note" placeholder="Admin note"></textarea>

              <div class="grid-2">
                <button class="btn btn-secondary" @click="runOrderAction('cancel')" :disabled="saving">Cancel Bill</button>
                <button class="btn btn-secondary" @click="runOrderAction('refund')" :disabled="saving">Refund</button>
              </div>
              <div class="grid-2">
                <button class="btn btn-secondary" @click="runOrderAction('resend-notice')" :disabled="saving">Resend Notice</button>
                <button class="btn btn-primary" @click="runOrderAction('apply-credit')" :disabled="saving">Apply Credit</button>
              </div>

              <div class="card muted-card">
                <h4 class="mb-sm">Audit Trail</h4>
                <div v-if="!selectedOrder.events?.length" class="text-xs text-muted">No events recorded yet.</div>
                <div v-else class="stack-xs">
                  <div v-for="event in selectedOrder.events" :key="event.id" class="audit-row">
                    <div class="text-sm">{{ event.event_type }}</div>
                    <div class="text-xs text-muted">{{ event.message }}</div>
                    <div class="text-xs text-muted">{{ new Date(event.created_at).toLocaleString('en-US') }}</div>
                  </div>
                </div>
              </div>
            </div>

            <div class="row mt-lg" style="gap: var(--space-sm)">
              <button class="btn btn-secondary" style="flex:1" @click="selectedOrder = null">Close</button>
              <button class="btn btn-primary" style="flex:2" @click="saveOrder" :disabled="saving || !selectedOrder">{{ saving ? 'Saving...' : 'Save Bill' }}</button>
            </div>
          </div>
        </div>
      </transition>
    </teleport>
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
}

.small-badge {
  font-size: 0.65rem;
  margin-left: 4px;
}

.config-grid {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.config-item,
.config-section,
.muted-card {
  background: var(--bg-glass);
}

.config-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: var(--space-sm);
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

.input-sm {
  padding: 4px 8px;
  font-size: 0.8125rem;
  flex: 1;
}

.stack-xs {
  display: flex;
  flex-direction: column;
  gap: 6px;
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
  max-height: 85vh;
  overflow-y: auto;
}

.order-note {
  min-height: 96px;
}

.audit-row {
  padding: var(--space-sm);
  border-radius: var(--radius-sm);
  background: var(--bg-card);
}

.fw-semibold {
  font-weight: 600;
}

.text-center {
  text-align: center;
}

.text-right {
  text-align: right;
}
</style>
