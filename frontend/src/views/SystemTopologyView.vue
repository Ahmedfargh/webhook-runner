<template>
  <div class="topology-view animate-fade-in">
    <!-- Action Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('topology.title') }}</h1>
        <p class="view-subtitle">{{ t('topology.subtitle') }}</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-secondary" @click="runLiveHealthCheck" :disabled="checking">
          <RefreshCw :size="14" :class="{ 'spin-anim': checking }" /> {{ t('topology.pingBtn') }}
        </button>
      </div>
    </div>

    <!-- Visual Architecture Diagram -->
    <div class="card topo-card">
      <div class="card-header">
        <div class="card-header-left">
          <Network :size="18" class="text-primary" />
          <h3 class="card-title">{{ t('topology.pipelineTitle') }}</h3>
        </div>
        <span class="badge badge-success">
          <ShieldCheck :size="12" /> {{ t('dashboard.protectedMesh') }}
        </span>
      </div>

      <div class="pipeline-diagram">
        <!-- Step 1: Vue Client -->
        <div class="pipeline-node">
          <div class="pipeline-badge">{{ t('dashboard.clientLayer') }}</div>
          <div class="pipeline-box">
            <Globe :size="24" class="text-primary" />
            <div class="box-title">Vue.js 3 App</div>
            <div class="box-desc">Zoho Projects UI</div>
            <div class="box-meta font-mono">Port 5173</div>
          </div>
        </div>

        <!-- Connector 1 -->
        <div class="pipeline-arrow">
          <span class="arrow-protocol">HTTP REST</span>
          <div class="arrow-line">
            <ArrowRight :size="16" />
          </div>
          <span class="arrow-auth">Bearer JWT</span>
        </div>

        <!-- Step 2: API Gateway -->
        <div class="pipeline-node">
          <div class="pipeline-badge">{{ t('dashboard.gatewayLayer') }}</div>
          <div class="pipeline-box active-layer">
            <Cpu :size="24" class="text-primary" />
            <div class="box-title">Webhook API Gateway</div>
            <div class="box-desc">Gin REST Router & CORS</div>
            <div class="box-meta font-mono">Port 8080</div>
          </div>
        </div>

        <!-- Connector 2 (Protected gRPC) -->
        <div class="pipeline-arrow protected-arrow">
          <span class="arrow-protocol">gRPC / Protobuf</span>
          <div class="arrow-line protected-line">
            <Lock :size="12" />
            <ArrowRight :size="16" />
          </div>
          <span class="arrow-auth font-mono">X-Service-Name: api-gateway</span>
        </div>

        <!-- Step 3: Accounts Service -->
        <div class="pipeline-node">
          <div class="pipeline-badge">{{ t('dashboard.serviceLayer') }}</div>
          <div class="pipeline-box">
            <Server :size="24" class="text-primary" />
            <div class="box-title">Accounts Service</div>
            <div class="box-desc">HMVC gRPC Microservice</div>
            <div class="box-meta font-mono">Port 50051</div>
          </div>
        </div>

        <!-- Connector 3 -->
        <div class="pipeline-arrow">
          <span class="arrow-protocol">TCP / SQL</span>
          <div class="arrow-line">
            <ArrowRight :size="16" />
          </div>
          <span class="arrow-auth">GORM Pool</span>
        </div>

        <!-- Step 4: Database -->
        <div class="pipeline-node">
          <div class="pipeline-badge">{{ t('dashboard.databaseLayer') }}</div>
          <div class="pipeline-box">
            <Database :size="24" class="text-primary" />
            <div class="box-title">MySQL 8.0</div>
            <div class="box-desc">webhook_accounts</div>
            <div class="box-meta font-mono">Port 3306</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Detailed Security & Diagnostics Grid -->
    <div class="diagnostics-grid">
      <!-- Security Enforcement Panel -->
      <div class="card diag-card">
        <div class="card-header">
          <div class="card-header-left">
            <ShieldCheck :size="18" class="text-success" />
            <h3 class="card-title">{{ t('topology.securityPolicies') }}</h3>
          </div>
        </div>

        <div class="sec-policy-list">
          <div class="policy-item">
            <div class="policy-bullet"><Check :size="14" /></div>
            <div>
              <div class="policy-title">{{ t('topology.whitelistPolicyTitle') }}</div>
              <div class="policy-desc">
                {{ t('topology.whitelistPolicyDesc') }}
              </div>
            </div>
          </div>

          <div class="policy-item">
            <div class="policy-bullet"><Check :size="14" /></div>
            <div>
              <div class="policy-title">{{ t('topology.secretPolicyTitle') }}</div>
              <div class="policy-desc">
                {{ t('topology.secretPolicyDesc') }}
              </div>
            </div>
          </div>

          <div class="policy-item">
            <div class="policy-bullet"><Check :size="14" /></div>
            <div>
              <div class="policy-title">{{ t('topology.jwtPolicyTitle') }}</div>
              <div class="policy-desc">
                {{ t('topology.jwtPolicyDesc') }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Live Diagnostics Response -->
      <div class="card diag-card">
        <div class="card-header">
          <div class="card-header-left">
            <Terminal :size="18" class="text-primary" />
            <h3 class="card-title">{{ t('topology.diagnosticsTitle') }}</h3>
          </div>
          <span class="badge badge-success">HTTP 200 OK</span>
        </div>

        <div class="raw-json-box">
          <pre><code>{{ formattedHealthData }}</code></pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '../locales'
import { healthService } from '../services/healthService'
import { useToastStore } from '../stores/toast'
import {
  Network,
  ShieldCheck,
  Globe,
  Cpu,
  Server,
  Database,
  ArrowRight,
  Lock,
  Check,
  Terminal,
  RefreshCw,
} from 'lucide-vue-next'

const { t } = useI18n()
const toastStore = useToastStore()
const checking = ref(false)
const healthData = ref(null)

const formattedHealthData = computed(() => {
  if (!healthData.value) return '{\n  "status": "Checking status..."\n}'
  return JSON.stringify(healthData.value, null, 2)
})

async function runLiveHealthCheck() {
  checking.value = true
  try {
    const res = await healthService.getHealth()
    healthData.value = res
    toastStore.success('Live service mesh ping succeeded!')
  } catch (err) {
    toastStore.error('Service ping returned an error')
  } finally {
    checking.value = false
  }
}

onMounted(() => {
  runLiveHealthCheck()
})
</script>

<style scoped>
.topology-view {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
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

.topo-card {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
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

.pipeline-diagram {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 1.5rem;
  background-color: var(--bg-card-muted);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  overflow-x: auto;
}

.pipeline-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  min-width: 140px;
}

.pipeline-badge {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
}

.pipeline-box {
  width: 100%;
  background-color: #ffffff;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 1rem 0.75rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 0.35rem;
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-fast);
}

.pipeline-box:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.pipeline-box.active-layer {
  border-color: #93c5fd;
  box-shadow: 0 0 0 2px var(--primary-focus);
}

.box-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
}

