<template>
  <div class="pagination-bar">
    <div class="pagination-info">
      Showing <strong>{{ fromItem }}</strong> to <strong>{{ toItem }}</strong> of <strong>{{ total }}</strong> records
    </div>

    <div class="pagination-actions">
      <div class="page-size-selector">
        <span>Rows:</span>
        <select :value="pageSize" @change="$emit('update:pageSize', Number($event.target.value))" class="page-select">
          <option :value="10">10</option>
          <option :value="25">25</option>
          <option :value="50">50</option>
        </select>
      </div>

      <div class="page-nav">
        <button
          class="page-btn"
          :disabled="page <= 1"
          @click="$emit('update:page', page - 1)"
          title="Previous Page"
        >
          <ChevronLeft :size="16" />
        </button>

        <span class="page-indicator">Page {{ page }} of {{ totalPages || 1 }}</span>

        <button
          class="page-btn"
          :disabled="page >= totalPages"
          @click="$emit('update:page', page + 1)"
          title="Next Page"
        >
          <ChevronRight :size="16" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

const props = defineProps({
  page: { type: Number, default: 1 },
  pageSize: { type: Number, default: 10 },
  total: { type: Number, default: 0 },
})

defineEmits(['update:page', 'update:pageSize'])

const totalPages = computed(() => Math.ceil(props.total / props.pageSize) || 1)
const fromItem = computed(() => (props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1))
const toItem = computed(() => Math.min(props.page * props.pageSize, props.total))
</script>

<style scoped>
.pagination-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.875rem 1.25rem;
  background-color: var(--bg-card);
  border-top: 1px solid var(--border-color);
  font-size: 13px;
  color: var(--text-secondary);
  flex-wrap: wrap;
  gap: 0.75rem;
}

.pagination-actions {
  display: flex;
  align-items: center;
  gap: 1.25rem;
}

.page-size-selector {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 12px;
}

.page-select {
  padding: 0.25rem 0.5rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  background-color: var(--bg-input);
  font-size: 12px;
  color: var(--text-primary);
  outline: none;
}

.page-nav {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.page-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border-color);
  background-color: var(--bg-card);
  color: var(--text-primary);
  cursor: pointer;
  transition: all var(--transition-fast);
}
.page-btn:hover:not(:disabled) {
  background-color: var(--bg-card-muted);
  border-color: #cbd5e1;
}
.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-indicator {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}
</style>
