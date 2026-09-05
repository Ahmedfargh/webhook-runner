<template>
  <div class="invoices-view animate-fade-in">
    <!-- Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('billing.invoicesTitle') }}</h1>
        <p class="view-subtitle">{{ t('billing.invoicesSubtitle') }}</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-sm btn-secondary" @click="fetchInvoices">
          <RefreshCw :size="13" :class="{ 'spin-anim': loading }" /> {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <!-- Filter Bar -->
    <div class="filter-card">
      <div class="search-box">
        <Search :size="15" class="search-icon" />
        <input
          type="text"
          v-model="searchQuery"
          @input="handleSearch"
          :placeholder="t('billing.searchInvoicesPlaceholder')"
          class="search-input"
        />
        <button v-if="searchQuery" class="clear-search" @click="clearSearch">
          <X :size="13" />
        </button>
      </div>

      <div class="status-filters">
        <button
          class="filter-pill"
          :class="{ active: statusFilter === '' }"
          @click="setStatusFilter('')"
        >
          {{ t('common.all') }}
        </button>
        <button
          class="filter-pill"
          :class="{ active: statusFilter === 'unpaid' }"
          @click="setStatusFilter('unpaid')"
        >
          {{ t('billing.unpaid') }}
        </button>
        <button
          class="filter-pill"
          :class="{ active: statusFilter === 'paid' }"
          @click="setStatusFilter('paid')"
        >
          {{ t('billing.paid') }}
        </button>
      </div>
    </div>

    <!-- Table Container -->
    <div class="zoho-table-container">
      <table class="zoho-table">
        <thead>
          <tr>
            <th>{{ t('billing.invoiceNumber') }}</th>
            <th>{{ t('billing.amount') }}</th>
            <th>{{ t('common.status') }}</th>
            <th>{{ t('billing.dueDate') }}</th>
            <th>{{ t('billing.paidDate') }}</th>
            <th class="text-right">{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="6" class="text-center py-8">
              <div class="loading-state">
                <RefreshCw :size="20" class="spin-anim text-primary" />
                <span>{{ t('common.loading') }}</span>
              </div>
            </td>
          </tr>

          <tr v-else-if="invoices.length === 0">
            <td colspan="6">
              <EmptyState
                :title="t('billing.noInvoicesTitle')"
                :description="t('billing.noInvoicesDesc')"
                :icon="Receipt"
              />
            </td>
          </tr>

          <tr v-else v-for="inv in invoices" :key="inv.id">
            <td>
              <div class="invoice-num-cell">
                <FileText :size="14" class="text-primary" />
                <span class="font-mono font-semibold">{{ inv.invoice_number || inv.invoiceNumber }}</span>
              </div>
            </td>
            <td>
              <strong class="text-primary">${{ Number(inv.total_amount ?? inv.totalAmount ?? inv.amount ?? 0).toFixed(2) }} {{ inv.currency || 'USD' }}</strong>
            </td>
            <td>
              <span class="badge" :class="getInvoiceBadge(inv.status)">
                {{ formatStatus(inv.status) }}
              </span>
            </td>
            <td>
              <span class="text-xs text-secondary">{{ formatDate(inv.due_date || inv.dueDate) }}</span>
            </td>
            <td>
              <span class="text-xs text-secondary">{{ (inv.paid_at || inv.paidAt) ? formatDate(inv.paid_at || inv.paidAt) : '-' }}</span>
            </td>
            <td>
              <div class="action-buttons-group">
                <button class="btn btn-xs btn-outline" @click="openInvoiceDetails(inv)">
                  <Eye :size="12" /> {{ t('billing.view') }}
                </button>
                <button
                  v-if="inv.status === 'unpaid' || inv.status === 'STATUS_UNPAID'"
                  class="btn btn-xs btn-primary"
                  @click="openPaymentModal(inv)"
                >
                  <Send :size="12" /> {{ t('billing.submitWireReference') }}
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <Pagination
        :page="page"
        :pageSize="pageSize"
        :total="totalInvoices"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <!-- Zoho Invoice Printable Modal -->
    <Modal
      :isOpen="isDetailModalOpen"
      :title="t('billing.invoiceDetailsTitle')"
      @close="isDetailModalOpen = false"
      width="640px"
    >
      <div v-if="selectedInvoice" class="invoice-printable-sheet">
        <div class="invoice-sheet-header">
          <div>
            <h2 class="invoice-brand">Webhook Platform</h2>
            <p class="text-xs text-muted">Cloud Webhooks & Invoicing Services</p>
          </div>
          <div class="text-right">
            <div class="invoice-tag">{{ selectedInvoice.invoice_number || selectedInvoice.invoiceNumber }}</div>
            <span class="badge mt-1" :class="getInvoiceBadge(selectedInvoice.status)">
              {{ formatStatus(selectedInvoice.status) }}
            </span>
          </div>
        </div>

        <div class="invoice-meta-grid">
          <div>
            <span class="text-xs text-muted">{{ t('billing.issueDate') }}:</span>
            <div class="text-xs font-semibold">{{ formatDate(selectedInvoice.created_at || selectedInvoice.createdAt) }}</div>
          </div>
          <div>
            <span class="text-xs text-muted">{{ t('billing.dueDate') }}:</span>
            <div class="text-xs font-semibold">{{ formatDate(selectedInvoice.due_date || selectedInvoice.dueDate) }}</div>
          </div>
          <div>
            <span class="text-xs text-muted">{{ t('billing.paymentMethod') }}:</span>
            <div class="text-xs font-semibold">{{ selectedInvoice.payment_method || selectedInvoice.paymentMethod || 'Offline Bank Transfer' }}</div>
          </div>
        </div>

        <!-- Line Items -->
        <table class="line-items-table mt-4">
          <thead>
            <tr>
              <th>{{ t('billing.description') }}</th>
              <th class="text-center">{{ t('billing.qty') }}</th>
              <th class="text-right">{{ t('billing.unitPrice') }}</th>
              <th class="text-right">{{ t('billing.total') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, idx) in selectedInvoice.items" :key="idx">
              <td>{{ item.description }}</td>
              <td class="text-center">{{ item.quantity }}</td>
              <td class="text-right">${{ Number(item.unit_price ?? item.unitPrice ?? 0).toFixed(2) }}</td>
              <td class="text-right font-semibold">${{ Number(item.total ?? 0).toFixed(2) }}</td>
            </tr>
          </tbody>
          <tfoot>
            <tr>
              <td colspan="3" class="text-right font-semibold">{{ t('billing.subtotal') }}:</td>
              <td class="text-right">${{ Number(selectedInvoice.amount ?? 0).toFixed(2) }}</td>
            </tr>
            <tr>
              <td colspan="3" class="text-right font-semibold">{{ t('billing.tax') }}:</td>
              <td class="text-right">${{ Number(selectedInvoice.tax ?? 0).toFixed(2) }}</td>
            </tr>
            <tr class="grand-total-row">
              <td colspan="3" class="text-right font-bold">{{ t('billing.totalDue') }}:</td>
              <td class="text-right font-bold text-primary">${{ Number(selectedInvoice.total_amount ?? selectedInvoice.totalAmount ?? selectedInvoice.amount ?? 0).toFixed(2) }} {{ selectedInvoice.currency || 'USD' }}</td>
            </tr>
          </tfoot>
        </table>

        <!-- Wire Instructions Section -->
        <div v-if="selectedInvoice.status === 'unpaid' || selectedInvoice.status === 'STATUS_UNPAID'" class="wire-sheet-box mt-4">
          <div class="wire-title"><Landmark :size="14" /> {{ t('billing.bankWireDetails') }}</div>
          <p class="wire-instruction-text">{{ selectedInvoice.bank_account_instructions || selectedInvoice.bankAccountInstructions }}</p>
        </div>
      </div>

      <template #footer>
        <button class="btn btn-secondary" @click="isDetailModalOpen = false">{{ t('common.close') }}</button>
        <button
          v-if="selectedInvoice?.status === 'unpaid' || selectedInvoice?.status === 'STATUS_UNPAID'"
          class="btn btn-primary"
          @click="openPaymentModal(selectedInvoice)"
        >
          <Send :size="13" /> {{ t('billing.submitWireReference') }}
        </button>
      </template>
    </Modal>

    <!-- Submit Offline Wire Transfer Modal -->
    <Modal
      :isOpen="isPaymentModalOpen"
      :title="t('billing.submitWireTitle')"
      @close="isPaymentModalOpen = false"
    >
      <form @submit.prevent="submitOfflinePayment">
        <div class="form-group">
          <label class="form-label">{{ t('billing.invoiceNumber') }}</label>
          <input type="text" disabled :value="paymentInvoice?.invoice_number || paymentInvoice?.invoiceNumber" class="form-control" />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('billing.transactionReference') }} *</label>
          <input
            v-model="paymentForm.transaction_reference"
            type="text"
            required
            class="form-control font-mono"
            placeholder="e.g. WIRE-889922 or Slip Ref"
          />
          <span class="form-hint">{{ t('billing.transactionReferenceHint') }}</span>
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('billing.payerName') }} *</label>
          <input
            v-model="paymentForm.payer_name"
            type="text"
            required
            class="form-control"
            placeholder="e.g. John Smith / Acme Corp"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('billing.payerNotes') }}</label>
          <textarea
            v-model="paymentForm.payer_notes"
            rows="2"
            class="form-control"
            placeholder="e.g. Transferred via Mobile App on Sept 5"
          ></textarea>
        </div>
      </form>

      <template #footer>
        <button class="btn btn-secondary" @click="isPaymentModalOpen = false" :disabled="submittingPayment">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="submitOfflinePayment" :disabled="submittingPayment">
          <RefreshCw v-if="submittingPayment" :size="14" class="spin-anim" />
          {{ t('billing.submitForVerification') }}
        </button>
      </template>
    </Modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from '../locales'
