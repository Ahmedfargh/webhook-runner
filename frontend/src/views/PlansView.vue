<template>
  <div class="plans-view animate-fade-in">
    <!-- Header -->
    <div class="view-header text-center">
      <div v-if="authStore.isAdmin" class="admin-top-banner">
        <router-link to="/admin/billing" class="btn btn-xs btn-secondary">
          <Sliders :size="13" /> {{ t('billing.plansManagement') }} & {{ t('nav.adminBilling') }}
        </router-link>
      </div>

      <div>
        <span class="badge badge-primary uppercase-badge">{{ t('billing.plansBadge') }}</span>
        <h1 class="view-title mt-2">{{ t('billing.plansTitle') }}</h1>
        <p class="view-subtitle">{{ t('billing.plansSubtitle') }}</p>
      </div>

      <!-- Monthly / Annual Toggle -->
      <div class="billing-toggle-container">
        <button
          class="toggle-btn"
          :class="{ active: billingCycle === 'monthly' }"
          @click="billingCycle = 'monthly'"
        >
          {{ t('billing.monthly') }}
        </button>
        <button
          class="toggle-btn"
          :class="{ active: billingCycle === 'annually' }"
          @click="billingCycle = 'annually'"
        >
          {{ t('billing.annually') }}
          <span class="discount-pill">{{ t('billing.save20') }}</span>
        </button>
      </div>
    </div>

    <!-- Pricing Cards Grid -->
    <div v-if="loading" class="text-center py-12">
      <RefreshCw :size="24" class="spin-anim text-primary" />
      <p class="mt-2 text-muted">{{ t('common.loading') }}</p>
    </div>

    <div v-else class="pricing-grid">
      <div
        v-for="plan in plans"
        :key="plan.id"
        class="pricing-card"
        :class="{ 'popular-card': plan.is_popular || plan.isPopular, 'current-plan-card': isCurrentPlan(plan.id) }"
      >
        <div v-if="plan.is_popular || plan.isPopular" class="popular-ribbon">
          <Sparkles :size="12" /> {{ t('billing.mostPopular') }}
        </div>

        <div class="card-top">
          <h3 class="plan-name">{{ plan.name }}</h3>
          <p class="plan-desc">{{ plan.description }}</p>

          <div class="price-box">
            <span class="currency-symbol">{{ plan.currency === 'USD' ? '$' : (plan.currency || '$') }}</span>
            <span class="price-amount">{{ getPrice(plan) }}</span>
            <span class="price-period">/ {{ billingCycle === 'annually' ? t('billing.year') : t('billing.month') }}</span>
          </div>

          <button
            class="btn w-full mt-4"
            :class="isCurrentPlan(plan.id) ? 'btn-secondary' : (plan.is_popular || plan.isPopular) ? 'btn-primary' : 'btn-outline'"
            :disabled="isCurrentPlan(plan.id) || ordering"
            @click="openOrderModal(plan)"
          >
            <span v-if="isCurrentPlan(plan.id)">{{ t('billing.currentPlan') }}</span>
            <span v-else-if="isFreePlan(plan)">{{ t('billing.getStartedFree') }}</span>
            <span v-else>{{ t('billing.upgradeTo') }} {{ plan.name }}</span>
          </button>
        </div>

        <div class="card-divider"></div>

        <div class="card-features">
          <div class="features-title">{{ t('billing.whatsIncluded') }}</div>
          <ul class="feature-list">
            <li class="feature-item">
              <Check :size="14" class="text-success" />
              <span>
                <strong>{{ (plan.max_webhooks ?? plan.maxWebhooks) === -1 ? t('billing.unlimited') : (plan.max_webhooks ?? plan.maxWebhooks ?? 100) }}</strong>
                {{ t('billing.webhookEndpoints') }}
              </span>
            </li>
            <li class="feature-item">
              <Check :size="14" class="text-success" />
              <span>
                <strong>{{ formatEvents(plan.max_events_per_month ?? plan.maxEventsPerMonth) }}</strong>
                {{ t('billing.monthlyEvents') }}
              </span>
            </li>
            <li class="feature-item">
              <Check :size="14" class="text-success" />
              <span>
                <strong>{{ (plan.max_team_members ?? plan.maxTeamMembers) === -1 ? t('billing.unlimited') : (plan.max_team_members ?? plan.maxTeamMembers ?? 1) }}</strong>
                {{ t('billing.teamSeats') }}
              </span>
            </li>
            <li v-for="(feat, idx) in plan.features" :key="idx" class="feature-item">
              <Check :size="14" class="text-success" />
              <span>{{ feat }}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- Offline Bank Wire Order Modal -->
    <Modal
      :isOpen="isOrderModalOpen"
      :title="t('billing.confirmOrderTitle')"
      @close="isOrderModalOpen = false"
    >
      <div v-if="selectedPlan" class="order-summary-box">
        <div class="order-row">
          <span class="text-secondary">{{ t('billing.selectedPlan') }}:</span>
          <strong>{{ selectedPlan.name }}</strong>
        </div>
        <div class="order-row">
          <span class="text-secondary">{{ t('billing.billingPeriod') }}:</span>
          <span>{{ billingCycle === 'annually' ? t('billing.annualBilling') : t('billing.monthlyBilling') }}</span>
        </div>
        <div class="order-row total-row">
          <span>{{ t('billing.amountDue') }}:</span>
          <span class="total-price">${{ getPrice(selectedPlan) }} {{ selectedPlan.currency || 'USD' }}</span>
        </div>

        <div v-if="!isFreePlan(selectedPlan)" class="manual-payment-notice">
          <div class="notice-title"><Info :size="14" /> {{ t('billing.offlinePaymentNoticeTitle') }}</div>
          <p class="notice-text">
            {{ t('billing.offlinePaymentNoticeDesc') }}
          </p>
        </div>
      </div>

      <template #footer>
        <button class="btn btn-secondary" @click="isOrderModalOpen = false" :disabled="ordering">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="confirmSubscribe" :disabled="ordering">
          <RefreshCw v-if="ordering" :size="14" class="spin-anim" />
          {{ t('billing.confirmAndGetInvoice') }}
        </button>
      </template>
    </Modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from '../locales'
