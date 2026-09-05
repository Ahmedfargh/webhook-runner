<template>
  <div class="admins-view animate-fade-in">
    <!-- Action Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('admins.title') }}</h1>
        <p class="view-subtitle">{{ t('admins.subtitle') }}</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-primary" @click="openCreateDrawer">
          <Plus :size="15" /> {{ t('admins.addAdmin') }}
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
          :placeholder="t('admins.filterPlaceholder')"
          class="search-input"
        />
        <button v-if="searchQuery" class="clear-search" @click="clearSearch">
          <X :size="13" />
        </button>
      </div>

      <div class="filter-actions">
        <button class="btn btn-sm btn-secondary" @click="fetchAdmins">
          <RefreshCw :size="13" :class="{ 'spin-anim': loading }" /> {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <!-- Admins Table -->
    <div class="zoho-table-container">
      <table class="zoho-table">
        <thead>
          <tr>
            <th>{{ t('admins.title') }}</th>
            <th>{{ t('users.contactDetails') }}</th>
            <th>{{ t('admins.assignedRoles') }}</th>
            <th>{{ t('admins.directPermissions') }}</th>
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

          <tr v-else-if="admins.length === 0">
            <td colspan="5">
              <EmptyState
                :title="t('common.noRecords')"
                :description="t('admins.subtitle')"
                :icon="Shield"
              >
                <template #action>
                  <button class="btn btn-primary btn-sm" @click="openCreateDrawer">
                    <Plus :size="13" /> {{ t('admins.addAdmin') }}
                  </button>
                </template>
              </EmptyState>
            </td>
          </tr>

          <tr v-else v-for="admin in admins" :key="admin.id">
            <td>
              <div class="admin-cell">
                <div class="admin-avatar">{{ getInitials(admin.name) }}</div>
                <div>
                  <div class="admin-name">{{ admin.name }}</div>
                  <div class="admin-id">ID: {{ admin.id.slice(0, 8) }}...</div>
                </div>
              </div>
            </td>
            <td>
              <div class="contact-meta">
                <span><Mail :size="12" /> {{ admin.email }}</span>
                <span><Phone :size="12" /> {{ admin.phone }}</span>
              </div>
            </td>
            <td>
              <div class="pill-group">
                <span v-if="admin.roles?.length === 0" class="text-muted text-xs">No roles assigned</span>
                <span v-else v-for="role in admin.roles" :key="role.id" class="badge badge-warning">
                  <KeyRound :size="11" /> {{ role.name }}
                </span>
              </div>
            </td>
            <td>
              <div class="pill-group">
                <span v-if="admin.permissions?.length === 0" class="text-muted text-xs">Inherited only</span>
                <span v-else v-for="perm in admin.permissions" :key="perm.id" class="badge badge-primary">
                  <Lock :size="11" /> {{ perm.name }}
                </span>
              </div>
            </td>
            <td>
              <div class="action-buttons-group">
                <button class="action-icon-btn" :title="t('common.edit')" @click="openEditDrawer(admin)">
                  <Edit2 :size="14" />
                </button>
                <button class="action-icon-btn text-danger" :title="t('common.delete')" @click="confirmDelete(admin)">
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
        :total="totalAdmins"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <!-- Create / Edit Drawer -->
    <Drawer
      :isOpen="isDrawerOpen"
      :title="isEditing ? t('admins.editAdmin') : t('admins.addAdmin')"
      :subtitle="isEditing ? t('admins.editSubtitle') : t('admins.createSubtitle')"
      @close="isDrawerOpen = false"
      width="540px"
    >
      <form @submit.prevent="saveAdmin">
        <div class="form-group">
          <label class="form-label">{{ t('auth.fullName') }} *</label>
          <input v-model="adminForm.name" type="text" required class="form-control" placeholder="e.g. Sarah Connor" />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('auth.email') }} *</label>
          <input v-model="adminForm.email" type="email" required class="form-control" placeholder="e.g. sarah@example.com" />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('auth.phoneNumber') }} *</label>
          <input v-model="adminForm.phone" type="tel" required class="form-control" placeholder="e.g. +1987654321" />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('auth.password') }} {{ isEditing ? t('auth.passwordHint') : '*' }}</label>
          <input
            v-model="adminForm.password"
            type="password"
            :required="!isEditing"
            minlength="6"
            class="form-control"
            placeholder="••••••••"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('common.country') }} *</label>
          <select v-model="adminForm.country_id" required class="form-control">
            <option value="" disabled>{{ t('common.selectCountry') }}</option>
            <option v-for="c in countries" :key="c.id" :value="c.id">
              {{ c.name_en }} ({{ c.country_code }}) - {{ c.name_ar }}
            </option>
          </select>
        </div>

        <!-- Role Assignment Selection -->
        <div class="form-group">
          <label class="form-label">{{ t('admins.assignedRoles') }}</label>
          <div class="checkbox-grid">
            <label v-for="role in availableRoles" :key="role.id" class="custom-checkbox-item">
              <input
                type="checkbox"
                :value="role.id"
                v-model="adminForm.role_ids"
              />
              <span class="checkbox-label">{{ role.name }}</span>
            </label>
          </div>
        </div>

        <!-- Direct Permissions Assignment -->
        <div class="form-group">
          <label class="form-label">{{ t('admins.directPermissions') }}</label>
          <div class="checkbox-grid">
            <label v-for="perm in availablePermissions" :key="perm.id" class="custom-checkbox-item">
              <input
                type="checkbox"
                :value="perm.id"
                v-model="adminForm.permission_ids"
              />
              <span class="checkbox-label">{{ perm.name }}</span>
            </label>
          </div>
        </div>
      </form>

      <template #footer>
        <button class="btn btn-secondary" @click="isDrawerOpen = false" :disabled="saving">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveAdmin" :disabled="saving">
          <RefreshCw v-if="saving" :size="14" class="spin-anim" />
          {{ isEditing ? t('common.save') : t('common.create') }}
        </button>
      </template>
    </Drawer>

    <!-- Delete Modal -->
    <Modal :isOpen="isDeleteModalOpen" :title="t('common.confirmDelete')" @close="isDeleteModalOpen = false">
      <p class="modal-text">
        {{ t('admins.deletePrompt', { name: adminToDelete?.name || '' }) }}
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
import { adminService } from '../services/adminService'
import { roleService } from '../services/roleService'
import { permissionService } from '../services/permissionService'
import { countryService } from '../services/countryService'
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
  Mail,
  Phone,
  Shield,
  KeyRound,
  Lock,
  Edit2,
  Trash2,
} from 'lucide-vue-next'

