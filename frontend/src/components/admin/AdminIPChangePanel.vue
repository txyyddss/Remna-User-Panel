<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api'
import { useConfirm } from '@/composables/useConfirm'
import { useToast } from '@/composables/useToast'
import type { AdminIPChangeRequest } from '@/types'

const toast = useToast()
const { confirm } = useConfirm()

const loading = ref(true)
const acting = ref<number | null>(null)
const requests = ref<AdminIPChangeRequest[]>([])
const total = ref(0)

const activeRequest = computed(() => requests.value.find((request) => request.status === 'PENDING' || request.status === 'CHANGING') || null)
const historyRequests = computed(() => requests.value.filter((request) => !activeRequest.value || request.id !== activeRequest.value.id))

function formatDateTime(value?: string) {
    if (!value) {
        return '—'
    }
    return new Date(value).toLocaleString('en-US')
}

function statusTone(status: AdminIPChangeRequest['status']) {
    switch (status) {
        case 'PENDING':
            return 'warning'
        case 'CHANGING':
            return 'success'
        case 'COMPLETED':
            return 'success'
        case 'REJECTED':
            return 'danger'
        default:
            return 'muted'
    }
}

async function loadRequests() {
    loading.value = true
    try {
        const resp = await api.adminListIPChangeRequests(20, 0)
        requests.value = resp.requests || []
        total.value = resp.total || 0
    } catch (e: any) {
        toast.error(e.message || 'Failed to load IP change requests.')
        requests.value = []
        total.value = 0
    } finally {
        loading.value = false
    }
}

async function runAction(requestId: number, action: 'approve' | 'decline' | 'complete') {
    acting.value = requestId
    try {
        await api.adminIPChangeAction(requestId, action)
        await loadRequests()
        toast.success(`IP change request ${action}d.`)
    } catch (e: any) {
        toast.error(e.message || `Failed to ${action} request.`)
    } finally {
        acting.value = null
    }
}

async function deleteRequest(request: AdminIPChangeRequest) {
    const ok = await confirm({
        title: 'Delete Request',
        message: `Delete IP change request ${request.username}? This removes the request and its votes.`,
    })
    if (!ok) {
        return
    }

    acting.value = request.id
    try {
        await api.adminDeleteIPChangeRequest(request.id)
        await loadRequests()
        toast.success('IP change request deleted.')
    } catch (e: any) {
        toast.error(e.message || 'Failed to delete request.')
    } finally {
        acting.value = null
    }
}

onMounted(() => {
    void loadRequests()
})
</script>

