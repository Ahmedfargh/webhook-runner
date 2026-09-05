<template>
  <div class="admin-billing-view animate-fade-in">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('billing.adminConsoleTitle') }}</h1>
        <p class="view-subtitle">{{ t('billing.adminConsoleSubtitle') }}</p>
      </div>

      <div class="header-actions">
        <button v-if="currentTab === 'plans'" class="btn btn-primary" @click="openCreatePlanDrawer">
          <Plus :size="15" /> {{ t('billing.createPlan') }}
        </button>
        <button v-else-if="currentTab === 'invoices'" class="btn btn-primary" @click="openCreateInvoiceDrawer">
          <Plus :size="15" /> {{ t('billing.createCustomInvoice') }}
        </button>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="tab-nav">
      <button
        class="tab-btn"
        :class="{ active: currentTab === 'payments' }"
        @click="currentTab = 'payments'"
      >
        <CheckCircle2 :size="15" />
        {{ t('billing.offlinePaymentQueue') }}
        <span v-if="pendingPaymentsCount > 0" class="badge badge-warning ml-1">
          {{ pendingPaymentsCount }}
        </span>
      </button>

      <button
        class="tab-btn"
        :class="{ active: currentTab === 'invoices' }"
        @click="currentTab = 'invoices'"
      >
        <Receipt :size="15" />
        {{ t('billing.allInvoices') }}
      </button>

      <button
        class="tab-btn"
        :class="{ active: currentTab === 'subscriptions' }"
        @click="currentTab = 'subscriptions'"
      >
        <CreditCard :size="15" />
        {{ t('billing.allSubscriptions') }}
      </button>

      <button
        class="tab-btn"
        :class="{ active: currentTab === 'plans' }"
        @click="currentTab = 'plans'"
      >
        <Sparkles :size="15" />
        {{ t('billing.plansManagement') }}
      </button>
    </div>

    <!-- Tab 1: Offline Payment Approvals Queue -->
    <div v-if="currentTab === 'payments'" class="tab-content">
      <div class="filter-card">
        <div class="search-box">
          <Search :size="15" class="search-icon" />
          <input
            type="text"
            v-model="searchPaymentQuery"
            @input="fetchPayments"
            :placeholder="t('billing.searchPaymentsPlaceholder')"
            class="search-input"
          />
        </div>

        <div class="filter-actions">
          <button class="btn btn-sm btn-secondary" @click="fetchPayments">
            <RefreshCw :size="13" :class="{ 'spin-anim': loadingPayments }" /> {{ t('common.refresh') }}
          </button>
        </div>
      </div>

      <div class="zoho-table-container mt-3">
        <table class="zoho-table">
          <thead>
            <tr>
              <th>{{ t('billing.transactionReference') }}</th>
              <th>{{ t('billing.payer') }}</th>
              <th>{{ t('billing.amount') }}</th>
              <th>{{ t('common.status') }}</th>
              <th>{{ t('billing.submittedDate') }}</th>
              <th class="text-right">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loadingPayments">
              <td colspan="6" class="text-center py-8">
                <div class="loading-state">
                  <RefreshCw :size="20" class="spin-anim text-primary" />
                  <span>{{ t('common.loading') }}</span>
                </div>
              </td>
            </tr>

            <tr v-else-if="payments.length === 0">
              <td colspan="6">
                <EmptyState
                  :title="t('billing.noPendingPaymentsTitle')"
                  :description="t('billing.noPendingPaymentsDesc')"
                  :icon="CheckCircle2"
                />
              </td>
            </tr>

            <tr v-else v-for="pmt in payments" :key="pmt.id">
              <td>
                <span class="font-mono font-bold text-primary">{{ pmt.transaction_reference || pmt.transactionReference }}</span>
              </td>
              <td>
                <div>
                  <div class="font-semibold">{{ pmt.payer_name || pmt.payerName || 'N/A' }}</div>
                  <div class="text-xs text-muted">{{ pmt.payer_notes || pmt.payerNotes }}</div>
                </div>
              </td>
              <td>
                <strong>${{ Number(pmt.amount ?? 0).toFixed(2) }} {{ pmt.currency || 'USD' }}</strong>
              </td>
              <td>
                <span class="badge" :class="getPaymentStatusBadge(pmt.status)">
                  {{ pmt.status?.toUpperCase() }}
                </span>
              </td>
              <td>
                <span class="text-xs text-secondary">{{ formatDate(pmt.created_at || pmt.createdAt) }}</span>
              </td>
              <td>
                <div v-if="pmt.status === 'submitted' || pmt.status === 'STATUS_SUBMITTED'" class="action-buttons-group">
                  <button class="btn btn-xs btn-success" @click="reviewPayment(pmt, true)">
                    <Check :size="12" /> {{ t('billing.approveAndActivate') }}
                  </button>
                  <button class="btn btn-xs btn-outline text-danger" @click="reviewPayment(pmt, false)">
                    <X :size="12" /> {{ t('billing.reject') }}
                  </button>
                </div>
                <span v-else class="text-xs text-muted">
                  {{ pmt.admin_notes || pmt.adminNotes || 'Reviewed' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Tab 2: All Invoices -->
    <div v-if="currentTab === 'invoices'" class="tab-content">
      <div class="zoho-table-container">
        <table class="zoho-table">
          <thead>
            <tr>
              <th>{{ t('billing.invoiceNumber') }}</th>
              <th>{{ t('billing.userID') }}</th>
              <th>{{ t('billing.amount') }}</th>
              <th>{{ t('common.status') }}</th>
              <th>{{ t('billing.dueDate') }}</th>
              <th class="text-right">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loadingInvoices">
              <td colspan="6" class="text-center py-8">
                <div class="loading-state">
                  <RefreshCw :size="20" class="spin-anim text-primary" />
                  <span>{{ t('common.loading') }}</span>
                </div>
              </td>
            </tr>

            <tr v-else v-for="inv in allInvoices" :key="inv.id">
              <td>
                <span class="font-mono font-bold">{{ inv.invoice_number || inv.invoiceNumber }}</span>
              </td>
              <td>
                <span class="text-xs font-mono text-muted">{{ (inv.user_id || inv.userId)?.slice(0, 8) }}...</span>
              </td>
              <td>
                <strong>${{ Number(inv.total_amount ?? inv.totalAmount ?? inv.amount ?? 0).toFixed(2) }} {{ inv.currency || 'USD' }}</strong>
              </td>
              <td>
                <span class="badge" :class="getInvoiceBadge(inv.status)">
                  {{ inv.status?.toUpperCase() }}
                </span>
              </td>
              <td>
                <span class="text-xs text-secondary">{{ formatDate(inv.due_date || inv.dueDate) }}</span>
              </td>
              <td>
                <div class="action-buttons-group">
                  <button
                    v-if="inv.status === 'unpaid' || inv.status === 'STATUS_UNPAID'"
                    class="btn btn-xs btn-outline"
                    @click="markInvoiceAsPaid(inv)"
                  >
                    <Check :size="12" /> {{ t('billing.markAsPaid') }}
                  </button>
                  <button
                    v-if="inv.status === 'unpaid' || inv.status === 'STATUS_UNPAID'"
                    class="btn btn-xs btn-outline text-danger"
                    @click="voidInvoice(inv)"
                  >
                    <X :size="12" /> {{ t('billing.void') }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Tab 3: All Subscriptions -->
    <div v-if="currentTab === 'subscriptions'" class="tab-content">
      <div class="zoho-table-container">
        <table class="zoho-table">
          <thead>
            <tr>
              <th>{{ t('billing.userID') }}</th>
              <th>{{ t('billing.plan') }}</th>
              <th>{{ t('billing.cycle') }}</th>
              <th>{{ t('common.status') }}</th>
              <th>{{ t('billing.periodEnd') }}</th>
              <th class="text-right">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loadingSubs">
              <td colspan="6" class="text-center py-8">
                <div class="loading-state">
                  <RefreshCw :size="20" class="spin-anim text-primary" />
                  <span>{{ t('common.loading') }}</span>
                </div>
              </td>
            </tr>

            <tr v-else v-for="sub in allSubscriptions" :key="sub.id">
              <td>
                <span class="font-mono text-xs">{{ sub.user_id || sub.userId }}</span>
              </td>
              <td>
                <strong>{{ sub.plan?.name || 'Free' }}</strong>
              </td>
              <td>
                <span class="text-xs uppercase">{{ sub.billing_cycle || sub.billingCycle }}</span>
              </td>
              <td>
                <span class="badge" :class="getSubStatusBadge(sub.status)">
                  {{ sub.status?.toUpperCase() }}
                </span>
              </td>
              <td>
                <span class="text-xs text-secondary">{{ formatDate(sub.current_period_end || sub.currentPeriodEnd) }}</span>
              </td>
              <td>
                <div class="action-buttons-group">
                  <button class="btn btn-xs btn-outline" @click="openOverrideModal(sub)">
                    <Edit2 :size="12" /> {{ t('billing.override') }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Tab 4: Plans Management (Admin CRUD) -->
    <div v-if="currentTab === 'plans'" class="tab-content">
      <div class="filter-card">
        <div class="flex items-center gap-2">
          <span class="text-xs font-semibold text-muted">{{ t('billing.allPlans') }} ({{ availablePlans.length }})</span>
        </div>

        <div class="filter-actions">
          <button class="btn btn-sm btn-secondary" @click="fetchSubscriptions">
            <RefreshCw :size="13" :class="{ 'spin-anim': loadingSubs }" /> {{ t('common.refresh') }}
          </button>
          <button class="btn btn-sm btn-primary" @click="openCreatePlanDrawer">
            <Plus :size="13" /> {{ t('billing.createPlan') }}
          </button>
        </div>
      </div>

      <div class="zoho-table-container mt-3">
        <table class="zoho-table">
          <thead>
            <tr>
              <th>{{ t('billing.planName') }}</th>
              <th>{{ t('billing.priceMonthly') }}</th>
              <th>{{ t('billing.priceAnnually') }}</th>
              <th>{{ t('billing.maxWebhooks') }}</th>
              <th>{{ t('billing.maxEvents') }}</th>
              <th>{{ t('billing.teamSeats') }}</th>
              <th>{{ t('common.status') }}</th>
              <th class="text-right">{{ t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loadingSubs">
              <td colspan="8" class="text-center py-8">
                <div class="loading-state">
                  <RefreshCw :size="20" class="spin-anim text-primary" />
                  <span>{{ t('common.loading') }}</span>
                </div>
              </td>
            </tr>

            <tr v-else-if="availablePlans.length === 0">
              <td colspan="8">
                <EmptyState
                  title="No plans found"
                  description="Create your first subscription tier to get started."
                  :icon="Sparkles"
                />
              </td>
            </tr>

            <tr v-else v-for="p in availablePlans" :key="p.id">
              <td>
                <div>
                  <div class="font-bold flex items-center gap-2">
                    {{ p.name }}
                    <span v-if="p.is_popular || p.isPopular" class="badge badge-warning text-2xs">
                      <Sparkles :size="10" /> {{ t('billing.mostPopular') }}
                    </span>
                  </div>
                  <div class="text-xs font-mono text-muted">{{ p.code || 'default' }}</div>
                </div>
              </td>
              <td>
                <strong>${{ Number(p.price_monthly ?? p.priceMonthly ?? 0).toFixed(2) }}</strong>
              </td>
              <td>
                <strong>${{ Number(p.price_annually ?? p.priceAnnually ?? 0).toFixed(2) }}</strong>
              </td>
              <td>
                <span class="badge badge-outline">
                  {{ (p.max_webhooks ?? p.maxWebhooks) === -1 ? '∞' : (p.max_webhooks ?? p.maxWebhooks ?? 100) }}
                </span>
              </td>
              <td>
                <span class="badge badge-outline">
                  {{ (p.max_events_per_month ?? p.maxEventsPerMonth) === -1 ? '∞' : (p.max_events_per_month ?? p.maxEventsPerMonth)?.toLocaleString() }}
                </span>
              </td>
              <td>
                <span class="badge badge-outline">
                  {{ (p.max_team_members ?? p.maxTeamMembers) === -1 ? '∞' : (p.max_team_members ?? p.maxTeamMembers ?? 1) }}
                </span>
              </td>
              <td>
                <span class="badge" :class="(p.is_active ?? p.isActive ?? true) ? 'badge-success' : 'badge-secondary'">
                  {{ (p.is_active ?? p.isActive ?? true) ? t('common.active') : 'Inactive' }}
                </span>
              </td>
              <td>
                <div class="action-buttons-group">
                  <button class="btn btn-xs btn-outline" @click="openEditPlanDrawer(p)">
                    <Edit2 :size="12" /> {{ t('common.edit') }}
                  </button>
                  <button class="btn btn-xs btn-outline text-danger" @click="openDeletePlanModal(p)">
                    <Trash2 :size="12" /> {{ t('common.delete') }}
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create Custom Invoice Drawer -->
    <Drawer
      :isOpen="isInvoiceDrawerOpen"
      :title="t('billing.createCustomInvoice')"
      :subtitle="t('billing.createCustomInvoiceDesc')"
      @close="isInvoiceDrawerOpen = false"
      width="540px"
    >
      <form @submit.prevent="createCustomInvoice">
        <div class="form-group">
          <label class="form-label">{{ t('billing.targetUserID') }} *</label>
          <input
            v-model="invoiceForm.user_id"
            type="text"
            required
            class="form-control font-mono"
            placeholder="User UUID"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('billing.itemDescription') }} *</label>
          <input
            v-model="invoiceForm.description"
            type="text"
            required
            class="form-control"
            placeholder="e.g. Enterprise Setup Fee or Pro Plan Annual"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('billing.amount') }} ($ USD) *</label>
          <input
            v-model.number="invoiceForm.amount"
            type="number"
            step="0.01"
            min="1"
            required
            class="form-control"
            placeholder="49.00"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('billing.dueDate') }}</label>
          <input
            v-model="invoiceForm.due_date"
            type="date"
            class="form-control"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('billing.notes') }}</label>
          <textarea
            v-model="invoiceForm.notes"
            rows="2"
            class="form-control"
            placeholder="Optional custom billing notes..."
          ></textarea>
        </div>
      </form>

      <template #footer>
        <button class="btn btn-secondary" @click="isInvoiceDrawerOpen = false" :disabled="savingInvoice">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="createCustomInvoice" :disabled="savingInvoice">
          <RefreshCw v-if="savingInvoice" :size="14" class="spin-anim" />
          {{ t('billing.generateInvoice') }}
        </button>
      </template>
    </Drawer>

    <!-- Plan CRUD Drawer (Create / Edit) -->
    <Drawer
      :isOpen="isPlanDrawerOpen"
      :title="editingPlanId ? t('billing.editPlan') : t('billing.createPlan')"
      :subtitle="editingPlanId ? 'Update pricing, limits, and features' : 'Define new tier pricing and feature quotas'"
      @close="isPlanDrawerOpen = false"
      width="580px"
    >
      <form @submit.prevent="savePlan">
        <div class="form-row">
          <div class="form-group flex-1">
            <label class="form-label">{{ t('billing.planName') }} *</label>
            <input
              v-model="planForm.name"
              type="text"
              required
              class="form-control"
              placeholder="e.g. Pro Developer"
            />
          </div>
          <div class="form-group flex-1">
            <label class="form-label">{{ t('billing.planCode') }}</label>
            <input
              v-model="planForm.code"
              type="text"
              class="form-control font-mono"
              placeholder="e.g. pro-developer"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('billing.planDescription') }}</label>
          <textarea
            v-model="planForm.description"
            rows="2"
            class="form-control"
            placeholder="Brief tier overview..."
          ></textarea>
        </div>

        <div class="form-row">
          <div class="form-group flex-1">
            <label class="form-label">{{ t('billing.priceMonthly') }} ($) *</label>
            <input
              v-model.number="planForm.price_monthly"
              type="number"
              step="0.01"
              min="0"
              required
              class="form-control"
              placeholder="29.00"
            />
          </div>
          <div class="form-group flex-1">
            <label class="form-label">{{ t('billing.priceAnnually') }} ($) *</label>
            <input
              v-model.number="planForm.price_annually"
              type="number"
              step="0.01"
              min="0"
              required
              class="form-control"
              placeholder="290.00"
            />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group flex-1">
            <label class="form-label">{{ t('billing.maxWebhooks') }}</label>
            <input
              v-model.number="planForm.max_webhooks"
              type="number"
              class="form-control"
              placeholder="100 (-1 = inf)"
            />
          </div>
          <div class="form-group flex-1">
            <label class="form-label">{{ t('billing.maxEvents') }}</label>
            <input
              v-model.number="planForm.max_events_per_month"
              type="number"
              class="form-control"
              placeholder="1000000"
            />
          </div>
          <div class="form-group flex-1">
            <label class="form-label">{{ t('billing.maxTeamMembers') }}</label>
            <input
              v-model.number="planForm.max_team_members"
              type="number"
              class="form-control"
              placeholder="5"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('billing.featuresList') }}</label>
          <textarea
            v-model="planFeaturesRaw"
            rows="4"
            class="form-control font-mono text-xs"
            placeholder="Dedicated webhook dispatchers&#10;99.99% SLA Guarantee&#10;Priority 24/7 Slack Support"
          ></textarea>
        </div>

        <div class="form-row items-center gap-6 mt-3">
          <label class="checkbox-label flex items-center gap-2 cursor-pointer">
            <input type="checkbox" v-model="planForm.is_popular" class="form-checkbox" />
            <span class="text-xs font-semibold">{{ t('billing.isPopular') }}</span>
          </label>

          <label class="checkbox-label flex items-center gap-2 cursor-pointer">
            <input type="checkbox" v-model="planForm.is_active" class="form-checkbox" />
            <span class="text-xs font-semibold">{{ t('billing.isActive') }}</span>
          </label>
        </div>
      </form>

      <template #footer>
        <button class="btn btn-secondary" @click="isPlanDrawerOpen = false" :disabled="savingPlan">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="savePlan" :disabled="savingPlan">
          <RefreshCw v-if="savingPlan" :size="14" class="spin-anim" />
          {{ t('common.save') }}
        </button>
      </template>
    </Drawer>

    <!-- Delete Plan Confirmation Modal -->
    <Modal
      :isOpen="isDeletePlanModalOpen"
      :title="t('billing.deletePlan')"
      @close="isDeletePlanModalOpen = false"
    >
      <p class="modal-text">
        {{ t('billing.deletePlanConfirm', { name: planToDelete?.name }) }}
      </p>

      <template #footer>
        <button class="btn btn-secondary" @click="isDeletePlanModalOpen = false" :disabled="deletingPlan">{{ t('common.cancel') }}</button>
        <button class="btn btn-danger" @click="executeDeletePlan" :disabled="deletingPlan">
          <RefreshCw v-if="deletingPlan" :size="14" class="spin-anim" />
          {{ t('common.delete') }}
        </button>
      </template>
    </Modal>

    <!-- Admin Override Subscription Modal -->
    <Modal
      :isOpen="isOverrideModalOpen"
      :title="t('billing.overrideSubscriptionTitle')"
      @close="isOverrideModalOpen = false"
    >
      <div v-if="selectedSub" class="form-group">
        <div class="form-group">
          <label class="form-label">{{ t('billing.userID') }}</label>
          <input type="text" disabled :value="selectedSub.user_id" class="form-control font-mono text-xs" />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('billing.selectPlan') }}</label>
          <select v-model="overrideForm.plan_id" class="form-control">
            <option v-for="p in availablePlans" :key="p.id" :value="p.id">
              {{ p.name }} (${{ p.price_monthly ?? p.priceMonthly }}/mo)
            </option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('common.status') }}</label>
          <select v-model="overrideForm.status" class="form-control">
            <option value="active">ACTIVE</option>
            <option value="pending_manual_payment">PENDING MANUAL PAYMENT</option>
            <option value="trialing">TRIALING</option>
            <option value="canceled">CANCELED</option>
            <option value="expired">EXPIRED</option>
          </select>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('billing.adminNotes') }}</label>
          <textarea
            v-model="overrideForm.admin_notes"
            rows="2"
            class="form-control"
            placeholder="Reason for override..."
          ></textarea>
        </div>
      </div>

      <template #footer>
        <button class="btn btn-secondary" @click="isOverrideModalOpen = false">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="executeOverride">
          {{ t('billing.applyOverride') }}
        </button>
      </template>
    </Modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from '../locales'