import { planService } from '../services/planService'
import { subscriptionService } from '../services/subscriptionService'
import { useToastStore } from '../stores/toast'
import { useAuthStore } from '../stores/auth'
import Modal from '../components/common/Modal.vue'
import {
  Check,
  Sparkles,
  RefreshCw,
  Info,
  Sliders,
} from 'lucide-vue-next'

const { t } = useI18n()
const router = useRouter()
const toastStore = useToastStore()
const authStore = useAuthStore()

const plans = ref([])
const currentSubscription = ref(null)
const loading = ref(false)
const ordering = ref(false)
const billingCycle = ref('monthly') // monthly | annually
const isOrderModalOpen = ref(false)
const selectedPlan = ref(null)

async function fetchPlans() {
  loading.value = true
  try {
    const [plansRes, subRes] = await Promise.all([
      planService.listPlans(false),
      subscriptionService.getCurrentSubscription().catch(() => ({ data: null })),
    ])
    plans.value = plansRes.data || []
    currentSubscription.value = subRes?.data || null
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to load plans')
  } finally {
    loading.value = false
  }
}

function isCurrentPlan(planId) {
  if (!currentSubscription.value) return false
  const currentPlanId = currentSubscription.value.plan_id || currentSubscription.value.planId || currentSubscription.value.plan?.id
  const status = currentSubscription.value.status
  return currentPlanId === planId && (status === 'active' || status === 'STATUS_ACTIVE')
}

function getPrice(plan) {
  if (!plan) return '0.00'
  const monthly = plan.price_monthly ?? plan.priceMonthly ?? 0
  const annually = plan.price_annually ?? plan.priceAnnually ?? 0
  if (billingCycle.value === 'annually') {
    return Number(annually).toFixed(2)
  }
  return Number(monthly).toFixed(2)
}

function isFreePlan(plan) {
  if (!plan) return false
  const monthly = plan.price_monthly ?? plan.priceMonthly ?? 0
  const annually = plan.price_annually ?? plan.priceAnnually ?? 0
  return Number(monthly) === 0 && Number(annually) === 0
}