const { t } = useI18n()
const toastStore = useToastStore()

const admins = ref([])
const countries = ref([])
const availableRoles = ref([])
const availablePermissions = ref([])
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const page = ref(1)
const pageSize = ref(10)
const totalAdmins = ref(0)
const searchQuery = ref('')

const isDrawerOpen = ref(false)
const isEditing = ref(false)
const isDeleteModalOpen = ref(false)
const adminToDelete = ref(null)

const adminForm = reactive({
  id: '',
  name: '',
  email: '',
  phone: '',
  password: '',
  country_id: '',
  role_ids: [],
  permission_ids: [],
})

async function fetchAdmins() {
  loading.value = true
  try {
    const res = await adminService.listAdmins({
      page: page.value,
      page_size: pageSize.value,
      search: searchQuery.value,
    })
    admins.value = res.data || []
    totalAdmins.value = res.pagination?.total_items || 0
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to load administrators')
  } finally {
    loading.value = false
  }
}

async function fetchRolesAndPermissions() {
  try {
    const [rolesRes, permsRes, countryRes] = await Promise.all([
      roleService.listRoles({ page: 1, page_size: 100 }),
      permissionService.listPermissions({ page: 1, page_size: 100 }),
      countryService.listCountries(),
    ])
    availableRoles.value = rolesRes.data || []
    availablePermissions.value = permsRes.data || []
    countries.value = countryRes.data || []
  } catch (err) {
    // Non-critical background fetch
  }
}

