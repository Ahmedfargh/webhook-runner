<template>
  <div class="apps-view animate-fade-in">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('apps.title') }}</h1>
        <p class="view-subtitle">{{ t('apps.subtitle') }}</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-secondary btn-sm" @click="fetchApps" :disabled="loading">
          <RefreshCw :size="14" :class="{ 'spin-anim': loading }" /> {{ t('common.refresh') }}
        </button>
        <button class="btn btn-primary btn-sm" @click="openCreateDrawer">
          <Plus :size="14" /> {{ t('apps.createApp') }}
        </button>
      </div>
    </div>

    <!-- Stats Bar -->
    <div class="stats-overview-grid mb-6">
      <div class="card stat-card">
        <div class="stat-icon-wrapper blue">
          <Layers :size="22" />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ apps.length }}</div>
          <div class="stat-label">{{ t('apps.registeredApps') }}</div>
        </div>
      </div>
      <div class="card stat-card">
        <div class="stat-icon-wrapper green">
          <ShieldCheck :size="22" />
        </div>
        <div class="stat-content">
          <div class="stat-value">HMAC-SHA256</div>
          <div class="stat-label">{{ t('apps.cryptoSignatures') }}</div>
        </div>
      </div>
      <div class="card stat-card">
        <div class="stat-icon-wrapper purple">
          <Zap :size="22" />
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ t('apps.instantDispatch') }}</div>
          <div class="stat-label">HTTP / gRPC Runner</div>
        </div>
      </div>
    </div>

    <!-- Search & Filter Bar -->
    <div class="card filter-card mb-6">
      <div class="filter-row">
        <div class="search-input-wrapper">
          <Search :size="16" class="search-icon" />
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('apps.filterPlaceholder')"
            class="form-control search-field"
          />
        </div>

        <!-- Admin: Filter Apps by User -->
        <div v-if="authStore.isAdmin && usersList.length > 0" class="user-filter-box">
          <select v-model="selectedUserFilter" class="form-control user-select" @change="fetchApps">
            <option value="">{{ t('apps.allUsers') }}</option>
            <option v-for="u in usersList" :key="u.id" :value="u.id">
              {{ u.name }} ({{ u.email }})
            </option>
          </select>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading && apps.length === 0" class="text-center py-12">
      <RefreshCw :size="28" class="spin-anim text-primary mx-auto" />
      <p class="mt-3 text-muted">{{ t('common.loading') }}</p>
    </div>

    <!-- Empty State -->
    <EmptyState
      v-else-if="filteredApps.length === 0"
      :title="t('apps.noAppsFound')"
      :description="t('apps.noAppsSubtitle')"
      :icon="Layers"
    >
      <template #action>
        <button class="btn btn-primary btn-sm" @click="openCreateDrawer">
          <Plus :size="14" /> {{ t('apps.createApp') }}
        </button>
      </template>
    </EmptyState>

    <!-- Apps Grid -->
    <div v-else class="apps-grid">
      <div v-for="app in filteredApps" :key="app.id" class="card app-card">
        <div class="app-card-header">
          <div class="app-title-area">
            <div class="app-avatar-circle">
              {{ getAppInitials(app.name) }}
            </div>
            <div class="app-name-meta">
              <h3 class="app-name">{{ app.name }}</h3>
              <div class="badges-row">
                <span class="badge badge-success">{{ t('common.active') }}</span>
                <!-- Admin: Owner Pill -->
                <span v-if="authStore.isAdmin && (app.user_id || app.userId)" class="badge badge-secondary owner-badge" :title="getUserName(app.user_id || app.userId)">
                  <UserIcon :size="11" />
                  <span>{{ getUserShortName(app.user_id || app.userId) }}</span>
                </span>
              </div>
            </div>
          </div>
          <div class="action-buttons-group">
            <button class="action-icon-btn" :title="t('apps.testWebhook')" @click="openTestModal(app)">
              <Send :size="14" />
            </button>
            <button class="action-icon-btn" :title="t('common.edit')" @click="openEditDrawer(app)">
              <Edit3 :size="14" />
            </button>
            <button class="action-icon-btn text-danger" :title="t('common.delete')" @click="confirmDelete(app)">
              <Trash2 :size="14" />
            </button>
          </div>
        </div>

        <div class="app-card-body">
          <!-- App ID -->
          <div class="credential-field">
            <label class="cred-label">{{ t('apps.appId') }}</label>
            <div class="cred-box" dir="ltr">
              <code class="cred-code">{{ app.app_id || app.appId }}</code>
              <button class="copy-btn" @click="copyToClipboard(app.app_id || app.appId, t('apps.appId'))" :title="t('apps.appId')">
                <Copy :size="13" />
              </button>
            </div>
          </div>

          <!-- App Secret -->
          <div class="credential-field">
            <label class="cred-label">{{ t('apps.appSecret') }}</label>
            <div class="cred-box" dir="ltr">
              <code class="cred-code">{{ isRevealed(app.id, 'app_secret') ? (app.app_secret || app.appSecret) : '••••••••••••••••••••••••••••••••' }}</code>
              <div class="cred-actions">
                <button class="copy-btn" @click="toggleReveal(app.id, 'app_secret')">
                  <EyeOff v-if="isRevealed(app.id, 'app_secret')" :size="13" />
                  <Eye v-else :size="13" />
                </button>
                <button class="copy-btn" @click="copyToClipboard(app.app_secret || app.appSecret, t('apps.appSecret'))">
                  <Copy :size="13" />
                </button>
              </div>
            </div>
          </div>

          <!-- Webhook HMAC Secret -->
          <div class="credential-field">
            <label class="cred-label">{{ t('apps.hmacSecret') }}</label>
            <div class="cred-box" dir="ltr">
              <code class="cred-code">{{ isRevealed(app.id, 'webhook_secret') ? (app.webhook_secret || app.webhookSecret) : '••••••••••••••••••••••••••••••••' }}</code>
              <div class="cred-actions">
                <button class="copy-btn" @click="toggleReveal(app.id, 'webhook_secret')">
                  <EyeOff v-if="isRevealed(app.id, 'webhook_secret')" :size="13" />
                  <Eye v-else :size="13" />
                </button>
                <button class="copy-btn" @click="copyToClipboard(app.webhook_secret || app.webhookSecret, t('apps.hmacSecret'))">
                  <Copy :size="13" />
                </button>
              </div>
            </div>
          </div>

          <!-- Webhook URL -->
          <div class="credential-field">
            <label class="cred-label">{{ t('apps.endpointUrl') }}</label>
            <div class="url-box" dir="ltr">
              <Globe :size="14" class="text-muted flex-shrink-0" />
              <span class="url-text" :title="app.webhook_url || app.webhookUrl">{{ app.webhook_url || app.webhookUrl || 'Not configured' }}</span>
            </div>
          </div>
        </div>

        <div class="app-card-footer">
          <button class="btn btn-sm btn-secondary" @click="rotateSecrets(app)">
            <Key :size="13" /> {{ t('apps.rotateSecrets') }}
          </button>
          <button class="btn btn-sm btn-primary" @click="openTestModal(app)">
            <Send :size="13" /> {{ t('apps.testWebhook') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Create / Edit Slide-Over Drawer -->
    <Drawer
      :isOpen="isDrawerOpen"
      :title="isEditing ? t('apps.editApp') : t('apps.createApp')"
      :subtitle="isEditing ? form.name : t('apps.subtitle')"
      @close="closeDrawer"
    >
      <form @submit.prevent="saveApp" class="drawer-form">
        <!-- Admin-Only: Target User Account Selection -->
        <div v-if="authStore.isAdmin && !isEditing" class="form-group">
          <label class="form-label">{{ t('apps.selectUser') }} *</label>
          <select v-model="form.user_id" required class="form-control">
            <option value="" disabled>{{ t('apps.selectUserPlaceholder') }}</option>
            <option v-for="u in usersList" :key="u.id" :value="u.id">
              {{ u.name }} ({{ u.email }})
            </option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('apps.appName') }} *</label>
          <input
            v-model="form.name"
            type="text"
            required
            class="form-control"
            :placeholder="t('apps.appNamePlaceholder')"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('apps.webhookUrl') }}</label>
          <input
            v-model="form.webhook_url"
            type="url"
            class="form-control"
            dir="ltr"
            placeholder="https://api.yourdomain.com/webhooks"
          />
          <span class="form-hint">{{ t('apps.webhookUrlHint') }}</span>
        </div>

        <div v-if="!isEditing" class="form-group">
          <label class="form-label">{{ t('apps.customHmacSecret') }}</label>
          <input
            v-model="form.webhook_secret"
            type="text"
            class="form-control"
            dir="ltr"
            placeholder="whsec_..."
          />
          <span class="form-hint">{{ t('apps.customHmacSecretHint') }}</span>
        </div>
      </form>

      <template #footer>
        <button class="btn btn-secondary" @click="closeDrawer" :disabled="saving">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveApp" :disabled="saving">
          <RefreshCw v-if="saving" :size="14" class="spin-anim" />
          {{ isEditing ? t('common.save') : t('common.create') }}
        </button>
      </template>
    </Drawer>

    <!-- Test Dispatch Modal -->
    <Modal
      :isOpen="isTestModalOpen"
      :title="`${t('apps.dispatchTitle')} - ${activeApp?.name || ''}`"
      width="680px"
      @close="closeTestModal"
    >
      <form @submit.prevent="executeTestDispatch" class="test-form">
        <div class="form-row-grid">
          <div class="form-group">
            <label class="form-label">{{ t('apps.eventName') }} *</label>
            <input
              v-model="testForm.event_name"
              type="text"
              required
              class="form-control"
              placeholder="e.g. order.created"
              dir="ltr"
            />
          </div>
          <div class="form-group">
            <label class="form-label">{{ t('apps.overrideUrl') }}</label>
            <input
              v-model="testForm.target_url_override"
              type="url"
              class="form-control"
              :placeholder="activeApp?.webhook_url || activeApp?.webhookUrl || 'https://webhook.site/...'"
              dir="ltr"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('apps.payloadJson') }} *</label>
          <textarea
            v-model="testForm.payload_str"
            rows="6"
            required
            class="form-control code-editor"
            dir="ltr"
          ></textarea>
        </div>

        <!-- Result Box -->
        <div v-if="testResult" class="test-result-box" :class="testResult.success ? 'result-success' : 'result-error'">
          <div class="result-header">
            <div class="result-title">
              <CheckCircle2 v-if="testResult.success" :size="16" class="text-success" />
              <AlertCircle v-else :size="16" class="text-danger" />
              <strong>Status: {{ testResult.data?.status || 'UNKNOWN' }}</strong>
              <span v-if="testResult.data?.response_status_code" class="badge ml-2" :class="testResult.data?.response_status_code < 400 ? 'badge-success' : 'badge-danger'">
                HTTP {{ testResult.data?.response_status_code }}
              </span>
              <span v-if="testResult.data?.latency_ms" class="badge badge-secondary ml-1">
                {{ testResult.data?.latency_ms }} ms
              </span>
            </div>
          </div>
          <div v-if="testResult.data?.signature" class="result-sig" dir="ltr">
            <span class="text-muted">HMAC-SHA256:</span>
            <code>{{ testResult.data?.signature }}</code>
          </div>
          <div v-if="testResult.data?.response_body" class="result-body-preview" dir="ltr">
            <span class="text-muted">Response:</span>
            <pre>{{ testResult.data?.response_body }}</pre>
          </div>
        </div>
      </form>

      <template #footer>
        <router-link to="/webhooks/logs" class="btn btn-outline btn-sm mr-auto" @click="closeTestModal">
          <FileText :size="14" /> {{ t('apps.viewAllLogs') }}
        </router-link>
        <button class="btn btn-secondary btn-sm" @click="closeTestModal">{{ t('common.close') }}</button>
        <button class="btn btn-primary btn-sm" @click="executeTestDispatch" :disabled="sending">
          <RefreshCw v-if="sending" :size="14" class="spin-anim" />
          <Send v-else :size="14" /> {{ t('apps.sendNow') }}
        </button>
      </template>
    </Modal>

    <!-- Delete Confirmation Modal -->
    <Modal
      :isOpen="isDeleteModalOpen"
      :title="t('common.confirmDelete')"
      @close="isDeleteModalOpen = false"
    >
      <p class="modal-text">
        {{ t('apps.deletePrompt', { name: appToDelete?.name || '' }) }}
      </p>

      <template #footer>
        <button class="btn btn-secondary" @click="isDeleteModalOpen = false" :disabled="deleting">{{ t('common.cancel') }}</button>
        <button class="btn btn-danger" @click="executeDelete" :disabled="deleting">
          <RefreshCw v-if="deleting" :size="14" class="spin-anim" />
          {{ t('common.delete') }}
        </button>
      </template>
    </Modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '../locales'