import { manualPaymentService } from '../services/manualPaymentService'
import { invoiceService } from '../services/invoiceService'
import { subscriptionService } from '../services/subscriptionService'
import { planService } from '../services/planService'
import { useToastStore } from '../stores/toast'
import Drawer from '../components/common/Drawer.vue'
import Modal from '../components/common/Modal.vue'
import EmptyState from '../components/common/EmptyState.vue'
import {
  Receipt,
  Plus,
  CheckCircle2,
  CreditCard,
  Search,
  RefreshCw,
  Check,
  X,
  Edit2,
  Sparkles,
  Trash2,
} from 'lucide-vue-next'

const { t } = useI18n()
const toastStore = useToastStore()

const currentTab = ref('payments') // payments | invoices | subscriptions | plans

// State
const payments = ref([])
const allInvoices = ref([])
const allSubscriptions = ref([])
const availablePlans = ref([])

const loadingPayments = ref(false)
const loadingInvoices = ref(false)
const loadingSubs = ref(false)
const savingInvoice = ref(false)
const savingPlan = ref(false)
const deletingPlan = ref(false)
const searchPaymentQuery = ref('')

const pendingPaymentsCount = computed(() => {
  return payments.value.filter((p) => p.status === 'submitted').length
})

// Invoice Drawer
const isInvoiceDrawerOpen = ref(false)
const invoiceForm = reactive({
  user_id: '',
  description: '',
  amount: 49.00,
  due_date: '',
  notes: '',
})

