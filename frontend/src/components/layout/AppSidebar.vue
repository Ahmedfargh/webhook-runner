<template>
  <aside class="sidebar" :class="{ 'sidebar-collapsed': isCollapsed }">
    <!-- Zoho Brand / Workspace Header -->
    <div class="sidebar-header">
      <div class="workspace-brand">
        <div class="workspace-logo">
          <Layers :size="20" class="logo-icon" />
        </div>
        <div v-if="!isCollapsed" class="workspace-info">
          <span class="workspace-name">{{ t('nav.brandSubtitle') }}</span>
          <span class="workspace-tag">{{ t('nav.brandTag') }}</span>
        </div>
      </div>
      <button class="collapse-toggle" @click="toggleCollapse" :title="isCollapsed ? 'Expand' : 'Collapse'">
        <PanelLeftClose v-if="!isCollapsed" :size="16" />
        <PanelLeftOpen v-else :size="16" />
      </button>
    </div>

    <!-- Navigation Menu -->
    <nav class="sidebar-nav">
      <!-- Admin Core Modules -->
      <template v-if="authStore.isAdmin">
        <div v-if="!isCollapsed" class="nav-section-title">{{ t('nav.coreModules') }}</div>
        
        <router-link to="/" class="nav-item" active-class="nav-item-active" exact>
          <LayoutDashboard :size="18" />
          <span v-if="!isCollapsed">{{ t('nav.dashboard') }}</span>
        </router-link>

        <router-link to="/users" class="nav-item" active-class="nav-item-active">
          <Users :size="18" />
          <span v-if="!isCollapsed">{{ t('nav.users') }}</span>
        </router-link>

        <router-link to="/admins" class="nav-item" active-class="nav-item-active">
          <ShieldCheck :size="18" />
          <span v-if="!isCollapsed">{{ t('nav.admins') }}</span>
        </router-link>

        <router-link to="/roles" class="nav-item" active-class="nav-item-active">
          <KeyRound :size="18" />
          <span v-if="!isCollapsed">{{ t('nav.roles') }}</span>
        </router-link>

        <router-link to="/permissions" class="nav-item" active-class="nav-item-active">
          <Lock :size="18" />
          <span v-if="!isCollapsed">{{ t('nav.permissions') }}</span>
        </router-link>
      </template>

      <!-- User Dashboard (if not admin) -->
      <template v-else>
        <router-link to="/" class="nav-item" active-class="nav-item-active" exact>
          <LayoutDashboard :size="18" />
          <span v-if="!isCollapsed">{{ t('nav.dashboard') }}</span>
        </router-link>
      </template>

      <!-- Apps & Webhooks Section -->
      <div v-if="!isCollapsed" class="nav-section-title">{{ t('nav.runnerAndWebhooks') }}</div>

      <router-link to="/apps" class="nav-item" active-class="nav-item-active">
        <Boxes :size="18" />
        <span v-if="!isCollapsed">{{ t('nav.apps') }}</span>
      </router-link>

      <router-link to="/webhooks/logs" class="nav-item" active-class="nav-item-active">
        <Activity :size="18" />
        <span v-if="!isCollapsed">{{ t('nav.webhookLogs') }}</span>
      </router-link>

      <!-- Billing and Plans Section -->
      <div v-if="!isCollapsed" class="nav-section-title">{{ t('nav.billingAndPlans') }}</div>

      <!-- Admin-Only Billing Management Console -->
      <template v-if="authStore.isAdmin">
        <router-link to="/admin/billing" class="nav-item" active-class="nav-item-active">
          <Sliders :size="18" />
          <span v-if="!isCollapsed">{{ t('nav.adminBilling') }}</span>
        </router-link>
      </template>

      <!-- Regular User: Plans, My Subscription, and Invoices -->
      <template v-else>
        <router-link to="/plans" class="nav-item" active-class="nav-item-active">
          <Sparkles :size="18" />
          <span v-if="!isCollapsed">{{ t('nav.plansPricing') }}</span>
        </router-link>

        <router-link to="/subscription" class="nav-item" active-class="nav-item-active">
          <CreditCard :size="18" />
          <span v-if="!isCollapsed">{{ t('nav.mySubscription') }}</span>
        </router-link>

        <router-link to="/invoices" class="nav-item" active-class="nav-item-active">
          <Receipt :size="18" />
          <span v-if="!isCollapsed">{{ t('nav.invoices') }}</span>
        </router-link>
      </template>

      <!-- Admin-Only System Security -->
      <template v-if="authStore.isAdmin">
        <div v-if="!isCollapsed" class="nav-section-title">{{ t('nav.systemSecurity') }}</div>

        <router-link to="/topology" class="nav-item" active-class="nav-item-active">
          <Network :size="18" />
          <span v-if="!isCollapsed">{{ t('nav.topology') }}</span>
        </router-link>
      </template>
    </nav>

    <!-- Bottom Service Health & User Widget -->
    <div class="sidebar-footer">
      <div v-if="!isCollapsed" class="service-pill" :class="healthStatusClass">
        <span class="pulse-dot" :class="healthStatusClass"></span>
        <span class="service-text">{{ healthLabel }}</span>
      </div>

      <div class="user-profile-widget">
        <div class="user-avatar">
          {{ userInitials }}
        </div>
        <div v-if="!isCollapsed" class="user-info-text">
          <div class="user-display-name">{{ authStore.user?.name || 'Admin User' }}</div>
          <div class="user-role-badge">{{ authStore.user?.role || 'administrator' }}</div>
        </div>
        <button v-if="!isCollapsed" class="logout-btn" @click="handleLogout" :title="t('nav.signOut')">
          <LogOut :size="15" />
        </button>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { t } from '../../locales'
