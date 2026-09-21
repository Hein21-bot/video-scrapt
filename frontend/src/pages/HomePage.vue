<template>
  <div class="app">
    <header class="header">
      <!-- Row 1: Logo + Search -->
      <div class="header-top">
        <div class="header-top-inner">
          <div class="logo">
            <img src="/logo.png" class="logo-img" alt="ChitNya" width="36" height="36" />
            <span class="logo-text">ChitNya</span>
          </div>

          <div class="search-wrap">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
              <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
            </svg>
            <input
              v-model="searchQuery"
              class="search-input"
              :placeholder="t('home.searchPlaceholder')"
              enterkeyhint="search"
              @keydown.enter="runSearch"
              @keydown.escape="clearSearch"
            />
            <button v-if="searchQuery" class="search-clear" @click="clearSearch">✕</button>
            <button class="search-go" :disabled="!searchQuery.trim()" @click="runSearch">{{ t('home.searchButton') }}</button>
          </div>
          <HeaderTools />
        </div>
      </div>

      <!-- Row 2: Site tabs (only shown when there is more than one) -->
      <div v-if="SITES.length > 1" class="header-tabs">
        <div class="header-tabs-inner">
          <button
            v-for="site in SITES"
            :key="site.key"
            :class="['tab', { active: activeSite === site.key }]"
            @click="switchSite(site.key)"
          >{{ site.label }}</button>
        </div>
      </div>

      <!-- Row 3: language sub-tabs (Myanmar channel) -->
      <div v-if="activeSite === 'channel1'" class="sub-header">
        <div class="sub-header-inner">
          <button
            :class="['sub-tab', { active: activeCategory === '' }]"
            @click="switchCategory('')"
          >{{ t('home.all') }}</button>
          <button
            v-for="c in mmCats"
            :key="c.value"
            :class="['sub-tab', { active: activeCategory === c.value }]"
            @click="switchCategory(c.value)"
          >{{ c.label }}</button>
        </div>
      </div>

      <!-- Row 4: filters -->
      <div v-if="hasActors || activeTag" class="sub-header">
        <div class="sub-header-inner filter-row">
          <RouterLink v-if="hasActors" :to="actorsLink" class="actors-link">{{ t('home.actresses') }}</RouterLink>
          <button v-if="activeTag" class="filter-chip" @click="clearTag">
            #{{ activeTagLabel }} ✕
          </button>
        </div>
      </div>
    </header>

    <main class="main">

      <!-- ── Search results ── -->
      <template v-if="isSearching">
        <div v-if="searchLoading" class="video-grid">
          <div v-for="i in 8" :key="i" class="skeleton-card">
            <div class="skeleton-thumb" /><div class="skeleton-line" /><div class="skeleton-line short" />
          </div>
        </div>
        <div v-else-if="searchResults.length === 0" class="empty-state">
          <p class="empty-msg">{{ t('home.noResults', { q: searchQuery }) }}</p>
        </div>
        <div v-else class="video-grid">
          <VideoCard
            v-for="v in searchResults" :key="v.id"
            :title="v.title" :thumbnail="v.thumbnail"
            @select="navigate(v, searchResults)" @hover="prefetch(v.id)"
          />
        </div>
      </template>

      <!-- ── Normal browsing ── -->
      <template v-else>

        <!-- Watch History -->
        <section v-if="watchHistory.length > 0" class="history-section">
          <div class="section-header">
            <h2 class="section-title">{{ t('home.recent') }}</h2>
            <button class="clear-history" @click="clearHistory">{{ t('home.clearHistory') }}</button>
          </div>
          <div class="history-scroll">
            <div v-for="v in watchHistory" :key="v.id" class="history-card" @click="navigate(v, watchHistory)">
              <div class="history-thumb-wrap">
                <img v-if="v.thumbnail" :src="v.thumbnail" class="history-thumb" loading="lazy" />
                <div v-else class="history-thumb-placeholder" />
              </div>
              <p class="history-card-title">{{ v.title }}</p>
            </div>
          </div>
        </section>

        <!-- Skeleton -->
        <div v-if="loading" class="video-grid">
          <div v-for="i in 20" :key="i" class="skeleton-card">
            <div class="skeleton-thumb" /><div class="skeleton-line" /><div class="skeleton-line short" />
          </div>
        </div>

        <!-- Error -->
        <div v-else-if="error" class="empty-state">
          <p class="empty-msg">{{ error }}</p>
          <button class="retry-btn" @click="load(page)">{{ t('common.tryAgain') }}</button>
        </div>

        <!-- Grid -->
        <template v-else>
          <div class="video-grid">
            <VideoCard
              v-for="v in videos" :key="v.id"
              :title="v.title" :thumbnail="v.thumbnail"
              :is-new="isNewVideo(v)"
              @select="navigate(v, videos)" @hover="prefetch(v.id)"
            />
          </div>

          <!-- Pagination -->
          <div v-if="totalPages > 1" class="pagination">
            <button class="nav-btn" :disabled="page <= 1" @click="goToPage(page - 1)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="14" height="14"><path d="M15 18l-6-6 6-6"/></svg>
            </button>

            <template v-for="p in pageNumbers" :key="p">
              <span v-if="p === '...'" class="ellipsis">…</span>
              <button
                v-else
                :class="['page-btn', { active: p === page }]"
                @click="goToPage(p as number)"
              >{{ p }}</button>
            </template>

            <button class="nav-btn" :disabled="page >= totalPages" @click="goToPage(page + 1)">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="14" height="14"><path d="M9 18l6-6-6-6"/></svg>
            </button>
          </div>

          <p class="page-info" v-if="totalPages > 1">{{ t('home.pageInfo', { page, total: totalPages, count: total.toLocaleString() }) }}</p>
        </template>

        <NativeBanner1 v-if="!loading && videos.length > 0" />
        <Ad728x90 v-if="!loading && videos.length > 0" />
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute, RouterLink } from 'vue-router'
import VideoCard from '../components/VideoCard.vue'
import HeaderTools from '../components/HeaderTools.vue'
import NativeBanner1 from '../components/ads/NativeBanner_1.vue'
import Ad728x90 from '../components/ads/Ad728x90.vue'
import { apiFetch } from '../utils/apiFetch'
import { t } from '../utils/i18n'