import { useAuthStore } from '../stores/auth'
import { appService } from '../services/appService'
import { webhookService } from '../services/webhookService'
import { userService } from '../services/userService'
import { useToastStore } from '../stores/toast'
import Drawer from '../components/common/Drawer.vue'
import Modal from '../components/common/Modal.vue'
import EmptyState from '../components/common/EmptyState.vue'
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
  Zap,
  ShieldCheck,
  CheckCircle2,
  AlertCircle,
  FileText,
  User as UserIcon,
} from 'lucide-vue-next'

const { t } = useI18n()
const authStore = useAuthStore()
const toastStore = useToastStore()

const loading = ref(false)
const saving = ref(false)
const sending = ref(false)
const deleting = ref(false)
const apps = ref([])
const usersList = ref([])
const selectedUserFilter = ref('')
const searchQuery = ref('')
const revealedSecrets = ref({})

// Drawer States
const isDrawerOpen = ref(false)
const isEditing = ref(false)
const form = ref({
  id: '',
  user_id: '',
  name: '',
  webhook_url: '',
  webhook_secret: '',
})

// Delete Modal States
const isDeleteModalOpen = ref(false)
const appToDelete = ref(null)

// Test Modal States
const isTestModalOpen = ref(false)
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

function getUserName(userId) {
  if (!userId) return ''
  const u = usersList.value.find((user) => user.id === userId)
  return u ? `${u.name} (${u.email})` : ''
}