import { invoiceService } from '../services/invoiceService'
import { manualPaymentService } from '../services/manualPaymentService'
import { useToastStore } from '../stores/toast'
import Modal from '../components/common/Modal.vue'
import Pagination from '../components/common/Pagination.vue'
import EmptyState from '../components/common/EmptyState.vue'
import {
  Receipt,
  Search,
  RefreshCw,
  X,
  FileText,
  Eye,
  Send,
  Landmark,
} from 'lucide-vue-next'

const { t } = useI18n()
const toastStore = useToastStore()

const invoices = ref([])
const loading = ref(false)
const submittingPayment = ref(false)
const page = ref(1)
const pageSize = ref(10)
const totalInvoices = ref(0)
const searchQuery = ref('')
const statusFilter = ref('')

const isDetailModalOpen = ref(false)
const selectedInvoice = ref(null)

const isPaymentModalOpen = ref(false)
const paymentInvoice = ref(null)
const paymentForm = reactive({
  transaction_reference: '',
  payer_name: '',
  payer_notes: '',
})

async function fetchInvoices() {
  loading.value = true
  try {
    const res = await invoiceService.getMyInvoices({
      page: page.value,
      page_size: pageSize.value,
      search: searchQuery.value,
      status: statusFilter.value,
    })
    invoices.value = res.data || []
    totalInvoices.value = res.pagination?.total_items || 0
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to load invoices')
  } finally {
    loading.value = false
  }
}

