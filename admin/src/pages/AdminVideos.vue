<template>
  <div class="admin-videos">
    <h1 class="page-title">Videos</h1>

    <!-- Add from Telegram (local bridge) -->
    <div class="add-card">
      <h2 class="add-title">📥 Add video — from Telegram</h2>
      <p v-if="!bridgeOnline" class="add-msg err">
        Bridge offline — start <code>bridge/server.py</code> and set <code>BRIDGE_URL</code>.
      </p>
      <form v-else @submit.prevent="tgAdd">
        <div class="add-grid">
          <input v-model="tg.link" class="add-input" placeholder="https://t.me/channel/7661 *" />
          <input v-model="tg.title" class="add-input" placeholder="Title (optional — uses caption)" />
          <select v-model="tg.channel" class="add-input">
            <option v-for="ch in channels" :key="ch.key" :value="ch.key">{{ ch.label }}</option>
          </select>
          <select v-if="mmCats.length" v-model="tg.category" class="add-input">
            <option v-for="c in mmCats" :key="c.value" :value="c.value">{{ c.label }}</option>
          </select>
          <input v-else v-model="tg.category" class="add-input" placeholder="Category (no sub-tabs set up yet)" />
          <input v-model="tg.actors" class="add-input" placeholder="Actress (optional — comma separated)" />
          <input v-model="tg.tags" class="add-input" placeholder="Tags (optional — comma separated)" />
        </div>
        <div class="add-row">
          <button type="submit" class="add-btn" :disabled="tgAdding">{{ tgAdding ? '…' : 'Queue' }}</button>
          <span v-if="tgMsg" :class="['add-msg', { err: tgErr }]">{{ tgMsg }}</span>
          <RouterLink to="/admin/channels" class="manage-link">Manage channels →</RouterLink>
          <RouterLink to="/admin/categories" class="manage-link">Manage sub-tabs →</RouterLink>
        </div>
      </form>

      <table v-if="jobs.length" class="job-table">
        <tr v-for="j in jobs" :key="j.id">
          <td><span :class="['job-status', j.status]">{{ j.status }}</span></td>
          <td class="job-msg">{{ j.message || j.link }}</td>
        </tr>
      </table>
    </div>

    <!-- Filters -->
    <div class="filters">
      <input v-model="q" class="filter-input" placeholder="Search title…" @input="onSearch" />
      <select v-model="selectedSite" class="filter-select" @change="load(1)">
        <option value="">All Sites</option>
        <option value="3xchina">Chinese AV (3xchina)</option>
        <option value="manual">Myanmar</option>
        <option value="muskuduu">Muskuduu</option>
        <option v-for="ch in customChannels" :key="ch.key" :value="ch.key">{{ ch.label }}</option>
      </select>
      <span class="total-label">{{ total.toLocaleString() }} videos</span>
    </div>

    <!-- Table -->
    <div class="table-wrap">
      <table class="table" v-if="items.length > 0">
        <thead>
          <tr>
            <th>Thumbnail</th>
            <th>Title</th>
            <th>Site</th>
            <th>Category</th>
            <th>Synced At</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="v in items" :key="v.page_url">
            <td class="thumb-cell">
              <img v-if="v.thumbnail" :src="v.thumbnail" class="thumb" loading="lazy" />
              <div v-else class="thumb-ph" />
            </td>
            <td class="title-cell">
              <a :href="v.page_url" target="_blank" class="title-link">{{ v.title }}</a>
            </td>
            <td><span class="badge">{{ v.site }}</span></td>
            <td class="dim">{{ v.category }}</td>
            <td class="dim">{{ formatDate(v.synced_at) }}</td>
            <td>
              <button class="del-btn" @click="deleteVideo(v.page_url)" title="Delete">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
                  <path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6"/>
                </svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else-if="loading" class="center"><div class="spinner" /></div>
      <div v-else class="center dim">No videos found.</div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="pagination">
      <button class="page-btn" :disabled="page <= 1" @click="load(page - 1)">←</button>
      <span class="page-info">{{ page }} / {{ totalPages }}</span>
      <button class="page-btn" :disabled="page >= totalPages" @click="load(page + 1)">→</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { authHeader } from '../stores/auth'

interface VideoItem {
  page_url: string
  title: string
  thumbnail: string
  site: string
  category: string
  synced_at: string
}

const q            = ref('')
const selectedSite = ref('')
const items        = ref<VideoItem[]>([])
const total        = ref(0)
const page         = ref(1)
const loading      = ref(false)

const totalPages = computed(() => Math.ceil(total.value / 20))