// Override Modal
const isOverrideModalOpen = ref(false)
const selectedSub = ref(null)
const overrideForm = reactive({
  plan_id: '',
  status: 'active',
  admin_notes: '',
})

// Plan CRUD Drawer
const isPlanDrawerOpen = ref(false)
const editingPlanId = ref(null)
const planFeaturesRaw = ref('')
const planForm = reactive({
  name: '',
  code: '',
  description: '',
  price_monthly: 29,
  price_annually: 290,
  currency: 'USD',
  max_webhooks: 100,
  max_events_per_month: 100000,
  max_team_members: 1,
  is_active: true,
  is_popular: false,
  tier_level: 1,
})

// Delete Plan Modal
const isDeletePlanModalOpen = ref(false)
const planToDelete = ref(null)

async function fetchPayments() {
  loadingPayments.value = true
  try {
    const res = await manualPaymentService.listAllPayments({
      search: searchPaymentQuery.value,
      page: 1,
      page_size: 50,
    })
    payments.value = res.data || []
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to load payments queue')
  } finally {
    loadingPayments.value = false
  }
}

async function fetchInvoices() {
  loadingInvoices.value = true
  try {
    const res = await invoiceService.listAllInvoices({ page: 1, page_size: 50 })
    allInvoices.value = res.data || []
  } catch (err) {
    // Non-blocking
  } finally {
    loadingInvoices.value = false
  }
}