let searchTimer = null
function handleSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchInvoices()
  }, 350)
}

function clearSearch() {
  searchQuery.value = ''
  page.value = 1
  fetchInvoices()
}

function setStatusFilter(status) {
  statusFilter.value = status
  page.value = 1
  fetchInvoices()
}

function handlePageChange(newPage) {
  page.value = newPage
  fetchInvoices()
}

function handlePageSizeChange(newSize) {
  pageSize.value = newSize
  page.value = 1
  fetchInvoices()
}

function formatDate(ts) {
  if (!ts) return 'N/A'
  return new Date(ts).toLocaleDateString()
}

function formatStatus(status) {
  if (!status) return ''
  return status.toUpperCase()
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

function openInvoiceDetails(inv) {
  selectedInvoice.value = inv
  isDetailModalOpen.value = true
}

function openPaymentModal(inv) {
  paymentInvoice.value = inv
  paymentForm.transaction_reference = ''
  paymentForm.payer_name = ''
  paymentForm.payer_notes = ''
  isDetailModalOpen.value = false
  isPaymentModalOpen.value = true
}

async function submitOfflinePayment() {
  if (!paymentForm.transaction_reference || !paymentForm.payer_name) {
    toastStore.warning('Please provide transaction reference and payer name.')
    return
  }

  submittingPayment.value = true
  try {
    await manualPaymentService.submitPayment({
      invoice_id: paymentInvoice.value.id,
      amount: paymentInvoice.value.total_amount,
      currency: paymentInvoice.value.currency,
      payment_method: 'bank_wire',
      transaction_reference: paymentForm.transaction_reference,
      payer_name: paymentForm.payer_name,
      payer_notes: paymentForm.payer_notes,
    })

    toastStore.success('Payment proof submitted! Admin will verify and activate your service.')
    isPaymentModalOpen.value = false
    fetchInvoices()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to submit payment proof')
  } finally {
    submittingPayment.value = false
  }
}

onMounted(() => {
  fetchInvoices()
})
</script>

<style scoped>
.invoices-view {
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
  padding: 0.45rem 2rem 0.45rem 2.25rem;
  font-size: 13px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  outline: none;
  background-color: var(--bg-input);
  color: var(--text-primary);
}

.clear-search {
  position: absolute;
  right: 0.5rem;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
}

.status-filters {
  display: flex;
  gap: 0.35rem;
}

.filter-pill {
  font-size: 12px;
  padding: 0.35rem 0.75rem;
  border-radius: var(--radius-full);
  border: 1px solid var(--border-color);
  background-color: var(--bg-card-muted);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.filter-pill.active {
  background-color: var(--primary);
  color: #ffffff;
  border-color: var(--primary);
}

.invoice-num-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 13px;
}

.invoice-printable-sheet {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 0.5rem 0;
}

.invoice-sheet-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  border-bottom: 2px solid var(--border-color);
  padding-bottom: 1rem;
}

