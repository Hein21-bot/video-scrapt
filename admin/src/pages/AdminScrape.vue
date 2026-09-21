<template>
  <div class="admin-scrape">
    <h1 class="page-title">Scrape / Sync</h1>

    <!-- Global status card -->
    <div class="status-card" :class="{ 'status-card--running': status.running }">
      <div class="status-left">
        <span class="dot" :class="status.running ? 'dot--pulse' : 'dot--gray'" />
        <div>
          <p class="status-label">{{ status.running ? 'Sync is running…' : 'Sync is idle' }}</p>
          <p class="status-sub">Last sync: {{ formatDate(status.last_sync_time) }}</p>
        </div>
      </div>
      <button class="sync-btn" :disabled="status.running || syncing" @click="triggerSyncAll">
        <span v-if="status.running || syncing" class="spinner" />
        <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="15" height="15">
          <path d="M23 4v6h-6M1 20v-6h6"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/>
        </svg>
        {{ status.running ? 'Running…' : 'Sync All' }}
      </button>
    </div>

    <p v-if="msg" class="msg" :class="{ 'msg--err': isErr }">{{ msg }}</p>

    <!-- Sync targets with per-row sync buttons -->
    <div class="section">
      <h2 class="section-title">Sync Targets</h2>
      <div class="targets">
        <div v-for="t in targets" :key="t.site + t.category" class="target-row">
          <span class="target-site">{{ t.site }}</span>
          <span class="target-cat">{{ t.category }}</span>
          <span class="target-url">{{ t.url }}</span>
          <button
            class="row-sync-btn"
            :disabled="status.running || runningTarget === t.site + '/' + t.category"
            @click="triggerSyncOne(t)"
            title="Sync this target only"
          >
            <span v-if="runningTarget === t.site + '/' + t.category" class="spinner-sm" />
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="12" height="12">
              <path d="M23 4v6h-6M1 20v-6h6"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Notes -->
    <div class="notes">
      <p>• Sync runs automatically every <strong>6 hours</strong> in the background.</p>
      <p>• Click <strong>Sync All</strong> to refresh every target at once.</p>
      <p>• Click the <strong>↻</strong> button on a row to sync only that category.</p>
      <p>• When deploying to Render, run sync from your <strong>local machine</strong> so it pushes to MongoDB Atlas.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { authHeader } from '../stores/auth'

interface SyncStatus {
  running: boolean
  last_sync_time: string
}

interface Target {
  site: string
  category: string
  url: string
}

const status        = ref<SyncStatus>({ running: false, last_sync_time: '' })
const syncing       = ref(false)
const runningTarget = ref('')   // "site/category" of the currently running single-target sync
const msg           = ref('')
const isErr         = ref(false)

const targets: Target[] = [
  { site: '3xchina', category: 'all', url: 'https://3xchina.page/' },
  { site: 'muskuduu', category: 'all', url: 'https://muskuduu.com/' },
]

function formatDate(iso: string) {
  if (!iso || iso.startsWith('0001')) return 'Never'
  return new Date(iso).toLocaleString()
}

async function fetchStatus() {
  try {
    const res  = await fetch('/api/admin/sync/status', { headers: authHeader() })
    const data = await res.json()
    if (res.ok) {
      status.value = data
      // clear runningTarget when backend says idle
      if (!data.running) runningTarget.value = ''
    }
  } catch { /* ignore */ }
}

async function triggerSyncAll() {
  syncing.value = true
  msg.value     = ''
  isErr.value   = false
  try {
    const res  = await fetch('/api/admin/sync', { method: 'POST', headers: authHeader() })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error ?? 'Failed')
    msg.value = data.message ?? 'Sync All started'
    fetchStatus()
  } catch (e) {
    msg.value = e instanceof Error ? e.message : 'Failed to start sync'
    isErr.value = true
  } finally {
    syncing.value = false
  }
}

async function triggerSyncOne(t: Target) {
  msg.value           = ''
  isErr.value         = false
  runningTarget.value = t.site + '/' + t.category
  try {
    const res  = await fetch('/api/admin/sync/target', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify({ site: t.site, category: t.category, url: t.url }),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error ?? 'Failed')
    msg.value = data.message ?? 'Sync started'
    fetchStatus()
  } catch (e) {
    msg.value   = e instanceof Error ? e.message : 'Failed'
    isErr.value = true
    runningTarget.value = ''
  }
}

