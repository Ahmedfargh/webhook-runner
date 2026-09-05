<template>
  <div class="apps-view animate-fade-in">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">Applications & Webhooks</h1>
        <p class="view-subtitle">Manage client applications, configure HMAC-SHA256 signing secrets, and trigger webhook dispatches.</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-secondary" @click="fetchApps" :disabled="loading">
          <RefreshCw :size="15" :class="{ 'spin-anim': loading }" /> Refresh
        </button>
        <button class="btn btn-primary" @click="openCreateModal">
          <Plus :size="15" /> Create Application
        </button>
      </div>
    </div>

    <!-- Stats Bar -->
    <div class="stats-overview-grid">
      <div class="stat-card">
        <div class="stat-icon-wrapper blue">
          <Layers :size="22" />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ apps.length }}</div>
          <div class="stat-label">Registered Applications</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrapper green">
          <ShieldCheck :size="22" />
        </div>
        <div class="stat-content">
          <div class="stat-value">HMAC-SHA256</div>
          <div class="stat-label">Cryptographic Signatures</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon-wrapper purple">
          <Zap :size="22" />
        </div>
        <div class="stat-content">
          <div class="stat-value">Instant Dispatch</div>
          <div class="stat-label">High Throughput Runner</div>
        </div>
      </div>
    </div>

    <!-- Search & Filters -->
    <div class="filter-bar card mb-4">
      <div class="search-input-wrapper">
        <Search :size="16" class="search-icon" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Filter applications by name or App ID..."
          class="form-input search-field"
        />
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading && apps.length === 0" class="text-center py-12">
      <RefreshCw :size="28" class="spin-anim text-primary mx-auto" />
      <p class="mt-3 text-muted">Loading applications...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="filteredApps.length === 0" class="empty-state-card card">
      <Layers :size="48" class="empty-icon text-muted" />
      <h3 class="empty-title">No Applications Found</h3>
      <p class="empty-subtitle">Get started by creating your first application to receive and dispatch webhooks.</p>
      <button class="btn btn-primary mt-4" @click="openCreateModal">
        <Plus :size="16" /> Create Application
      </button>
    </div>

    <!-- App Cards Grid -->
    <div v-else class="apps-grid">
      <div v-for="app in filteredApps" :key="app.id" class="card app-card">
        <div class="app-card-header">
          <div class="app-title-area">
            <div class="app-avatar">
              {{ getAppInitials(app.name) }}
            </div>
            <div>
              <h3 class="app-name">{{ app.name }}</h3>
              <span class="badge badge-active">Active</span>
            </div>
          </div>
          <div class="app-card-actions">
            <button class="btn-icon" @click="openTestWebhookModal(app)" title="Send Test Webhook">
              <Send :size="15" />
            </button>
            <button class="btn-icon" @click="openEditModal(app)" title="Edit App">
              <Edit3 :size="15" />
            </button>
            <button class="btn-icon text-danger" @click="confirmDelete(app)" title="Delete App">
              <Trash2 :size="15" />
            </button>
          </div>
        </div>

        <div class="app-card-body">
          <!-- App ID -->
          <div class="credential-row">
            <span class="cred-label">App ID</span>
            <div class="cred-value-box">
              <code>{{ app.app_id || app.appId }}</code>
              <button class="btn-copy" @click="copyToClipboard(app.app_id || app.appId, 'App ID')">
                <Copy :size="13" />
              </button>
            </div>
          </div>

          <!-- App Secret -->
          <div class="credential-row">
            <span class="cred-label">App Secret</span>
            <div class="cred-value-box">
              <code>{{ isRevealed(app.id, 'app_secret') ? (app.app_secret || app.appSecret) : '••••••••••••••••••••••••••••••••' }}</code>
              <button class="btn-copy" @click="toggleReveal(app.id, 'app_secret')">
                <EyeOff v-if="isRevealed(app.id, 'app_secret')" :size="13" />
                <Eye v-else :size="13" />
              </button>
              <button class="btn-copy" @click="copyToClipboard(app.app_secret || app.appSecret, 'App Secret')">
                <Copy :size="13" />
              </button>
            </div>
          </div>

          <!-- Webhook Secret -->
          <div class="credential-row">
            <span class="cred-label">HMAC Secret</span>
            <div class="cred-value-box">
              <code>{{ isRevealed(app.id, 'webhook_secret') ? (app.webhook_secret || app.webhookSecret) : '••••••••••••••••••••••••••••••••' }}</code>
              <button class="btn-copy" @click="toggleReveal(app.id, 'webhook_secret')">
                <EyeOff v-if="isRevealed(app.id, 'webhook_secret')" :size="13" />
                <Eye v-else :size="13" />
              </button>
              <button class="btn-copy" @click="copyToClipboard(app.webhook_secret || app.webhookSecret, 'Webhook Secret')">
                <Copy :size="13" />
              </button>
            </div>
          </div>

          <!-- Webhook URL -->
          <div class="credential-row">
            <span class="cred-label">Endpoint URL</span>
            <div class="cred-url-box">
              <Globe :size="14" class="text-muted" />
              <span class="url-text" :title="app.webhook_url || app.webhookUrl">{{ app.webhook_url || app.webhookUrl || 'Not configured' }}</span>
            </div>
          </div>
        </div>

        <div class="app-card-footer">
          <button class="btn btn-sm btn-outline" @click="rotateSecrets(app)">
            <Key :size="13" /> Rotate Secrets
          </button>
          <button class="btn btn-sm btn-primary" @click="openTestWebhookModal(app)">
            <Send :size="13" /> Test Webhook
          </button>
        </div>
      </div>
    </div>

    <!-- Create / Edit App Modal -->
    <div v-if="showModal" class="modal-backdrop" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <h3 class="modal-title">{{ isEditing ? 'Edit Application' : 'Create New Application' }}</h3>
          <button class="modal-close" @click="closeModal">
            <X :size="18" />
          </button>
        </div>

        <form @submit.prevent="handleSubmitApp" class="modal-body">
          <div class="form-group">
            <label class="form-label">Application Name <span class="text-danger">*</span></label>
            <input
              v-model="form.name"
              type="text"
              class="form-input"
              placeholder="e.g. Stripe Sync Service, E-Commerce Store"
              required
            />
          </div>

          <div class="form-group">
            <label class="form-label">Webhook Destination URL</label>
            <input
              v-model="form.webhook_url"
              type="url"
              class="form-input"
              placeholder="https://api.yourdomain.com/webhooks"
            />
            <span class="form-hint">The HTTP/HTTPS endpoint where webhook events will be POSTed.</span>
          </div>

          <div v-if="!isEditing" class="form-group">
            <label class="form-label">Custom HMAC Secret (Optional)</label>
            <input
              v-model="form.webhook_secret"
              type="text"
              class="form-input"
              placeholder="Leave empty to auto-generate secure secret"
            />
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeModal">Cancel</button>
            <button type="submit" class="btn btn-primary" :disabled="saving">
              <RefreshCw v-if="saving" :size="15" class="spin-anim" />
              {{ isEditing ? 'Update Application' : 'Create Application' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Test Webhook Modal -->
    <div v-if="showTestModal" class="modal-backdrop" @click.self="closeTestModal">
      <div class="modal-card modal-large">
        <div class="modal-header">
          <div class="modal-title-with-app">
            <Send :size="18" class="text-primary" />
            <h3 class="modal-title">Dispatch Webhook &bull; {{ activeApp?.name }}</h3>
          </div>
          <button class="modal-close" @click="closeTestModal">
            <X :size="18" />
          </button>
        </div>

        <form @submit.prevent="handleSendWebhook" class="modal-body">
          <div class="grid-2-col">
            <div class="form-group">
              <label class="form-label">Event Name <span class="text-danger">*</span></label>
              <input
                v-model="testForm.event_name"
                type="text"
                class="form-input"
                placeholder="e.g. payment.succeeded, user.created"
                required
              />
            </div>
            <div class="form-group">
              <label class="form-label">Override Target URL (Optional)</label>
              <input
                v-model="testForm.target_url_override"
                type="url"
                class="form-input"
                :placeholder="activeApp?.webhook_url || activeApp?.webhookUrl || 'https://webhook.site/...'"
              />
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">Payload JSON Data <span class="text-danger">*</span></label>
            <textarea
              v-model="testForm.payload_str"
              rows="6"
              class="form-input code-font"
              placeholder='{\n  "event": "user.created",\n  "data": {\n    "id": "usr_123",\n    "email": "user@example.com"\n  }\n}'
              required
            ></textarea>
          </div>

          <!-- Dispatch Result Preview Box -->
          <div v-if="testResult" class="test-result-box" :class="testResult.success ? 'result-success' : 'result-error'">
            <div class="result-header">
              <div class="result-title">
                <CheckCircle2 v-if="testResult.success" :size="16" class="text-success" />
                <AlertCircle v-else :size="16" class="text-danger" />
                <strong>Status: {{ testResult.data?.status || 'UNKNOWN' }}</strong>
                <span v-if="testResult.data?.response_status_code" class="badge ml-2">
                  HTTP {{ testResult.data?.response_status_code }}
                </span>
                <span v-if="testResult.data?.latency_ms" class="badge ml-1">
                  {{ testResult.data?.latency_ms }} ms
                </span>
              </div>
            </div>
            <div v-if="testResult.data?.signature" class="result-sig">
              <span class="text-muted">HMAC-SHA256 Signature:</span>
              <code>{{ testResult.data?.signature }}</code>
            </div>
            <div v-if="testResult.data?.response_body" class="result-body-preview">
              <span class="text-muted">Response Body:</span>
              <pre>{{ testResult.data?.response_body }}</pre>
            </div>
          </div>

          <div class="modal-footer">
            <router-link to="/webhooks/logs" class="btn btn-outline mr-auto">
              <FileText :size="14" /> View All Logs
            </router-link>
            <button type="button" class="btn btn-secondary" @click="closeTestModal">Close</button>
            <button type="submit" class="btn btn-primary" :disabled="sending">
              <RefreshCw v-if="sending" :size="15" class="spin-anim" />
              <Send v-else :size="15" /> Send Webhook Now
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { appService } from '../services/appService'
import { webhookService } from '../services/webhookService'
import { useToastStore } from '../stores/toast'
import {
  Layers,
  Plus,
  RefreshCw,
  Search,
  Copy,
  Eye,
  EyeOff,
  Globe,
  Key,
  Send,
  Edit3,
  Trash2,
  X,
  Zap,
  ShieldCheck,
  CheckCircle2,
  AlertCircle,
  FileText,
} from 'lucide-vue-next'

const toastStore = useToastStore()

const loading = ref(false)
const saving = ref(false)
const sending = ref(false)
const apps = ref([])
const searchQuery = ref('')

const revealedSecrets = ref({})

// Modal States
const showModal = ref(false)
const isEditing = ref(false)
const form = ref({
  id: '',
  name: '',
  webhook_url: '',
  webhook_secret: '',
})

// Test Webhook Modal States
const showTestModal = ref(false)
const activeApp = ref(null)
const testResult = ref(null)
const testForm = ref({
  app_id: '',
  event_name: 'order.created',
  target_url_override: '',
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

const filteredApps = computed(() => {
  if (!searchQuery.value) return apps.value
  const q = searchQuery.value.toLowerCase()
  return apps.value.filter(
    (a) =>
      a.name?.toLowerCase().includes(q) ||
      (a.app_id || a.appId || '').toLowerCase().includes(q) ||
      (a.webhook_url || a.webhookUrl || '').toLowerCase().includes(q)
  )
})

function getAppInitials(name) {
  if (!name) return 'AP'
  return name
    .split(' ')
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

function isRevealed(appId, field) {
  return !!revealedSecrets.value[`${appId}_${field}`]
}

function toggleReveal(appId, field) {
  const key = `${appId}_${field}`
  revealedSecrets.value[key] = !revealedSecrets.value[key]
}

async function copyToClipboard(text, label) {
  if (!text) return
  try {
    await navigator.clipboard.writeText(text)
    toastStore.success(`${label} copied to clipboard!`)
  } catch (err) {
    toastStore.error('Failed to copy to clipboard.')
  }
}

async function fetchApps() {
  loading.value = true
  try {
    const res = await appService.listApps()
    apps.value = res.data || []
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to load applications.')
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  isEditing.value = false
  form.value = {
    id: '',
    name: '',
    webhook_url: '',
    webhook_secret: '',
  }
  showModal.value = true
}

function openEditModal(app) {
  isEditing.value = true
  form.value = {
    id: app.id,
    name: app.name,
    webhook_url: app.webhook_url || app.webhookUrl || '',
    webhook_secret: '',
  }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

async function handleSubmitApp() {
  saving.value = true
  try {
    if (isEditing.value) {
      await appService.updateApp(form.value.id, {
        name: form.value.name,
        webhook_url: form.value.webhook_url,
      })
      toastStore.success('Application updated successfully!')
    } else {
      await appService.createApp({
        name: form.value.name,
        webhook_url: form.value.webhook_url,
        webhook_secret: form.value.webhook_secret,
      })
      toastStore.success('Application created successfully!')
    }
    closeModal()
    await fetchApps()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to save application.')
  } finally {
    saving.value = false
  }
}

async function confirmDelete(app) {
  if (confirm(`Are you sure you want to delete "${app.name}"? This action cannot be undone.`)) {
    try {
      await appService.deleteApp(app.id)
      toastStore.success('Application deleted.')
      await fetchApps()
    } catch (err) {
      toastStore.error(err.response?.data?.error || 'Failed to delete application.')
    }
  }
}

async function rotateSecrets(app) {
  if (confirm(`Rotate secrets for "${app.name}"? All existing webhook integrations must be updated.`)) {
    try {
      await appService.rotateSecrets(app.id, { rotate_app_secret: true, rotate_webhook_secret: true })
      toastStore.success('Secrets regenerated successfully!')
      await fetchApps()
    } catch (err) {
      toastStore.error(err.response?.data?.error || 'Failed to rotate secrets.')
    }
  }
}

function openTestWebhookModal(app) {
  activeApp.value = app
  testResult.value = null
  testForm.value.app_id = app.app_id || app.appId || app.id
  testForm.value.target_url_override = app.webhook_url || app.webhookUrl || ''
  showTestModal.value = true
}

function closeTestModal() {
  showTestModal.value = false
  activeApp.value = null
  testResult.value = null
}

async function handleSendWebhook() {
  sending.value = true
  testResult.value = null
  try {
    let parsedPayload = {}
    try {
      parsedPayload = JSON.parse(testForm.value.payload_str)
    } catch (e) {
      toastStore.error('Invalid JSON payload formatting.')
      sending.value = false
      return
    }

    const res = await webhookService.sendWebhook({
      app_id: testForm.value.app_id,
      event_name: testForm.value.event_name,
      target_url_override: testForm.value.target_url_override,
      payload: parsedPayload,
    })

    testResult.value = res
    if (res.success) {
      toastStore.success(`Webhook delivered! Status: ${res.data?.status}`)
    } else {
      toastStore.warning(`Webhook execution completed with status: ${res.data?.status}`)
    }
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to dispatch webhook.')
  } finally {
    sending.value = false
  }
}

onMounted(() => {
  fetchApps()
})
</script>

<style scoped>
.apps-view {
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

.header-actions {
  display: flex;
  gap: 0.75rem;
}

.stats-overview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.stat-card {
  background: var(--zoho-card-bg);
  border: 1px solid var(--zoho-border-color);
  border-radius: var(--radius-lg);
  padding: 1.25rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.stat-icon-wrapper {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-icon-wrapper.blue {
  background: #eff6ff;
  color: #2563eb;
}
.stat-icon-wrapper.green {
  background: #f0fdf4;
  color: #16a34a;
}
.stat-icon-wrapper.purple {
  background: #faf5ff;
  color: #9333ea;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--zoho-text-main);
}
.stat-label {
  font-size: 0.75rem;
  color: var(--zoho-text-muted);
}

.filter-bar {
  padding: 0.75rem 1rem;
  display: flex;
  align-items: center;
}

.search-input-wrapper {
  position: relative;
  width: 100%;
  max-width: 400px;
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

.apps-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 1.25rem;
}

.app-card {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  border: 1px solid var(--zoho-border-color);
  border-radius: var(--radius-lg);
  background: var(--zoho-card-bg);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.04);
  transition: transform 0.2s, box-shadow 0.2s;
}

.app-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.08);
}

.app-card-header {
  padding: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--zoho-border-color);
}

.app-title-area {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.app-avatar {
  width: 38px;
  height: 38px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: #ffffff;
  font-weight: 700;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.app-name {
  font-size: 1rem;
  font-weight: 600;
  color: var(--zoho-text-main);
  margin-bottom: 0.125rem;
}

.app-card-actions {
  display: flex;
  gap: 0.35rem;
}

.btn-icon {
  background: transparent;
  border: 1px solid var(--zoho-border-color);
  border-radius: var(--radius-sm);
  padding: 0.4rem;
  cursor: pointer;
  color: var(--zoho-text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
}
.btn-icon:hover {
  background: var(--zoho-bg-light);
  color: var(--zoho-text-main);
}
.btn-icon.text-danger:hover {
  background: #fef2f2;
  color: #dc2626;
  border-color: #fecaca;
}

.app-card-body {
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
  flex: 1;
}

.credential-row {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.cred-label {
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--zoho-text-muted);
}

.cred-value-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-sm);
  padding: 0.35rem 0.6rem;
}

.cred-value-box code {
  font-family: monospace;
  font-size: 0.75rem;
  color: #0f172a;
  word-break: break-all;
}

.cred-url-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-sm);
  padding: 0.35rem 0.6rem;
  font-size: 0.75rem;
}

.url-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #334155;
}

.btn-copy {
  background: transparent;
  border: none;
  color: var(--zoho-text-muted);
  cursor: pointer;
  padding: 0.2rem;
  display: flex;
  align-items: center;
}
.btn-copy:hover {
  color: #2563eb;
}

.app-card-footer {
  padding: 0.875rem 1.25rem;
  background: #fafafa;
  border-top: 1px solid var(--zoho-border-color);
  border-radius: 0 0 var(--radius-lg) var(--radius-lg);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.empty-state-card {
  text-align: center;
  padding: 4rem 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* Modal Styles */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  padding: 1rem;
}

.modal-card {
  background: #ffffff;
  border-radius: var(--radius-xl);
  width: 100%;
  max-width: 500px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.modal-card.modal-large {
  max-width: 680px;
}

.modal-header {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--zoho-border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.modal-title-with-app {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--zoho-text-main);
}

.modal-close {
  background: transparent;
  border: none;
  cursor: pointer;
  color: var(--zoho-text-muted);
}

.modal-body {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1rem;
}

.grid-2-col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.code-font {
  font-family: monospace;
  font-size: 0.8rem;
}

.test-result-box {
  border-radius: var(--radius-md);
  padding: 1rem;
  margin-top: 0.5rem;
}

.result-success {
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
}

.result-error {
  background: #fef2f2;
  border: 1px solid #fecaca;
}

.result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.result-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.result-sig {
  font-size: 0.75rem;
  margin-bottom: 0.5rem;
  display: flex;
  gap: 0.5rem;
}
.result-sig code {
  font-family: monospace;
  word-break: break-all;
}

.result-body-preview pre {
  margin-top: 0.25rem;
  padding: 0.5rem;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  max-height: 120px;
  overflow-y: auto;
}
</style>
