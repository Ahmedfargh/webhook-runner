<template>
  <header class="app-header">
    <div class="header-left">
      <div class="breadcrumb">
        <span class="breadcrumb-root">Accounts</span>
        <span class="breadcrumb-separator">/</span>
        <span class="breadcrumb-current">{{ currentRouteName }}</span>
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
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { t } from '../../locales'
import LanguageSelector from '../common/LanguageSelector.vue'
import { Shield, Activity } from 'lucide-vue-next'

const route = useRoute()

const currentRouteName = computed(() => {
  if (route.path === '/') return t('nav.dashboard')
  if (route.path === '/users') return t('nav.users')
  if (route.path === '/admins') return t('nav.admins')
  if (route.path === '/roles') return t('nav.roles')
  if (route.path === '/permissions') return t('nav.permissions')
  if (route.path === '/topology') return t('nav.topology')
  if (route.path === '/audit-logs') return t('nav.auditLogs')
  if (route.path === '/admin/billing') return t('nav.adminBilling')
  if (route.path === '/plans') return t('nav.plansPricing')
  if (route.path === '/subscription') return t('nav.mySubscription')
  if (route.path === '/invoices') return t('nav.invoices')
  if (route.path === '/apps') return t('nav.apps')
  if (route.path === '/webhooks') return t('nav.webhookLogs')
  return route.name || 'Overview'
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