function getUserShortName(userId) {
  if (!userId) return ''
  const u = usersList.value.find((user) => user.id === userId)
  return u ? u.name : `${userId.slice(0, 8)}...`
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
    toastStore.success(`${label} copied!`)
  } catch (err) {
    toastStore.error('Failed to copy to clipboard.')
  }
}

async function fetchUsers() {
  if (authStore.isAdmin) {
    try {
      const res = await userService.getUsers({ limit: 100 })
      usersList.value = res.data || res.users || []
    } catch (e) {
      // silent
    }
  }
}

async function fetchApps() {
  loading.value = true
  try {
    const params = {}
    if (authStore.isAdmin && selectedUserFilter.value) {
      params.user_id = selectedUserFilter.value
    }
    const res = await appService.listApps(params)
    apps.value = res.data || []
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to load applications.')
  } finally {
    loading.value = false
  }
}

function openCreateDrawer() {
  isEditing.value = false
  form.value = {
    id: '',
    user_id: usersList.value[0]?.id || '',
    name: '',
    webhook_url: '',
    webhook_secret: '',
  }
  isDrawerOpen.value = true
}

function openEditDrawer(app) {
  isEditing.value = true
  form.value = {
    id: app.id,
    user_id: app.user_id || app.userId || '',
    name: app.name,
    webhook_url: app.webhook_url || app.webhookUrl || '',
    webhook_secret: '',
  }
  isDrawerOpen.value = true
}

