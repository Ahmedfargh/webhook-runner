<template>
  <div class="audit-view animate-fade-in">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('audit.title') }}</h1>
        <p class="view-subtitle">{{ t('audit.subtitle') }}</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-secondary btn-sm" @click="fetchLogs" :disabled="loading">
          <RefreshCw :size="14" :class="{ 'spin-anim': loading }" /> {{ t('audit.refreshStream') }}
        </button>
      </div>
    </div>

    <!-- Quick Stats Cards -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon-wrapper icon-blue">
          <History :size="18" />
        </div>
        <div class="stat-info">
          <span class="stat-label">{{ t('audit.totalEvents') }}</span>
          <span class="stat-value">{{ total }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrapper icon-green">
          <CheckCircle2 :size="18" />
        </div>
        <div class="stat-info">
          <span class="stat-label">{{ t('audit.successfulOps') }}</span>
          <span class="stat-value text-success">{{ successCount }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrapper icon-red">
          <AlertTriangle :size="18" />
        </div>
        <div class="stat-info">
          <span class="stat-label">{{ t('audit.failedOps') }}</span>
          <span class="stat-value" :class="failedCount > 0 ? 'text-danger' : 'text-muted'">{{ failedCount }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrapper icon-purple">
          <Server :size="18" />
        </div>
        <div class="stat-info">
          <span class="stat-label">{{ t('audit.activeServices') }}</span>
          <span class="stat-value">{{ t('audit.servicesCount', { count: 5 }) }}</span>
        </div>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="card filter-card">
      <div class="filter-controls-row">
        <div class="search-input-wrapper">
          <Search :size="15" class="search-icon" />
          <input
            v-model="filters.search"
            type="text"
            class="form-control search-field"
            :placeholder="t('audit.searchPlaceholder')"
            @input="debounceFetch"
          />
        </div>

        <div class="select-filters-group">
          <select v-model="filters.service" class="form-control select-field" @change="fetchLogs">
            <option value="">{{ t('audit.allServices') }}</option>
            <option value="api-gateway">API Gateway</option>
            <option value="accounts">Accounts Service</option>
            <option value="subscriptions">Subscriptions Service</option>
            <option value="webhook-runner">Webhook Runner</option>
            <option value="audit-service">Audit Service</option>
          </select>

          <select v-model="filters.action" class="form-control select-field" @change="fetchLogs">
            <option value="">{{ t('audit.allActions') }}</option>
            <option value="CREATE">CREATE</option>
            <option value="UPDATE">UPDATE</option>
            <option value="DELETE">DELETE</option>
            <option value="LOGIN">LOGIN</option>
            <option value="OVERRIDE">OVERRIDE</option>
            <option value="ROTATE_SECRET">ROTATE_SECRET</option>
            <option value="DISPATCH">DISPATCH</option>
            <option value="RETRY">RETRY</option>
          </select>

          <select v-model="filters.status" class="form-control select-field" @change="fetchLogs">
            <option value="">{{ t('audit.allStatuses') }}</option>
            <option value="SUCCESS">SUCCESS</option>
            <option value="FAILED">FAILED</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Logs Table Container -->
    <div class="zoho-table-container">
      <table class="zoho-table">
        <thead>
          <tr>
            <th style="width: 170px;">{{ t('audit.timestamp') }}</th>
            <th style="width: 140px;">{{ t('audit.service') }}</th>
            <th>{{ t('audit.actor') }}</th>
            <th style="width: 120px;">{{ t('audit.action') }}</th>
            <th style="width: 150px;">{{ t('audit.resource') }}</th>
            <th style="width: 130px;">{{ t('audit.ipAddress') }}</th>
            <th style="width: 100px;">{{ t('audit.status') }}</th>
            <th style="width: 100px; text-align: right;">{{ t('audit.details') }}</th>
          </tr>
        </thead>
        <tbody>
          <!-- Loading State -->
          <tr v-if="loading && logs.length === 0">
            <td colspan="8" class="text-center py-12">
              <div class="loading-wrapper">
                <RefreshCw :size="26" class="spin-anim text-primary" />
                <span class="mt-2 text-muted">{{ t('audit.streaming') }}</span>
              </div>
            </td>
          </tr>

          <!-- Empty State -->
          <tr v-else-if="logs.length === 0">
            <td colspan="8" class="text-center py-12">
              <EmptyState
                :title="t('audit.noRecordsTitle')"
                :description="t('audit.noRecordsDesc')"
                :icon="History"
              />
            </td>
          </tr>

          <!-- Data Rows -->
          <tr v-else v-for="log in logs" :key="log.id" class="table-row hoverable" @click="openInspectModal(log)">
            <td>
              <span class="font-mono text-xs text-muted" dir="ltr">
                {{ formatDateTime(log.created_at) }}
              </span>
            </td>
            <td>
              <span class="badge badge-service font-mono">
                {{ log.service_name || 'api-gateway' }}
              </span>
            </td>
            <td>
              <div class="actor-cell">
                <div class="actor-avatar">
                  {{ (log.actor_name || log.actor_email || 'A').charAt(0).toUpperCase() }}
                </div>
                <div class="actor-info">
                  <div class="actor-name">{{ log.actor_name || 'System Actor' }}</div>
                  <div class="actor-email font-mono">{{ log.actor_email || log.actor_type || 'system' }}</div>
                </div>
              </div>
            </td>
            <td>
              <span class="badge" :class="getActionBadgeClass(log.action)">
                {{ log.action }}
              </span>
            </td>
            <td>
              <div class="resource-cell">
                <span class="resource-name">{{ log.resource }}</span>
                <span v-if="log.resource_id" class="resource-id font-mono">
                  #{{ log.resource_id.substring(0, 8) }}
                </span>
              </div>
            </td>
            <td>
              <span class="font-mono text-xs text-secondary" dir="ltr">
                {{ log.ip_address || '127.0.0.1' }}
              </span>
            </td>
            <td>
              <span class="badge" :class="log.status === 'SUCCESS' ? 'badge-success' : 'badge-danger'">
                <span class="status-dot"></span>
                {{ log.status }}
              </span>
            </td>
            <td style="text-align: right;">
              <button class="btn btn-sm btn-secondary" @click.stop="openInspectModal(log)">
                <Eye :size="13" /> {{ t('audit.inspect') }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Shared Pagination -->
      <Pagination
        v-if="total > 0"
        :page="filters.page"
        :page-size="filters.limit"
        :total="total"
        @update:page="changePage"
        @update:pageSize="changePageSize"
      />
    </div>

    <!-- Inspector Modal -->
    <Modal
      :is-open="!!activeModalLog"
      :title="t('audit.modalTitle')"
      width="720px"
      @close="activeModalLog = null"
    >
      <div v-if="activeModalLog" class="modal-body-wrapper">
        <!-- Metadata Overview -->
        <div class="meta-grid">
          <div class="meta-item">
            <span class="meta-label">{{ t('audit.actorIdentity') }}</span>
            <span class="meta-value">{{ activeModalLog.actor_name || activeModalLog.actor_email || 'System User' }}</span>
            <span class="meta-sub font-mono">{{ activeModalLog.actor_type || 'SYSTEM' }} &bull; {{ activeModalLog.actor_email || 'n/a' }}</span>
          </div>

          <div class="meta-item">
            <span class="meta-label">{{ t('audit.originService') }}</span>
            <span class="meta-value">
              <span class="badge badge-service font-mono">{{ activeModalLog.service_name }}</span>
            </span>
            <span class="meta-sub font-mono">
              {{ t('audit.status') }}:
              <strong :class="activeModalLog.status === 'SUCCESS' ? 'text-success' : 'text-danger'">
                {{ activeModalLog.status }}
              </strong>
            </span>
          </div>

          <div class="meta-item">
            <span class="meta-label">{{ t('audit.actionAndResource') }}</span>
            <span class="meta-value">
              <span class="badge" :class="getActionBadgeClass(activeModalLog.action)">{{ activeModalLog.action }}</span>
              <strong class="ml-1 font-mono">{{ activeModalLog.resource }}</strong>
            </span>
            <span v-if="activeModalLog.resource_id" class="meta-sub font-mono">ID: {{ activeModalLog.resource_id }}</span>
          </div>

          <div class="meta-item">
            <span class="meta-label">{{ t('audit.networkAndTime') }}</span>
            <span class="meta-value font-mono text-xs" dir="ltr">{{ activeModalLog.ip_address || '127.0.0.1' }}</span>
            <span class="meta-sub font-mono" dir="ltr">{{ formatDateTime(activeModalLog.created_at) }}</span>
          </div>
        </div>

        <!-- Failure / Error Banner if failed -->
        <div v-if="activeModalLog.status === 'FAILED' || activeModalLog.error_message" class="alert-box alert-danger mt-3">
          <div class="alert-header">
            <AlertTriangle :size="16" />
            <strong>{{ t('audit.execError') }}</strong>
          </div>
          <p class="alert-body font-mono text-xs mt-1">
            {{ activeModalLog.error_message || 'Operation returned HTTP 4xx/5xx failure status.' }}
          </p>
        </div>

        <!-- Client User Agent -->
        <div v-if="activeModalLog.user_agent" class="user-agent-box mt-3 font-mono">
          <span class="text-muted">{{ t('audit.userAgent') }}:</span> {{ activeModalLog.user_agent }}
        </div>

        <!-- Captured Request / State Payload -->
        <div class="payload-box mt-3">
          <div class="payload-header">
            <span class="payload-title">{{ t('audit.changePayload') }}</span>
            <span class="payload-type font-mono">{{ t('audit.payloadType') }}</span>
          </div>
          <pre class="json-code font-mono" dir="ltr">{{ formatJSON(activeModalLog.after_json || activeModalLog.before_json || '{}') }}</pre>
        </div>

        <!-- Before State Diff if present -->
        <div v-if="activeModalLog.before_json && activeModalLog.after_json" class="payload-box mt-3">
          <div class="payload-header">
            <span class="payload-title">Previous State Snapshot (Before Change)</span>
            <span class="payload-type font-mono">BEFORE</span>
          </div>
          <pre class="json-code font-mono" dir="ltr">{{ formatJSON(activeModalLog.before_json) }}</pre>
        </div>
      </div>

      <template #footer>
        <button class="btn btn-secondary" @click="activeModalLog = null">{{ t('common.close') }}</button>
      </template>
    </Modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api from '../services/api'
import { t } from '../locales'
import EmptyState from '../components/common/EmptyState.vue'
import Pagination from '../components/common/Pagination.vue'
import Modal from '../components/common/Modal.vue'
import {
  History,
  Search,
  RefreshCw,
  Eye,
  CheckCircle2,
  AlertTriangle,
  Server,
} from 'lucide-vue-next'

const logs = ref([])
const total = ref(0)
const loading = ref(false)
const activeModalLog = ref(null)

const filters = reactive({
  search: '',
  service: '',
  action: '',
  status: '',
  page: 1,
  limit: 15,
})

const successCount = computed(() => {
  return logs.value.filter(l => l.status === 'SUCCESS').length
})

const failedCount = computed(() => {
  return logs.value.filter(l => l.status === 'FAILED').length
})

let debounceTimer = null
function debounceFetch() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    filters.page = 1
    fetchLogs()
  }, 300)
}

async function fetchLogs() {
  loading.value = true
  try {
    const params = {
      page: filters.page,
      limit: filters.limit,
      search: filters.search || undefined,
      service_name: filters.service || undefined,
      action: filters.action || undefined,
      status: filters.status || undefined,
    }
    const res = await api.get('/audit-logs', { params })
    logs.value = res.data.data || []
    total.value = res.data.total || 0
  } catch (err) {
    console.error('Failed to fetch audit logs:', err)
  } finally {
    loading.value = false
  }
}

function changePage(page) {
  filters.page = page
  fetchLogs()
}

function changePageSize(size) {
  filters.limit = size
  filters.page = 1
  fetchLogs()
}

function openInspectModal(log) {
  activeModalLog.value = log
}

function formatDateTime(iso) {
  if (!iso) return '-'
  const d = new Date(iso)
  return d.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: true,
  })
}