async function fetchSubscriptions() {
  loadingSubs.value = true
  try {
    const [subRes, planRes] = await Promise.all([
      subscriptionService.listAllSubscriptions({ page: 1, page_size: 50 }),
      planService.listPlans(true),
    ])
    allSubscriptions.value = subRes.data || []
    availablePlans.value = planRes.data || []
  } catch (err) {
    // Non-blocking
  } finally {
    loadingSubs.value = false
  }
}

async function reviewPayment(pmt, approve) {
  try {
    await manualPaymentService.reviewPayment(pmt.id, {
      approve,
      admin_notes: approve ? 'Payment verified in bank account' : 'Payment rejected by admin',
    })
    toastStore.success(approve ? 'Payment approved & subscription activated!' : 'Payment rejected')
    fetchPayments()
    fetchInvoices()
    fetchSubscriptions()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to review payment')
  }
}

async function markInvoiceAsPaid(inv) {
  try {
    await invoiceService.markInvoicePaid(inv.id, {
      payment_reference: 'ADMIN-MANUAL-OVERRIDE',
      payment_method: 'manual_admin',
      admin_notes: 'Marked paid by system administrator',
    })
    toastStore.success('Invoice marked as paid!')
    fetchInvoices()
    fetchSubscriptions()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to mark invoice as paid')
  }
}

