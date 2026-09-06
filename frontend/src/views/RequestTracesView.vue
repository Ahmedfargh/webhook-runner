<template>
  <div class="traces-view animate-fade-in">
    <!-- View Header -->
    <div class="view-header">
      <div class="view-title-group">
        <h1 class="view-title">تتبع الطلبات وأداء النظام (APM)</h1>
        <p class="view-subtitle">مراقبة دورة حياة طلبات الـ HTTP من البداية حتى النهاية، زمن الاستجابة، وحمولات الردود</p>
      </div>
      <div class="header-actions">
        <label class="auto-refresh-toggle">
          <input type="checkbox" v-model="autoRefresh" @change="toggleAutoRefresh" />
          <span class="toggle-label">تحديث تلقائي (5ث)</span>
        </label>
        <button class="btn btn-secondary btn-sm" @click="fetchData" :disabled="loading">
          <RefreshCw :size="14" :class="{ 'spin-anim': loading }" /> تحديث
        </button>
      </div>
    </div>

    <!-- Quick Stats Cards -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-icon-wrapper icon-blue">
          <Activity :size="20" />
        </div>
        <div class="stat-info">
          <span class="stat-label">إجمالي الطلبات</span>
          <span class="stat-value font-mono" dir="ltr">{{ stats.total_requests || total }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrapper icon-cyan">
          <Timer :size="20" />
        </div>
        <div class="stat-info">
          <span class="stat-label">متوسط زمن الاستجابة</span>
          <span class="stat-value font-mono" :class="getLatencyColor(stats.avg_lifetime_ms || avgLifetime)" dir="ltr">
            {{ (stats.avg_lifetime_ms || avgLifetime).toFixed(1) }} <span class="unit">ms</span>
          </span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrapper icon-purple">
          <Gauge :size="20" />
        </div>
        <div class="stat-info">
          <span class="stat-label">مؤشر P95 / P99</span>
          <span class="stat-value font-mono text-secondary" dir="ltr">
            {{ (stats.p95_lifetime_ms || 0).toFixed(0) }} / {{ (stats.p99_lifetime_ms || 0).toFixed(0) }} <span class="unit">ms</span>
          </span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon-wrapper" :class="(stats.error_count > 0 || failedCount > 0) ? 'icon-red' : 'icon-green'">
          <AlertTriangle v-if="stats.error_count > 0 || failedCount > 0" :size="20" />
          <CheckCircle2 v-else :size="20" />
        </div>
        <div class="stat-info">
          <span class="stat-label">أخطاء (4xx / 5xx)</span>
          <span class="stat-value font-mono" :class="(stats.error_count > 0 || failedCount > 0) ? 'text-danger' : 'text-success'" dir="ltr">
            {{ stats.error_count || failedCount }} <span class="unit">({{ (stats.error_rate || 0).toFixed(1) }}%)</span>
          </span>
        </div>
      </div>
    </div>

    <!-- Filters Bar -->
    <div class="card filter-card">
      <div class="filter-controls-row">
        <div class="search-input-wrapper">
          <Search :size="15" class="search-icon" />
          <input
            v-model="filters.search"
            type="text"
            class="form-control search-field"
            placeholder="بحث بواسطة Trace ID، المسار، أو البريد الإلكتروني..."
            @input="debounceFetch"
          />
        </div>

        <div class="select-filters-group">
          <!-- Method Filter -->
          <select v-model="filters.method" class="form-control select-field" @change="fetchTraces">
            <option value="">كل الطرق (Methods)</option>
            <option value="GET">GET</option>
            <option value="POST">POST</option>
            <option value="PUT">PUT</option>
            <option value="PATCH">PATCH</option>
            <option value="DELETE">DELETE</option>
          </select>

          <!-- Status Code Filter -->
          <select v-model="filters.statusFilter" class="form-control select-field" @change="onStatusFilterChange">
            <option value="">كل الحالات</option>
            <option value="200">200 OK</option>
            <option value="201">201 Created</option>
            <option value="400">400 Bad Request</option>
            <option value="401">401 Unauthorized</option>
            <option value="403">403 Forbidden</option>
            <option value="404">404 Not Found</option>
            <option value="500">500 Server Error</option>
          </select>

          <!-- Actor Type Filter -->
          <select v-model="filters.actorType" class="form-control select-field" @change="fetchTraces">
            <option value="">كل الفئات</option>
            <option value="ADMIN">مسؤول (ADMIN)</option>
            <option value="USER">مستخدم (USER)</option>
            <option value="ANONYMOUS">مجهول (ANONYMOUS)</option>
          </select>

          <!-- Min Lifetime Filter -->
          <select v-model="filters.minLifetime" class="form-control select-field" @change="fetchTraces">
            <option :value="0">كل السرعات</option>
            <option :value="20">&gt; 20 ms</option>
            <option :value="50">&gt; 50 ms</option>
            <option :value="100">&gt; 100 ms (بطيء)</option>
            <option :value="500">&gt; 500 ms (حرج)</option>
          </select>

          <button
            v-if="hasActiveFilters"
            class="btn btn-ghost btn-sm text-danger"
            @click="resetFilters"
            title="إعادة تعيين الفلاتر"
          >
            مسح
          </button>
        </div>
      </div>
    </div>

    <!-- Zoho-Styled Traces Table Container -->
    <div class="zoho-table-container">
      <table class="zoho-table">
        <thead>
          <tr>
            <th style="width: 130px;">الوقت</th>
            <th style="width: 80px;">الحالة</th>
            <th style="width: 80px;">الطريقة</th>
            <th style="min-width: 240px; max-width: 320px;">المسار (Endpoint Route)</th>
            <th style="width: 150px;">مدة المعالجة (Lifetime)</th>
            <th style="width: 180px;">المستخدم / الفاعل</th>
            <th style="width: 120px;">عنوان IP</th>
            <th style="width: 70px; text-align: center;">معاينة</th>
          </tr>
        </thead>
        <tbody>
          <!-- Loading State -->
          <tr v-if="loading && traces.length === 0">
            <td colspan="8" class="text-center py-10 text-muted">
              <RefreshCw :size="24" class="spin-anim mb-2 text-primary" />
              <div>جاري تحميل سجلات التتبع...</div>
            </td>
          </tr>

          <!-- Empty State -->
          <tr v-else-if="traces.length === 0">
            <td colspan="8" class="text-center py-10 text-muted">
              <Activity :size="32" class="mb-2 text-muted" />
              <div class="font-bold text-base">لا توجد سجلات تتبع حالياً</div>
              <div class="text-xs text-secondary mt-1">قم بإرسال أي طلب إلى بوابة الـ API للبدء في جمع بيانات التتبع اللحظية.</div>
            </td>
          </tr>

          <!-- Data Rows -->
          <tr
            v-else
            v-for="trace in traces"
            :key="trace.id || trace.trace_id"
            class="table-row hoverable"
            @click="openInspectModal(trace)"
          >
            <td>
              <span class="font-mono text-xs text-muted" dir="ltr">
                {{ formatDateTime(trace.received_at) }}
              </span>
            </td>
            <td>
              <span class="badge font-mono font-bold" :class="getStatusBadgeClass(trace.status_code)">
                {{ trace.status_code }}
              </span>
            </td>
            <td>
              <span class="badge font-mono font-bold" :class="getMethodBadgeClass(trace.method)">
                {{ trace.method }}
              </span>
            </td>
            <td>
              <div class="route-cell" dir="ltr">
                <span class="route-main font-mono text-xs font-bold" :title="trace.path">
                  {{ trace.path }}
                </span>
                <span v-if="trace.route && trace.route !== trace.path" class="route-sub font-mono text-secondary" :title="trace.route">
                  {{ trace.route }}
                </span>
              </div>
            </td>
            <td>
              <div class="lifetime-wrapper" dir="ltr">
                <span class="lifetime-text font-mono font-bold text-xs" :class="getLatencyColor(trace.lifetime_ms)">
                  {{ trace.lifetime_ms.toFixed(1) }} ms
                </span>
                <div class="latency-bar-track">
                  <div
                    class="latency-bar-fill"
                    :class="getLatencyBgColor(trace.lifetime_ms)"
                    :style="{ width: Math.min(100, (trace.lifetime_ms / 150) * 100) + '%' }"
                  ></div>
                </div>
              </div>
            </td>
            <td>
              <div class="actor-cell">
                <div class="actor-avatar" :class="trace.actor_type === 'ADMIN' ? 'avatar-admin' : ''">
                  {{ (trace.actor_name || trace.actor_email || trace.actor_type || 'A').charAt(0).toUpperCase() }}
                </div>
                <div class="actor-info">
                  <div class="actor-name">{{ trace.actor_name || (trace.actor_type === 'ADMIN' ? 'مسؤول النظام' : (trace.actor_type === 'USER' ? 'مستخدم' : 'زائر مجهول')) }}</div>
                  <div class="actor-email font-mono" dir="ltr">{{ trace.actor_email || trace.actor_type }}</div>
                </div>
              </div>
            </td>
            <td>
              <span class="font-mono text-xs text-secondary" dir="ltr">
                {{ trace.client_ip || '127.0.0.1' }}
              </span>
            </td>
            <td class="text-center" @click.stop>
              <button class="btn btn-ghost btn-xs inspect-btn" @click="openInspectModal(trace)" title="معاينة تفاصيل الطلب">
                <Eye :size="15" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <Pagination
        :page="filters.page"
        :page-size="filters.limit"
        :total="total"
        @update:page="changePage"
        @update:pageSize="changePageSize"
      />
    </div>

    <!-- Inspect Request Trace Modal -->
    <Modal
      :is-open="!!activeTrace"
      title="مستكشف دورة حياة الطلب (Lifecycle Inspector)"
      width="820px"
      @close="activeTrace = null"
    >
      <div v-if="activeTrace" class="modal-body-wrapper">
        <!-- Trace Header Banner -->
        <div class="trace-header-card">
          <div class="trace-header-left" dir="ltr">
            <span class="badge font-mono text-sm font-bold" :class="getMethodBadgeClass(activeTrace.method)">
              {{ activeTrace.method }}
            </span>
            <span class="font-mono text-sm font-bold ml-2">{{ activeTrace.path }}</span>
          </div>
          <div class="trace-header-right" dir="ltr">
            <span class="badge font-mono text-sm font-bold" :class="getStatusBadgeClass(activeTrace.status_code)">
              {{ activeTrace.status_code }} {{ getStatusText(activeTrace.status_code) }}
            </span>
          </div>
        </div>

        <!-- Trace Meta Grid -->
        <div class="meta-grid mt-3">
          <div class="meta-item">
            <span class="meta-label">معرف التتبع (Trace ID)</span>
            <div class="copy-box" dir="ltr">
              <span class="meta-value font-mono text-xs font-semibold">{{ activeTrace.trace_id }}</span>
              <button class="copy-btn" @click="copyText(activeTrace.trace_id)" title="نسخ المعرف">
                <Copy :size="13" />
              </button>
            </div>
          </div>

          <div class="meta-item">
            <span class="meta-label">الزمن الكلي في النظام (Lifetime)</span>
            <span class="meta-value font-mono font-bold text-sm" :class="getLatencyColor(activeTrace.lifetime_ms)" dir="ltr">
              {{ activeTrace.lifetime_ms.toFixed(2) }} ms
            </span>
            <span class="meta-sub font-mono text-xs" dir="ltr">
              {{ formatDateTime(activeTrace.received_at) }} &rarr; {{ formatDateTime(activeTrace.completed_at) }}
            </span>
          </div>

          <div class="meta-item">
            <span class="meta-label">هوية الفاعل (Actor)</span>
            <span class="meta-value">
              <span class="badge badge-service font-mono mr-1">{{ activeTrace.actor_type }}</span>
              <strong>{{ activeTrace.actor_name || activeTrace.actor_email || 'غير مسجل' }}</strong>
            </span>
            <span v-if="activeTrace.actor_email" class="meta-sub font-mono text-xs" dir="ltr">{{ activeTrace.actor_email }}</span>
          </div>

          <div class="meta-item">
            <span class="meta-label">الشبكة والعميل</span>
            <span class="meta-value font-mono text-xs" dir="ltr">{{ activeTrace.client_ip }}</span>
            <span class="meta-sub text-xs truncate" :title="activeTrace.user_agent" dir="ltr">{{ activeTrace.user_agent }}</span>
          </div>
        </div>

        <!-- Query String if present -->
        <div v-if="activeTrace.query_params" class="param-box mt-3" dir="ltr">
          <span class="param-title">Query:</span>
          <code class="param-code font-mono text-xs">?{{ activeTrace.query_params }}</code>
        </div>

        <!-- Error banner if failed -->
        <div v-if="activeTrace.status_code >= 400 || activeTrace.error_message" class="alert-box alert-danger mt-3">
          <div class="alert-header">
            <AlertTriangle :size="16" />
            <strong>خطأ أثناء المعالجة (كود {{ activeTrace.status_code }})</strong>
          </div>
          <p class="alert-body font-mono text-xs mt-1" dir="ltr">
            {{ activeTrace.error_message || 'The request returned an HTTP 4xx or 5xx status code.' }}
          </p>
        </div>

        <!-- Request Trip, Response & Request JSON Inspector Tabs -->
        <div class="payload-tabs-section mt-4">
          <div class="payload-tab-headers">
            <button
              class="payload-tab-btn"
              :class="{ active: activePayloadTab === 'trip' }"
              @click="activePayloadTab = 'trip'"
            >
              <GitCommit :size="15" />
              رحلة الطلب ومخطط المسار (Request Trip) <span class="tab-badge font-mono">{{ activeTraceTrip.length }} مراحل</span>
            </button>
            <button
              class="payload-tab-btn"
              :class="{ active: activePayloadTab === 'response' }"
              @click="activePayloadTab = 'response'"
            >
              حمولة الرد (Response Payload) <span class="tab-badge font-mono">{{ activeTrace.status_code }}</span>
            </button>
            <button
              class="payload-tab-btn"
              :class="{ active: activePayloadTab === 'request' }"
              @click="activePayloadTab = 'request'"
            >
              حمولة الطلب (Request Body) <span v-if="activeTrace.request_body" class="tab-dot"></span>
            </button>
          </div>

          <!-- Trip Journey & Waterfall Panel -->
          <div v-if="activePayloadTab === 'trip'" class="trip-panel">
            <div class="trip-summary-bar">
              <div class="trip-summary-item">
                <span class="trip-summary-label">معرف الطلب الموحد:</span>
                <code class="font-mono text-xs font-bold" dir="ltr">{{ activeTrace.request_id || activeTrace.trace_id }}</code>
              </div>
              <div class="trip-summary-item">
                <span class="trip-summary-label">الزمن الإجمالي:</span>
                <span class="font-mono text-xs font-bold" :class="getLatencyColor(activeTrace.lifetime_ms)" dir="ltr">
                  {{ activeTrace.lifetime_ms.toFixed(2) }} ms
                </span>
              </div>
            </div>

            <!-- Waterfall Steps Timeline -->
            <div class="trip-timeline">
              <div
                v-for="(hop, index) in activeTraceTrip"
                :key="index"
                class="timeline-hop-card"
              >
                <!-- Hop Index / Dot -->
                <div class="hop-indicator">
                  <span class="hop-index font-mono">{{ index + 1 }}</span>
                  <div v-if="index < activeTraceTrip.length - 1" class="hop-line"></div>
                </div>

                <!-- Hop Content -->
                <div class="hop-content-box">
                  <div class="hop-header-row">
                    <div class="hop-title-group">
                      <span class="badge font-mono font-bold" :class="getProtocolBadgeClass(hop.protocol)">
                        {{ hop.protocol || 'REST' }}
                      </span>
                      <span class="hop-name font-mono text-xs font-bold" dir="ltr">{{ hop.name }}</span>
                    </div>
                    <div class="hop-timing-group" dir="ltr">
                      <span class="badge badge-service font-mono text-xs">{{ hop.service }}</span>
                      <span class="hop-duration font-mono text-xs font-bold" :class="getLatencyColor(hop.duration_ms)">
                        {{ hop.duration_ms.toFixed(2) }} ms
                      </span>
                      <span class="badge font-mono text-xs" :class="hop.status === 'OK' ? 'badge-success' : 'badge-danger'">
                        {{ hop.status }}
                      </span>
                    </div>
                  </div>

                  <!-- Details & Relative Gantt Progress Bar -->
                  <div class="hop-details-row">
                    <div v-if="hop.details" class="hop-details text-xs font-mono text-secondary" dir="ltr">
                      {{ hop.details }}
                    </div>
                    <div class="hop-progress-track">
                      <div
                        class="hop-progress-bar"
                        :class="getProtocolBgClass(hop.protocol)"
                        :style="{
                          width: Math.max(8, Math.min(100, (hop.duration_ms / (activeTrace.lifetime_ms || 1)) * 100)) + '%',
                          marginRight: ((hop.offset_ms || 0) / (activeTrace.lifetime_ms || 1)) * 100 + '%'
                        }"
                      ></div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Response Body Panel -->
          <div v-if="activePayloadTab === 'response'" class="payload-panel">
            <div class="payload-panel-toolbar">
              <span class="toolbar-title font-mono text-xs">Response Body</span>
              <button class="btn btn-ghost btn-xs" @click="copyText(activeTrace.response_body)">
                <Copy :size="12" /> نسخ الرد
              </button>
            </div>
            <pre class="json-code font-mono text-xs" dir="ltr"><code>{{ formatJSON(activeTrace.response_body) || '// Empty response body' }}</code></pre>
          </div>

          <!-- Request Body Panel -->
          <div v-if="activePayloadTab === 'request'" class="payload-panel">
            <div class="payload-panel-toolbar">
              <span class="toolbar-title font-mono text-xs">Request Body (Sanitized)</span>
              <button v-if="activeTrace.request_body" class="btn btn-ghost btn-xs" @click="copyText(activeTrace.request_body)">
                <Copy :size="12" /> نسخ الطلب
              </button>
            </div>
            <pre class="json-code font-mono text-xs" dir="ltr"><code>{{ formatJSON(activeTrace.request_body) || '// لا توجد حمولة مرسلة مع الطلب' }}</code></pre>
          </div>
        </div>
      </div>
    </Modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import apiClient from '../services/api'