let searchTimer: ReturnType<typeof setTimeout> | null = null
function onSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => load(1), 350)
}

function formatDate(iso: string) {
  if (!iso || iso.startsWith('0001')) return '-'
  return new Date(iso).toLocaleDateString()
}

async function load(pg: number) {
  page.value    = pg
  loading.value = true
  try {
    const params = new URLSearchParams({ page: String(pg) })
    if (selectedSite.value) params.set('site', selectedSite.value)
    if (q.value.trim()) params.set('q', q.value.trim())
    const res  = await fetch(`/api/admin/videos?${params}`, { headers: authHeader() })
    const data = await res.json()
    items.value = data.items ?? []
    total.value = data.total ?? 0
  } finally {
    loading.value = false
  }
}

async function deleteVideo(pageUrl: string) {
  if (!confirm('Delete this video?')) return
  const res = await fetch('/api/admin/videos', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json', ...authHeader() },
    body: JSON.stringify({ page_url: pageUrl }),
  })
  if (res.ok) {
    items.value = items.value.filter(v => v.page_url !== pageUrl)
    total.value--
  }
}

// ── Telegram bridge ──────────────────────────────────────────────────────────
interface Job { id: number; status: string; message: string; link: string }
const bridgeOnline = ref(false)
const jobs     = ref<Job[]>([])
const tg       = ref({ link: '', title: '', channel: 'channel1', category: '', actors: '', tags: '' })
const tgAdding = ref(false)
const tgMsg    = ref('')
const tgErr    = ref(false)
let jobTimer: ReturnType<typeof setInterval> | null = null

interface ChannelOpt { key: string; label: string; builtin: boolean }
const channels       = ref<ChannelOpt[]>([])
const customChannels = computed(() => channels.value.filter(c => !c.builtin))
async function loadChannels() {
  try {
    const r = await fetch('/api/admin/channels', { headers: authHeader() })
    if (r.ok) channels.value = await r.json()
  } catch { /* keep empty — filter/channel selects just show the built-ins */ }
}

interface MMCat { value: string; label: string }
const mmCats = ref<MMCat[]>([])
async function loadMMCats() {
  try {
    const r = await fetch(`/api/admin/categories?site=${tg.value.channel}`, { headers: authHeader() })
    mmCats.value = r.ok ? await r.json() : []
  } catch {
    mmCats.value = []
  }
}
watch(mmCats, list => { if (list.length && !tg.value.category) tg.value.category = list[0].value })
watch(() => tg.value.channel, () => { tg.value.category = ''; loadMMCats() })

async function checkBridge() {
  try {
    const r = await fetch('/api/admin/bridge/status', { headers: authHeader() })
    const d = await r.json()
    bridgeOnline.value = !!(d.enabled && d.online)
  } catch { bridgeOnline.value = false }
  if (bridgeOnline.value && !jobTimer) {
    loadJobs()
    jobTimer = setInterval(loadJobs, 3000)
  }
}

async function loadJobs() {
  try {
    const r = await fetch('/api/admin/bridge/jobs', { headers: authHeader() })
    if (r.ok) jobs.value = (await r.json()).jobs ?? []
  } catch { /* bridge went offline */ }
}

async function tgAdd() {
  tgAdding.value = true
  tgMsg.value = ''
  try {
    const r = await fetch('/api/admin/bridge/add', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify(tg.value),
    })
    const d = await r.json()
    if (!r.ok) throw new Error(d.error ?? 'Failed to queue')
    tgErr.value = false
    tgMsg.value = `Queued #${d.id}`
    tg.value = { link: '', title: '', channel: tg.value.channel, category: tg.value.category, actors: '', tags: '' }
    loadJobs()
  } catch (e) {
    tgErr.value = true
    tgMsg.value = e instanceof Error ? e.message : 'Failed'
  } finally {
    tgAdding.value = false
  }
}

onMounted(() => { load(1); checkBridge(); loadChannels(); loadMMCats() })
onUnmounted(() => { if (jobTimer) clearInterval(jobTimer) })
</script>

