<template>
  <div class="webhook-logs-view animate-fade-in">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">Webhook Delivery Logs</h1>
        <p class="view-subtitle">Inspect real-time webhook executions, delivery status codes, latencies, and payload signatures.</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-secondary" @click="fetchLogs" :disabled="loading">
          <RefreshCw :size="15" :class="{ 'spin-anim': loading }" /> Refresh
        </button>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="card filter-card mb-4">
      <div class="filter-controls">
        <div class="search-box">
          <Search :size="15" class="search-icon" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by Event, URL, or ID..."
            class="form-input search-field"
            @keyup.enter="fetchLogs"
          />
        </div>

        <div class="filter-selects">
          <select v-model="selectedStatus" class="form-select" @change="fetchLogs">
            <option value="">All Statuses</option>
            <option value="SUCCESS">Success (2xx)</option>
            <option value="FAILED">Failed</option>
            <option value="TIMEOUT">Timeout</option>
            <option value="PENDING">Pending</option>
          </select>

          <select v-model="selectedAppId" class="form-select" @change="fetchLogs">
            <option value="">All Applications</option>
            <option v-for="app in apps" :key="app.id" :value="app.id">
              {{ app.name }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <!-- Logs Table -->
    <div class="card table-card">
      <div v-if="loading && logs.length === 0" class="text-center py-12">
        <RefreshCw :size="28" class="spin-anim text-primary mx-auto" />
        <p class="mt-3 text-muted">Loading delivery logs...</p>
      </div>

      <div v-else-if="logs.length === 0" class="empty-state">
        <Activity :size="48" class="text-muted mb-2" />
        <h4 class="empty-title">No Webhook Logs Found</h4>
        <p class="empty-subtitle">Dispatch a webhook from the Applications tab or your API to see delivery traces here.</p>
        <router-link to="/apps" class="btn btn-primary mt-3">
          <Layers :size="15" /> Go to Applications
        </router-link>
      </div>

      <div v-else class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th>Status</th>
              <th>Event</th>
              <th>Destination URL</th>
              <th>HTTP Code</th>
              <th>Latency</th>
              <th>Timestamp</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="call in logs" :key="call.id" class="table-row hoverable" @click="openDetail(call)">
              <td>
                <span class="badge" :class="getStatusBadgeClass(call.status)">
                  <span class="status-dot"></span>
                  {{ call.status }}
                </span>
              </td>
              <td>
                <div class="event-cell">
                  <Zap :size="13" class="text-primary" />
                  <span class="event-name">{{ call.event_name || call.eventName }}</span>
                </div>
              </td>
              <td>
                <span class="url-cell" :title="call.target_url || call.targetUrl">
                  {{ call.target_url || call.targetUrl }}
                </span>
              </td>
              <td>
                <span v-if="call.response_status_code || call.responseStatusCode" class="code-badge" :class="getHttpCodeClass(call.response_status_code || call.responseStatusCode)">
                  {{ call.response_status_code || call.responseStatusCode }}
                </span>
                <span v-else class="text-muted">&mdash;</span>
              </td>
              <td>
                <span class="latency-text">{{ call.latency_ms || call.latencyMs || 0 }} ms</span>
              </td>
              <td>
                <span class="time-text">{{ formatDate(call.created_at || call.createdAt) }}</span>
              </td>
              <td class="text-right" @click.stop>
                <button class="btn btn-xs btn-outline mr-1" @click="openDetail(call)">
                  <Eye :size="13" /> Details
                </button>
                <button class="btn btn-xs btn-secondary" @click="retryCall(call)" :disabled="retryingId === call.id">
                  <RotateCw :size="13" :class="{ 'spin-anim': retryingId === call.id }" /> Retry
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="pagination-bar">
        <span class="page-info">Page {{ page }} of {{ totalPages }} ({{ total }} total deliveries)</span>
        <div class="page-controls">
          <button class="btn btn-sm btn-outline" :disabled="page <= 1" @click="changePage(page - 1)">
            Previous
          </button>
          <button class="btn btn-sm btn-outline" :disabled="page >= totalPages" @click="changePage(page + 1)">
            Next
          </button>
        </div>
      </div>
    </div>

    <!-- Detail Drawer / Modal -->
    <div v-if="activeLog" class="modal-backdrop" @click.self="activeLog = null">
      <div class="modal-card modal-detail">
        <div class="modal-header">
          <div class="modal-title-box">
            <h3 class="modal-title">Webhook Delivery Details</h3>
            <span class="badge ml-2" :class="getStatusBadgeClass(activeLog.status)">{{ activeLog.status }}</span>
          </div>
          <button class="modal-close" @click="activeLog = null">
            <X :size="18" />
          </button>
        </div>

        <div class="modal-body">
          <div class="meta-info-grid">
            <div class="meta-item">
              <span class="meta-lbl">Delivery ID</span>
              <code>{{ activeLog.id }}</code>
            </div>
            <div class="meta-item">
              <span class="meta-lbl">Event</span>
              <strong>{{ activeLog.event_name || activeLog.eventName }}</strong>
            </div>
            <div class="meta-item">
              <span class="meta-lbl">Target URL</span>
              <span class="url-val">{{ activeLog.target_url || activeLog.targetUrl }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-lbl">Attempts</span>
              <span>{{ activeLog.attempt_count || activeLog.attemptCount || 1 }}</span>
            </div>
          </div>

          <!-- HMAC Signature -->
          <div class="detail-section">
            <span class="section-title">Cryptographic Signature Header (HMAC-SHA256)</span>
            <div class="code-box">
              <code>{{ activeLog.signature || 'None' }}</code>
            </div>
          </div>

          <!-- Payload JSON -->
          <div class="detail-section">
            <span class="section-title">Request Payload</span>
            <pre class="code-box-pre">{{ formatJSON(activeLog.payload_json || activeLog.payloadJson) }}</pre>
          </div>

          <!-- Response Body -->
          <div class="detail-section">
            <div class="section-header-flex">
              <span class="section-title">Destination Response</span>
              <span class="badge" :class="getHttpCodeClass(activeLog.response_status_code || activeLog.responseStatusCode)">
                HTTP {{ activeLog.response_status_code || activeLog.responseStatusCode || 'None' }}
              </span>
            </div>
            <pre class="code-box-pre">{{ activeLog.response_body || activeLog.responseBody || activeLog.error_message || 'No response recorded.' }}</pre>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" @click="activeLog = null">Close</button>
          <button class="btn btn-primary" @click="retryCall(activeLog)" :disabled="retryingId === activeLog.id">
            <RotateCw :size="14" :class="{ 'spin-anim': retryingId === activeLog.id }" /> Retry Delivery
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { webhookService } from '../services/webhookService'
import { appService } from '../services/appService'
import { useToastStore } from '../stores/toast'
import {
  RefreshCw,
  Search,
  Zap,
  Activity,
  Eye,
  RotateCw,
  X,
  Layers,
} from 'lucide-vue-next'

const toastStore = useToastStore()

const loading = ref(false)
const retryingId = ref(null)
const logs = ref([])
const apps = ref([])
const total = ref(0)
const page = ref(1)
const limit = ref(15)

const searchQuery = ref('')
const selectedStatus = ref('')
const selectedAppId = ref('')

const activeLog = ref(null)

const totalPages = computed(() => Math.ceil(total.value / limit.value) || 1)

function getStatusBadgeClass(status) {
  switch (status) {
    case 'SUCCESS':
      return 'badge-success'
    case 'FAILED':
      return 'badge-danger'
    case 'TIMEOUT':
      return 'badge-warning'
    default:
      return 'badge-secondary'
  }
}

function getHttpCodeClass(code) {
  if (code >= 200 && code < 300) return 'code-2xx'
  if (code >= 400 && code < 500) return 'code-4xx'
  if (code >= 500) return 'code-5xx'
  return 'code-other'
}

function formatDate(isoStr) {
  if (!isoStr) return '-'
  const d = new Date(isoStr)
  return d.toLocaleString()
}

function formatJSON(raw) {
  if (!raw) return '{}'
  try {
    const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
    return JSON.stringify(parsed, null, 2)
  } catch (e) {
    return raw
  }
}

async function fetchApps() {
  try {
    const res = await appService.listApps()
    apps.value = res.data || []
  } catch (e) {
    // silent
  }
}

async function fetchLogs() {
  loading.value = true
  try {
    const res = await webhookService.listWebhookCalls({
      page: page.value,
      limit: limit.value,
      status: selectedStatus.value,
      app_id: selectedAppId.value,
      search: searchQuery.value,
    })
    logs.value = res.data || []
    total.value = res.total || 0
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to load webhook logs.')
  } finally {
    loading.value = false
  }
}

function changePage(newPage) {
  page.value = newPage
  fetchLogs()
}

function openDetail(call) {
  activeLog.value = call
}

async function retryCall(call) {
  retryingId.value = call.id
  try {
    const res = await webhookService.retryWebhookCall(call.id)
    if (res.success) {
      toastStore.success('Webhook delivery retried successfully!')
    } else {
      toastStore.warning(`Retry execution status: ${res.data?.status}`)
    }
    await fetchLogs()
    if (activeLog.value && activeLog.value.id === call.id) {
      activeLog.value = res.data
    }
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to retry webhook.')
  } finally {
    retryingId.value = null
  }
}

onMounted(() => {
  fetchApps()
  fetchLogs()
})
</script>

<style scoped>
.webhook-logs-view {
  padding: 1.5rem;
}

.view-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
}

.view-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--zoho-text-main);
  margin-bottom: 0.25rem;
}

