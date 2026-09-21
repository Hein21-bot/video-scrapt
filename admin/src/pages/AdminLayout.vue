<template>
  <div class="admin-shell">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="sidebar-logo">
        <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
          <path d="M4 6a2 2 0 012-2h12a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V6z"/>
          <path fill="#6366f1" d="M10 9l5 3-5 3V9z"/>
        </svg>
        <span>Admin</span>
      </div>

      <nav class="nav">
        <RouterLink to="/admin" class="nav-link" exact-active-class="active">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
          Dashboard
        </RouterLink>
        <RouterLink to="/admin/videos" class="nav-link" active-class="active">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M15 10l4.553-2.276A1 1 0 0121 8.723v6.554a1 1 0 01-1.447.894L15 14M3 8a2 2 0 012-2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V8z"/></svg>
          Videos
        </RouterLink>
        <RouterLink to="/admin/scrape" class="nav-link" active-class="active">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M23 4v6h-6M1 20v-6h6"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg>
          Scrape
        </RouterLink>
        <RouterLink to="/admin/categories" class="nav-link" active-class="active">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M4 6h16M4 12h10M4 18h6"/></svg>
          Sub-tabs
        </RouterLink>
        <RouterLink to="/admin/channels" class="nav-link" active-class="active">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><rect x="3" y="4" width="18" height="16" rx="2"/><path d="M3 10h18"/></svg>
          Channels
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <a :href="userSite" class="footer-link" target="_blank" rel="noopener">↗ User Site</a>
        <button class="logout-btn" @click="logout">Logout</button>
      </div>
    </aside>

    <!-- Main content -->
    <main class="content">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRouter, RouterLink, RouterView } from 'vue-router'
import { clearToken } from '../stores/auth'

const router = useRouter()
const userSite = import.meta.env.VITE_USER_URL ?? 'http://localhost:5173'

function logout() {
  clearToken()
  router.replace('/admin/login')
}
</script>

<style scoped>
.admin-shell {
  min-height: 100vh;
  display: flex;
  background: #0f172a;
  color: #f1f5f9;
}

.sidebar {
  width: 220px;
  min-height: 100vh;
  background: #1e293b;
  border-right: 1px solid #334155;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 20px;
  font-size: .95rem;
  font-weight: 700;
  border-bottom: 1px solid #334155;
}

.nav {
  display: flex;
  flex-direction: column;
  padding: 12px 8px;
  gap: 2px;
  flex: 1;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border-radius: 7px;
  text-decoration: none;
  color: #94a3b8;
  font-size: .875rem;
  transition: all .15s;
}
.nav-link:hover { color: #f1f5f9; background: #273549; }
.nav-link.active { color: #f1f5f9; background: #6366f1; }

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #334155;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.footer-link {
  color: #64748b;
  font-size: .8rem;
  text-decoration: none;
  transition: color .15s;
}
.footer-link:hover { color: #94a3b8; }
.logout-btn {
  background: transparent;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #94a3b8;
  font-size: .8rem;
  padding: 6px 10px;
  cursor: pointer;
  text-align: left;
  transition: all .15s;
}
.logout-btn:hover { border-color: #f87171; color: #f87171; }

.content {
  flex: 1;
  padding: 32px;
  overflow-y: auto;
}

@media (max-width: 640px) {
  .admin-shell { flex-direction: column; }
  .sidebar { width: 100%; min-height: auto; flex-direction: row; flex-wrap: wrap; }
  .nav { flex-direction: row; padding: 8px; }
  .content { padding: 16px; }
}
</style>