import Pagination from '../components/common/Pagination.vue'
import Modal from '../components/common/Modal.vue'
import {
  Activity,
  Timer,
  Gauge,
  Search,
  RefreshCw,
  Eye,
  CheckCircle2,
  AlertTriangle,
  Copy,
  GitCommit,
} from 'lucide-vue-next'

const traces = ref([])
const total = ref(0)
const loading = ref(false)
const autoRefresh = ref(false)
const activeTrace = ref(null)
const activePayloadTab = ref('trip')

const stats = reactive({
  total_requests: 0,
  avg_lifetime_ms: 0,
  p95_lifetime_ms: 0,
  p99_lifetime_ms: 0,
  error_count: 0,
  error_rate: 0,
})

const filters = reactive({
  search: '',
  method: '',
  statusFilter: '',
  actorType: '',
  minLifetime: 0,
  page: 1,
  limit: 20,
})

const hasActiveFilters = computed(() => {
  return (
    filters.search !== '' ||
    filters.method !== '' ||
    filters.statusFilter !== '' ||
    filters.actorType !== '' ||
    filters.minLifetime > 0
  )
})

const avgLifetime = computed(() => {
  if (traces.value.length === 0) return 0
  const sum = traces.value.reduce((acc, t) => acc + (t.lifetime_ms || 0), 0)
  return sum / traces.value.length
})

