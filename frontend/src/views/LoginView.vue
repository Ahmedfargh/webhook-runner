<template>
  <div class="auth-page">
    <div class="auth-card animate-fade-in">
      <div class="auth-top-bar">
        <LanguageSelector />
      </div>

      <!-- Zoho Projects Style Auth Header -->
      <div class="auth-header">
        <div class="brand-logo">
          <Layers :size="24" />
        </div>
        <h2 class="auth-title">{{ t('auth.signInTitle') }}</h2>
        <p class="auth-subtitle">{{ t('auth.signInSubtitle') }}</p>
      </div>

      <!-- Quick Demo Login Option -->
      <div class="demo-quick-fill">
        <div class="demo-info">
          <Zap :size="14" class="text-warning" />
          <span>{{ t('auth.quickEval') }}</span>
        </div>
        <button type="button" class="btn btn-sm btn-secondary" @click="fillDemoCreds">
          {{ t('auth.fillDemo') }}
        </button>
      </div>

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="auth-form">
        <div class="form-group">
          <label class="form-label">{{ t('auth.email') }}</label>
          <div class="input-with-icon">
            <Mail :size="15" class="input-icon" />
            <input
              v-model="email"
              type="email"
              required
              class="form-control"
              placeholder="admin@example.com"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('auth.password') }}</label>
          <div class="input-with-icon">
            <Lock :size="15" class="input-icon" />
            <input
              v-model="password"
              type="password"
              required
              class="form-control"
              placeholder="••••••••"
            />
          </div>
        </div>

        <button type="submit" class="btn btn-primary btn-lg w-full" :disabled="authStore.loading">
          <RefreshCw v-if="authStore.loading" :size="16" class="spin-anim" />
          {{ authStore.loading ? t('auth.signingIn') : t('auth.signInBtn') }}
        </button>
      </form>

      <div class="auth-footer">
        <span>{{ t('auth.noAccount') }}</span>
        <router-link to="/register" class="auth-link">{{ t('auth.createAccountTitle') }}</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { t } from '../locales'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'
import LanguageSelector from '../components/common/LanguageSelector.vue'
import { Layers, Mail, Lock, Zap, RefreshCw } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()
const toastStore = useToastStore()

const email = ref('')
const password = ref('')

function fillDemoCreds() {
  email.value = 'admin@example.com'
  password.value = 'Secret123456'
  toastStore.info('Demo credentials prefilled!')
}

async function handleLogin() {
  try {
    await authStore.login({
      email: email.value,
      password: password.value,
    })
    toastStore.success('Welcome back!')
    router.push('/')
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Invalid credentials')
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
  padding: 1.5rem;
}

.auth-card {
  width: 100%;
  max-width: 420px;
  background-color: #ffffff;
  border-radius: var(--radius-xl);
  padding: 2.25rem;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.auth-top-bar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 0.75rem;
}

.auth-header {
  text-align: center;
  margin-bottom: 1.5rem;
}

.brand-logo {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #1e40af, #3b82f6);
  border-radius: var(--radius-lg);
  color: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1rem;
  box-shadow: var(--shadow-md);
}

.auth-title {
  font-size: 20px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.auth-subtitle {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 0.25rem;
}

.demo-quick-fill {
  background-color: #f8fafc;
  border: 1px dashed var(--border-color);
  border-radius: var(--radius-md);
  padding: 0.625rem 0.875rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.25rem;
}

.demo-info {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
}

.input-with-icon {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 0.75rem;
  color: var(--text-muted);
}

.input-with-icon .form-control {
  padding-left: 2.25rem;
}

.w-full {
  width: 100%;
}

.auth-footer {
  margin-top: 1.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--border-color);
  text-align: center;
  font-size: 13px;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
}

.auth-link {
  color: var(--primary);
  font-weight: 600;
  text-decoration: none;
}
.auth-link:hover {
  text-decoration: underline;
}

.text-warning { color: var(--warning); }

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}
</style>