function closeDrawer() {
  isDrawerOpen.value = false
}

async function saveApp() {
  if (!form.value.name.trim()) return
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
        user_id: form.value.user_id,
        name: form.value.name,
        webhook_url: form.value.webhook_url,
        webhook_secret: form.value.webhook_secret,
      })
      toastStore.success('Application created successfully!')
    }
    closeDrawer()
    await fetchApps()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to save application.')
  } finally {
    saving.value = false
  }
}

function confirmDelete(app) {
  appToDelete.value = app
  isDeleteModalOpen.value = true
}

async function executeDelete() {
  if (!appToDelete.value) return
  deleting.value = true
  try {
    await appService.deleteApp(appToDelete.value.id)
    toastStore.success('Application deleted.')
    isDeleteModalOpen.value = false
    await fetchApps()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to delete application.')
  } finally {
    deleting.value = false
  }
}

async function rotateSecrets(app) {
  if (confirm(t('apps.rotateConfirm', { name: app.name }))) {
    try {
      await appService.rotateSecrets(app.id, { rotate_app_secret: true, rotate_webhook_secret: true })
      toastStore.success('Secrets rotated successfully!')
      await fetchApps()
    } catch (err) {
      toastStore.error(err.response?.data?.error || 'Failed to rotate secrets.')
    }
  }
}

