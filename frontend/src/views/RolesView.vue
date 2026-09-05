<template>
  <div class="roles-view animate-fade-in">
    <!-- Action Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('roles.title') }}</h1>
        <p class="view-subtitle">{{ t('roles.subtitle') }}</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-primary" @click="openCreateDrawer">
          <Plus :size="15" /> {{ t('roles.createRole') }}
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
          :placeholder="t('roles.filterPlaceholder')"
          class="search-input"
        />
        <button v-if="searchQuery" class="clear-search" @click="clearSearch">
          <X :size="13" />
        </button>
      </div>

      <div class="filter-actions">
        <button class="btn btn-sm btn-secondary" @click="fetchRoles">
          <RefreshCw :size="13" :class="{ 'spin-anim': loading }" /> {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <!-- Roles Matrix Grid / Table -->
    <div class="zoho-table-container">
      <table class="zoho-table">
        <thead>
          <tr>
            <th>{{ t('roles.roleName') }}</th>
            <th>{{ t('roles.roleID') }}</th>
            <th>{{ t('roles.grantedPermissions') }}</th>
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

          <tr v-else-if="roles.length === 0">
            <td colspan="5">
              <EmptyState
                :title="t('common.noRecords')"
                :description="t('roles.subtitle')"
                :icon="KeyRound"
              >
                <template #action>
                  <button class="btn btn-primary btn-sm" @click="openCreateDrawer">
                    <Plus :size="13" /> {{ t('roles.createRole') }}
                  </button>
                </template>
              </EmptyState>
            </td>
          </tr>

          <tr v-else v-for="role in roles" :key="role.id">
            <td>
              <div class="role-title-cell">
                <div class="role-icon-box">
                  <KeyRound :size="14" />
                </div>
                <span class="role-name">{{ role.name }}</span>
              </div>
            </td>
            <td>
              <span class="role-id-code">{{ role.id }}</span>
            </td>
            <td>
              <div class="permissions-pill-wrap">
                <span v-if="role.permissions?.length === 0" class="text-muted text-xs">No permissions assigned</span>
                <span
                  v-else
                  v-for="perm in role.permissions"
                  :key="perm.id"
                  class="badge badge-primary"
                >
                  <Lock :size="10" /> {{ perm.name }}
                </span>
              </div>
            </td>
            <td>
              <span class="date-text">{{ formatDate(role.created_at) }}</span>
            </td>
            <td>
              <div class="action-buttons-group">
                <button class="action-icon-btn" :title="t('common.edit')" @click="openEditDrawer(role)">
                  <Edit2 :size="14" />
                </button>
                <button class="action-icon-btn text-danger" :title="t('common.delete')" @click="confirmDelete(role)">
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
        :total="totalRoles"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <!-- Create / Edit Drawer -->
    <Drawer
      :isOpen="isDrawerOpen"
      :title="isEditing ? t('roles.configureRole') : t('roles.createRole')"
      :subtitle="isEditing ? t('roles.subtitle') : t('roles.subtitle')"
      @close="isDrawerOpen = false"
      width="520px"
    >
      <form @submit.prevent="saveRole">
        <div class="form-group">
          <label class="form-label">{{ t('roles.roleName') }} *</label>
          <input
            v-model="roleForm.name"
            type="text"
            required
            class="form-control"
            placeholder="e.g. billing_manager or security_auditor"
          />
        </div>

        <div class="form-group">
          <div class="perm-header-flex">
            <label class="form-label">{{ t('roles.permissionsChecklist') }}</label>
            <span class="perm-count-selected">{{ roleForm.permission_ids.length }} {{ t('roles.selected') }}</span>
          </div>

          <div class="permissions-matrix-list">
            <label
              v-for="perm in availablePermissions"
              :key="perm.id"
              class="matrix-item"
              :class="{ 'matrix-item-selected': roleForm.permission_ids.includes(perm.id) }"
            >
              <input
                type="checkbox"
                :value="perm.id"
                v-model="roleForm.permission_ids"
              />
              <div class="matrix-info">
                <span class="matrix-name">{{ perm.name }}</span>
                <span class="matrix-id">ID: {{ perm.id }}</span>
              </div>
            </label>
          </div>
        </div>
      </form>

      <template #footer>
        <button class="btn btn-secondary" @click="isDrawerOpen = false" :disabled="saving">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveRole" :disabled="saving">
          <RefreshCw v-if="saving" :size="14" class="spin-anim" />
          {{ isEditing ? t('common.save') : t('common.create') }}
        </button>
      </template>
    </Drawer>

    <!-- Delete Modal -->
    <Modal :isOpen="isDeleteModalOpen" :title="t('common.confirmDelete')" @close="isDeleteModalOpen = false">
      <p class="modal-text">
        {{ t('roles.deletePrompt', { name: roleToDelete?.name || '' }) }}
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
import { roleService } from '../services/roleService'
import { permissionService } from '../services/permissionService'
import { useToastStore } from '../stores/toast'
import Drawer from '../components/common/Drawer.vue'
import Modal from '../components/common/Modal.vue'
import Pagination from '../components/common/Pagination.vue'
import EmptyState from '../components/common/EmptyState.vue'
import {
  Plus,
  Search,
  RefreshCw,
  X,
  KeyRound,
  Lock,
  Edit2,
  Trash2,
} from 'lucide-vue-next'

const { t } = useI18n()
const toastStore = useToastStore()

const roles = ref([])
const availablePermissions = ref([])
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const page = ref(1)
const pageSize = ref(10)
const totalRoles = ref(0)
const searchQuery = ref('')

const isDrawerOpen = ref(false)
const isEditing = ref(false)
const isDeleteModalOpen = ref(false)
const roleToDelete = ref(null)

const roleForm = reactive({
  id: '',
  name: '',
  permission_ids: [],
})

async function fetchRoles() {
  loading.value = true
  try {
    const res = await roleService.listRoles({
      page: page.value,
      page_size: pageSize.value,
      search: searchQuery.value,
    })
    roles.value = res.data || []
    totalRoles.value = res.pagination?.total_items || 0
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to load roles')
  } finally {
    loading.value = false
  }
}

async function fetchPermissions() {
  try {
    const res = await permissionService.listPermissions({ page: 1, page_size: 100 })
    availablePermissions.value = res.data || []
  } catch (err) {
    // Non-critical
  }
}

let searchTimer = null
function handleSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchRoles()
  }, 350)
}