async function voidInvoice(inv) {
  try {
    await invoiceService.voidInvoice(inv.id, 'Voided by administrator')
    toastStore.success('Invoice voided')
    fetchInvoices()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to void invoice')
  }
}

function openCreateInvoiceDrawer() {
  invoiceForm.user_id = ''
  invoiceForm.description = ''
  invoiceForm.amount = 49.00
  invoiceForm.due_date = ''
  invoiceForm.notes = ''
  isInvoiceDrawerOpen.value = true
}

async function createCustomInvoice() {
  if (!invoiceForm.user_id || !invoiceForm.description || invoiceForm.amount <= 0) {
    toastStore.warning('Please fill in user ID, description, and amount.')
    return
  }

  savingInvoice.value = true
  try {
    await invoiceService.createManualInvoice({
      user_id: invoiceForm.user_id,
      amount: invoiceForm.amount,
      currency: 'USD',
      due_date: invoiceForm.due_date ? new Date(invoiceForm.due_date).toISOString() : '',
      notes: invoiceForm.notes,
      items: [
        {
          description: invoiceForm.description,
          quantity: 1,
          unit_price: invoiceForm.amount,
        },
      ],
    })
    toastStore.success('Custom invoice generated successfully')
    isInvoiceDrawerOpen.value = false
    fetchInvoices()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to create invoice')
  } finally {
    savingInvoice.value = false
  }
}

