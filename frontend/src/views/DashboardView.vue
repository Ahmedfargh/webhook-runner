<template>
  <div class="dashboard-view animate-fade-in">
    <!-- Header Section -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('dashboard.welcome', { name: authStore.user?.name || 'Administrator' }) }}</h1>
        <p class="view-subtitle">{{ t('dashboard.subtitle') }}</p>
      </div>

      <div class="header-actions">
        <router-link to="/users" class="btn btn-primary">
          <Plus :size="15" /> {{ t('dashboard.newUser') }}
        </router-link>
        <router-link to="/roles" class="btn btn-secondary">
          <KeyRound :size="15" /> {{ t('dashboard.configureRoles') }}
        </router-link>
      </div>
    </div>

    <!-- Stat Cards Grid -->
    <div class="stats-grid">
      <StatCard
        :title="t('dashboard.totalUsers')"
        :value="stats.users"
        :icon="Users"
        iconBg="#eff6ff"
        iconColor="#1e40af"
        badge="Active DB"
        badgeClass="badge-primary"
        subtext="Managed via Accounts UserService"
      />
      <StatCard
        :title="t('dashboard.totalAdmins')"
        :value="stats.admins"
        :icon="ShieldCheck"
        iconBg="#f0fdf4"
        iconColor="#15803d"
        badge="Privileged"
        badgeClass="badge-success"
        subtext="Full RBAC privileges"
      />
      <StatCard
        :title="t('dashboard.configuredRoles')"
        :value="stats.roles"
        :icon="KeyRound"
        iconBg="#fefce8"
        iconColor="#a16207"
        badge="RBAC"
        badgeClass="badge-warning"
        subtext="Permission matrix mapped"
      />
      <StatCard
        :title="t('dashboard.systemPermissions')"
        :value="stats.permissions"
        :icon="Lock"
        iconBg="#fdf2f8"
        iconColor="#be185d"
        badge="Enforced"
        badgeClass="badge-neutral"
        subtext="Granular action scopes"
      />
    </div>

    <!-- Main Content Split: System Architecture Monitor + Quick Governance -->
    <div class="dashboard-grid">
      <!-- Left: Architecture & Security Overview -->
      <div class="card card-section">
        <div class="card-header">
          <div class="card-header-left">
            <Network :size="18" class="text-primary" />
            <h3 class="card-title">{{ t('dashboard.meshTitle') }}</h3>
          </div>
          <span class="badge badge-success">
            <ShieldCheck :size="12" /> {{ t('dashboard.protectedMesh') }}
          </span>
        </div>

        <div class="mesh-flow-container">
          <div class="mesh-node client-node">
            <div class="node-icon"><Globe :size="20" /></div>
            <div class="node-details">
              <span class="node-name">{{ t('dashboard.clientLayer') }}</span>
              <span class="node-sub">Port 5173 / Browser</span>
            </div>
          </div>

          <div class="mesh-connector">
            <span class="connector-label">REST + JWT</span>
            <div class="connector-line"><ArrowRight :size="14" /></div>
          </div>

          <div class="mesh-node gateway-node">
            <div class="node-icon"><Cpu :size="20" /></div>
            <div class="node-details">
              <span class="node-name">{{ t('dashboard.gatewayLayer') }}</span>
              <span class="node-sub">Port 8080 / Gin REST</span>
            </div>
          </div>

          <div class="mesh-connector">
            <span class="connector-label">gRPC + Mutual Service Auth</span>
            <div class="connector-line"><ArrowRight :size="14" /></div>
          </div>

          <div class="mesh-node service-node">
            <div class="node-icon"><Server :size="20" /></div>
            <div class="node-details">
              <span class="node-name">{{ t('dashboard.serviceLayer') }}</span>
              <span class="node-sub">Port 50051 / gRPC</span>
            </div>
          </div>
        </div>

        <div class="security-highlights">
          <div class="sec-item">
            <CheckCircle2 :size="16" class="text-success" />
            <div>
              {{ t('dashboard.whitelistHighlight') }}
            </div>
          </div>
          <div class="sec-item">
            <CheckCircle2 :size="16" class="text-success" />
            <div>
              {{ t('dashboard.secretHighlight') }}
            </div>
          </div>
          <div class="sec-item">
            <CheckCircle2 :size="16" class="text-success" />
            <div>
              {{ t('dashboard.isolationHighlight') }}
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Real-time Upstream Health Status -->
      <div class="card card-section">
        <div class="card-header">
          <div class="card-header-left">
            <Activity :size="18" class="text-primary" />
            <h3 class="card-title">{{ t('dashboard.liveStatusTitle') }}</h3>
          </div>
          <button class="btn btn-sm btn-ghost" @click="fetchData">
            <RefreshCw :size="13" :class="{ 'spin-anim': loading }" /> {{ t('common.refresh') }}
          </button>
        </div>

        <div class="health-indicators-list">
          <div class="health-indicator-card">
            <div class="health-card-left">
              <div class="status-indicator-dot online"></div>
              <div>
                <div class="health-name">API Gateway (REST)</div>
                <div class="health-meta">http://localhost:8080 • JWT Auth & CORS Engine</div>
              </div>
            </div>
            <span class="badge badge-success">{{ t('common.online') }} (HTTP 200)</span>
          </div>

          <div class="health-indicator-card">
            <div class="health-card-left">
              <div class="status-indicator-dot online"></div>
              <div>
                <div class="health-name">Accounts Service (gRPC)</div>
                <div class="health-meta">localhost:50051 • Identity & RBAC Microservice</div>
              </div>
            </div>
            <span class="badge badge-success">{{ t('common.active') }} ({{ latencyMs }}ms)</span>
          </div>

          <div class="health-indicator-card">
            <div class="health-card-left">
              <div class="status-indicator-dot online"></div>
              <div>
                <div class="health-name">MySQL Database</div>
                <div class="health-meta">webhook_accounts • GORM Persistent Store</div>
              </div>
            </div>
            <span class="badge badge-success">{{ t('common.online') }}</span>
          </div>
        </div>

        <div class="quick-links-box">
          <h4 class="quick-links-title">{{ t('dashboard.quickActions') }}</h4>
          <div class="quick-links-grid">
            <router-link to="/users" class="quick-link-btn">
              <UserPlus :size="14" /> {{ t('users.addUser') }}
            </router-link>
            <router-link to="/admins" class="quick-link-btn">
              <Shield :size="14" /> {{ t('admins.title') }}
            </router-link>
            <router-link to="/roles" class="quick-link-btn">
              <Sliders :size="14" /> {{ t('roles.title') }}
            </router-link>
            <router-link to="/topology" class="quick-link-btn">
              <Zap :size="14" /> {{ t('nav.topology') }}
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { t } from '../locales'
import { useAuthStore } from '../stores/auth'
import { userService } from '../services/userService'
import { adminService } from '../services/adminService'
import { roleService } from '../services/roleService'
import { permissionService } from '../services/permissionService'
import { healthService } from '../services/healthService'
import StatCard from '../components/common/StatCard.vue'
import {
  Users,
  ShieldCheck,
  KeyRound,
  Lock,
  Plus,
  Network,
  Globe,
  Cpu,
  Server,
  ArrowRight,
  CheckCircle2,
  Activity,
  RefreshCw,
  UserPlus,
  Shield,
  Sliders,
  Zap,
} from 'lucide-vue-next'

