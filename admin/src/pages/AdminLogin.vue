<template>
  <div class="login-page">
    <form class="login-card" @submit.prevent="submit">
      <div class="login-logo">
        <svg viewBox="0 0 24 24" fill="currentColor" width="28" height="28">
          <path d="M4 6a2 2 0 012-2h12a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V6z"/>
          <path fill="#6366f1" d="M10 9l5 3-5 3V9z"/>
        </svg>
        <span>Admin Panel</span>
      </div>

      <div class="field">
        <label>Password</label>
        <input
          v-model="password"
          type="password"
          placeholder="Enter admin password"
          class="input"
          :disabled="loading"
          autofocus
        />
      </div>

      <p v-if="error" class="error-msg">{{ error }}</p>

      <button type="submit" class="login-btn" :disabled="loading || !password">
        <span v-if="loading" class="spinner" />
        <span v-else>Login</span>
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { setToken } from '../stores/auth'

const router   = useRouter()
const password = ref('')
const loading  = ref(false)
const error    = ref('')

async function submit() {
  loading.value = true
  error.value   = ''
  try {
    const res  = await fetch('/api/admin/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ password: password.value }),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error ?? 'Login failed')
    setToken(data.token)
    router.replace('/admin')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0f172a;
}
.login-card {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 12px;
  padding: 40px;
  width: 100%;
  max-width: 380px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.login-logo {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 1.2rem;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 4px;
}
.field { display: flex; flex-direction: column; gap: 6px; }
.field label { font-size: .8rem; color: #94a3b8; font-weight: 500; }
.input {
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 10px 14px;
  color: #f1f5f9;
  font-size: .9rem;
  outline: none;
  transition: border-color .15s;
}
.input:focus { border-color: #6366f1; }
.error-msg { color: #f87171; font-size: .85rem; margin: 0; }
.login-btn {
  padding: 11px;
  border-radius: 8px;
  border: none;
  background: #6366f1;
  color: #fff;
  font-size: .9rem;
  font-weight: 600;
  cursor: pointer;
  transition: background .15s;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}
.login-btn:hover:not(:disabled) { background: #4f46e5; }
.login-btn:disabled { opacity: .6; cursor: not-allowed; }
.spinner {
  width: 16px; height: 16px;
  border: 2px solid rgba(255,255,255,.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
</style>
