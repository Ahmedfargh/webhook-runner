<template>
  <header class="app-header">
    <div class="header-left">
      <div class="breadcrumb">
        <span class="breadcrumb-root">Accounts</span>
        <span class="breadcrumb-separator">/</span>
        <span class="breadcrumb-current">{{ currentRouteName }}</span>
      </div>
    </div>

    <div class="header-center">
      <div class="global-search-bar">
        <Search :size="15" class="search-icon" />
        <input
          type="text"
          :placeholder="t('header.searchPlaceholder')"
          class="global-search-input"
          ref="searchInput"
          @keydown.stop
        />
      </div>
    </div>

    <div class="header-right">
      <!-- Language Selector with Country Flags -->
      <LanguageSelector />

      <div class="service-badge-header">
        <span class="badge badge-success">
          <Shield :size="12" /> {{ t('header.gatewayActive') }}
        </span>
      </div>

      <div class="header-actions">
        <router-link to="/topology" class="header-icon-btn" :title="t('nav.topology')">
          <Activity :size="17" />
        </router-link>
      </div>
    </div>
  </header>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { t } from '../../locales'
import LanguageSelector from '../common/LanguageSelector.vue'
import { Search, Shield, Activity } from 'lucide-vue-next'

const route = useRoute()
const searchInput = ref(null)

const currentRouteName = computed(() => {
  if (route.path === '/') return t('nav.dashboard')
  if (route.path === '/users') return t('nav.users')
  if (route.path === '/admins') return t('nav.admins')
  if (route.path === '/roles') return t('nav.roles')
  if (route.path === '/permissions') return t('nav.permissions')
  if (route.path === '/topology') return t('nav.topology')
  return route.name || 'Overview'
})

function handleGlobalKeydown(e) {
  if (e.key === '/' && document.activeElement !== searchInput.value) {
    e.preventDefault()
    searchInput.value?.focus()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleGlobalKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeydown)
})
</script>

<style scoped>
.app-header {
  height: 56px;
  background-color: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 1.5rem;
  gap: 1rem;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 13px;
}

.breadcrumb-root {
  color: var(--text-muted);
  font-weight: 500;
}

.breadcrumb-separator {
  color: var(--border-color);
}

.breadcrumb-current {
  color: var(--text-primary);
  font-weight: 600;
}

.header-center {
  flex: 1;
  max-width: 440px;
}

.global-search-bar {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  color: var(--text-muted);
  pointer-events: none;
}

[dir="rtl"] .search-icon {
  left: auto;
  right: 0.75rem;
}

.global-search-input {
  width: 100%;
  padding: 0.45rem 0.75rem 0.45rem 2.25rem;
  font-size: 12px;
  background-color: var(--bg-card-muted);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-full);
  color: var(--text-primary);
  outline: none;
  transition: all var(--transition-fast);
}

[dir="rtl"] .global-search-input {
  padding: 0.45rem 2.25rem 0.45rem 0.75rem;
}

.global-search-input:focus {
  background-color: #ffffff;
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px var(--primary-focus);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 0.875rem;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.header-icon-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  background: transparent;
  text-decoration: none;
  transition: all var(--transition-fast);
}
.header-icon-btn:hover {
  background-color: var(--bg-card-muted);
  color: var(--text-primary);
}
</style>
