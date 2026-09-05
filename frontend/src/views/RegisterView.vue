<template>
  <div class="auth-page">
    <div class="auth-card animate-fade-in">
      <div class="auth-top-bar">
        <LanguageSelector />
      </div>

      <div class="auth-header">
        <div class="brand-logo">
          <Layers :size="24" />
        </div>
        <h2 class="auth-title">{{ t('auth.createAccountTitle') }}</h2>
        <p class="auth-subtitle">{{ t('auth.createAccountSubtitle') }}</p>
      </div>

      <form @submit.prevent="handleRegister" class="auth-form">
        <div class="form-group">
          <label class="form-label">{{ t('auth.fullName') }} *</label>
          <div class="input-with-icon">
            <User :size="15" class="input-icon" />
            <input
              v-model="name"
              type="text"
              required
              class="form-control"
              placeholder="e.g. Alex Morgan"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('auth.email') }} *</label>
          <div class="input-with-icon">
            <Mail :size="15" class="input-icon" />
            <input
              v-model="email"
              type="email"
              required
              class="form-control"
              placeholder="alex@example.com"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('auth.phoneNumber') }} *</label>
          <div class="input-with-icon">
            <Phone :size="15" class="input-icon" />
            <input
              v-model="phone"
              type="tel"
              required
              class="form-control"
              placeholder="01128242012"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('common.country') }} *</label>
          <select v-model="country_id" required class="form-control">
            <option value="" disabled>{{ t('common.selectCountry') }}</option>
            <option v-for="c in countries" :key="c.id" :value="c.id">
              {{ c.name_en }} ({{ c.country_code }}) - {{ c.name_ar }}
            </option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('auth.password') }} *</label>
          <div class="input-with-icon">
            <Lock :size="15" class="input-icon" />
            <input
              v-model="password"
              type="password"
              required
              minlength="6"
              class="form-control"
              placeholder="••••••••"
            />
          </div>
        </div>

        <button type="submit" class="btn btn-primary btn-lg w-full" :disabled="authStore.loading">
          <RefreshCw v-if="authStore.loading" :size="16" class="spin-anim" />
          {{ authStore.loading ? t('auth.creatingAccount') : t('auth.registerBtn') }}
        </button>
      </form>

      <div class="auth-footer">
        <span>{{ t('auth.haveAccount') }}</span>
        <router-link to="/login" class="auth-link">{{ t('auth.signInBtn') }}</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { t } from '../locales'
import { useAuthStore } from '../stores/auth'
import { useToastStore } from '../stores/toast'
import { countryService } from '../services/countryService'
import LanguageSelector from '../components/common/LanguageSelector.vue'
import { Layers, User, Mail, Phone, Lock, RefreshCw } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()
const toastStore = useToastStore()

const name = ref('')
const email = ref('')
const phone = ref('')
const password = ref('')
const country_id = ref('')
const countries = ref([])

onMounted(async () => {
  try {
    const res = await countryService.listCountries()
    countries.value = res.data || []
    const egypt = countries.value.find((c) => c.country_code === 'EG')
    country_id.value = egypt ? egypt.id : (countries.value[0]?.id || '')
  } catch (err) {
    // Non-critical
  }
})

async function handleRegister() {
  if (!country_id.value) {
    toastStore.warning('Please select a country.')
    return
  }

  try {
    await authStore.register({
      name: name.value,
      email: email.value,
      phone: phone.value,
      country_id: country_id.value,
      password: password.value,
    })
    toastStore.success('Account created successfully!')
    router.push('/')
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Registration failed')
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
  max-width: 440px;
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

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}
</style>
