<template>
  <div class="webhook-logs-view animate-fade-in">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('webhooks.logsTitle') }}</h1>
        <p class="view-subtitle">{{ t('webhooks.logsSubtitle') }}</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-secondary btn-sm" @click="fetchLogs" :disabled="loading">
          <RefreshCw :size="14" :class="{ 'spin-anim': loading }" /> {{ t('common.refresh') }}
        </button>
        <button class="btn btn-primary btn-sm" @click="openDispatchModal">
          <Send :size="14" /> {{ t('apps.testWebhook') }}
        </button>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="card filter-card mb-6">
      <div class="filter-controls-row">
        <div class="search-input-wrapper">
          <Search :size="15" class="search-icon" />
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('webhooks.searchPlaceholder')"
            class="form-control search-field"
            @keyup.enter="handleFilterChange"
          />
        </div>

        <div class="select-filters-group">
          <select v-model="selectedStatus" class="form-control select-field" @change="handleFilterChange">
            <option value="">{{ t('webhooks.allStatuses') }}</option>
            <option value="SUCCESS">Success (2xx)</option>
            <option value="FAILED">Failed</option>
            <option value="TIMEOUT">Timeout</option>
            <option value="PENDING">Pending</option>
          </select>

          <select v-model="selectedAppId" class="form-control select-field" @change="handleFilterChange">
            <option value="">{{ t('webhooks.allApps') }}</option>
            <option v-for="app in apps" :key="app.id" :value="app.id">
              {{ app.name }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <!-- Logs Table Card -->
    <div class="card table-card">
      <div v-if="loading && logs.length === 0" class="text-center py-12">
        <RefreshCw :size="28" class="spin-anim text-primary mx-auto" />
        <p class="mt-3 text-muted">{{ t('common.loading') }}</p>
      </div>

      <EmptyState
        v-else-if="logs.length === 0"
        :title="t('webhooks.noLogs')"
        :description="t('webhooks.noLogsSubtitle')"
        :icon="Activity"
      >
        <template #action>
          <div class="d-flex gap-2">
            <button class="btn btn-primary btn-sm" @click="openDispatchModal">
              <Send :size="14" /> {{ t('apps.testWebhook') }}
            </button>
            <router-link to="/apps" class="btn btn-secondary btn-sm">
              <Layers :size="14" /> {{ t('webhooks.goToApps') }}
            </router-link>
          </div>
        </template>
      </EmptyState>

      <div v-else class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th>{{ t('webhooks.status') }}</th>
              <th>{{ t('webhooks.event') }}</th>
              <th>{{ t('webhooks.destinationUrl') }}</th>
              <th>{{ t('webhooks.httpCode') }}</th>
              <th>{{ t('webhooks.latency') }}</th>
              <th>{{ t('webhooks.timestamp') }}</th>
              <th class="text-right">{{ t('common.actions') }}</th>
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
                <div class="event-cell" dir="ltr">
                  <Zap :size="13" class="text-primary flex-shrink-0" />
                  <span class="event-name">{{ call.event_name || call.eventName }}</span>
                </div>
              </td>
              <td>
                <span class="url-cell" dir="ltr" :title="call.target_url || call.targetUrl">
                  {{ call.target_url || call.targetUrl }}
                </span>
              </td>
              <td>
                <span v-if="call.response_status_code || call.responseStatusCode" class="code-badge" :class="getHttpCodeClass(call.response_status_code || call.responseStatusCode)">
                  HTTP {{ call.response_status_code || call.responseStatusCode }}
                </span>
                <span v-else class="text-muted">&mdash;</span>
              </td>
              <td>
                <span class="latency-text" dir="ltr">{{ call.latency_ms || call.latencyMs || 0 }} ms</span>
              </td>
              <td>
                <span class="time-text">{{ formatDate(call.created_at || call.createdAt) }}</span>
              </td>
              <td class="text-right" @click.stop>
                <div class="action-buttons-group">
                  <button class="btn btn-xs btn-outline" @click="openDetail(call)">
                    <Eye :size="12" /> {{ t('webhooks.details') }}
                  </button>
                  <button class="btn btn-xs btn-secondary" @click="retryCall(call)" :disabled="retryingId === call.id">
                    <RotateCw :size="12" :class="{ 'spin-anim': retryingId === call.id }" /> {{ t('webhooks.retry') }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <Pagination
        v-if="total > 0"
        :page="page"
        :pageSize="limit"
        :total="total"
        @update:page="changePage"
        @update:pageSize="changeLimit"
      />
    </div>

    <!-- Detail Modal -->
    <Modal
      :isOpen="!!activeLog"
      :title="t('webhooks.deliveryDetails')"
      width="720px"
      @close="activeLog = null"
    >
      <div v-if="activeLog" class="detail-content">
        <div class="meta-info-grid">
          <div class="meta-item">
            <span class="meta-lbl">{{ t('webhooks.deliveryId') }}</span>
            <code dir="ltr">{{ activeLog.id }}</code>
          </div>
          <div class="meta-item">
            <span class="meta-lbl">{{ t('webhooks.event') }}</span>
            <strong dir="ltr">{{ activeLog.event_name || activeLog.eventName }}</strong>
          </div>
          <div class="meta-item">
            <span class="meta-lbl">{{ t('webhooks.destinationUrl') }}</span>
            <span class="url-val" dir="ltr">{{ activeLog.target_url || activeLog.targetUrl }}</span>
          </div>
          <div class="meta-item">
            <span class="meta-lbl">{{ t('webhooks.attempts') }}</span>
            <span>{{ activeLog.attempt_count || activeLog.attemptCount || 1 }}</span>
          </div>
        </div>

        <!-- HMAC Signature -->
        <div class="detail-section">
          <span class="section-title">{{ t('webhooks.hmacHeader') }}</span>
          <div class="code-box" dir="ltr">
            <code>{{ activeLog.signature || 'None' }}</code>
          </div>
        </div>

        <!-- Payload JSON -->
        <div class="detail-section">
          <span class="section-title">{{ t('webhooks.requestPayload') }}</span>
          <pre class="code-box-pre" dir="ltr">{{ formatJSON(activeLog.payload_json || activeLog.payloadJson) }}</pre>
        </div>

        <!-- Response Body -->
        <div class="detail-section">
          <div class="section-header-flex">
            <span class="section-title">{{ t('webhooks.destinationResponse') }}</span>
            <span class="badge" :class="getHttpCodeClass(activeLog.response_status_code || activeLog.responseStatusCode)">
              HTTP {{ activeLog.response_status_code || activeLog.responseStatusCode || 'None' }}
            </span>
          </div>
          <pre class="code-box-pre" dir="ltr">{{ activeLog.response_body || activeLog.responseBody || activeLog.error_message || 'No response recorded.' }}</pre>
        </div>
      </div>

      <template #footer>
        <button class="btn btn-secondary btn-sm" @click="activeLog = null">{{ t('common.close') }}</button>
        <button class="btn btn-primary btn-sm" @click="retryCall(activeLog)" :disabled="retryingId === activeLog?.id">
          <RotateCw :size="13" :class="{ 'spin-anim': retryingId === activeLog?.id }" /> {{ t('webhooks.retry') }}
        </button>
      </template>
    </Modal>

    <!-- Quick Test / Send Webhook Modal -->
    <Modal
      :isOpen="isDispatchModalOpen"
      :title="t('apps.dispatchTitle')"
      width="680px"
      @close="closeDispatchModal"
    >
      <form @submit.prevent="executeDispatch" class="dispatch-form">
        <div class="form-row-grid">
          <div class="form-group">
            <label class="form-label">{{ t('apps.registeredApps') }} *</label>
            <select v-model="dispatchForm.app_id" required class="form-control" @change="onDispatchAppChange">
              <option value="" disabled>Select target app...</option>
              <option v-for="app in apps" :key="app.id" :value="app.app_id || app.appId || app.id">
                {{ app.name }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">{{ t('apps.eventName') }} *</label>
            <input
              v-model="dispatchForm.event_name"
              type="text"
              required
              class="form-control"
              placeholder="e.g. order.created"
              dir="ltr"
            />
          </div>
        </div>

        <div class="form-group">
          <div class="d-flex justify-between items-center mb-1">
            <label class="form-label mb-0">{{ t('apps.overrideUrl') }}</label>
            <button
              type="button"
              class="btn-text-link text-primary text-xs"
              @click="dispatchForm.target_url_override = 'http://localhost:8080/api/v1/webhooks/test-receiver'"
            >
              ⚡ Use Local Receiver
            </button>
          </div>
          <input
            v-model="dispatchForm.target_url_override"
            type="url"
            class="form-control"
            placeholder="http://localhost:8080/api/v1/webhooks/test-receiver"
            dir="ltr"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('apps.payloadJson') }} *</label>
          <textarea
            v-model="dispatchForm.payload_str"
            rows="5"
            required
            class="form-control code-editor"
            dir="ltr"
          ></textarea>
        </div>

        <!-- Result Box -->
        <div v-if="dispatchResult" class="test-result-box" :class="dispatchResult.success ? 'result-success' : 'result-error'">
          <div class="result-header">
            <div class="result-title">
              <CheckCircle2 v-if="dispatchResult.success" :size="16" class="text-success" />
              <AlertCircle v-else :size="16" class="text-danger" />
              <strong>Status: {{ dispatchResult.data?.status || 'UNKNOWN' }}</strong>
              <span v-if="dispatchResult.data?.response_status_code" class="badge ml-2" :class="dispatchResult.data?.response_status_code < 400 ? 'badge-success' : 'badge-danger'">
                HTTP {{ dispatchResult.data?.response_status_code }}
              </span>
              <span v-if="dispatchResult.data?.latency_ms" class="badge badge-secondary ml-1">
                {{ dispatchResult.data?.latency_ms }} ms
              </span>
            </div>
          </div>
          <div v-if="dispatchResult.data?.signature" class="result-sig" dir="ltr">
            <span class="text-muted">HMAC-SHA256:</span>
            <code>{{ dispatchResult.data?.signature }}</code>
          </div>
          <div v-if="dispatchResult.data?.response_body" class="result-body-preview" dir="ltr">
            <span class="text-muted">Response:</span>
            <pre>{{ dispatchResult.data?.response_body }}</pre>
          </div>
        </div>
      </form>

      <template #footer>
        <button class="btn btn-secondary btn-sm" @click="closeDispatchModal">{{ t('common.close') }}</button>
        <button class="btn btn-primary btn-sm" @click="executeDispatch" :disabled="sending">
          <RefreshCw v-if="sending" :size="14" class="spin-anim" />
          <Send v-else :size="14" /> {{ t('apps.sendNow') }}
        </button>
      </template>
    </Modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '../locales'
import { webhookService } from '../services/webhookService'
import { appService } from '../services/appService'
import { useToastStore } from '../stores/toast'
import Modal from '../components/common/Modal.vue'
import Pagination from '../components/common/Pagination.vue'
import EmptyState from '../components/common/EmptyState.vue'
import {
  RefreshCw,
  Search,
  Zap,
  Activity,
  Eye,
  RotateCw,
  Layers,
  Send,
  CheckCircle2,
  AlertCircle,
  Copy,
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()

const { t } = useI18n()
const toastStore = useToastStore()

const loading = ref(false)
const sending = ref(false)
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

// Dispatch Modal state
const isDispatchModalOpen = ref(false)
const dispatchResult = ref(null)
const dispatchForm = ref({
  app_id: '',
  event_name: 'order.created',
  target_url_override: 'http://localhost:8080/api/v1/webhooks/test-receiver',
  payload_str: JSON.stringify(
    {
      event: 'order.created',
      timestamp: new Date().toISOString(),
      order_id: 'ord_' + Math.floor(Math.random() * 900000 + 100000),
      amount: 149.99,
      currency: 'USD',
      customer: {
        id: 'cust_8821',
        email: 'customer@acme-corp.com',
      },
    },
    null,
    2
  ),
})

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
    if (selectedAppId.value && !dispatchForm.value.app_id) {
      dispatchForm.value.app_id = selectedAppId.value
    } else if (apps.value.length > 0 && !dispatchForm.value.app_id) {
      dispatchForm.value.app_id = apps.value[0].app_id || apps.value[0].appId || apps.value[0].id
    }
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

function syncFromQuery() {
  if (route.query.app_id) selectedAppId.value = String(route.query.app_id)
  if (route.query.status) selectedStatus.value = String(route.query.status)
  if (route.query.search) searchQuery.value = String(route.query.search)
  if (route.query.page) page.value = parseInt(route.query.page, 10) || 1
  if (route.query.limit) limit.value = parseInt(route.query.limit, 10) || 15
  if (route.query.send === 'true' || route.query.test === 'true') {
    openDispatchModal()
  }
}

function handleFilterChange() {
  page.value = 1
  updateUrlQuery()
  fetchLogs()
}

function updateUrlQuery() {
  const query = {}
  if (selectedAppId.value) query.app_id = selectedAppId.value
  if (selectedStatus.value) query.status = selectedStatus.value
  if (searchQuery.value) query.search = searchQuery.value
  if (page.value > 1) query.page = page.value
  if (limit.value !== 15) query.limit = limit.value
  router.replace({ query })
}

function changePage(newPage) {
  page.value = newPage
  updateUrlQuery()
  fetchLogs()
}

function changeLimit(newLimit) {
  limit.value = newLimit
  page.value = 1
  updateUrlQuery()
  fetchLogs()
}

function openDetail(call) {
  activeLog.value = call
}

function openDispatchModal() {
  dispatchResult.value = null
  if (selectedAppId.value) {
    dispatchForm.value.app_id = selectedAppId.value
  } else if (apps.value.length > 0) {
    dispatchForm.value.app_id = apps.value[0].app_id || apps.value[0].appId || apps.value[0].id
  }
  isDispatchModalOpen.value = true
}

function closeDispatchModal() {
  isDispatchModalOpen.value = false
  dispatchResult.value = null
}

function onDispatchAppChange() {
  const app = apps.value.find(a => (a.app_id || a.appId || a.id) === dispatchForm.value.app_id)
  if (app && (app.webhook_url || app.webhookUrl)) {
    dispatchForm.value.target_url_override = app.webhook_url || app.webhookUrl
  }
}

async function executeDispatch() {
  sending.value = true
  dispatchResult.value = null
  try {
    let parsedPayload = {}
    try {
      parsedPayload = JSON.parse(dispatchForm.value.payload_str)
    } catch {
      toastStore.error('Invalid JSON payload formatting.')
      sending.value = false
      return
    }

    const res = await webhookService.sendWebhook({
      app_id: dispatchForm.value.app_id,
      event_name: dispatchForm.value.event_name,
      target_url_override: dispatchForm.value.target_url_override,
      payload: parsedPayload,
    })

    dispatchResult.value = res
    if (res.success) {
      toastStore.success(`Webhook delivered! Status: ${res.data?.status}`)
    } else {
      toastStore.warning(`Webhook completed with status: ${res.data?.status}`)
    }
    await fetchLogs()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to dispatch webhook.')
  } finally {
    sending.value = false
  }
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

watch(() => route.query, () => {
  syncFromQuery()
  fetchLogs()
})

onMounted(async () => {
  syncFromQuery()
  await fetchApps()
  await fetchLogs()
})
</script>

<style scoped>
.webhook-logs-view {
  padding: 1.5rem;
}

.filter-card {
  padding: 0.875rem 1.25rem;
}

.filter-controls-row {
  display: flex;
  gap: 1rem;
  align-items: center;
  flex-wrap: wrap;
}

.search-input-wrapper {
  position: relative;
  flex: 1;
  min-width: 260px;
}

.search-icon {
  position: absolute;
  inset-inline-start: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
}

.search-field {
  padding-inline-start: 2.25rem;
}

.select-filters-group {
  display: flex;
  gap: 0.75rem;
}

.select-field {
  min-width: 160px;
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
  text-align: start;
  font-size: 0.875rem;
}

.data-table th {
  background: #f8fafc;
  padding: 0.75rem 1.25rem;
  font-weight: 700;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border-color);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.data-table td {
  padding: 0.875rem 1.25rem;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
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
  max-width: 240px;
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

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  display: inline-block;
  margin-inline-end: 4px;
}

.action-buttons-group {
  display: flex;
  gap: 0.35rem;
  justify-content: flex-end;
}

/* Detail Modal */
.detail-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.meta-info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem 1.5rem;
  background: #f8fafc;
  padding: 1rem;
  border-radius: var(--radius-md);
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.meta-lbl {
  font-size: 0.7rem;
  text-transform: uppercase;
  font-weight: 700;
  color: var(--text-muted);
}

.detail-section {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.section-title {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
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
