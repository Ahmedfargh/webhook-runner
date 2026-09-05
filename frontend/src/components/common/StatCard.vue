<template>
  <div class="stat-card">
    <div class="stat-header">
      <div class="stat-icon-wrapper" :style="{ backgroundColor: iconBg, color: iconColor }">
        <component :is="icon" :size="20" />
      </div>
      <span v-if="badge" class="stat-badge" :class="badgeClass">{{ badge }}</span>
    </div>
    <div class="stat-body">
      <div class="stat-value">{{ value }}</div>
      <div class="stat-title">{{ title }}</div>
    </div>
    <div v-if="subtext" class="stat-footer">
      <span class="stat-subtext">{{ subtext }}</span>
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, required: true },
  value: { type: [String, Number], default: 0 },
  icon: { type: Object, required: true },
  iconBg: { type: String, default: '#eff6ff' },
  iconColor: { type: String, default: '#1e40af' },
  badge: { type: String, default: '' },
  badgeClass: { type: String, default: 'badge-primary' },
  subtext: { type: String, default: '' },
})
</script>

<style scoped>
.stat-card {
  background-color: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  box-shadow: var(--shadow-sm);
  transition: transform var(--transition-fast), box-shadow var(--transition-fast);
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.stat-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 0.2rem 0.5rem;
  border-radius: var(--radius-full);
}

.stat-body {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 24px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.02em;
}

.stat-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.stat-footer {
  padding-top: 0.5rem;
  border-top: 1px solid var(--border-light);
  font-size: 12px;
  color: var(--text-muted);
}
</style>