interface VideoItem {
  id: string
  title: string
  thumbnail: string
  site: string
  synced_at?: string
}

// Main channel tabs — the two built-ins render immediately; admin-created
// manual channels (see /api/channels) are appended once they load.
const SITES = ref([
  { key: 'channel2', label: 'Chinese AV' },
  { key: 'channel1', label: 'Myanmar' },
])
async function loadChannels() {
  try {
    const res  = await apiFetch('/api/channels')
    const data = await res.json()
    if (res.ok && Array.isArray(data) && data.length) SITES.value = data
  } catch { /* keep the built-in two */ }
}

const DEFAULT_SITE = 'channel2'

// Myanmar channel language sub-tabs (all videos have MM subtitles; this splits
// by the source video's language). Admin-managed — see /api/categories.
const mmCats = ref<{ label: string; value: string }[]>([])
let mmCatsLoaded = false
async function loadMMCats() {
  if (mmCatsLoaded) return
  mmCatsLoaded = true
  try {
    const res  = await apiFetch('/api/categories?site=channel1')
    const data = await res.json()
    if (res.ok && Array.isArray(data)) mmCats.value = data
  } catch { /* keep the default "All" tab */ }
}

const router = useRouter()
const route  = useRoute()

const VALID_OVERRIDES = new Set([''])

// Trusts the URL rather than a fixed list, so a direct link to an
// admin-created channel works before /api/channels has loaded.
function validSite(s: string) { return s || DEFAULT_SITE }
function validOver(s: string) { return VALID_OVERRIDES.has(s) ? s : '' }

