<template>
  <div class="actor-app">
    <header class="header">
      <div class="header-inner">
        <button class="back-btn" @click="router.back()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
            <path d="M19 12H5M12 5l-7 7 7 7"/>
          </svg>
          Back
        </button>
        <RouterLink to="/" class="logo">
          <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
            <path d="M4 6a2 2 0 012-2h12a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V6z"/>
            <path fill="#6366f1" d="M10 9l5 3-5 3V9z"/>
          </svg>
          NightMM
        </RouterLink>
      </div>
    </header>

    <main class="main">
      <NativeBanner1 />

      <div class="actor-head">
        <div class="actor-avatar" v-if="headThumb"><img :src="headThumb" alt="" /></div>
        <div>
          <p class="crumb">
            <RouterLink :to="site === 'channel2' ? '/actors' : '/actors?site=channel1'" class="crumb-link">🎭 Actresses</RouterLink> ›
          </p>
          <h1 class="actor-name">{{ name }}</h1>
          <p class="actor-sub">{{ channelLabel }} · {{ total.toLocaleString() }} videos</p>
        </div>
      </div>

      <div v-if="loading" class="video-grid">
        <div v-for="i in 20" :key="i" class="skeleton-card">
          <div class="skeleton-thumb" /><div class="skeleton-line" /><div class="skeleton-line short" />
        </div>
      </div>

      <div v-else-if="error" class="empty-state">
        <p class="empty-msg">{{ error }}</p>
        <button class="retry-btn" @click="load(page)">Try again</button>
      </div>

      <div v-else-if="videos.length === 0" class="empty-state">
        <p class="empty-msg">ဒီ actress အတွက် video မတွေ့ပါ</p>
      </div>

      <template v-else>
        <div class="video-grid">
          <VideoCard
            v-for="v in videos" :key="v.id"
            :title="v.title" :thumbnail="v.thumbnail"
            @select="navigate(v)" @hover="prefetch(v.id)"
          />
        </div>

        <div v-if="totalPages > 1" class="pagination">
          <button class="nav-btn" :disabled="page <= 1" @click="goToPage(page - 1)">‹ Prev</button>
          <span class="page-info">Page {{ page }} / {{ totalPages }}</span>
          <button class="nav-btn" :disabled="page >= totalPages" @click="goToPage(page + 1)">Next ›</button>
        </div>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import VideoCard from '../components/VideoCard.vue'
import NativeBanner1 from '../components/ads/NativeBanner_1.vue'
import { apiFetch } from '../utils/apiFetch'

interface VideoItem { id: string; title: string; thumbnail: string; site: string; synced_at?: string }

const route  = useRoute()
const router = useRouter()

const slug = computed(() => String(route.params.slug || ''))
const site = computed(() => (route.query.site === 'channel1' ? 'channel1' : 'channel2'))
const channelLabel = computed(() => (site.value === 'channel1' ? 'Myanmar' : 'Chinese AV'))
const name = computed(() =>
  slug.value.split('-')
    .map(w => (w ? w[0].toUpperCase() + w.slice(1) : w))
    .join(' '),
)

const videos  = ref<VideoItem[]>([])
const total   = ref(0)
const page    = ref(Number(route.query.p) || 1)
const loading = ref(true)
const error   = ref('')

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / 20)))
const headThumb  = computed(() => videos.value[0]?.thumbnail ?? '')

const prefetched = new Set<string>()
function prefetch(id: string) {
  if (prefetched.has(id)) return
  prefetched.add(id)
  apiFetch(`/api/video-url?id=${encodeURIComponent(id)}`).catch(() => {})
}