let pollTimer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  fetchStatus()
  pollTimer = setInterval(fetchStatus, 4000)
})

onBeforeUnmount(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
.admin-scrape { display: flex; flex-direction: column; gap: 24px; max-width: 860px; }
.page-title { margin: 0; font-size: 1.4rem; font-weight: 700; color: #f1f5f9; }

.status-card {
  background: #1e293b; border: 1px solid #334155; border-radius: 10px;
  padding: 20px 24px; display: flex; align-items: center;
  justify-content: space-between; gap: 16px; transition: border-color .3s;
}
.status-card--running { border-color: #6366f1; }
.status-left { display: flex; align-items: center; gap: 14px; }
.dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
.dot--gray  { background: #475569; }
.dot--pulse {
  background: #6366f1;
  box-shadow: 0 0 0 0 rgba(99,102,241,.6);
  animation: pulse 1.4s infinite;
}
@keyframes pulse {
  0%   { box-shadow: 0 0 0 0 rgba(99,102,241,.6); }
  70%  { box-shadow: 0 0 0 8px rgba(99,102,241,0); }
  100% { box-shadow: 0 0 0 0 rgba(99,102,241,0); }
}
.status-label { margin: 0; font-size: .9rem; color: #f1f5f9; font-weight: 500; }
.status-sub   { margin: 2px 0 0; font-size: .78rem; color: #64748b; }

.sync-btn {
  display: flex; align-items: center; gap: 7px;
  padding: 9px 18px; border-radius: 8px; border: none;
  background: #6366f1; color: #fff; font-size: .875rem; font-weight: 600;
  cursor: pointer; white-space: nowrap; transition: background .15s;
}
.sync-btn:hover:not(:disabled) { background: #4f46e5; }
.sync-btn:disabled { opacity: .6; cursor: not-allowed; }

.spinner    { width: 14px; height: 14px; border: 2px solid rgba(255,255,255,.3); border-top-color: #fff; border-radius: 50%; animation: spin .7s linear infinite; }
.spinner-sm { width: 11px; height: 11px; border: 2px solid rgba(255,255,255,.3); border-top-color: #fff; border-radius: 50%; animation: spin .7s linear infinite; display: inline-block; }
@keyframes spin { to { transform: rotate(360deg); } }

.msg { font-size: .875rem; padding: 10px 14px; border-radius: 7px; background: rgba(99,102,241,.1); border: 1px solid rgba(99,102,241,.2); color: #a5b4fc; margin: 0; }
.msg--err { background: rgba(248,113,113,.1); border-color: rgba(248,113,113,.2); color: #f87171; }

.section { display: flex; flex-direction: column; gap: 10px; }
.section-title { margin: 0; font-size: .8rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: .05em; }

.targets { display: flex; flex-direction: column; gap: 2px; background: #1e293b; border: 1px solid #334155; border-radius: 8px; overflow: hidden; }
.target-row {
  display: flex; align-items: center; gap: 10px;
  padding: 8px 14px; font-size: .82rem;
  border-bottom: 1px solid #1a2744;
}
.target-row:last-child { border-bottom: none; }
.target-site { width: 110px; color: #818cf8; font-weight: 500; flex-shrink: 0; }
.target-cat  { width: 80px;  color: #64748b; flex-shrink: 0; }
.target-url  { flex: 1; color: #475569; font-size: .78rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.row-sync-btn {
  flex-shrink: 0;
  display: flex; align-items: center; justify-content: center; gap: 4px;
  padding: 4px 10px; border-radius: 5px; border: 1px solid #334155;
  background: transparent; color: #64748b;
  font-size: .75rem; cursor: pointer; transition: all .15s;
  white-space: nowrap;
}
.row-sync-btn:hover:not(:disabled) { border-color: #6366f1; color: #a5b4fc; background: rgba(99,102,241,.08); }
.row-sync-btn:disabled { opacity: .4; cursor: not-allowed; }

.notes { display: flex; flex-direction: column; gap: 6px; }
.notes p { margin: 0; font-size: .82rem; color: #64748b; }
.notes strong { color: #94a3b8; }
</style>