const failedCount = computed(() => {
  return traces.value.filter(t => t.status_code >= 400).length
})

const activeTraceTrip = computed(() => {
  if (!activeTrace.value) return []
  if (activeTrace.value.spans_json) {
    try {
      const parsed = JSON.parse(activeTrace.value.spans_json)
      if (Array.isArray(parsed) && parsed.length > 0) return parsed
    } catch (e) {}
  }

  // Fallback: reconstruct the trip from metadata
  const totalMs = activeTrace.value.lifetime_ms || 1.0
  const path = activeTrace.value.path || ''
  let targetService = 'accounts-service (gRPC :50051)'
  if (path.includes('/plans') || path.includes('/subscriptions') || path.includes('/invoices') || path.includes('/manual-payments')) {
    targetService = 'subscriptions-service (gRPC :50052)'
  } else if (path.includes('/webhooks') || path.includes('/apps')) {
    targetService = 'webhook-runner (Kafka & gRPC :50053)'
  } else if (path.includes('/audit-logs')) {
    targetService = 'audit-service (gRPC :50054)'
  } else if (path.includes('/request-traces')) {
    targetService = 'request-tracker-service (gRPC :50055)'
  }

  const hops = [
    {
      name: 'REST Ingress & Gateway Routing',
      service: 'api-gateway',
      protocol: 'REST',
      type: 'INGRESS',
      offset_ms: 0,
      duration_ms: Math.max(0.2, totalMs * 0.15),
      status: 'OK',
      details: `${activeTrace.value.method} ${activeTrace.value.path}`
    },
    {
      name: `gRPC Invocations (${targetService})`,
      service: targetService,
      protocol: 'GRPC',
      type: 'DOWNSTREAM_RPC',
      offset_ms: totalMs * 0.15,
      duration_ms: Math.max(0.3, totalMs * 0.72),
      status: activeTrace.value.status_code < 400 ? 'OK' : 'ERROR',
      details: `Propagated Request ID: ${activeTrace.value.request_id || activeTrace.value.trace_id}`
    }
  ]

  if (path.includes('dispatch') || path.includes('send')) {
    hops.push({
      name: 'Kafka Message Dispatch (webhook-dispatches)',
      service: 'kafka-broker',
      protocol: 'KAFKA',
      type: 'EVENT_STREAM',
      offset_ms: totalMs * 0.65,
      duration_ms: Math.max(0.3, totalMs * 0.2),
      status: 'OK',
      details: `Partition Message with key: ${activeTrace.value.request_id || activeTrace.value.trace_id}`
    })
  }

  hops.push({
    name: 'REST Response Serialization & Egress',
    service: 'api-gateway',
    protocol: 'REST',
    type: 'EGRESS',
    offset_ms: totalMs * 0.88,
    duration_ms: Math.max(0.1, totalMs * 0.12),
    status: 'OK',
    details: `HTTP Response Status: ${activeTrace.value.status_code}`
  })

  return hops
})