.box-desc {
  font-size: 11px;
  color: var(--text-secondary);
}

.box-meta {
  font-size: 10px;
  color: var(--accent-cyan);
  background-color: var(--accent-cyan-light);
  padding: 0.15rem 0.45rem;
  border-radius: var(--radius-xs);
  margin-top: 0.25rem;
}

.pipeline-arrow {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  flex: 1;
  min-width: 100px;
}

.arrow-protocol {
  font-size: 10px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
}

.arrow-line {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
}

.protected-line {
  color: var(--success);
  gap: 0.25rem;
}

.arrow-auth {
  font-size: 9px;
  color: var(--text-secondary);
  text-align: center;
}

.diagnostics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(380px, 1fr));
  gap: 1.25rem;
}

.diag-card {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.sec-policy-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.policy-item {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}

.policy-bullet {
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  background-color: var(--success-light);
  color: var(--success);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 0.1rem;
}

.policy-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.15rem;
}

.policy-desc {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.policy-desc code {
  background-color: var(--bg-card-muted);
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  color: var(--primary);
}

.raw-json-box {
  background-color: #0f172a;
  border-radius: var(--radius-md);
  padding: 1rem;
  overflow-x: auto;
  font-size: 12px;
  color: #38bdf8;
  border: 1px solid #1e293b;
}

.text-primary { color: var(--primary); }
.text-success { color: var(--success); }

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}
</style>
