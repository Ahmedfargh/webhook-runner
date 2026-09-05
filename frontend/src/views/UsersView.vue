<template>
  <div class="users-view animate-fade-in">
    <!-- Action Header -->
    <div class="view-header">
      <div>
        <h1 class="view-title">{{ t('users.title') }}</h1>
        <p class="view-subtitle">{{ t('users.subtitle') }}</p>
      </div>

      <div class="header-actions">
        <button class="btn btn-primary" @click="openCreateDrawer">
          <Plus :size="15" /> {{ t('users.addUser') }}
        </button>
      </div>
    </div>

    <!-- Filter & Search Bar (Zoho Projects Style) -->
    <div class="filter-card">
      <div class="search-box">
        <Search :size="15" class="search-icon" />
        <input
          type="text"
          v-model="searchQuery"
          @input="handleSearch"
          :placeholder="t('users.filterPlaceholder')"
          class="search-input"
        />
        <button v-if="searchQuery" class="clear-search" @click="clearSearch">
          <X :size="13" />
        </button>
      </div>

      <div class="filter-actions">
        <button class="btn btn-sm btn-secondary" @click="fetchUsers">
          <RefreshCw :size="13" :class="{ 'spin-anim': loading }" /> {{ t('common.refresh') }}
        </button>
      </div>
    </div>

    <!-- Users Table Container -->
    <div class="zoho-table-container">
      <table class="zoho-table">
        <thead>
          <tr>
            <th>{{ t('nav.users') }}</th>
            <th>{{ t('users.contactDetails') }}</th>
            <th>{{ t('common.country') }}</th>
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

          <tr v-else-if="users.length === 0">
            <td colspan="5">
              <EmptyState
                :title="t('common.noRecords')"
                :description="t('users.subtitle')"
                :icon="UsersIcon"
              >
                <template #action>
                  <button class="btn btn-primary btn-sm" @click="openCreateDrawer">
                    <Plus :size="13" /> {{ t('users.addUser') }}
                  </button>
                </template>
              </EmptyState>
            </td>
          </tr>

          <tr v-else v-for="user in users" :key="user.id">
            <td>
              <div class="user-cell">
                <div class="user-avatar-circle">{{ getInitials(user.name) }}</div>
                <div class="user-cell-meta">
                  <span class="user-name">{{ user.name }}</span>
                  <span class="user-id">ID: {{ user.id.slice(0, 8) }}...</span>
                </div>
              </div>
            </td>
            <td>
              <div class="contact-meta">
                <span class="contact-email"><Mail :size="12" /> {{ user.email }}</span>
                <span class="contact-phone"><Phone :size="12" /> {{ user.phone }}</span>
              </div>
            </td>
            <td>
              <span v-if="user.country" class="country-pill">
                <Globe :size="12" /> {{ user.country.name_en || user.country.name_ar || user.country.country_code }}
              </span>
              <span v-else class="text-muted text-xs">N/A</span>
            </td>
            <td>
              <span class="date-text">{{ formatDate(user.created_at) }}</span>
            </td>
            <td>
              <div class="action-buttons-group">
                <button class="action-icon-btn" :title="t('common.edit')" @click="openEditDrawer(user)">
                  <Edit2 :size="14" />
                </button>
                <button class="action-icon-btn text-danger" :title="t('common.delete')" @click="confirmDelete(user)">
                  <Trash2 :size="14" />
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
        :total="totalUsers"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <!-- Create / Edit Slide-Over Drawer -->
    <Drawer
      :isOpen="isDrawerOpen"
      :title="isEditing ? t('users.editUser') : t('users.addUser')"
      :subtitle="isEditing ? t('users.editUserSubtitle') : t('users.createUserSubtitle')"
      @close="closeDrawer"
    >
      <form @submit.prevent="saveUser">
        <div class="form-group">
          <label class="form-label">{{ t('auth.fullName') }} *</label>
          <input
            v-model="userForm.name"
            type="text"
            required
            class="form-control"
            placeholder="e.g. John Doe"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('auth.email') }} *</label>
          <input
            v-model="userForm.email"
            type="email"
            required
            class="form-control"
            placeholder="e.g. john@example.com"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('auth.phoneNumber') }} *</label>
          <input
            v-model="userForm.phone"
            type="tel"
            required
            class="form-control"
            placeholder="e.g. +1234567890"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('auth.password') }} {{ isEditing ? t('auth.passwordHint') : '*' }}</label>
          <input
            v-model="userForm.password"
            type="password"
            :required="!isEditing"
            minlength="6"
            class="form-control"
            placeholder="••••••••"
          />
        </div>

        <div class="form-group">
          <label class="form-label">{{ t('common.country') }} *</label>
          <select v-model="userForm.country_id" required class="form-control">
            <option value="" disabled>{{ t('common.selectCountry') }}</option>
            <option v-for="c in countries" :key="c.id" :value="c.id">
              {{ c.name_en }} ({{ c.country_code }}) - {{ c.name_ar }}
            </option>
          </select>
        </div>
      </form>

      <template #footer>
        <button class="btn btn-secondary" @click="closeDrawer" :disabled="saving">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" @click="saveUser" :disabled="saving">
          <RefreshCw v-if="saving" :size="14" class="spin-anim" />
          {{ isEditing ? t('common.save') : t('common.create') }}
        </button>
      </template>
    </Drawer>

    <!-- Delete Confirmation Modal -->
    <Modal
      :isOpen="isDeleteModalOpen"
      :title="t('common.confirmDelete')"
      @close="isDeleteModalOpen = false"
    >
      <p class="modal-text">
        {{ t('users.deletePrompt', { name: userToDelete?.name || '', email: userToDelete?.email || '' }) }}
      </p>

      <template #footer>
        <button class="btn btn-secondary" @click="isDeleteModalOpen = false" :disabled="deleting">{{ t('common.cancel') }}</button>
        <button class="btn btn-danger" @click="executeDelete" :disabled="deleting">
          <RefreshCw v-if="deleting" :size="14" class="spin-anim" />
          {{ t('common.delete') }}
        </button>
      </template>
    </Modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from '../locales'