const activeSite         = ref(validSite((route.query.site as string) || ''))
const activeCategory     = ref((route.query.cat  as string) || '')
const activeSiteOverride = ref(validOver((route.query.over as string) || ''))
const activeTag          = ref((route.query.tag as string) || '')

const activeTagLabel = computed(() =>
  activeTag.value.split('-')
    .map(w => (w.length <= 2 || /\d/.test(w) ? w.toUpperCase() : w[0].toUpperCase() + w.slice(1)))
    .join(' '),
)

// Only Chinese AV and Myanmar have actress data; other channels hide the link.
const hasActors = computed(() => activeSite.value === 'channel1' || activeSite.value === 'channel2')

// "Actresses" link stays scoped to the channel you're browsing.
const actorsLink = computed(() =>
  activeSite.value === 'channel1' ? '/actors?site=channel1' : '/actors',
)

function clearTag() {
  if (!activeTag.value) return
  activeTag.value = ''
  syncURL()
  load(1)
}

function switchCategory(v: string) {
  if (activeCategory.value === v) return
  activeCategory.value = v
  syncURL()
  load(1)
}

function syncURL(pg = 1) {
  const q: Record<string, string> = { site: activeSite.value }
  if (activeCategory.value)     q.cat  = activeCategory.value
  if (activeSiteOverride.value) q.over = activeSiteOverride.value
  if (activeTag.value)          q.tag  = activeTag.value
  if (pg > 1)                   q.p    = String(pg)
  router.replace({ query: q })
}

function switchSite(key: string) {
  if (activeSite.value === key) return
  activeSite.value         = key
  activeCategory.value     = ''
  activeSiteOverride.value = ''
  activeTag.value          = ''
  clearSearch()
  router.replace({ query: { site: key } })
  load(1)
}

const videos  = ref<VideoItem[]>([])
const loading = ref(false)
const error   = ref('')
const page    = ref(1)
const total   = ref(0)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / 20)))

const pageNumbers = computed(() => {
  const tp  = totalPages.value
  const cur = page.value
  if (tp <= 7) return Array.from({ length: tp }, (_, i) => i + 1)

  const result: (number | '...')[] = [1]
  if (cur > 3) result.push('...')
  const start = Math.max(2, cur - 2)
  const end   = Math.min(tp - 1, cur + 2)
  for (let i = start; i <= end; i++) result.push(i)
  if (cur < tp - 2) result.push('...')
  result.push(tp)
  return result
})

const DAY_MS = 24 * 60 * 60 * 1000
function isNewVideo(v: VideoItem) {
  if (!v.synced_at) return false
  return Date.now() - new Date(v.synced_at).getTime() < DAY_MS
}

// Search — runs only when the user presses Enter or taps the ရှာ button.
const searchQuery   = ref('')
const searchResults = ref<VideoItem[]>([])
const searchLoading = ref(false)
const searchActive  = ref(false)
const isSearching   = computed(() => searchActive.value)

async function runSearch() {
  const q = searchQuery.value.trim()
  if (!q) return
  searchActive.value  = true
  searchLoading.value = true
  searchResults.value = []
  try {
    const res  = await apiFetch(`/api/search?q=${encodeURIComponent(q)}&site=${encodeURIComponent(activeSiteOverride.value || activeSite.value)}`)
    const data = await res.json()
    searchResults.value = res.ok ? (data as VideoItem[]) : []
  } catch {
    searchResults.value = []
  } finally {
    searchLoading.value = false
  }
}

function clearSearch() {
  searchQuery.value   = ''
  searchResults.value = []
  searchActive.value  = false
}

// Leaving the box empty drops back to normal browsing.
watch(searchQuery, v => { if (!v.trim()) clearSearch() })

// Watch history
const watchHistory = ref<VideoItem[]>([])

function loadHistory() {
  const raw = localStorage.getItem('watch_history')
  watchHistory.value = raw ? JSON.parse(raw).slice(0, 12) : []
}