function navigate(v: VideoItem) {
  const watchId = Math.random().toString(36).slice(2, 10)
  sessionStorage.setItem(`vid:${watchId}`, JSON.stringify({
    videoId: v.id, title: v.title, thumbnail: v.thumbnail, site: v.site,
  }))
  const idx = videos.value.findIndex(x => x.id === v.id)
  sessionStorage.setItem('playlist', JSON.stringify(videos.value.map(x => ({
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
  const q: Record<string, string> = {}
  if (site.value === 'channel1') q.site = 'channel1'
  if (pg > 1) q.p = String(pg)
  router.replace({ query: q })
  try {
    const params = new URLSearchParams({ site: site.value, actor: slug.value, page: String(pg) })
    const res  = await apiFetch(`/api/videos?${params}`)
    const data = await res.json()
    if (!res.ok) throw new Error(data.error ?? 'Failed to load videos')
    videos.value = data.items ?? []
    total.value  = data.total ?? 0
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Unknown error'
  } finally {
    loading.value = false
  }
}

function goToPage(pg: number) {
  if (pg < 1 || pg > totalPages.value || pg === page.value) return
  window.scrollTo({ top: 0, behavior: 'smooth' })
  load(pg)
}

watch(slug, () => load(1))
onMounted(() => load(page.value))
</script>

<style scoped>
.actor-app { min-height: 100vh; display: flex; flex-direction: column; background: #0f172a; }

.header { background: rgba(15,23,42,0.95); border-bottom: 1px solid #1e293b; position: sticky; top: 0; z-index: 50; }
.header-inner { max-width: 1400px; margin: 0 auto; padding: 0 20px; height: 56px; display: flex; align-items: center; gap: 20px; }
.back-btn { display: flex; align-items: center; gap: 6px; padding: 6px 14px; border-radius: 6px; border: 1px solid #334155; background: transparent; color: #94a3b8; font-size: .875rem; cursor: pointer; transition: all .15s; }
.back-btn:hover { color: #f1f5f9; border-color: #6366f1; }
.logo { display: flex; align-items: center; gap: 8px; font-size: .95rem; font-weight: 700; color: #f1f5f9; text-decoration: none; }

.main { max-width: 1400px; width: 100%; margin: 0 auto; padding: 20px; flex: 1; }

.actor-head { display: flex; align-items: center; gap: 16px; margin: 12px 0 20px; }
.actor-avatar { width: 68px; height: 88px; border-radius: 8px; overflow: hidden; background: #1e293b; flex-shrink: 0; }
.actor-avatar img { width: 100%; height: 100%; object-fit: cover; }
.crumb { margin: 0 0 2px; font-size: .78rem; color: #64748b; }
.crumb-link { color: #818cf8; text-decoration: none; }
.crumb-link:hover { text-decoration: underline; }
.actor-name { margin: 0; font-size: 1.35rem; font-weight: 700; color: #f1f5f9; }
.actor-sub { margin: 2px 0 0; font-size: .85rem; color: #94a3b8; }

.video-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 16px 14px; }
.skeleton-card { display: flex; flex-direction: column; gap: 8px; }
.skeleton-thumb { aspect-ratio: 16/9; border-radius: 6px; background: linear-gradient(90deg,#1e293b 25%,#273549 50%,#1e293b 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.skeleton-line { height: 10px; border-radius: 4px; background: linear-gradient(90deg,#1e293b 25%,#273549 50%,#1e293b 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.skeleton-line.short { width: 60%; }
@keyframes shimmer { 0%{background-position:200% 0} 100%{background-position:-200% 0} }

.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 200px; gap: 16px; }
.empty-msg { color: #94a3b8; text-align: center; }
.retry-btn { padding: 8px 22px; border-radius: 6px; border: 1px solid #6366f1; background: transparent; color: #6366f1; cursor: pointer; font-size: .875rem; }
.retry-btn:hover { background: #6366f1; color: #fff; }

.pagination { display: flex; align-items: center; justify-content: center; gap: 14px; padding-top: 20px; }
.nav-btn { padding: 7px 16px; border-radius: 6px; border: 1px solid #334155; background: #1e293b; color: #cbd5e1; font-size: .85rem; cursor: pointer; transition: all .15s; }
.nav-btn:hover:not(:disabled) { border-color: #6366f1; color: #f1f5f9; }
.nav-btn:disabled { opacity: .4; cursor: not-allowed; }
.page-info { font-size: .82rem; color: #64748b; }

@media (max-width: 1024px) { .video-grid { grid-template-columns: repeat(4, 1fr); } }
@media (max-width: 768px)  { .video-grid { grid-template-columns: repeat(3, 1fr); } }
@media (max-width: 480px)  { .video-grid { grid-template-columns: repeat(2, 1fr); gap: 10px 8px; } }
</style>