function getProtocolBadgeClass(protocol) {
  switch (protocol) {
    case 'REST': return 'badge-protocol-rest'
    case 'GRPC': return 'badge-protocol-grpc'
    case 'KAFKA': return 'badge-protocol-kafka'
    default: return 'badge-secondary'
  }
}

function getProtocolBgClass(protocol) {
  switch (protocol) {
    case 'REST': return 'bg-protocol-rest'
    case 'GRPC': return 'bg-protocol-grpc'
    case 'KAFKA': return 'bg-protocol-kafka'
    default: return 'bg-primary'
  }
}

let debounceTimer = null
function debounceFetch() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    filters.page = 1
    fetchTraces()
  }, 300)
}

function onStatusFilterChange() {
  filters.page = 1
  fetchTraces()
}

function resetFilters() {
  filters.search = ''
  filters.method = ''
  filters.statusFilter = ''
  filters.actorType = ''
  filters.minLifetime = 0
  filters.page = 1
  fetchTraces()
}

async function fetchStats() {
  try {
    const res = await apiClient.get('/request-traces/stats')
    if (res.data && res.data.data) {
      Object.assign(stats, res.data.data)
    }
  } catch (err) {
    // Fall back to table-level computation
  }
}

async function fetchTraces() {
  loading.value = true
  try {
    const params = {
      page: filters.page,
      limit: filters.limit,
    }
    if (filters.search) params.search = filters.search
    if (filters.method) params.method = filters.method
    if (filters.statusFilter) params.status_code = parseInt(filters.statusFilter)
    if (filters.actorType) params.actor_type = filters.actorType
    if (filters.minLifetime > 0) params.min_lifetime_ms = filters.minLifetime

    const res = await apiClient.get('/request-traces', { params })
    if (res.data) {
      traces.value = res.data.data || []
      total.value = res.data.total || 0
    }
  } catch (err) {
    console.error('Failed to fetch request traces:', err)
  } finally {
    loading.value = false
  }
}