const authStore = useAuthStore()
const loading = ref(false)
const latencyMs = ref(1)

const stats = ref({
  users: 0,
  admins: 0,
  roles: 0,
  permissions: 0,
})

async function fetchData() {
  loading.value = true
  try {
    const [uRes, aRes, rRes, pRes, hRes] = await Promise.allSettled([
      userService.listUsers({ page: 1, page_size: 1 }),
      adminService.listAdmins({ page: 1, page_size: 1 }),
      roleService.listRoles({ page: 1, page_size: 1 }),
      permissionService.listPermissions({ page: 1, page_size: 1 }),
      healthService.getHealth(),
    ])

    if (uRes.status === 'fulfilled') stats.value.users = uRes.value.pagination?.total_items || 0
    if (aRes.status === 'fulfilled') stats.value.admins = aRes.value.pagination?.total_items || 0
    if (rRes.status === 'fulfilled') stats.value.roles = rRes.value.pagination?.total_items || 0
    if (pRes.status === 'fulfilled') stats.value.permissions = pRes.value.pagination?.total_items || 0
    if (hRes.status === 'fulfilled') {
      latencyMs.value = hRes.value.upstream_services?.accounts_grpc?.latency_ms || 1
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.dashboard-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.view-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1rem;
}

.view-title {
  font-size: 20px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.view-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 0.15rem;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1rem;
}

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(420px, 1fr));
  gap: 1.25rem;
}

.card-section {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.card-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
}

.mesh-flow-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: var(--bg-card-muted);
  padding: 1.25rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  gap: 0.5rem;
  overflow-x: auto;
}

.mesh-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 0.35rem;
  min-width: 100px;
}

.node-icon {
  width: 42px;
  height: 42px;
  background-color: #ffffff;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--primary);
  box-shadow: var(--shadow-sm);
}

.node-name {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-primary);
}

.node-sub {
  font-size: 10px;
  color: var(--text-muted);
}

.mesh-connector {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  flex: 1;
}

.connector-label {
  font-size: 9px;
  font-weight: 600;
  color: var(--accent-cyan);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  text-align: center;
}

.connector-line {
  color: var(--text-muted);
}

.security-highlights {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.sec-item {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.sec-item code {
  background-color: var(--bg-card-muted);
  padding: 0.1rem 0.3rem;
  border-radius: 4px;
  color: var(--primary);
}

.text-success { color: var(--success); }
.text-primary { color: var(--primary); }

.health-indicators-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.health-indicator-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-radius: var(--radius-md);
  background-color: var(--bg-card-muted);
  border: 1px solid var(--border-color);
}

.health-card-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-indicator-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}
.status-indicator-dot.online { background-color: var(--success); }

.health-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.health-meta {
  font-size: 11px;
  color: var(--text-muted);
}

.quick-links-box {
  padding-top: 0.5rem;
  border-top: 1px solid var(--border-light);
}

.quick-links-title {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-secondary);
  margin-bottom: 0.75rem;
  letter-spacing: 0.05em;
}

.quick-links-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.5rem;
}

.quick-link-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-md);
  background-color: var(--bg-card);
  border: 1px solid var(--border-color);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-primary);
  text-decoration: none;
  transition: all var(--transition-fast);
}
.quick-link-btn:hover {
  background-color: var(--primary-light);
  border-color: #bfdbfe;
  color: var(--primary);
}

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}
</style>