<style scoped>
.admin-videos { display: flex; flex-direction: column; gap: 20px; }
.page-title { margin: 0; font-size: 1.4rem; font-weight: 700; color: #f1f5f9; }

.add-card { background: #1e293b; border: 1px solid #334155; border-radius: 10px; padding: 16px; display: flex; flex-direction: column; gap: 12px; }
.add-title { margin: 0; font-size: .95rem; font-weight: 600; color: #f1f5f9; }
.add-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.add-input { background: #0f172a; border: 1px solid #334155; border-radius: 7px; padding: 8px 12px; color: #f1f5f9; font-size: .85rem; outline: none; transition: border-color .15s; }
.add-input:focus { border-color: #6366f1; }
.add-row { display: flex; align-items: center; gap: 12px; }
.add-btn { background: #6366f1; border: none; border-radius: 7px; color: #fff; font-size: .85rem; font-weight: 600; padding: 8px 20px; cursor: pointer; transition: background .15s; }
.add-btn:hover:not(:disabled) { background: #4f46e5; }
.add-btn:disabled { opacity: .6; cursor: not-allowed; }
.add-msg { font-size: .82rem; color: #4ade80; }
.add-msg.err { color: #f87171; }
.manage-link { margin-left: auto; font-size: .78rem; color: #818cf8; text-decoration: none; }
.manage-link:hover { text-decoration: underline; }
@media (max-width: 700px) { .add-grid { grid-template-columns: 1fr; } }

.tg-grid { grid-template-columns: 2fr 1.5fr auto; align-items: center; }
.job-table { width: 100%; border-collapse: collapse; margin-top: 4px; font-size: .8rem; }
.job-table td { padding: 5px 8px; border-top: 1px solid #334155; color: #94a3b8; }
.job-msg { color: #cbd5e1; }
.job-status { font-size: .7rem; font-weight: 700; text-transform: uppercase; padding: 2px 7px; border-radius: 4px; }
.job-status.queued     { background: #334155; color: #cbd5e1; }
.job-status.processing { background: rgba(251,191,36,.15); color: #fbbf24; }
.job-status.done       { background: rgba(74,222,128,.15); color: #4ade80; }
.job-status.failed     { background: rgba(248,113,113,.15); color: #f87171; }

.filters { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.filter-input {
  background: #1e293b; border: 1px solid #334155; border-radius: 7px;
  padding: 8px 12px; color: #f1f5f9; font-size: .85rem; outline: none;
  transition: border-color .15s; width: 220px;
}
.filter-input:focus { border-color: #6366f1; }
.filter-select {
  background: #1e293b; border: 1px solid #334155; border-radius: 7px;
  padding: 8px 12px; color: #f1f5f9; font-size: .85rem; outline: none; cursor: pointer;
}
.filter-select option { background: #1e293b; }
.total-label { font-size: .8rem; color: #64748b; margin-left: auto; }

.table-wrap { background: #1e293b; border: 1px solid #334155; border-radius: 10px; overflow: auto; }
.table { width: 100%; border-collapse: collapse; font-size: .83rem; }
.table th {
  padding: 10px 14px; text-align: left; font-size: .72rem;
  color: #64748b; text-transform: uppercase; letter-spacing: .05em;
  border-bottom: 1px solid #334155; white-space: nowrap;
}
.table td { padding: 10px 14px; border-bottom: 1px solid #1e293b; vertical-align: middle; color: #cbd5e1; }
.table tr:last-child td { border-bottom: none; }
.table tr:hover td { background: rgba(99,102,241,.05); }

.thumb-cell { width: 80px; }
.thumb { width: 72px; height: 40px; object-fit: cover; border-radius: 4px; display: block; }
.thumb-ph { width: 72px; height: 40px; background: #334155; border-radius: 4px; }
.title-cell { max-width: 280px; }
.title-link { color: #a5b4fc; text-decoration: none; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; line-height: 1.4; }
.title-link:hover { text-decoration: underline; }
.badge { background: rgba(99,102,241,.15); color: #818cf8; font-size: .72rem; padding: 2px 7px; border-radius: 4px; white-space: nowrap; }
.dim { color: #64748b; }
.del-btn {
  background: transparent; border: 1px solid #334155; border-radius: 5px;
  color: #64748b; cursor: pointer; padding: 5px 7px;
  display: flex; align-items: center; transition: all .15s;
}
.del-btn:hover { border-color: #f87171; color: #f87171; background: rgba(248,113,113,.08); }

.center { display: flex; justify-content: center; padding: 40px; }
.spinner { width: 28px; height: 28px; border: 3px solid rgba(99,102,241,.3); border-top-color: #6366f1; border-radius: 50%; animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.pagination { display: flex; align-items: center; gap: 12px; justify-content: center; }
.page-btn {
  background: #1e293b; border: 1px solid #334155; border-radius: 6px;
  color: #94a3b8; padding: 6px 14px; cursor: pointer; font-size: .875rem; transition: all .15s;
}
.page-btn:hover:not(:disabled) { border-color: #6366f1; color: #f1f5f9; }
.page-btn:disabled { opacity: .4; cursor: not-allowed; }
.page-info { font-size: .85rem; color: #64748b; }
</style>