function clearSearch() {
  searchQuery.value = ''
  page.value = 1
  fetchRoles()
}

function handlePageChange(newPage) {
  page.value = newPage
  fetchRoles()
}

function handlePageSizeChange(newSize) {
  pageSize.value = newSize
  page.value = 1
  fetchRoles()
}

function openCreateDrawer() {
  isEditing.value = false
  roleForm.id = ''
  roleForm.name = ''
  roleForm.permission_ids = []
  isDrawerOpen.value = true
}

function openEditDrawer(role) {
  isEditing.value = true
  roleForm.id = role.id
  roleForm.name = role.name
  roleForm.permission_ids = (role.permissions || []).map((p) => p.id)
  isDrawerOpen.value = true
}

async function saveRole() {
  if (!roleForm.name) {
    toastStore.warning('Role name is required.')
    return
  }

  saving.value = true
  try {
    if (isEditing.value) {
      await roleService.updateRole(roleForm.id, {
        name: roleForm.name,
        permission_ids: roleForm.permission_ids,
      })
      toastStore.success('Role updated successfully')
    } else {
      await roleService.createRole({
        name: roleForm.name,
        permission_ids: roleForm.permission_ids,
      })
      toastStore.success('Role created successfully')
    }
    isDrawerOpen.value = false
    fetchRoles()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to save role')
  } finally {
    saving.value = false
  }
}

function confirmDelete(role) {
  roleToDelete.value = role
  isDeleteModalOpen.value = true
}

async function executeDelete() {
  if (!roleToDelete.value) return
  deleting.value = true
  try {
    await roleService.deleteRole(roleToDelete.value.id)
    toastStore.success(`Role ${roleToDelete.value.name} deleted`)
    isDeleteModalOpen.value = false
    roleToDelete.value = null
    fetchRoles()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to delete role')
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
  fetchRoles()
  fetchPermissions()
})
</script>

<style scoped>
.roles-view {
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

.role-title-cell {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.role-icon-box {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  background-color: var(--warning-light);
  color: var(--warning);
  display: flex;
  align-items: center;
  justify-content: center;
}

.role-name {
  font-weight: 600;
  color: var(--text-primary);
}

.role-id-code {
  font-family: monospace;
  font-size: 11px;
  color: var(--text-muted);
}

.permissions-pill-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  max-width: 350px;
}

.perm-header-flex {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.perm-count-selected {
  font-size: 11px;
  color: var(--primary);
  font-weight: 600;
}

.permissions-matrix-list {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  max-height: 260px;
  overflow-y: auto;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 0.5rem;
  background-color: var(--bg-card-muted);
}

.matrix-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.45rem 0.625rem;
  border-radius: var(--radius-sm);
  background-color: #ffffff;
  border: 1px solid var(--border-color);
  cursor: pointer;
  transition: all var(--transition-fast);
}
.matrix-item:hover {
  background-color: #f8fafc;
}
.matrix-item-selected {
  border-color: #93c5fd;
  background-color: var(--primary-light);
}

.matrix-info {
  display: flex;
  flex-direction: column;
}

.matrix-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
}

.matrix-id {
  font-size: 10px;
  color: var(--text-muted);
  font-family: monospace;
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
