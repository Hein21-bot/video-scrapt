import { ref } from 'vue'

export type Theme = 'dark' | 'light'

const KEY = 'theme'

function stored(): Theme {
  try {
    return localStorage.getItem(KEY) === 'light' ? 'light' : 'dark'
  } catch {
    return 'dark'
  }
}

export const theme = ref<Theme>(stored())

function apply(t: Theme) {
  document.documentElement.dataset.theme = t
  document.querySelector('meta[name="theme-color"]')?.setAttribute('content', t === 'light' ? '#ff9000' : '#000000')
}

export function initTheme() {
  apply(theme.value)
}

export function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  apply(theme.value)
  try {
    localStorage.setItem(KEY, theme.value)
  } catch { /* private mode — choice just isn't remembered */ }
}