<template>
  <div class="stack-sm">
    <div class="card">
      <div class="row-between">
        <div>
          <h3>IP Change Requests</h3>
          <p class="text-xs text-muted mt-sm">Manage the active queue and recent IP swap requests without relying on Telegram votes.</p>
        </div>
        <button class="btn btn-sm btn-secondary" @click="loadRequests" :disabled="loading">Refresh</button>
      </div>
      <div class="text-xs text-muted mt-sm">Showing {{ requests.length }} of {{ total }} request(s)</div>
    </div>

    <div v-if="loading" class="card text-center">
      <div class="loading-spinner"></div>
    </div>

    <template v-else>
      <div v-if="activeRequest" class="card active-card">
        <div class="row-between card-head">
          <div>
            <div class="eyebrow">Active Request</div>
            <h3>{{ activeRequest.username }}</h3>
          </div>
          <span class="badge" :class="`badge-${statusTone(activeRequest.status)}`">{{ activeRequest.status }}</span>
        </div>

        <p class="request-reason">{{ activeRequest.reason }}</p>

        <div class="metrics-grid">
          <div class="metric-box">
            <span class="metric-label">Votes</span>
            <strong class="metric-value">{{ activeRequest.agree_count }}/5</strong>
          </div>
          <div class="metric-box">
            <span class="metric-label">Declines</span>
            <strong class="metric-value">{{ activeRequest.decline_count }}/2</strong>
          </div>
          <div class="metric-box">
            <span class="metric-label">Requested</span>
            <strong class="metric-value">{{ formatDateTime(activeRequest.requested_at) }}</strong>
          </div>
        </div>

        <div class="meta-list">
          <div class="meta-row">
            <span class="text-xs text-muted">Request Key</span>
            <span class="mono text-sm">{{ activeRequest.request_key }}</span>
          </div>
          <div class="meta-row">
            <span class="text-xs text-muted">Short UUID</span>
            <span class="mono text-sm">{{ activeRequest.short_uuid || '—' }}</span>
          </div>
          <div class="meta-row" v-if="activeRequest.message_link">
            <span class="text-xs text-muted">Telegram Thread</span>
            <a class="thread-link" :href="activeRequest.message_link" target="_blank" rel="noreferrer">Open thread</a>
          </div>
        </div>

        <div class="action-grid">
          <button
            v-if="activeRequest.status === 'PENDING'"
            class="btn btn-primary"
            :disabled="acting === activeRequest.id"
            @click="runAction(activeRequest.id, 'approve')"
          >
            Approve
          </button>
          <button
            v-if="activeRequest.status === 'PENDING'"
            class="btn btn-secondary"
            :disabled="acting === activeRequest.id"
            @click="runAction(activeRequest.id, 'decline')"
          >
            Decline
          </button>
          <button
            v-if="activeRequest.status === 'CHANGING'"
            class="btn btn-primary"
            :disabled="acting === activeRequest.id"
            @click="runAction(activeRequest.id, 'complete')"
          >
            Mark Completed
          </button>
          <button class="btn btn-danger" :disabled="acting === activeRequest.id" @click="deleteRequest(activeRequest)">
            Delete
          </button>
        </div>
      </div>

      <div v-else class="card text-sm text-muted">
        No active request. The public queue is currently idle.
      </div>

      <div class="card">
        <div class="row-between">
          <h3>Recent Requests</h3>
          <span class="text-xs text-muted">{{ historyRequests.length }} item(s)</span>
        </div>

        <div v-if="historyRequests.length === 0" class="text-sm text-muted mt-md">
          No recent request history beyond the active queue item.
        </div>

        <div v-else class="stack-sm mt-md">
          <div v-for="request in historyRequests" :key="request.id" class="history-item">
            <div class="row-between history-head">
              <div>
                <div class="text-sm fw-semibold">{{ request.username }}</div>
                <div class="text-xs text-muted">{{ formatDateTime(request.requested_at) }}</div>
              </div>
              <span class="badge" :class="`badge-${statusTone(request.status)}`">{{ request.status }}</span>
            </div>

            <p class="history-reason">{{ request.reason }}</p>

            <div class="history-meta">
              <span class="text-xs text-muted">Votes {{ request.agree_count }}/5</span>
              <span class="text-xs text-muted">Declines {{ request.decline_count }}/2</span>
              <span class="text-xs text-muted">Updated {{ formatDateTime(request.updated_at) }}</span>
            </div>

            <div class="row-between mt-sm">
              <a v-if="request.message_link" class="thread-link" :href="request.message_link" target="_blank" rel="noreferrer">Open thread</a>
              <span v-else class="text-xs text-muted">No Telegram thread</span>
              <button
                v-if="request.status !== 'COMPLETED'"
                class="btn btn-xs btn-danger"
                :disabled="acting === request.id"
                @click="deleteRequest(request)"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.active-card,
.history-item {
  border: 1px solid var(--border-subtle);
}

.card-head {
  align-items: flex-start;
}

.eyebrow {
  font-size: 0.72rem;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--accent-secondary);
}

.request-reason,
.history-reason {
  margin-top: var(--space-md);
  color: var(--text-secondary);
  line-height: 1.6;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-sm);
  margin-top: var(--space-md);
}

.metric-box {
  padding: var(--space-sm);
  border-radius: var(--radius-sm);
  background: var(--bg-glass);
}

.metric-label {
  display: block;
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--text-muted);
}

.metric-value {
  display: block;
  margin-top: 6px;
  font-size: 0.95rem;
}

.meta-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  margin-top: var(--space-md);
}

.meta-row,
.history-meta {
  display: flex;
  justify-content: space-between;
  gap: var(--space-sm);
  flex-wrap: wrap;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-sm);
  margin-top: var(--space-md);
}

.history-item {
  padding: var(--space-md);
  border-radius: var(--radius-md);
  background: var(--bg-glass);
}

.history-head {
  align-items: flex-start;
}

.thread-link {
  color: var(--accent-secondary);
  text-decoration: none;
}

.fw-semibold {
  font-weight: 600;
}

.text-center {
  text-align: center;
}

.btn-danger {
  background: rgba(214, 48, 49, 0.15);
  color: #d63031;
  border: 1px solid rgba(214, 48, 49, 0.3);
}

.btn-xs {
  padding: 4px 8px;
  font-size: 0.75rem;
}

@media (max-width: 720px) {
  .metrics-grid,
  .action-grid {
    grid-template-columns: 1fr;
  }
}
</style>