function fetchData() {
  fetchStats()
  fetchTraces()
}

let refreshInterval = null
function toggleAutoRefresh() {
  if (autoRefresh.value) {
    refreshInterval = setInterval(fetchData, 5000)
  } else {
    clearInterval(refreshInterval)
  }
}

function changePage(p) {
  filters.page = p
  fetchTraces()
}

function changePageSize(size) {
  filters.limit = size
  filters.page = 1
  fetchTraces()
}

function openInspectModal(trace) {
  activeTrace.value = trace
  activePayloadTab.value = 'trip'
}

function formatDateTime(val) {
  if (!val) return 'n/a'
  const d = new Date(val)
  return d.toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit', fractionalSecondDigits: 3 })
}

function formatJSON(str) {
  if (!str) return ''
  try {
    const parsed = JSON.parse(str)
    return JSON.stringify(parsed, null, 2)
  } catch (e) {
    return str
  }
}

function copyText(text) {
  if (!text) return
  navigator.clipboard.writeText(text)
}

function getStatusText(code) {
  const map = {
    200: 'OK',
    201: 'Created',
    204: 'No Content',
    400: 'Bad Request',
    401: 'Unauthorized',
    403: 'Forbidden',
    404: 'Not Found',
    409: 'Conflict',
    422: 'Unprocessable Entity',
    500: 'Internal Server Error',
    502: 'Bad Gateway',
    503: 'Service Unavailable',
  }
  return map[code] || ''
}