function formatEvents(count) {
  const num = Number(count || 0)
  if (num <= 0) return '0'
  if (num >= 1000000) return (num / 1000000).toFixed(0) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(0) + 'K'
  return num.toString()
}

function openOrderModal(plan) {
  selectedPlan.value = plan
  isOrderModalOpen.value = true
}

async function confirmSubscribe() {
  if (!selectedPlan.value) return
  ordering.value = true
  try {
    const res = await subscriptionService.subscribe({
      plan_id: selectedPlan.value.id,
      billing_cycle: billingCycle.value,
      payment_method: 'bank_transfer',
      notes: `Subscribed via portal for ${billingCycle.value} cycle`,
    })

    toastStore.success('Subscription order placed successfully!')
    isOrderModalOpen.value = false

    if (selectedPlan.value.price_monthly > 0) {
      router.push('/invoices')
    } else {
      fetchPlans()
    }
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to process subscription')
  } finally {
    ordering.value = false
  }
}

onMounted(() => {
  fetchPlans()
})
</script>

<style scoped>
.plans-view {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.view-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}

.uppercase-badge {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.view-title {
  font-size: 26px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.view-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  max-width: 540px;
}

.billing-toggle-container {
  display: inline-flex;
  background-color: var(--bg-card-muted);
  padding: 0.3rem;
  border-radius: var(--radius-full);
  border: 1px solid var(--border-color);
  margin-top: 0.5rem;
}

.toggle-btn {
  padding: 0.45rem 1.25rem;
  font-size: 13px;
  font-weight: 600;
  border-radius: var(--radius-full);
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: all var(--transition-fast);
}

.toggle-btn.active {
  background-color: #ffffff;
  color: var(--text-primary);
  box-shadow: var(--shadow-sm);
}

.discount-pill {
  font-size: 10px;
  font-weight: 700;
  color: #059669;
  background-color: #d1fae5;
  padding: 0.1rem 0.4rem;
  border-radius: var(--radius-full);
}

.pricing-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 1.5rem;
  align-items: stretch;
}

.pricing-card {
  background-color: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 1.75rem 1.5rem;
  display: flex;
  flex-direction: column;
  position: relative;
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-fast);
}

.pricing-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

.popular-card {
  border-color: var(--primary);
  box-shadow: 0 0 0 2px var(--primary-focus);
}

.popular-ribbon {
  position: absolute;
  top: -12px;
  left: 50%;
  transform: translateX(-50%);
  background: linear-gradient(135deg, #1e40af, #3b82f6);
  color: #ffffff;
  font-size: 10px;
  font-weight: 700;
  padding: 0.25rem 0.75rem;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  gap: 0.35rem;
  box-shadow: var(--shadow-sm);
}

.plan-name {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

.plan-desc {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 0.35rem;
  min-height: 36px;
  line-height: 1.4;
}

.price-box {
  margin-top: 1rem;
  display: flex;
  align-items: baseline;
  gap: 0.2rem;
}

.currency-symbol {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

.price-amount {
  font-size: 32px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.03em;
}

.price-period {
  font-size: 12px;
  color: var(--text-muted);
}

.card-divider {
  height: 1px;
  background-color: var(--border-color);
  margin: 1.5rem 0;
}

.card-features {
  flex: 1;
}

.features-title {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--text-secondary);
  margin-bottom: 0.75rem;
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.feature-item {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.35;
}

.order-summary-box {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.order-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}

.total-row {
  border-top: 1px dashed var(--border-color);
  padding-top: 0.75rem;
  font-weight: 700;
  font-size: 15px;
}

.total-price {
  color: var(--primary);
}

.manual-payment-notice {
  background-color: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: var(--radius-md);
  padding: 0.75rem 1rem;
  margin-top: 0.5rem;
}

.notice-title {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 12px;
  font-weight: 700;
  color: #1e40af;
}

.notice-text {
  font-size: 11px;
  color: #1e3a8a;
  margin-top: 0.25rem;
  line-height: 1.4;
}

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}
</style>
