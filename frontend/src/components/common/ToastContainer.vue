<template>
  <div class="toast-stack">
    <transition-group name="toast-anim">
      <div
        v-for="toast in toastStore.toasts"
        :key="toast.id"
        class="toast-item"
        :class="`toast-${toast.type}`"
      >
        <div class="toast-content">
          <span class="toast-icon">
            <CheckCircle2 v-if="toast.type === 'success'" :size="18" />
            <AlertCircle v-else-if="toast.type === 'danger'" :size="18" />
            <AlertTriangle v-else-if="toast.type === 'warning'" :size="18" />
            <Info v-else :size="18" />
          </span>
          <span class="toast-message">{{ toast.message }}</span>
        </div>
        <button class="toast-close" @click="toastStore.remove(toast.id)">
          <X :size="14" />
        </button>
      </div>
    </transition-group>
  </div>
</template>

<script setup>
import { useToastStore } from '../../stores/toast'
import { CheckCircle2, AlertCircle, AlertTriangle, Info, X } from 'lucide-vue-next'

const toastStore = useToastStore()
</script>

<style scoped>
.toast-stack {
  position: fixed;
  top: 1.5rem;
  right: 1.5rem;
  z-index: 9999;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  max-width: 400px;
  width: calc(100vw - 3rem);
  pointer-events: none;
}

.toast-item {
  pointer-events: auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.875rem 1rem;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  background-color: #ffffff;
  border: 1px solid var(--border-color);
  font-size: 13px;
  color: var(--text-primary);
  backdrop-filter: blur(8px);
}

.toast-content {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.toast-success {
  border-left: 4px solid var(--success);
}
.toast-success .toast-icon {
  color: var(--success);
}

.toast-danger {
  border-left: 4px solid var(--danger);
}
.toast-danger .toast-icon {
  color: var(--danger);
}

.toast-warning {
  border-left: 4px solid var(--warning);
}
.toast-warning .toast-icon {
  color: var(--warning);
}

.toast-info {
  border-left: 4px solid var(--info);
}
.toast-info .toast-icon {
  color: var(--info);
}

.toast-close {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}
.toast-close:hover {
  color: var(--text-primary);
  background-color: #f1f5f9;
}

/* Animations */
.toast-anim-enter-active,
.toast-anim-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.toast-anim-enter-from {
  opacity: 0;
  transform: translateX(30px);
}
.toast-anim-leave-to {
  opacity: 0;
  transform: scale(0.9);
}
</style>