.invoice-brand {
  font-size: 20px;
  font-weight: 800;
  color: var(--primary);
  margin: 0;
}

.invoice-tag {
  font-family: monospace;
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
}

.invoice-meta-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  background-color: var(--bg-card-muted);
  padding: 0.75rem 1rem;
  border-radius: var(--radius-md);
}

.line-items-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.line-items-table th {
  background-color: var(--bg-card-muted);
  padding: 0.5rem;
  font-weight: 600;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-color);
}

.line-items-table td {
  padding: 0.625rem 0.5rem;
  border-bottom: 1px solid var(--border-color);
}

.grand-total-row td {
  font-size: 14px;
  border-top: 2px solid var(--border-color);
  padding-top: 0.75rem;
}

.wire-sheet-box {
  background-color: #f8fafc;
  border: 1px dashed #cbd5e1;
  border-radius: var(--radius-md);
  padding: 0.75rem 1rem;
}

.wire-title {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 12px;
  font-weight: 700;
  color: var(--text-primary);
}

.wire-instruction-text {
  font-size: 11px;
  color: var(--text-secondary);
  margin-top: 0.35rem;
  white-space: pre-wrap;
  font-family: monospace;
  line-height: 1.4;
}

.action-buttons-group {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.35rem;
}

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}
</style>
