<template>
  <div class="permissions-view animate-fade-in">
    <!-- Action Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('permissions.title') }}</h1>
        <p class="view-subtitle">{{ t('permissions.subtitle') }}</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-primary" @click="openCreateModal">
          <Plus :size="15" /> {{ t('permissions.addPermission') }}
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
          :placeholder="t('permissions.filterPlaceholder')"
          class="search-input"
        />
        <button v-if="searchQuery" class="clear-search" @click="clearSearch">
          <X :size="13" />
        </button>
      </div>

      <div class="filter-actions">
        <button class="btn btn-sm btn-secondary" @click="fetchPermissions">
          <RefreshCw :size="13" :class="{ 'spin-anim': loading }" /> {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <!-- Permissions Table -->
    <div class="zoho-table-container">
      <table class="zoho-table">
        <thead>
          <tr>
            <th>{{ t('permissions.permissionKey') }}</th>
            <th>{{ t('common.id') }}</th>
            <th>{{ t('permissions.scope') }}</th>
            <th>{{ t('users.createdDate') }}</th>
            <th class="text-right">{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td colspan="5" class="text-center py-8">
              <div class="loading-state">
                <RefreshCw :size="20" class="spin-anim text-primary" />
                <span>{{ t('common.loading') }}</span>
              </div>
            </td>
          </tr>

          <tr v-else-if="permissions.length === 0">
            <td colspan="5">
              <EmptyState
                :title="t('common.noRecords')"
                :description="t('permissions.subtitle')"
                :icon="Lock"
              >
                <template #action>
                  <button class="btn btn-primary btn-sm" @click="openCreateModal">
                    <Plus :size="13" /> {{ t('permissions.addPermission') }}
                  </button>
                </template>
              </EmptyState>
            </td>
          </tr>

          <tr v-else v-for="perm in permissions" :key="perm.id">
            <td>
              <div class="perm-title-cell">
                <div class="perm-icon-box">
                  <Lock :size="13" />
                </div>
                <span class="perm-name font-mono">{{ perm.name }}</span>
              </div>
            </td>
            <td>
              <span class="perm-id-code">{{ perm.id }}</span>
            </td>
            <td>
              <span class="badge badge-primary">Accounts Service</span>
            </td>
            <td>
              <span class="date-text">{{ formatDate(perm.created_at) }}</span>
            </td>
            <td>
              <div class="action-buttons-group">
                <button class="action-icon-btn" :title="t('common.edit')" @click="openEditModal(perm)">
                  <Edit2 :size="14" />
                </button>
                <button class="action-icon-btn text-danger" :title="t('common.delete')" @click="confirmDelete(perm)">
                  <Trash2 :size="14" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <Pagination
        :page="page"
        :pageSize="pageSize"
        :total="totalPermissions"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <!-- Create / Edit Modal -->
    <Modal
      :isOpen="isModalOpen"
      :title="isEditing ? t('permissions.editPermission') : t('permissions.addPermission')"
      @close="isModalOpen = false"
    >
      <form @submit.prevent="savePermission">
        <div class="form-group">
          <label class="form-label">{{ t('permissions.permissionKey') }} *</label>
          <input
            v-model="permForm.name"
            type="text"
            required
            class="form-control font-mono"
            placeholder="e.g. users:create or roles:assign"
          />
          <span class="form-hint">{{ t('permissions.permissionHint') }}</span>
        </div>
      </form>

      <template #footer>
        <button class="btn btn-secondary" @click="isModalOpen = false" :disabled="saving">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="savePermission" :disabled="saving">
          <RefreshCw v-if="saving" :size="14" class="spin-anim" />
          {{ isEditing ? t('common.save') : t('common.create') }}
        </button>
      </template>
    </Modal>

    <!-- Delete Modal -->
    <Modal :isOpen="isDeleteModalOpen" :title="t('common.confirmDelete')" @close="isDeleteModalOpen = false">
      <p class="modal-text">
        {{ t('permissions.deletePrompt', { name: permToDelete?.name || '' }) }}
      </p>
      <template #footer>
        <button class="btn btn-secondary" @click="isDeleteModalOpen = false" :disabled="deleting">{{ t('common.cancel') }}</button>
        <button class="btn btn-danger" @click="executeDelete" :disabled="deleting">
          <RefreshCw v-if="deleting" :size="14" class="spin-anim" /> {{ t('common.delete') }}
        </button>
      </template>
    </Modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from '../locales'
import { permissionService } from '../services/permissionService'
import { useToastStore } from '../stores/toast'
import Modal from '../components/common/Modal.vue'
import Pagination from '../components/common/Pagination.vue'
import EmptyState from '../components/common/EmptyState.vue'
import {
  Plus,
  Search,
  RefreshCw,
  X,
  Lock,
  Edit2,
  Trash2,
} from 'lucide-vue-next'

const { t } = useI18n()
const toastStore = useToastStore()

const permissions = ref([])
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const page = ref(1)
const pageSize = ref(10)
const totalPermissions = ref(0)
const searchQuery = ref('')

const isModalOpen = ref(false)
const isEditing = ref(false)
const isDeleteModalOpen = ref(false)
const permToDelete = ref(null)

const permForm = reactive({
  id: '',
  name: '',
})

async function fetchPermissions() {
  loading.value = true
  try {
    const res = await permissionService.listPermissions({
      page: page.value,
      page_size: pageSize.value,
      search: searchQuery.value,
    })
    permissions.value = res.data || []
    totalPermissions.value = res.pagination?.total_items || 0
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to load permissions')
  } finally {
    loading.value = false
  }
}

let searchTimer = null
function handleSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchPermissions()
  }, 350)
}