let searchTimer = null
function handleSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchAdmins()
  }, 350)
}

function clearSearch() {
  searchQuery.value = ''
  page.value = 1
  fetchAdmins()
}

function handlePageChange(newPage) {
  page.value = newPage
  fetchAdmins()
}

function handlePageSizeChange(newSize) {
  pageSize.value = newSize
  page.value = 1
  fetchAdmins()
}

function openCreateDrawer() {
  isEditing.value = false
  adminForm.id = ''
  adminForm.name = ''
  adminForm.email = ''
  adminForm.phone = ''
  adminForm.password = ''
  
  const egypt = countries.value.find((c) => c.country_code === 'EG')
  adminForm.country_id = egypt ? egypt.id : (countries.value[0]?.id || '')
  adminForm.role_ids = []
  adminForm.permission_ids = []
  isDrawerOpen.value = true
}

function openEditDrawer(admin) {
  isEditing.value = true
  adminForm.id = admin.id
  adminForm.name = admin.name
  adminForm.email = admin.email
  adminForm.phone = admin.phone
  adminForm.password = ''
  adminForm.country_id = admin.country_id || admin.country?.id || ''
  adminForm.role_ids = (admin.roles || []).map((r) => r.id)
  adminForm.permission_ids = (admin.permissions || []).map((p) => p.id)
  isDrawerOpen.value = true
}

async function saveAdmin() {
  if (!adminForm.name || !adminForm.email || !adminForm.phone) {
    toastStore.warning('Please fill in required fields.')
    return
  }
  if (!isEditing.value && !adminForm.password) {
    toastStore.warning('Password is required for new administrator.')
    return
  }

  saving.value = true
  try {
    if (isEditing.value) {
      const payload = {
        name: adminForm.name,
        email: adminForm.email,
        phone: adminForm.phone,
        country_id: adminForm.country_id,
        role_ids: adminForm.role_ids,
        permission_ids: adminForm.permission_ids,
      }
      if (adminForm.password) payload.password = adminForm.password
      await adminService.updateAdmin(adminForm.id, payload)
      toastStore.success('Administrator updated successfully')
    } else {
      await adminService.createAdmin({
        name: adminForm.name,
        email: adminForm.email,
        phone: adminForm.phone,
        password: adminForm.password,
        country_id: adminForm.country_id,
        role_ids: adminForm.role_ids,
        permission_ids: adminForm.permission_ids,
      })
      toastStore.success('Administrator created successfully')
    }
    isDrawerOpen.value = false
    fetchAdmins()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to save administrator')
  } finally {
    saving.value = false
  }
}

function confirmDelete(admin) {
  adminToDelete.value = admin
  isDeleteModalOpen.value = true
}

async function executeDelete() {
  if (!adminToDelete.value) return
  deleting.value = true
  try {
    await adminService.deleteAdmin(adminToDelete.value.id)
    toastStore.success(`Administrator ${adminToDelete.value.name} deleted`)
    isDeleteModalOpen.value = false
    adminToDelete.value = null
    fetchAdmins()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to delete admin')
  } finally {
    deleting.value = false
  }
}

function getInitials(name) {
  if (!name) return 'AD'
  return name.slice(0, 2).toUpperCase()
}

onMounted(() => {
  fetchAdmins()
  fetchRolesAndPermissions()
})
</script>

<style scoped>
.admins-view {
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

.admin-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.admin-avatar {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, #059669, #10b981);
  color: #ffffff;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.admin-name {
  font-weight: 600;
  color: var(--text-primary);
}

.admin-id {
  font-size: 11px;
  color: var(--text-muted);
  font-family: monospace;
}

.contact-meta {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  font-size: 12px;
  color: var(--text-secondary);
}

.contact-meta span {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.pill-group {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.checkbox-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.5rem;
  max-height: 160px;
  overflow-y: auto;
  padding: 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background-color: var(--bg-card-muted);
}

.custom-checkbox-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 12px;
  color: var(--text-primary);
  cursor: pointer;
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

.spin-anim {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  100% { transform: rotate(360deg); }
}
</style>
