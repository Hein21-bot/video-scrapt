<template>
  <div class="actor-app">
    <header class="header">
      <div class="header-inner">
        <button class="back-btn" @click="router.back()">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
            <path d="M19 12H5M12 5l-7 7 7 7"/>
          </svg>
          {{ t('common.back') }}
        </button>
        <RouterLink to="/" class="logo">
          <img src="/logo.png" class="logo-img" alt="ChitNya" width="36" height="36" />
          <span class="logo-text">ChitNya</span>
        </RouterLink>
        <HeaderTools />
      </div>
    </header>

    <main class="main">
      <div class="actor-head">
        <div class="actor-avatar" v-if="headThumb"><img :src="headThumb" alt="" /></div>
        <div>
          <p class="crumb">
            <RouterLink :to="site === 'channel2' ? '/actors' : '/actors?site=channel1'" class="crumb-link">{{ t('home.actresses') }}</RouterLink> ›
          </p>
          <h1 class="actor-name">{{ name }}</h1>
          <p class="actor-sub">{{ channelLabel }} · {{ t('actors.videoCount', { n: total.toLocaleString() }) }}</p>
        </div>
      </div>

      <div v-if="loading" class="video-grid">
        <div v-for="i in 20" :key="i" class="skeleton-card">
          <div class="skeleton-thumb" /><div class="skeleton-line" /><div class="skeleton-line short" />
        </div>
      </div>

      <div v-else-if="error" class="empty-state">
        <p class="empty-msg">{{ error }}</p>
        <button class="retry-btn" @click="load(page)">{{ t('common.tryAgain') }}</button>
      </div>

      <div v-else-if="videos.length === 0" class="empty-state">
        <p class="empty-msg">{{ t('actors.noVideos') }}</p>
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
          <button class="nav-btn" :disabled="page <= 1" @click="goToPage(page - 1)">{{ t('pager.prev') }}</button>
          <span class="page-info">{{ t('pager.page', { p: page, t: totalPages }) }}</span>
          <button class="nav-btn" :disabled="page >= totalPages" @click="goToPage(page + 1)">{{ t('pager.next') }}</button>
        </div>
      </template>

      <NativeBanner1 v-if="!loading && videos.length > 0" />
      <Ad728x90 v-if="!loading && videos.length > 0" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import VideoCard from '../components/VideoCard.vue'
import HeaderTools from '../components/HeaderTools.vue'
import { t } from '../utils/i18n'
import NativeBanner1 from '../components/ads/NativeBanner_1.vue'
import Ad728x90 from '../components/ads/Ad728x90.vue'
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
  load(pg)
}

watch(slug, () => load(1))
onMounted(() => load(page.value))
</script>

<style scoped>
.actor-app { min-height: 100vh; display: flex; flex-direction: column; background: var(--bg); }

.header { background: rgb(var(--bg-rgb) / 0.95); border-bottom: 1px solid var(--surface); position: sticky; top: 0; z-index: 50; }
.header-inner { max-width: 1400px; margin: 0 auto; padding: 0 20px; height: 56px; display: flex; align-items: center; gap: 20px; }
.back-btn { display: flex; align-items: center; gap: 6px; padding: 6px 14px; border-radius: 6px; border: 1px solid var(--border); background: transparent; color: var(--text-3); font-size: .875rem; cursor: pointer; transition: all .15s; }
.back-btn:hover { color: var(--text); border-color: var(--accent); }
.logo { display: flex; align-items: center; gap: 8px; font-size: .95rem; font-weight: 700; color: var(--text); text-decoration: none; }

.main { max-width: 1400px; width: 100%; margin: 0 auto; padding: 20px; flex: 1; }

.actor-head { display: flex; align-items: center; gap: 16px; margin: 12px 0 20px; }
.actor-avatar { width: 68px; height: 88px; border-radius: 8px; overflow: hidden; background: var(--surface); flex-shrink: 0; }
.actor-avatar img { width: 100%; height: 100%; object-fit: cover; }
.crumb { margin: 0 0 2px; font-size: .78rem; color: var(--text-4); }
.crumb-link { color: var(--accent-text); text-decoration: none; }
.crumb-link:hover { text-decoration: underline; }
.actor-name { margin: 0; font-size: 1.35rem; font-weight: 700; color: var(--text); }
.actor-sub { margin: 2px 0 0; font-size: .85rem; color: var(--text-3); }

.video-grid { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 16px 14px; }
.skeleton-card { display: flex; flex-direction: column; gap: 8px; }
.skeleton-thumb { aspect-ratio: 16/9; border-radius: 6px; background: linear-gradient(90deg,var(--surface) 25%,var(--surface-2) 50%,var(--surface) 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.skeleton-line { height: 10px; border-radius: 4px; background: linear-gradient(90deg,var(--surface) 25%,var(--surface-2) 50%,var(--surface) 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.skeleton-line.short { width: 60%; }
@keyframes shimmer { 0%{background-position:200% 0} 100%{background-position:-200% 0} }

.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 200px; gap: 16px; }
.empty-msg { color: var(--text-3); text-align: center; }
.retry-btn { padding: 8px 22px; border-radius: 6px; border: 1px solid var(--accent); background: transparent; color: var(--accent); cursor: pointer; font-size: .875rem; }
.retry-btn:hover { background: var(--accent); color: var(--on-accent); }

.pagination { display: flex; align-items: center; justify-content: center; gap: 14px; padding-top: 20px; }
.nav-btn { padding: 7px 16px; border-radius: 6px; border: 1px solid var(--border); background: var(--surface); color: var(--text-2); font-size: .85rem; cursor: pointer; transition: all .15s; }
.nav-btn:hover:not(:disabled) { border-color: var(--accent); color: var(--text); }
.nav-btn:disabled { opacity: .4; cursor: not-allowed; }
.page-info { font-size: .82rem; color: var(--text-4); }

@media (max-width: 1024px) { .video-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); } }
@media (max-width: 768px)  { .video-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
@media (max-width: 480px)  { .video-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px 8px; } }
@media (max-width: 400px) {
  .logo-text { display: none; }
  .header-inner { padding: 0 12px; gap: 10px; }
}
</style>