function openOverrideModal(sub) {
  selectedSub.value = sub
  overrideForm.plan_id = sub.plan_id || ''
  overrideForm.status = sub.status || 'active'
  overrideForm.admin_notes = ''
  isOverrideModalOpen.value = true
}

async function executeOverride() {
  try {
    await subscriptionService.adminOverrideSubscription({
      user_id: selectedSub.value.user_id,
      plan_id: overrideForm.plan_id,
      status: overrideForm.status,
      admin_notes: overrideForm.admin_notes,
    })
    toastStore.success('Subscription updated by admin override')
    isOverrideModalOpen.value = false
    fetchSubscriptions()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to override subscription')
  }
}

// Plan Management CRUD Functions
function openCreatePlanDrawer() {
  editingPlanId.value = null
  planForm.name = ''
  planForm.code = ''
  planForm.description = ''
  planForm.price_monthly = 29
  planForm.price_annually = 290
  planForm.currency = 'USD'
  planForm.max_webhooks = 100
  planForm.max_events_per_month = 500000
  planForm.max_team_members = 3
  planForm.is_active = true
  planForm.is_popular = false
  planForm.tier_level = 2
  planFeaturesRaw.value = '100 Webhook Endpoints\n500,000 Monthly Events\n3 Team Seats\nFast Retry Queue\nStandard Support'
  isPlanDrawerOpen.value = true
}

