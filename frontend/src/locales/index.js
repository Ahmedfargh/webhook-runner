import { ref, computed } from 'vue'
import en from './en'
import ar from './ar'
import fr from './fr'
import de from './de'
import ru from './ru'

const messages = { en, ar, fr, de, ru }

export const languages = [
  { code: 'en', name: 'English', flag: '🇬🇧', dir: 'ltr' },
  { code: 'ar', name: 'العربية', flag: '🇸🇦', dir: 'rtl' },
  { code: 'fr', name: 'Français', flag: '🇫🇷', dir: 'ltr' },
  { code: 'de', name: 'Deutsch', flag: '🇩🇪', dir: 'ltr' },
  { code: 'ru', name: 'Русский', flag: '🇷🇺', dir: 'ltr' },
]

const savedLocale = localStorage.getItem('app_locale') || 'en'
export const currentLocale = ref(savedLocale)

export const currentLanguage = computed(() => {
  return languages.find((l) => l.code === currentLocale.value) || languages[0]
})

export const isRTL = computed(() => currentLanguage.value.dir === 'rtl')

export function setLocale(langCode) {
  const target = languages.find((l) => l.code === langCode)
  if (!target) return

  currentLocale.value = target.code
  localStorage.setItem('app_locale', target.code)

  // Update HTML document direction and language attributes
  document.documentElement.setAttribute('dir', target.dir)
  document.documentElement.setAttribute('lang', target.code)
  if (target.dir === 'rtl') {
    document.documentElement.classList.add('rtl-mode')
  } else {
    document.documentElement.classList.remove('rtl-mode')
  }
}

// Initial setup of document direction
document.documentElement.setAttribute('dir', currentLanguage.value.dir)
document.documentElement.setAttribute('lang', currentLanguage.value.code)
if (currentLanguage.value.dir === 'rtl') {
  document.documentElement.classList.add('rtl-mode')
}

// Recursive key lookup helper
function getNestedValue(obj, keyPath) {
  if (!obj || !keyPath) return null
  return keyPath.split('.').reduce((prev, curr) => (prev ? prev[curr] : null), obj)
}

export function t(key, params = {}) {
  const currentMsg = messages[currentLocale.value]
  const fallbackMsg = messages.en

  let val = getNestedValue(currentMsg, key)
  if (val === undefined || val === null) {
    val = getNestedValue(fallbackMsg, key)
  }

  if (typeof val !== 'string') {
    return key
  }

  // Replace {param} placeholders
  let result = val
  for (const [pKey, pVal] of Object.entries(params)) {
    result = result.replaceAll(`{${pKey}}`, pVal)
  }

  return result
}

// Composable hook for Vue 3 Script Setup
export function useI18n() {
  return {
    t,
    currentLocale,
    setLocale,
    languages,
    currentLanguage,
    isRTL,
  }
}

// Vue plugin installer
export default {
  install(app) {
    app.config.globalProperties.$t = t
    app.config.globalProperties.$currentLocale = currentLocale
    app.provide('i18n', { t, currentLocale, setLocale, languages, currentLanguage, isRTL })
  },
}

