<template>
  <div class="empty-state">
    <div class="empty-icon-wrapper">
      <component :is="icon || Inbox" :size="32" />
    </div>
    <h4 class="empty-title">{{ title }}</h4>
    <p class="empty-desc">{{ description }}</p>
    <div v-if="$slots.action" class="empty-action">
      <slot name="action" />
    </div>
  </div>
</template>

<script setup>
import { Inbox } from 'lucide-vue-next'

defineProps({
  title: { type: String, default: 'No records found' },
  description: { type: String, default: 'Get started by creating a new entry.' },
  icon: { type: Object, default: null },
})
</script>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3.5rem 1.5rem;
  text-align: center;
}

.empty-icon-wrapper {
  width: 56px;
  height: 56px;
  border-radius: var(--radius-full);
  background-color: var(--bg-card-muted);
  color: var(--text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

.empty-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.25rem;
}

.empty-desc {
  font-size: 13px;
  color: var(--text-secondary);
  max-width: 320px;
  margin-bottom: 1.25rem;
}
</style>
