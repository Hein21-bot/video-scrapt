<template>
  <div class="dashboard">
    <h1 class="page-title">Dashboard</h1>

    <div v-if="loading" class="loading-row">
      <div class="spinner" />
    </div>

    <template v-else-if="stats">
      <!-- Stat cards -->
      <div class="stat-grid">
        <div class="stat-card">
          <p class="stat-label">Total Videos</p>
          <p class="stat-value">{{ stats.total_videos.toLocaleString() }}</p>
        </div>
        <div class="stat-card">
          <p class="stat-label">Cache Entries</p>
          <p class="stat-value">{{ stats.cache_entries }}</p>
        </div>
        <div class="stat-card">
          <p class="stat-label">Share Links</p>
          <p class="stat-value">{{ stats.share_links }}</p>
        </div>
        <div class="stat-card" :class="{ 'stat-card--active': stats.sync_running }">
          <p class="stat-label">Sync Status</p>
          <p class="stat-value sync-status">
            <span class="dot" :class="stats.sync_running ? 'dot--green' : 'dot--gray'" />
            {{ stats.sync_running ? 'Running' : 'Idle' }}
          </p>
        </div>
      </div>

      <!-- By site breakdown -->
      <div class="section">
        <h2 class="section-title">Videos by Site</h2>
        <div class="site-list">
          <div v-for="row in stats.by_site" :key="row.site" class="site-row">
            <span class="site-name">{{ row.site }}</span>
            <div class="site-bar-wrap">
              <div class="site-bar" :style="{ width: barWidth(row.count) + '%' }" />
            </div>
            <span class="site-count">{{ row.count.toLocaleString() }}</span>
          </div>
        </div>
      </div>

      <!-- Last sync -->
      <p v-if="stats.last_sync" class="last-sync">
        Last sync: {{ formatDate(stats.last_sync) }}
      </p>
    </template>

    <p v-else-if="error" class="error-msg">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { authHeader } from '../stores/auth'

interface Stats {
  total_videos: number
  cache_entries: number
  share_links: number
  sync_running: boolean
  last_sync: string
  by_site: { site: string; count: number }[]
}

const stats   = ref<Stats | null>(null)
const loading = ref(true)
const error   = ref('')

function barWidth(count: number) {
  if (!stats.value) return 0
  const max = Math.max(...stats.value.by_site.map(r => r.count))
  return max ? Math.round((count / max) * 100) : 0
}

function formatDate(iso: string) {
  if (!iso || iso.startsWith('0001')) return 'Never'
  return new Date(iso).toLocaleString()
}

onMounted(async () => {
  try {
    const res  = await fetch('/api/admin/stats', { headers: authHeader() })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error)
    stats.value = data
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load stats'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.dashboard { display: flex; flex-direction: column; gap: 28px; max-width: 800px; }
.page-title { margin: 0; font-size: 1.4rem; font-weight: 700; color: #f1f5f9; }

.loading-row { display: flex; justify-content: center; padding: 40px 0; }
.spinner { width: 32px; height: 32px; border: 3px solid rgba(99,102,241,.3); border-top-color: #6366f1; border-radius: 50%; animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.stat-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 16px; }
.stat-card {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 10px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.stat-card--active { border-color: #22c55e; }
.stat-label { margin: 0; font-size: .78rem; color: #64748b; text-transform: uppercase; letter-spacing: .05em; }
.stat-value { margin: 0; font-size: 1.6rem; font-weight: 700; color: #f1f5f9; }

.sync-status { display: flex; align-items: center; gap: 8px; font-size: 1rem; }
.dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.dot--green { background: #22c55e; box-shadow: 0 0 6px #22c55e; }
.dot--gray  { background: #475569; }

.section { display: flex; flex-direction: column; gap: 12px; }
.section-title { margin: 0; font-size: .9rem; font-weight: 600; color: #94a3b8; text-transform: uppercase; letter-spacing: .05em; }

.site-list { display: flex; flex-direction: column; gap: 10px; }
.site-row { display: flex; align-items: center; gap: 12px; }
.site-name { width: 120px; font-size: .85rem; color: #cbd5e1; flex-shrink: 0; }
.site-bar-wrap { flex: 1; height: 8px; background: #334155; border-radius: 4px; overflow: hidden; }
.site-bar { height: 100%; background: #6366f1; border-radius: 4px; transition: width .4s; }
.site-count { width: 60px; text-align: right; font-size: .85rem; color: #94a3b8; flex-shrink: 0; }

.last-sync { margin: 0; font-size: .8rem; color: #475569; }
.error-msg { color: #f87171; font-size: .875rem; }
</style>