function clearHistory() {
  localStorage.removeItem('watch_history')
  watchHistory.value = []
}

const prefetched = new Set<string>()
function prefetch(id: string) {
  if (prefetched.has(id)) return
  prefetched.add(id)
  apiFetch(`/api/video-url?id=${encodeURIComponent(id)}`).catch(() => {})
}

function navigate(v: VideoItem, list: VideoItem[]) {
  const watchId = Math.random().toString(36).slice(2, 10)
  sessionStorage.setItem(`vid:${watchId}`, JSON.stringify({
    videoId: v.id, title: v.title, thumbnail: v.thumbnail, site: v.site,
  }))
  const idx = list.findIndex(x => x.id === v.id)
  sessionStorage.setItem('playlist', JSON.stringify(list.map(x => ({
    id: x.id, title: x.title, thumbnail: x.thumbnail, site: x.site,
  }))))
  sessionStorage.setItem('playlist_index', String(idx))
  router.push(`/watch/${watchId}`)
}

async function load(pg: number) {
  page.value    = pg
  videos.value  = []
  error.value   = ''
  loading.value = true
  prefetched.clear()

  try {
    const effectiveSite = activeSiteOverride.value || activeSite.value
    const params = new URLSearchParams({ site: effectiveSite, page: String(pg) })
    if (activeCategory.value) params.set('category', activeCategory.value)
    if (activeTag.value)      params.set('tag', activeTag.value)
    const res  = await apiFetch(`/api/videos?${params}`)
    const data = await res.json()
    if (!res.ok) throw new Error(data.error ?? t('home.failedLoad'))
    videos.value = data.items ?? []
    total.value  = data.total ?? 0
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('common.unknownError')
  } finally {
    loading.value = false
  }
}

function goToPage(pg: number) {
  if (pg < 1 || pg > totalPages.value || pg === page.value) return
  window.scrollTo({ top: 0, behavior: 'smooth' })
  syncURL(pg)
  load(pg)
}

onMounted(() => {
  loadHistory()
  loadChannels()
  loadMMCats()
  // Replace URL if it contained invalid/old site names
  syncURL(Number(route.query.p) || 1)
  load(Number(route.query.p) || 1)
})
</script>

<style scoped>
.app { min-height: 100vh; display: flex; flex-direction: column; }

.header {
  position: sticky; top: 0; z-index: 50;
  background: rgb(var(--bg-rgb) / 0.97);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--surface);
}

/* Row 1: Logo + Search */
.header-top { border-bottom: 1px solid var(--border-soft); }
.header-top-inner {
  max-width: 1400px; margin: 0 auto; padding: 0 20px;
  height: 52px; display: flex; align-items: center; gap: 16px;
}
.logo {
  display: flex; align-items: center; gap: 8px;
  font-size: 1rem; font-weight: 700; color: var(--text);
  white-space: nowrap; flex-shrink: 0;
}

.search-wrap {
  min-width: 0;
  display: flex; align-items: center; gap: 8px;
  background: var(--surface); border: 1px solid var(--border);
  border-radius: 8px; padding: 0 14px; height: 36px;
  flex: 1; max-width: 520px;
  transition: border-color .15s;
}
.search-wrap:focus-within { border-color: var(--accent); }
.search-icon { color: var(--text-4); flex-shrink: 0; }
.search-input {
  background: transparent; border: none; outline: none;
  color: var(--text); font-size: .875rem; flex: 1; min-width: 0;
}
.search-input::placeholder { color: var(--border-strong); }
.search-clear {
  background: transparent; border: none; color: var(--text-4);
  cursor: pointer; font-size: .8rem; padding: 0; transition: color .15s;
}
.search-clear:hover { color: var(--text); }
.search-go {
  flex-shrink: 0; border: none; border-radius: 6px;
  background: var(--accent); color: var(--on-accent);
  font-size: .8rem; font-weight: 600; padding: 0 14px; height: 28px;
  cursor: pointer; transition: background .15s;
}
.search-go:hover:not(:disabled) { background: var(--accent-hover); }
.search-go:disabled { background: var(--border); color: var(--text-4); cursor: not-allowed; }