function formatJSON(raw) {
  if (!raw) return '{}'
  try {
    const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
    return JSON.stringify(parsed, null, 2)
  } catch {
    return raw
  }
}

function getActionBadgeClass(action) {
  switch (action) {
    case 'CREATE':
      return 'badge-success'
    case 'UPDATE':
      return 'badge-warning'
    case 'DELETE':
      return 'badge-danger'
    case 'LOGIN':
      return 'badge-info'
    case 'DISPATCH':
    case 'ROTATE_SECRET':
    case 'OVERRIDE':
    case 'RETRY':
      return 'badge-primary'
    default:
      return 'badge-secondary'
  }
}

onMounted(() => {
  fetchLogs()
})
</script>

<style scoped>
.audit-view {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  padding: 1.5rem;
}

.view-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.view-title {
  font-size: 1.35rem;
  font-weight: 800;
  color: var(--text-primary);
  margin-bottom: 0.25rem;
}

.view-subtitle {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

/* Stats Row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(210px, 1fr));
  gap: 1rem;
}

.stat-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 1rem 1.25rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  box-shadow: var(--shadow-sm);
}

.stat-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.icon-blue { background: #eff6ff; color: #2563eb; }
.icon-green { background: #ecfdf5; color: #059669; }
.icon-red { background: #fef2f2; color: #dc2626; }
.icon-purple { background: #faf5ff; color: #9333ea; }

.stat-info {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.stat-label {
  font-size: 0.725rem;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--text-primary);
}

/* Filters Bar */
.filter-card {
  padding: 0.875rem 1.25rem;
  background: var(--bg-card);
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
  pointer-events: none;
}

.search-field {
  padding-inline-start: 2.25rem;
  height: 38px;
}

.select-filters-group {
  display: flex;
  gap: 0.75rem;
}

.select-field {
  min-width: 150px;
  height: 38px;
}

/* Table Enhancements */
.zoho-table-container {
  overflow-x: auto;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  background-color: var(--bg-card);
  box-shadow: var(--shadow-sm);
}

.zoho-table {
  width: 100%;
  border-collapse: collapse;
  text-align: start;
}

.zoho-table th {
  background-color: #f8fafc;
  padding: 0.75rem 1rem;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-color);
  white-space: nowrap;
  text-align: start;
}

