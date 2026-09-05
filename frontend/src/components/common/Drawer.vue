<template>
  <teleport to="body">
    <transition name="drawer-backdrop">
      <div v-if="isOpen" class="drawer-backdrop" @click="close">
        <div class="drawer-panel" :style="{ maxWidth: width }" @click.stop>
          <div class="drawer-header">
            <div class="drawer-title-group">
              <h3 class="drawer-title">{{ title }}</h3>
              <p v-if="subtitle" class="drawer-subtitle">{{ subtitle }}</p>
            </div>
            <button class="drawer-close-btn" @click="close">
              <X :size="18" />
            </button>
          </div>

          <div class="drawer-body">
            <slot />
          </div>

          <div v-if="$slots.footer" class="drawer-footer">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup>
import { X } from 'lucide-vue-next'

const props = defineProps({
  isOpen: { type: Boolean, default: false },
  title: { type: String, default: '' },
  subtitle: { type: String, default: '' },
  width: { type: String, default: '480px' },
})

const emit = defineEmits(['close'])

function close() {
  emit('close')
}
</script>

<style scoped>
.drawer-backdrop {
  position: fixed;
  inset: 0;
  background-color: rgba(15, 23, 42, 0.5);
  backdrop-filter: blur(4px);
  z-index: 1000;
  display: flex;
  justify-content: flex-end;
}

.drawer-panel {
  width: 100%;
  height: 100%;
  background-color: var(--bg-card);
  box-shadow: var(--shadow-drawer);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: slideInRight 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.drawer-header {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  background-color: #fafbfc;
}

.drawer-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
}

.drawer-subtitle {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 0.2rem;
}

.drawer-close-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.35rem;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
}
.drawer-close-btn:hover {
  color: var(--text-primary);
  background-color: #f1f5f9;
}

.drawer-body {
  padding: 1.5rem;
  flex: 1;
  overflow-y: auto;
}

.drawer-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border-color);
  background-color: #fafbfc;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.75rem;
}

/* Transitions */
.drawer-backdrop-enter-active,
.drawer-backdrop-leave-active {
  transition: opacity 0.25s ease;
}
.drawer-backdrop-enter-from,
.drawer-backdrop-leave-to {
  opacity: 0;
}
</style>