/* Row 2: Site tabs */
.header-tabs { position: relative; }
.header-tabs::after {
  content: ''; pointer-events: none;
  position: absolute; right: 0; top: 0; bottom: 0; width: 32px;
  background: linear-gradient(to right, transparent, rgb(var(--bg-rgb) / .95));
}
.header-tabs-inner {
  max-width: 1400px; margin: 0 auto; padding: 0 16px;
  height: 44px; display: flex; align-items: center; gap: 2px;
  overflow-x: auto; scrollbar-width: none;
  -webkit-overflow-scrolling: touch;
}
.header-tabs-inner::-webkit-scrollbar { display: none; }
.tab {
  padding: 7px 20px; border-radius: 6px; border: none;
  background: transparent; color: var(--text-3);
  font-size: .875rem; font-weight: 500;
  cursor: pointer; transition: all .15s; white-space: nowrap; flex-shrink: 0;
}
.tab:hover  { color: var(--text); background: var(--surface); }
.tab.active { color: var(--text); background: var(--accent); }

/* Row 3: Sub-tabs */
.sub-header { background: var(--bg); border-top: 1px solid var(--border-soft); position: relative; }
.sub-header::after {
  content: ''; pointer-events: none;
  position: absolute; right: 0; top: 0; bottom: 0; width: 32px;
  background: linear-gradient(to right, transparent, rgb(var(--bg-rgb) / .95));
}
.sub-header-inner {
  max-width: 1400px; margin: 0 auto; padding: 0 16px;
  height: 38px; display: flex; align-items: center; gap: 4px;
  overflow-x: auto; scrollbar-width: none;
  -webkit-overflow-scrolling: touch;
}
.sub-header-inner::-webkit-scrollbar { display: none; }
.sub-tab {
  padding: 4px 14px; border-radius: 5px; border: none;
  background: transparent; color: var(--text-4);
  font-size: .82rem; cursor: pointer; transition: all .15s; white-space: nowrap; flex-shrink: 0;
}
.sub-tab:hover  { color: var(--text); background: var(--surface); }
.sub-tab.active { color: var(--accent-text); background: rgb(var(--accent-rgb) / .15); font-weight: 600; }

.filter-row { gap: 10px; }
.actors-link {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 5px 14px; border-radius: 6px;
  border: 1px solid var(--border); background: var(--surface); color: var(--text-2);
  font-size: .82rem; font-weight: 500; text-decoration: none;
  white-space: nowrap; flex-shrink: 0; transition: all .15s;
}
.actors-link:hover { border-color: var(--accent); color: var(--text); }
.filter-chip {
  display: inline-flex; align-items: center; gap: 4px;
  background: rgb(var(--accent-rgb) / .15); color: var(--accent-text);
  border: 1px solid rgb(var(--accent-rgb) / .4); border-radius: 999px;
  padding: 4px 12px; font-size: .78rem; font-weight: 600;
  cursor: pointer; white-space: nowrap; flex-shrink: 0;
}
.filter-chip:hover { background: rgb(var(--accent-rgb) / .25); }

.main { max-width: 1400px; margin: 0 auto; padding: 24px 20px 80px; width: 100%; display: flex; flex-direction: column; gap: 20px; }