.zoho-table td {
  padding: 0.75rem 1rem;
  font-size: 13px;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-light);
  vertical-align: middle;
  text-align: start;
}

.table-row.hoverable {
  cursor: pointer;
  transition: background-color var(--transition-fast);
}

.table-row.hoverable:hover {
  background-color: #f8fafc;
}

/* Cell Elements */
.badge-service {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #cbd5e1;
  font-size: 11px;
}

.actor-cell {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.actor-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: #e0f2fe;
  color: #0284c7;
  font-weight: 700;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 1px solid #bae6fd;
}

.actor-info {
  display: flex;
  flex-direction: column;
}

.actor-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 13px;
  line-height: 1.2;
}

.actor-email {
  font-size: 11px;
  color: var(--text-muted);
}

.resource-cell {
  display: flex;
  flex-direction: column;
}

.resource-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 13px;
}

.resource-id {
  font-size: 11px;
  color: var(--text-muted);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  display: inline-block;
  margin-inline-end: 4px;
}

.loading-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}

/* Modal Content */
.modal-body-wrapper {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.meta-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  background: #f8fafc;
  padding: 1rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.meta-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-muted);
  letter-spacing: 0.05em;
}

.meta-value {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.meta-sub {
  font-size: 11px;
  color: var(--text-muted);
}

.user-agent-box {
  background: #f1f5f9;
  padding: 0.6rem 0.85rem;
  border-radius: var(--radius-sm);
  font-size: 11px;
  color: var(--text-secondary);
  word-break: break-all;
}

.alert-box {
  padding: 0.75rem 1rem;
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.alert-danger {
  background: #fef2f2;
  color: #991b1b;
  border: 1px solid #fecaca;
}

.payload-box {
  background: #0f172a;
  border-radius: var(--radius-md);
  overflow: hidden;
}

.payload-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 0.85rem;
  background: #1e293b;
  border-bottom: 1px solid #334155;
}

.payload-title {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: #94a3b8;
  letter-spacing: 0.05em;
}

.payload-type {
  font-size: 10px;
  font-weight: 700;
  background: #334155;
  color: #38bdf8;
  padding: 0.15rem 0.4rem;
  border-radius: 4px;
}

.json-code {
  margin: 0;
  padding: 0.85rem;
  color: #38bdf8;
  font-size: 12px;
  line-height: 1.5;
  max-height: 250px;
  overflow-y: auto;
}

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
