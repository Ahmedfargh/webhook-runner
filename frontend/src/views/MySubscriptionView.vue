<template>
  <div class="my-subscription-view animate-fade-in">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('billing.mySubscriptionTitle') }}</h1>
        <p class="view-subtitle">{{ t('billing.mySubscriptionSubtitle') }}</p>
      </div>

      <div class="header-actions">
        <router-link to="/plans" class="btn btn-primary">
          <Sparkles :size="15" /> {{ t('billing.changePlan') }}
        </router-link>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12">
      <RefreshCw :size="24" class="spin-anim text-primary" />
      <p class="mt-2 text-muted">{{ t('common.loading') }}</p>
    </div>

    <div v-else class="subscription-content-grid">
      <!-- Active Subscription Overview Card -->
      <div class="card current-plan-overview">
        <div class="card-header">
          <div class="card-header-left">
            <CreditCard :size="18" class="text-primary" />
            <h3 class="card-title">{{ t('billing.activePlanOverview') }}</h3>
          </div>
          <span
            class="badge"
            :class="getStatusBadgeClass(subscription?.status)"
          >
            {{ formatStatus(subscription?.status) }}
          </span>
        </div>

        <div class="plan-hero-box">
          <div class="plan-details">
            <div class="plan-tier-label">{{ subscription?.plan?.name || 'Free Developer' }}</div>
            <div class="plan-price-tag">
              <span class="price-val">
                ${{ (subscription?.billing_cycle === 'annually' || subscription?.billingCycle === 'annually')
                    ? (subscription?.plan?.price_annually ?? subscription?.plan?.priceAnnually ?? 0)
                    : (subscription?.plan?.price_monthly ?? subscription?.plan?.priceMonthly ?? 0) }}
              </span>
              <span class="price-sub">/ {{ subscription?.billing_cycle || subscription?.billingCycle || 'monthly' }}</span>
            </div>
          </div>

          <div class="plan-dates-meta">
            <div class="meta-date-item">
              <span class="text-muted">{{ t('billing.periodStart') }}:</span>
              <span>{{ formatDate(subscription?.current_period_start || subscription?.currentPeriodStart) }}</span>
            </div>
            <div class="meta-date-item">
              <span class="text-muted">{{ t('billing.renewalDate') }}:</span>
              <strong>{{ formatDate(subscription?.current_period_end || subscription?.currentPeriodEnd) }}</strong>
            </div>
          </div>
        </div>

        <!-- Pending Manual Payment Alert -->
        <div v-if="subscription?.status === 'pending_manual_payment' || subscription?.status === 'STATUS_PENDING_MANUAL_PAYMENT'" class="pending-payment-banner">
          <div class="banner-icon"><AlertTriangle :size="18" /></div>
          <div class="banner-text">
            <h4>{{ t('billing.pendingPaymentHeader') }}</h4>
            <p>{{ t('billing.pendingPaymentBody') }}</p>
            <router-link to="/invoices" class="btn btn-sm btn-warning mt-2">
              <Receipt :size="13" /> {{ t('billing.viewPendingInvoices') }}
            </router-link>
          </div>
        </div>

        <!-- Quota & Usage Progress Bars -->
        <div class="quota-section">
          <h4 class="quota-title">{{ t('billing.usageAndLimits') }}</h4>
          <div class="quota-grid">
            <div class="quota-card">
              <div class="quota-header">
                <span class="quota-name">{{ t('billing.webhookEndpoints') }}</span>
                <span class="quota-numbers">
                  1 / {{ (subscription?.plan?.max_webhooks ?? subscription?.plan?.maxWebhooks) === -1 ? '∞' : (subscription?.plan?.max_webhooks ?? subscription?.plan?.maxWebhooks ?? 100) }}
                </span>
              </div>
              <div class="progress-track">
                <div class="progress-fill" style="width: 20%;"></div>
              </div>
            </div>

            <div class="quota-card">
              <div class="quota-header">
                <span class="quota-name">{{ t('billing.monthlyEvents') }}</span>
                <span class="quota-numbers">
                  420 / {{ (subscription?.plan?.max_events_per_month ?? subscription?.plan?.maxEventsPerMonth)?.toLocaleString() || '10,000' }}
                </span>
              </div>
              <div class="progress-track">
                <div class="progress-fill" style="width: 5%;"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Action Footer -->
        <div class="card-footer-actions">
          <button
            v-if="(subscription?.status === 'active' || subscription?.status === 'STATUS_ACTIVE') && !subscription?.cancel_at_period_end && !subscription?.cancelAtPeriodEnd"
            class="btn btn-sm btn-outline text-danger"
            @click="openCancelModal"
          >
            {{ t('billing.cancelSubscription') }}
          </button>
          <span v-else-if="subscription?.cancel_at_period_end || subscription?.cancelAtPeriodEnd" class="text-xs text-danger font-semibold">
            {{ t('billing.cancelsAtPeriodEnd') }}
          </span>
        </div>
      </div>

      <!-- Offline Bank Details Box -->
      <div class="card offline-payment-info-card">
        <div class="card-header">
          <div class="card-header-left">
            <Landmark :size="18" class="text-primary" />
            <h3 class="card-title">{{ t('billing.bankWireDetails') }}</h3>
          </div>
          <span class="badge badge-outline">{{ t('billing.manualOffline') }}</span>
        </div>

        <div class="bank-details-body">
          <p class="bank-desc">{{ t('billing.bankWireDescription') }}</p>

          <div class="bank-item-row">
            <span class="bank-lbl">{{ t('billing.bankName') }}:</span>
            <strong class="bank-val">Standard International Bank</strong>
          </div>
          <div class="bank-item-row">
            <span class="bank-lbl">{{ t('billing.accountName') }}:</span>
            <strong class="bank-val">Webhook Cloud Services Inc.</strong>
          </div>
          <div class="bank-item-row">
            <span class="bank-lbl">{{ t('billing.iban') }}:</span>
            <code class="bank-code">US92 SIBK 8820 9931 4402 1198</code>
          </div>
          <div class="bank-item-row">
            <span class="bank-lbl">{{ t('billing.swift') }}:</span>
            <code class="bank-code">SIBKUS33XXX</code>
          </div>

          <div class="help-box mt-3">
            <HelpCircle :size="14" />
            <span>{{ t('billing.bankHelpTip') }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Cancel Modal -->
    <Modal
      :isOpen="isCancelModalOpen"
      :title="t('billing.cancelConfirmTitle')"
      @close="isCancelModalOpen = false"
    >
      <p class="modal-text">
        {{ t('billing.cancelConfirmBody') }}
      </p>

      <div class="form-group mt-3">
        <label class="form-label">{{ t('billing.cancelReasonLabel') }}</label>
        <textarea
          v-model="cancelReason"
          rows="2"
          class="form-control"
          placeholder="e.g. Downsizing or project finished"
        ></textarea>
      </div>

      <template #footer>
        <button class="btn btn-secondary" @click="isCancelModalOpen = false" :disabled="canceling">{{ t('common.cancel') }}</button>
        <button class="btn btn-danger" @click="executeCancel" :disabled="canceling">
          <RefreshCw v-if="canceling" :size="14" class="spin-anim" />
          {{ t('billing.confirmCancellation') }}
        </button>
      </template>
    </Modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from '../locales'
import { subscriptionService } from '../services/subscriptionService'
import { useToastStore } from '../stores/toast'
import Modal from '../components/common/Modal.vue'
import {
  Sparkles,
  CreditCard,
  RefreshCw,
  AlertTriangle,
  Receipt,
  Landmark,
  HelpCircle,
} from 'lucide-vue-next'

const { t } = useI18n()
const toastStore = useToastStore()

const subscription = ref(null)
const loading = ref(false)
const canceling = ref(false)
const isCancelModalOpen = ref(false)
const cancelReason = ref('')

async function fetchSubscription() {
  loading.value = true
  try {
    const res = await subscriptionService.getCurrentSubscription()
    subscription.value = res.data || null
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to load subscription')
  } finally {
    loading.value = false
  }
}

function formatDate(ts) {
  if (!ts) return 'N/A'
  return new Date(ts).toLocaleDateString()
}

function getStatusBadgeClass(status) {
  switch (status) {
    case 'active':
      return 'badge-success'
    case 'pending_manual_payment':
      return 'badge-warning'
    case 'canceled':
      return 'badge-danger'
    default:
      return 'badge-secondary'
  }
}

function formatStatus(status) {
  if (!status) return 'Inactive'
  return status.replace(/_/g, ' ').toUpperCase()
}

function openCancelModal() {
  cancelReason.value = ''
  isCancelModalOpen.value = true
}

async function executeCancel() {
  canceling.value = true
  try {
    await subscriptionService.cancelSubscription({
      reason: cancelReason.value,
      immediately: false,
    })
    toastStore.success('Subscription scheduled for cancellation at period end')
    isCancelModalOpen.value = false
    fetchSubscription()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to cancel subscription')
  } finally {
    canceling.value = false
  }
}

onMounted(() => {
  fetchSubscription()
})
</script>

<style scoped>
.my-subscription-view {
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

.subscription-content-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 1.5rem;
}

@media (max-width: 900px) {
  .subscription-content-grid {
    grid-template-columns: 1fr;
  }
}

.current-plan-overview {
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

.plan-hero-box {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: var(--bg-card-muted);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 1.25rem 1.5rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.plan-tier-label {
  font-size: 20px;
  font-weight: 800;
  color: var(--text-primary);
}

.plan-price-tag {
  display: flex;
  align-items: baseline;
  gap: 0.25rem;
  margin-top: 0.25rem;
}

.price-val {
  font-size: 24px;
  font-weight: 800;
  color: var(--primary);
}

.price-sub {
  font-size: 12px;
  color: var(--text-muted);
}

.plan-dates-meta {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  font-size: 13px;
}

.meta-date-item {
  display: flex;
  gap: 0.5rem;
}

.pending-payment-banner {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  background-color: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: var(--radius-md);
  padding: 1rem;
  color: #92400e;
}

.pending-payment-banner h4 {
  font-size: 13px;
  font-weight: 700;
  margin: 0;
}

.pending-payment-banner p {
  font-size: 12px;
  margin: 0.25rem 0 0 0;
  line-height: 1.4;
}

.quota-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.quota-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.quota-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

@media (max-width: 600px) {
  .quota-grid {
    grid-template-columns: 1fr;
  }
}

.quota-card {
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 0.75rem 1rem;
  background-color: #ffffff;
}

.quota-header {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 0.5rem;
}

.progress-track {
  height: 6px;
  background-color: var(--border-color);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background-color: var(--primary);
  border-radius: var(--radius-full);
}

.card-footer-actions {
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid var(--border-color);
  padding-top: 1rem;
}

.offline-payment-info-card {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.bank-desc {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.bank-item-row {
  display: flex;
  flex-direction: column;
  margin-top: 0.5rem;
  font-size: 12px;
}

.bank-lbl {
  color: var(--text-muted);
  font-size: 11px;
}

.bank-val {
  color: var(--text-primary);
  margin-top: 0.1rem;
}

.bank-code {
  background-color: var(--bg-card-muted);
  padding: 0.2rem 0.4rem;
  border-radius: var(--radius-xs);
  font-family: monospace;
  color: var(--primary);
  font-weight: 700;
  margin-top: 0.1rem;
}

.help-box {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 11px;
  color: var(--text-muted);
  background-color: var(--bg-card-muted);
  padding: 0.5rem;
  border-radius: var(--radius-sm);
}
</style>