/* History */
.history-section { display: flex; flex-direction: column; gap: 12px; }
.section-header { display: flex; align-items: center; justify-content: space-between; }
.section-title { margin: 0; font-size: .85rem; font-weight: 600; color: var(--text-4); text-transform: uppercase; letter-spacing: .05em; }
.clear-history { background: transparent; border: none; color: var(--border-strong); font-size: .78rem; cursor: pointer; padding: 2px 6px; border-radius: 4px; transition: color .15s; }
.clear-history:hover { color: #f87171; }
.history-scroll { display: flex; gap: 12px; overflow-x: auto; padding-bottom: 8px; scrollbar-width: thin; scrollbar-color: var(--border) transparent; }
.history-scroll::-webkit-scrollbar { height: 4px; }
.history-scroll::-webkit-scrollbar-thumb { background: var(--border); border-radius: 2px; }
.history-card { flex-shrink: 0; width: 160px; cursor: pointer; display: flex; flex-direction: column; gap: 6px; transition: opacity .15s; }
.history-card:hover { opacity: .8; }
.history-thumb-wrap { width: 160px; height: 90px; border-radius: 6px; overflow: hidden; background: var(--surface); }
.history-thumb { width: 100%; height: 100%; object-fit: cover; }
.history-thumb-placeholder { width: 100%; height: 100%; background: var(--surface); }
.history-card-title { margin: 0; font-size: .78rem; color: var(--text-2); line-height: 1.3; overflow: hidden; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; }

.video-grid { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 16px 14px; }

.skeleton-card { display: flex; flex-direction: column; gap: 8px; }
.skeleton-thumb { aspect-ratio: 16/9; border-radius: 6px; background: linear-gradient(90deg,var(--surface) 25%,var(--surface-2) 50%,var(--surface) 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.skeleton-line  { height: 10px; border-radius: 4px; background: linear-gradient(90deg,var(--surface) 25%,var(--surface-2) 50%,var(--surface) 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.skeleton-line.short { width: 60%; }
@keyframes shimmer { 0%{background-position:200% 0} 100%{background-position:-200% 0} }

.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 200px; gap: 16px; }
.empty-msg   { color: var(--text-3); text-align: center; }
.retry-btn   { padding: 8px 22px; border-radius: 6px; border: 1px solid var(--accent); background: transparent; color: var(--accent); cursor: pointer; font-size: .875rem; transition: all .15s; }
.retry-btn:hover { background: var(--accent); color: var(--on-accent); }

/* Pagination */
.pagination {
  display: flex; align-items: center; justify-content: center;
  gap: 4px; flex-wrap: wrap; padding-top: 8px;
}
.page-btn, .nav-btn {
  min-width: 36px; height: 36px; padding: 0 8px;
  border-radius: 6px; border: 1px solid var(--border);
  background: var(--surface); color: var(--text-3);
  font-size: .85rem; cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: all .15s;
}
.page-btn:hover:not(:disabled):not(.active),
.nav-btn:hover:not(:disabled) { border-color: var(--accent); color: var(--text); }
.page-btn.active { background: var(--accent); border-color: var(--accent); color: var(--on-accent); font-weight: 600; }
.page-btn:disabled, .nav-btn:disabled { opacity: .4; cursor: not-allowed; }
.ellipsis {
  min-width: 36px; height: 36px;
  display: flex; align-items: center; justify-content: center;
  color: var(--border-strong); font-size: .85rem;
}
.page-info { text-align: center; font-size: .78rem; color: var(--border-strong); margin: 0; }

@media (max-width: 1024px) {
  .video-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); }
}
@media (max-width: 768px) {
  .logo-text { display: none; }
  .search-wrap { max-width: 100%; }
  .header-tabs-inner { height: 40px; }
  .tab { padding: 6px 14px; font-size: .8rem; }
  .sub-header-inner { height: 36px; }
  .sub-tab { padding: 4px 11px; font-size: .78rem; }
  .video-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
}
@media (max-width: 480px) {
  .header-top-inner { height: 46px; padding: 0 12px; gap: 8px; }
  .header-tabs-inner { height: 38px; padding: 0 8px; }
  .tab { padding: 5px 12px; font-size: .78rem; }
  .sub-header-inner { height: 34px; padding: 0 8px; }
  .sub-tab { padding: 3px 10px; font-size: .75rem; }
  .main { padding: 16px 12px 80px; }
  .video-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px 8px; }
  .history-card { width: 130px; }
  .history-thumb-wrap { width: 130px; height: 73px; }
}
</style>
