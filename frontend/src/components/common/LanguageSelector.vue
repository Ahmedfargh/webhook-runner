<template>
  <div class="lang-selector-wrapper" ref="wrapperRef">
    <button class="lang-trigger-btn" @click="isOpen = !isOpen" :title="currentLanguage.name">
      <span class="flag-icon">{{ currentLanguage.flag }}</span>
      <span class="lang-code">{{ currentLanguage.code.toUpperCase() }}</span>
      <ChevronDown :size="13" class="chevron-icon" :class="{ 'chevron-rotated': isOpen }" />
    </button>

    <transition name="dropdown-anim">
      <div v-if="isOpen" class="lang-dropdown-menu">
        <button
          v-for="lang in languages"
          :key="lang.code"
          class="lang-option-btn"
          :class="{ 'lang-option-active': lang.code === currentLocale }"
          @click="selectLanguage(lang.code)"
        >
          <div class="lang-option-left">
            <span class="flag-icon-large">{{ lang.flag }}</span>
            <span class="lang-name">{{ lang.name }}</span>
          </div>
          <Check v-if="lang.code === currentLocale" :size="14" class="check-icon" />
        </button>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { languages, currentLocale, currentLanguage, setLocale } from '../../locales'
import { ChevronDown, Check } from 'lucide-vue-next'

const isOpen = ref(false)
const wrapperRef = ref(null)

function selectLanguage(code) {
  setLocale(code)
  isOpen.value = false
}

function handleClickOutside(e) {
  if (wrapperRef.value && !wrapperRef.value.contains(e.target)) {
    isOpen.value = false
  }
}

onMounted(() => {
  window.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  window.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.lang-selector-wrapper {
  position: relative;
  display: inline-block;
}

.lang-trigger-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.6rem;
  background-color: var(--bg-card-muted);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.lang-trigger-btn:hover {
  background-color: #ffffff;
  border-color: #cbd5e1;
}

.flag-icon {
  font-size: 15px;
  line-height: 1;
}

.lang-code {
  font-size: 11px;
  letter-spacing: 0.05em;
}

.chevron-icon {
  color: var(--text-muted);
  transition: transform var(--transition-fast);
}

.chevron-rotated {
  transform: rotate(180deg);
}

.lang-dropdown-menu {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  width: 170px;
  background-color: #ffffff;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  padding: 0.35rem;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

/* For RTL */
[dir="rtl"] .lang-dropdown-menu {
  right: auto;
  left: 0;
}

.lang-option-btn {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 0.625rem;
  border-radius: var(--radius-sm);
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background-color var(--transition-fast);
  width: 100%;
  text-align: left;
}

[dir="rtl"] .lang-option-btn {
  text-align: right;
}

.lang-option-btn:hover {
  background-color: var(--bg-card-muted);
}

.lang-option-active {
  background-color: var(--primary-light);
  color: var(--primary);
  font-weight: 600;
}

.lang-option-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.flag-icon-large {
  font-size: 17px;
  line-height: 1;
}

.check-icon {
  color: var(--primary);
}

/* Transitions */
.dropdown-anim-enter-active,
.dropdown-anim-leave-active {
  transition: all 0.15s cubic-bezier(0.4, 0, 0.2, 1);
}
.dropdown-anim-enter-from {
  opacity: 0;
  transform: translateY(-6px);
}
.dropdown-anim-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
