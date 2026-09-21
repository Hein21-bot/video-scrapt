import { ref, computed } from 'vue'

const TOKEN_KEY = 'admin_token'

const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))

export const isAdmin = computed(() => !!token.value)

export function setToken(t: string) {
  token.value = t
  localStorage.setItem(TOKEN_KEY, t)
}

export function clearToken() {
  token.value = null
  localStorage.removeItem(TOKEN_KEY)
}

export function authHeader(): Record<string, string> {
  return token.value ? { Authorization: `Bearer ${token.value}` } : {}
}