function getStatusBadgeClass(code) {
  if (code >= 200 && code < 300) return 'badge-success'
  if (code >= 300 && code < 400) return 'badge-info'
  if (code >= 400 && code < 500) return 'badge-warning'
  return 'badge-danger'
}

function getMethodBadgeClass(method) {
  switch (method) {
    case 'GET': return 'badge-method-get'
    case 'POST': return 'badge-method-post'
    case 'PUT': return 'badge-method-put'
    case 'PATCH': return 'badge-method-patch'
    case 'DELETE': return 'badge-method-delete'
    default: return 'badge-secondary'
  }
}

function getLatencyColor(ms) {
  if (!ms) return 'text-secondary'
  if (ms < 50) return 'text-success'
  if (ms < 150) return 'text-warning'
  return 'text-danger'
}

function getLatencyBgColor(ms) {
  if (!ms) return 'bg-success'
  if (ms < 50) return 'bg-success'
  if (ms < 150) return 'bg-warning'
  return 'bg-danger'
}

onMounted(() => {
  fetchData()
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
})
</script>

<style scoped>
.traces-view {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.view-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.view-title-group {
  display: flex;
  flex-direction: column;
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

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.auto-refresh-toggle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
  color: var(--text-secondary);
  cursor: pointer;
  user-select: none;
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
  width: 42px;
  height: 42px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.icon-blue { background: #eff6ff; color: #2563eb; }
.icon-cyan { background: #ecfeff; color: #0891b2; }
.icon-purple { background: #faf5ff; color: #9333ea; }
.icon-green { background: #ecfdf5; color: #059669; }
.icon-red { background: #fef2f2; color: #dc2626; }

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

.unit {
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-secondary);
}

/* Filter Card */
.filter-card {
  padding: 0.875rem 1.25rem;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
}

.filter-controls-row {
  display: flex;
  gap: 0.75rem;
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
  padding-inline-start: 2.25rem !important;
  height: 38px;
}

.select-filters-group {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
}

.select-field {
  height: 38px;
  min-width: 130px;
  font-size: 0.8125rem;
}

/* Zoho Table Styles */
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
.route-cell {
  display: flex;
  flex-direction: column;
  max-width: 320px;
  overflow: hidden;
}

.route-main {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #1e293b;
}

.route-sub {
  font-size: 0.7rem;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.lifetime-wrapper {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.lifetime-text {
  min-width: 55px;
  display: inline-block;
}

.latency-bar-track {
  width: 50px;
  height: 5px;
  background-color: #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
  flex-shrink: 0;
}

.latency-bar-fill {
  height: 100%;
  border-radius: 3px;
}

.bg-success { background-color: #10b981; }
.bg-warning { background-color: #f59e0b; }
.bg-danger { background-color: #ef4444; }

.actor-cell {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.actor-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background-color: #e2e8f0;
  color: var(--text-primary);
  font-size: 0.75rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.avatar-admin {
  background: #f3e8ff;
  color: #9333ea;
  border: 1px solid #d8b4fe;
}

.actor-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.actor-name {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.actor-email {
  font-size: 0.7rem;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.inspect-btn {
  color: var(--text-muted);
}
.inspect-btn:hover {
  color: var(--primary);
}

/* Method Badges */
.badge-method-get { background: #eff6ff; color: #2563eb; border: 1px solid #bfdbfe; }
.badge-method-post { background: #ecfdf5; color: #059669; border: 1px solid #a7f3d0; }
.badge-method-put { background: #fffbeb; color: #d97706; border: 1px solid #fde68a; }
.badge-method-patch { background: #faf5ff; color: #9333ea; border: 1px solid #e9d5ff; }
.badge-method-delete { background: #fef2f2; color: #dc2626; border: 1px solid #fecaca; }

.badge-service {
  background: #f1f5f9;
  color: #475569;
  border: 1px solid #e2e8f0;
}

/* Modal Styling */
.trace-header-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: #f8fafc;
  padding: 0.75rem 1rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
}

.meta-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.85rem;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  background-color: var(--bg-card);
  border: 1px solid var(--border-color);
  padding: 0.75rem 1rem;
  border-radius: var(--radius-md);
}

.meta-label {
  font-size: 0.7rem;
  text-transform: uppercase;
  color: var(--text-muted);
  font-weight: 600;
  letter-spacing: 0.04em;
}

.copy-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.copy-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 3px;
  border-radius: 4px;
}
.copy-btn:hover {
  color: var(--primary);
  background-color: #f1f5f9;
}

.param-box {
  background-color: #f8fafc;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 0.5rem 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.param-title {
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--text-secondary);
}

.alert-box {
  border-radius: var(--radius-md);
  padding: 0.75rem 1rem;
}

.alert-danger {
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  color: #dc2626;
}

.alert-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

/* Tabs & Code Panels */
.payload-tabs-section {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.payload-tab-headers {
  display: flex;
  background-color: #f8fafc;
  border-bottom: 1px solid var(--border-color);
}

.payload-tab-btn {
  padding: 0.65rem 1.25rem;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--text-secondary);
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: all var(--transition-fast);
}

.payload-tab-btn.active {
  color: var(--primary);
  border-bottom-color: var(--primary);
  background-color: #ffffff;
}

.tab-badge {
  font-size: 0.7rem;
  padding: 1px 6px;
  border-radius: 4px;
  background: #ecfdf5;
  color: #059669;
  border: 1px solid #a7f3d0;
}

.tab-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: #2563eb;
}

.payload-panel {
  background-color: #0f172a;
  padding: 0.875rem 1rem;
}

.payload-panel-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
  border-bottom: 1px solid #1e293b;
  padding-bottom: 0.35rem;
}

.toolbar-title {
  color: #94a3b8;
  font-weight: 600;
  letter-spacing: 0.05em;
}

.payload-panel-toolbar .btn {
  color: #94a3b8;
}
.payload-panel-toolbar .btn:hover {
  color: #f8fafc;
}

.json-code {
  color: #38bdf8;
  margin: 0;
  overflow-x: auto;
  max-height: 280px;
  line-height: 1.5;
}

/* Request Trip & Waterfall Timeline Styling */
.trip-panel {
  background-color: #f8fafc;
  border-top: 1px solid var(--border-color);
  padding: 1rem 1.25rem;
}

.trip-summary-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-md);
  margin-bottom: 1.25rem;
}

.trip-summary-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.trip-summary-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-weight: 600;
}

.trip-timeline {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  position: relative;
}

.timeline-hop-card {
  display: flex;
  align-items: stretch;
  gap: 1rem;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-md);
  padding: 0.875rem 1rem;
  transition: all var(--transition-fast);
}

.timeline-hop-card:hover {
  border-color: #cbd5e1;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.04);
}

.hop-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  width: 28px;
  flex-shrink: 0;
}

.hop-index {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #f1f5f9;
  border: 2px solid #94a3b8;
  color: #334155;
  font-size: 0.75rem;
  font-weight: 700;
  z-index: 2;
}

.hop-line {
  position: absolute;
  top: 24px;
  bottom: -20px;
  width: 2px;
  background-color: #e2e8f0;
  z-index: 1;
}

.hop-content-box {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 0;
}

.hop-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.hop-title-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.hop-name {
  color: var(--text-primary);
  font-weight: 700;
}

.hop-timing-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.badge-service {
  background-color: #f1f5f9;
  color: #475569;
  border: 1px solid #cbd5e1;
}

.badge-protocol-rest {
  background-color: #e0f2fe;
  color: #0369a1;
  border: 1px solid #7dd3fc;
}

.badge-protocol-grpc {
  background-color: #f3e8ff;
  color: #6b21a8;
  border: 1px solid #d8b4fe;
}

.badge-protocol-kafka {
  background-color: #ffedd5;
  color: #c2410c;
  border: 1px solid #fdba74;
}

.bg-protocol-rest {
  background-color: #0284c7;
}

.bg-protocol-grpc {
  background-color: #9333ea;
}

.bg-protocol-kafka {
  background-color: #ea580c;
}

.hop-details-row {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.hop-details {
  background-color: #f8fafc;
  border: 1px solid #f1f5f9;
  border-radius: 4px;
  padding: 0.25rem 0.5rem;
  color: #64748b;
  word-break: break-all;
}

.hop-progress-track {
  width: 100%;
  height: 6px;
  background-color: #f1f5f9;
  border-radius: 9999px;
  overflow: hidden;
  position: relative;
}

.hop-progress-bar {
  height: 100%;
  border-radius: 9999px;
  transition: width 0.3s ease;
}

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