function openEditPlanDrawer(plan) {
  editingPlanId.value = plan.id
  planForm.name = plan.name || ''
  planForm.code = plan.code || ''
  planForm.description = plan.description || ''
  planForm.price_monthly = Number(plan.price_monthly ?? plan.priceMonthly ?? 0)
  planForm.price_annually = Number(plan.price_annually ?? plan.priceAnnually ?? 0)
  planForm.currency = plan.currency || 'USD'
  planForm.max_webhooks = Number(plan.max_webhooks ?? plan.maxWebhooks ?? 100)
  planForm.max_events_per_month = Number(plan.max_events_per_month ?? plan.maxEventsPerMonth ?? 100000)
  planForm.max_team_members = Number(plan.max_team_members ?? plan.maxTeamMembers ?? 1)
  planForm.is_active = plan.is_active ?? plan.isActive ?? true
  planForm.is_popular = plan.is_popular ?? plan.isPopular ?? false
  planForm.tier_level = Number(plan.tier_level ?? plan.tierLevel ?? 1)

  const feats = plan.features || []
  planFeaturesRaw.value = feats.join('\n')
  isPlanDrawerOpen.value = true
}

async function savePlan() {
  if (!planForm.name) {
    toastStore.warning('Please specify plan name.')
    return
  }

  savingPlan.value = true
  const features = planFeaturesRaw.value
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean)

  const payload = {
    name: planForm.name,
    code: planForm.code || planForm.name.toLowerCase().replace(/[^a-z0-9]+/g, '-'),
    description: planForm.description,
    price_monthly: Number(planForm.price_monthly),
    price_annually: Number(planForm.price_annually),
    currency: planForm.currency || 'USD',
    max_webhooks: Number(planForm.max_webhooks),
    max_events_per_month: Number(planForm.max_events_per_month),
    max_team_members: Number(planForm.max_team_members),
    features,
    is_active: planForm.is_active,
    is_popular: planForm.is_popular,
    tier_level: Number(planForm.tier_level),
  }

  try {
    if (editingPlanId.value) {
      await planService.updatePlan(editingPlanId.value, payload)
      toastStore.success('Plan updated successfully')
    } else {
      await planService.createPlan(payload)
      toastStore.success('Plan created successfully')
    }
    isPlanDrawerOpen.value = false
    fetchSubscriptions()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to save plan')
  } finally {
    savingPlan.value = false
  }
}