.view-subtitle {
  font-size: 0.875rem;
  color: var(--zoho-text-muted);
}

.filter-card {
  padding: 1rem;
}

.filter-controls {
  display: flex;
  gap: 1rem;
  align-items: center;
  flex-wrap: wrap;
}

.search-box {
  position: relative;
  flex: 1;
  min-width: 260px;
}

.search-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--zoho-text-muted);
}

.search-field {
  padding-left: 2rem;
}

.filter-selects {
  display: flex;
  gap: 0.75rem;
}

.form-select {
  padding: 0.5rem 1rem;
  border: 1px solid var(--zoho-border-color);
  border-radius: var(--radius-md);
  background-color: #ffffff;
  font-size: 0.875rem;
  color: var(--zoho-text-main);
}

.table-card {
  overflow: hidden;
}

.table-responsive {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
  font-size: 0.875rem;
}

.data-table th {
  background: #f8fafc;
  padding: 0.75rem 1rem;
  font-weight: 600;
  color: var(--zoho-text-muted);
  border-bottom: 1px solid var(--zoho-border-color);
  text-transform: uppercase;
  font-size: 0.75rem;
  letter-spacing: 0.05em;
}

.data-table td {
  padding: 0.875rem 1rem;
  border-bottom: 1px solid var(--zoho-border-color);
  color: var(--zoho-text-main);
}