function openTestModal(app) {
  activeApp.value = app
  testResult.value = null
  testForm.value.app_id = app.app_id || app.appId || app.id
  testForm.value.target_url_override = app.webhook_url || app.webhookUrl || ''
  isTestModalOpen.value = true
}

function closeTestModal() {
  isTestModalOpen.value = false
  activeApp.value = null
  testResult.value = null
}

async function executeTestDispatch() {
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

onMounted(async () => {
  await fetchUsers()
  await fetchApps()
})
</script>

<style scoped>
.apps-view {
  padding: 1.5rem;
}

.stats-overview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.stat-card {
  padding: 1.25rem;
  display: flex;
  align-items: center;
  gap: 1rem;
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
  color: var(--text-primary);
}
.stat-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.filter-card {
  padding: 0.875rem 1.25rem;
}

.filter-row {
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

.user-filter-box {
  min-width: 220px;
}

.apps-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.25rem;
}

.app-card {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  transition: transform 0.2s, box-shadow 0.2s;
}

.app-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.app-card-header {
  padding: 1.125rem 1.25rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-color);
}

.app-title-area {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.app-avatar-circle {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #ffffff;
  font-weight: 700;
  font-size: 0.875rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.app-name-meta {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.app-name {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.badges-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.owner-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 11px;
}

.action-buttons-group {
  display: flex;
  gap: 0.35rem;
}

.action-icon-btn {
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 0.4rem;
  cursor: pointer;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
}
.action-icon-btn:hover {
  background: #f8fafc;
  color: var(--text-primary);
}
.action-icon-btn.text-danger:hover {
  background: #fef2f2;
  color: #ef4444;
  border-color: #fecaca;
}

.app-card-body {
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
  flex: 1;
}

.credential-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.cred-label {
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.cred-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-sm);
  padding: 0.4rem 0.6rem;
  gap: 0.5rem;
}

.cred-code {
  font-family: monospace;
  font-size: 0.75rem;
  color: #0f172a;
  word-break: break-all;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cred-actions {
  display: flex;
  align-items: center;
  gap: 0.2rem;
}

.copy-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.2rem;
  display: flex;
  align-items: center;
}
.copy-btn:hover {
  color: #2563eb;
}

.url-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: var(--radius-sm);
  padding: 0.4rem 0.6rem;
  font-size: 0.75rem;
}

.url-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #334155;
  font-family: monospace;
}

.app-card-footer {
  padding: 0.875rem 1.25rem;
  background: #fafbfc;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 0.5rem;
}

/* Forms and Modals */
.drawer-form, .test-form {
  display: flex;
  flex-direction: column;
  gap: 1.125rem;
}

.form-row-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.code-editor {
  font-family: monospace;
  font-size: 0.8rem;
  background: #f8fafc;
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
  word-break: break-all;
}
.result-sig code {
  font-family: monospace;
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

.modal-text {
  font-size: 0.875rem;
  color: var(--text-primary);
  line-height: 1.5;
}
</style>