function openDeletePlanModal(plan) {
  planToDelete.value = plan
  isDeletePlanModalOpen.value = true
}

async function executeDeletePlan() {
  if (!planToDelete.value) return
  deletingPlan.value = true
  try {
    await planService.deletePlan(planToDelete.value.id)
    toastStore.success('Plan deleted')
    isDeletePlanModalOpen.value = false
    fetchSubscriptions()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to delete plan')
  } finally {
    deletingPlan.value = false
  }
}

function formatDate(ts) {
  if (!ts) return 'N/A'
  return new Date(ts).toLocaleDateString()
}

function getPaymentStatusBadge(status) {
  switch (status) {
    case 'approved':
      return 'badge-success'
    case 'submitted':
      return 'badge-warning'
    case 'rejected':
      return 'badge-danger'
    default:
      return 'badge-secondary'
  }
}

function getInvoiceBadge(status) {
  switch (status) {
    case 'paid':
      return 'badge-success'
    case 'unpaid':
      return 'badge-warning'
    case 'overdue':
      return 'badge-danger'
    case 'void':
      return 'badge-secondary'
    default:
      return 'badge-outline'
  }
}

function getSubStatusBadge(status) {
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

onMounted(() => {
  fetchPayments()
  fetchInvoices()
  fetchSubscriptions()
})
</script>

<style scoped>
.admin-billing-view {
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

.tab-nav {
  display: flex;
  gap: 0.5rem;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 0.5rem;
  overflow-x: auto;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  background: transparent;
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.tab-btn:hover {
  background-color: var(--bg-card-muted);
  color: var(--text-primary);
}

.tab-btn.active {
  background-color: var(--primary-light);
  color: var(--primary);
}

.filter-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background-color: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  flex-wrap: wrap;
}

.search-box {
  position: relative;
  display: flex;
  align-items: center;
  flex: 1;
  max-width: 340px;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  color: var(--text-muted);
}

.search-input {
  width: 100%;
  padding: 0.45rem 1rem 0.45rem 2.25rem;
  font-size: 13px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  outline: none;
  background-color: var(--bg-input);
  color: var(--text-primary);
}

.action-buttons-group {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.35rem;
}

.form-row {
  display: flex;
  gap: 1rem;
  margin-bottom: 0.5rem;
}

@media (max-width: 600px) {
  .form-row {
    flex-direction: column;
    gap: 0;
  }
}

.text-2xs {
  font-size: 9px;
  padding: 0.1rem 0.3rem;
}

.btn-success {
  background-color: #059669;
  color: #ffffff;
}
.btn-success:hover {
  background-color: #047857;
}

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}
</style>