function clearSearch() {
  searchQuery.value = ''
  page.value = 1
  fetchPermissions()
}

function handlePageChange(newPage) {
  page.value = newPage
  fetchPermissions()
}

function handlePageSizeChange(newSize) {
  pageSize.value = newSize
  page.value = 1
  fetchPermissions()
}

function openCreateModal() {
  isEditing.value = false
  permForm.id = ''
  permForm.name = ''
  isModalOpen.value = true
}

function openEditModal(perm) {
  isEditing.value = true
  permForm.id = perm.id
  permForm.name = perm.name
  isModalOpen.value = true
}

async function savePermission() {
  if (!permForm.name) {
    toastStore.warning('Permission name is required.')
    return
  }

  saving.value = true
  try {
    if (isEditing.value) {
      await permissionService.updatePermission(permForm.id, { name: permForm.name })
      toastStore.success('Permission updated')
    } else {
      await permissionService.createPermission({ name: permForm.name })
      toastStore.success('Permission created')
    }
    isModalOpen.value = false
    fetchPermissions()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to save permission')
  } finally {
    saving.value = false
  }
}

function confirmDelete(perm) {
  permToDelete.value = perm
  isDeleteModalOpen.value = true
}

async function executeDelete() {
  if (!permToDelete.value) return
  deleting.value = true
  try {
    await permissionService.deletePermission(permToDelete.value.id)
    toastStore.success(`Permission ${permToDelete.value.name} deleted`)
    isDeleteModalOpen.value = false
    permToDelete.value = null
    fetchPermissions()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to delete permission')
  } finally {
    deleting.value = false
  }
}

function formatDate(ts) {
  if (!ts) return 'Just now'
  if (typeof ts === 'string') return new Date(ts).toLocaleDateString()
  if (ts.seconds) return new Date(ts.seconds * 1000).toLocaleDateString()
  return 'Recently'
}

onMounted(() => {
  fetchPermissions()
})
</script>

<style scoped>
.permissions-view {
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
  max-width: 380px;
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
.search-input:focus {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px var(--primary-focus);
}

.clear-search {
  position: absolute;
  right: 0.5rem;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
}

.perm-title-cell {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.perm-icon-box {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  background-color: var(--primary-light);
  color: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.perm-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 13px;
}

.perm-id-code {
  font-family: monospace;
  font-size: 11px;
  color: var(--text-muted);
}

.form-hint {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 0.35rem;
  display: block;
}

.action-buttons-group {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.35rem;
}

.action-icon-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  background-color: var(--bg-card);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}
.action-icon-btn:hover {
  background-color: var(--bg-card-muted);
  color: var(--text-primary);
}
.action-icon-btn.text-danger:hover {
  background-color: var(--danger-light);
  color: var(--danger);
  border-color: var(--danger-border);
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  color: var(--text-muted);
  font-size: 13px;
}

.modal-text {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.modal-text code {
  background-color: var(--bg-card-muted);
  padding: 0.1rem 0.35rem;
  border-radius: 4px;
  color: var(--primary);
}

.date-text {
  font-size: 12px;
  color: var(--text-secondary);
}

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}
</style>