import { userService } from '../services/userService'
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
  Globe,
  Edit2,
  Trash2,
  Users as UsersIcon,
} from 'lucide-vue-next'

const { t } = useI18n()
const toastStore = useToastStore()

const users = ref([])
const countries = ref([])
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const page = ref(1)
const pageSize = ref(10)
const totalUsers = ref(0)
const searchQuery = ref('')

// Drawer & Modal States
const isDrawerOpen = ref(false)
const isEditing = ref(false)
const isDeleteModalOpen = ref(false)
const userToDelete = ref(null)

const userForm = reactive({
  id: '',
  name: '',
  email: '',
  phone: '',
  password: '',
  country_id: '',
})

async function fetchUsers() {
  loading.value = true
  try {
    const res = await userService.listUsers({
      page: page.value,
      page_size: pageSize.value,
      search: searchQuery.value,
    })
    users.value = res.data || []
    totalUsers.value = res.pagination?.total_items || 0
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to load users list')
  } finally {
    loading.value = false
  }
}

let searchTimer = null
function handleSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchUsers()
  }, 350)
}

function clearSearch() {
  searchQuery.value = ''
  page.value = 1
  fetchUsers()
}

function handlePageChange(newPage) {
  page.value = newPage
  fetchUsers()
}

function handlePageSizeChange(newSize) {
  pageSize.value = newSize
  page.value = 1
  fetchUsers()
}

function openCreateDrawer() {
  isEditing.value = false
  userForm.id = ''
  userForm.name = ''
  userForm.email = ''
  userForm.phone = ''
  userForm.password = ''
  
  // Default to Egypt if available, otherwise first country
  const egypt = countries.value.find((c) => c.country_code === 'EG')
  userForm.country_id = egypt ? egypt.id : (countries.value[0]?.id || '')
  isDrawerOpen.value = true
}

function openEditDrawer(user) {
  isEditing.value = true
  userForm.id = user.id
  userForm.name = user.name
  userForm.email = user.email
  userForm.phone = user.phone
  userForm.password = ''
  userForm.country_id = user.country_id || user.country?.id || ''
  isDrawerOpen.value = true
}

function closeDrawer() {
  isDrawerOpen.value = false
}

async function saveUser() {
  if (!userForm.name || !userForm.email || !userForm.phone) {
    toastStore.warning('Please fill all required fields.')
    return
  }
  if (!userForm.country_id) {
    toastStore.warning('Please select a country.')
    return
  }
  if (!isEditing.value && !userForm.password) {
    toastStore.warning('Password is required for new user.')
    return
  }

  saving.value = true
  try {
    if (isEditing.value) {
      const payload = {
        name: userForm.name,
        email: userForm.email,
        phone: userForm.phone,
        country_id: userForm.country_id,
      }
      if (userForm.password) payload.password = userForm.password
      await userService.updateUser(userForm.id, payload)
      toastStore.success('User updated successfully')
    } else {
      await userService.createUser({
        name: userForm.name,
        email: userForm.email,
        phone: userForm.phone,
        password: userForm.password,
        country_id: userForm.country_id,
      })
      toastStore.success('User created successfully')
    }
    closeDrawer()
    fetchUsers()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to save user')
  } finally {
    saving.value = false
  }
}

function confirmDelete(user) {
  userToDelete.value = user
  isDeleteModalOpen.value = true
}

async function executeDelete() {
  if (!userToDelete.value) return
  deleting.value = true
  try {
    await userService.deleteUser(userToDelete.value.id)
    toastStore.success(`User ${userToDelete.value.name} deleted`)
    isDeleteModalOpen.value = false
    userToDelete.value = null
    fetchUsers()
  } catch (err) {
    toastStore.error(err.response?.data?.error || 'Failed to delete user')
  } finally {
    deleting.value = false
  }
}

function getInitials(name) {
  if (!name) return 'U'
  return name.slice(0, 2).toUpperCase()
}

function formatDate(ts) {
  if (!ts) return 'Just now'
  if (typeof ts === 'string') return new Date(ts).toLocaleDateString()
  if (ts.seconds) return new Date(ts.seconds * 1000).toLocaleDateString()
  return 'Recently'
}

async function fetchCountries() {
  try {
    const res = await countryService.listCountries()
    countries.value = res.data || []
  } catch (err) {
    console.error('Failed to load countries', err)
  }
}

onMounted(() => {
  fetchUsers()
  fetchCountries()
})
</script>

<style scoped>
.users-view {
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

.user-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-avatar-circle {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, #1e40af, #3b82f6);
  color: #ffffff;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-cell-meta {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-weight: 600;
  color: var(--text-primary);
}

.user-id {
  font-size: 11px;
  color: var(--text-muted);
  font-family: monospace;
}

.contact-meta {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  font-size: 12px;
}

.contact-email, .contact-phone {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  color: var(--text-secondary);
}

.country-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 12px;
  color: var(--text-secondary);
  background-color: var(--bg-card-muted);
  padding: 0.2rem 0.5rem;
  border-radius: var(--radius-sm);
}

.date-text {
  font-size: 12px;
  color: var(--text-secondary);
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