import { useAuthStore } from '../../stores/auth'
import { useToastStore } from '../../stores/toast'
import { healthService } from '../../services/healthService'
import {
  Layers,
  LayoutDashboard,
  Users,
  ShieldCheck,
  KeyRound,
  Lock,
  Network,
  LogOut,
  PanelLeftClose,
  PanelLeftOpen,
  Sparkles,
  CreditCard,
  Receipt,
  Sliders,
  Boxes,
  Activity,
} from 'lucide-vue-next'

const authStore = useAuthStore()
const toastStore = useToastStore()
const router = useRouter()

const isCollapsed = ref(false)
const isHealthy = ref(true)

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value
}

const userInitials = computed(() => {
  const name = authStore.user?.name || 'AD'
  return name.slice(0, 2).toUpperCase()
})

const healthLabel = computed(() => (isHealthy.value ? 'Gateway & gRPC Active' : 'Service Connecting...'))
const healthStatusClass = computed(() => (isHealthy.value ? 'success' : 'warning'))

async function checkHealth() {
  try {
    const res = await healthService.getHealth()
    isHealthy.value = res.status === 'ok' && res.upstream_services?.accounts_grpc?.status === 'healthy'
  } catch (err) {
    isHealthy.value = false
  }
}

onMounted(() => {
  checkHealth()
  setInterval(checkHealth, 15000)
})

function handleLogout() {
  authStore.logout()
  toastStore.info('You have logged out successfully.')
  router.push('/login')
}
</script>

<style scoped>
.sidebar {
  width: 240px;
  height: 100vh;
  background-color: var(--zoho-sidebar-bg);
  border-right: 1px solid var(--zoho-sidebar-border);
  display: flex;
  flex-direction: column;
  transition: width var(--transition-normal);
  flex-shrink: 0;
  z-index: 50;
}

.sidebar-collapsed {
  width: 64px;
}

.sidebar-header {
  padding: 1.125rem 1rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--zoho-sidebar-border);
}

.workspace-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  overflow: hidden;
}

.workspace-logo {
  width: 34px;
  height: 34px;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  flex-shrink: 0;
}

.workspace-info {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.workspace-name {
  font-size: 13px;
  font-weight: 700;
  color: var(--zoho-sidebar-text-active);
  white-space: nowrap;
}

.workspace-tag {
  font-size: 10px;
  color: #00a8cc;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.collapse-toggle {
  background: transparent;
  border: none;
  color: var(--zoho-sidebar-text);
  cursor: pointer;
  padding: 0.35rem;
  border-radius: var(--radius-xs);
  display: flex;
  align-items: center;
  justify-content: center;
}
.collapse-toggle:hover {
  background-color: var(--zoho-sidebar-active);
  color: #ffffff;
}

.sidebar-nav {
  flex: 1;
  padding: 1rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  overflow-y: auto;
}

.nav-section-title {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #64748b;
  padding: 0.75rem 0.5rem 0.25rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 0.75rem;
  border-radius: var(--radius-md);
  color: var(--zoho-sidebar-text);
  font-size: 13px;
  font-weight: 500;
  text-decoration: none;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.nav-item:hover {
  background-color: var(--zoho-sidebar-hover);
  color: var(--zoho-sidebar-text-active);
}

.nav-item-active {
  background-color: var(--zoho-sidebar-active);
  color: #ffffff;
  font-weight: 600;
  border-left: 3px solid #3b82f6;
}

.sidebar-footer {
  padding: 1rem 0.75rem;
  border-top: 1px solid var(--zoho-sidebar-border);
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.service-pill {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.35rem 0.625rem;
  border-radius: var(--radius-full);
  background-color: #1e293b;
  font-size: 11px;
  font-weight: 500;
  color: #94a3b8;
}

.user-profile-widget {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.35rem;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-full);
  background-color: #3b82f6;
  color: #ffffff;
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-info-text {
  flex: 1;
  overflow: hidden;
}

.user-display-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--zoho-sidebar-text-active);
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
}

.user-role-badge {
  font-size: 10px;
  color: var(--zoho-sidebar-text);
  text-transform: capitalize;
}

.logout-btn {
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 0.35rem;
  border-radius: 4px;
}
.logout-btn:hover {
  color: #ef4444;
  background-color: #1e293b;
}
</style>