.table-row.hoverable {
  cursor: pointer;
  transition: background 0.15s;
}

.table-row.hoverable:hover {
  background: #f8fafc;
}

.event-cell {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-weight: 600;
}

.url-cell {
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  font-family: monospace;
  font-size: 0.8rem;
  color: #475569;
}

.code-badge {
  font-family: monospace;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.2rem 0.5rem;
  border-radius: var(--radius-xs);
}

.code-2xx {
  background: #dcfce7;
  color: #166534;
}

.code-4xx {
  background: #fef9c3;
  color: #854d0e;
}

.code-5xx {
  background: #fee2e2;
  color: #991b1b;
}

.badge-success {
  background: #dcfce7;
  color: #166534;
}

.badge-danger {
  background: #fee2e2;
  color: #991b1b;
}

.badge-warning {
  background: #fef9c3;
  color: #854d0e;
}

.badge-secondary {
  background: #f1f5f9;
  color: #475569;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  display: inline-block;
  margin-right: 4px;
}

.pagination-bar {
  padding: 0.875rem 1.25rem;
  border-top: 1px solid var(--zoho-border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.8rem;
  color: var(--zoho-text-muted);
}

.page-controls {
  display: flex;
  gap: 0.5rem;
}

.empty-state {
  padding: 3.5rem 1rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* Detail Modal */
.modal-detail {
  max-width: 720px;
}

.modal-title-box {
  display: flex;
  align-items: center;
}

.meta-info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem 1.5rem;
  background: #f8fafc;
  padding: 1rem;
  border-radius: var(--radius-md);
  margin-bottom: 1rem;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.meta-lbl {
  font-size: 0.7rem;
  text-transform: uppercase;
  font-weight: 600;
  color: var(--zoho-text-muted);
}

.detail-section {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  margin-top: 0.75rem;
}

.section-title {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--zoho-text-muted);
}

.section-header-flex {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.code-box {
  background: #0f172a;
  color: #f8fafc;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  overflow-x: auto;
}

.code-box-pre {
  background: #0f172a;
  color: #38bdf8;
  padding: 0.75rem;
  border-radius: var(--radius-sm);
  font-family: monospace;
  font-size: 0.75rem;
  max-height: 180px;
  overflow-y: auto;
}
</style>
