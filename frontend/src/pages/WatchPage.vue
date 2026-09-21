<template>
  <div class="watch-app">
    <!-- Header -->
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
      <!-- ① Top native banner ad -->
      <NativeBanner1 />

      <div class="watch-layout">
        <!-- ── Left: Player column ── -->
        <section class="player-col">

          <!-- Loading -->
          <div v-if="fetching" class="state-box">
            <div class="spinner-ring" />
            <p>Fetching video…</p>
          </div>

          <!-- No session (direct URL / new tab) -->
          <div v-else-if="!meta?.videoId" class="state-box error-box">
            <p>Video session expired or link opened in a new tab.</p>
            <RouterLink to="/" class="retry-btn" style="text-decoration:none;text-align:center">← Go to Home</RouterLink>
          </div>

          <!-- Error -->
          <div v-else-if="fetchError" class="state-box error-box">
            <p>{{ fetchError }}</p>
            <button class="retry-btn" @click="fetchVideo">Retry</button>
          </div>

          <!-- Player -->
          <template v-else-if="result">
            <div class="player-wrap">
              <VideoPlayer
                :url="activeUrl"
                :type="result.type"
                @error="onPlayerError"
                @ended="onPlayerEnded"
              />
            </div>

            <div class="video-meta">
              <h1 class="video-title">{{ title }}</h1>

              <!-- Category / Actress / Tags -->
              <div v-if="hasTaxonomy" class="taxonomy">
                <p v-if="result.categories?.length" class="tax-line">
                  <span class="tax-key">Category:</span>
                  <span class="tax-val">{{ result.categories.map(c => c.label).join(' · ') }}</span>
                </p>
                <p v-if="result.actors?.length" class="tax-line">
                  <span class="tax-key">Actress:</span>
                  <button
                    v-for="a in result.actors" :key="a.slug"
                    class="tax-chip actress" @click="openActress(a.slug)"
                  >{{ a.name }}</button>
                </p>
                <p v-if="result.tags?.length" class="tax-line">
                  <span class="tax-key">Tags:</span>
                  <button
                    v-for="t in result.tags" :key="t.slug"
                    class="tax-chip" @click="goFilter('tag', t.slug)"
                  >{{ t.label }}</button>
                </p>
              </div>

              <!-- Auto-retry notice -->
              <p v-if="autoRetrying" class="retry-notice">
                Server {{ activeMirrorIndex + 1 }} failed — trying Server {{ activeMirrorIndex + 2 }}…
              </p>

              <div class="meta-actions">
                <!-- Mirror server buttons -->
                <div v-if="result.mirrors && result.mirrors.length > 1" class="mirrors">
                  <span class="mirrors-label">Servers</span>
                  <button
                    v-for="(m, i) in result.mirrors"
                    :key="m"
                    :class="['mirror-btn', { active: m === activeUrl }]"
                    @click="switchMirror(i)"
                  >
                    {{ i + 1 }}
                  </button>
                </div>

                <!-- Share button -->
                <button class="share-btn" @click="shareVideo">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
                    <circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/>
                    <path d="M8.59 13.51l6.83 3.98M15.41 6.51l-6.82 3.98"/>
                  </svg>
                  {{ shareCopied ? 'Link Copied!' : 'Share' }}
                </button>
              </div>

              <!-- Auto-play next -->
              <div v-if="hasNextVideo" class="autoplay-row">
                <span class="autoplay-label">Next:</span>
                <span class="autoplay-title">{{ nextVideo?.title }}</span>
                <button class="autoplay-skip" @click="goToNext">Play now</button>
              </div>
            </div>

            <!-- ② Below-player 300×250 ad — key forces re-mount on server switch -->
            <Ad300x250 :key="activeUrl" />
          </template>
        </section>

        <!-- ── Right: Sidebar ── -->
        <aside class="sidebar">
          <!-- ③ Sidebar 300×250 ad -->
          <Ad300x250 />

          <!-- ④ Related videos -->
          <div v-if="relatedVideos.length > 0" class="related-section">
            <h3 class="related-title">ဆင်တူသော ဗီဒီယိုများ</h3>
            <div class="related-list">
              <div
                v-for="v in relatedVideos"
                :key="v.id"
                class="related-card"
                @click="navigateRelated(v)"
              >
                <div class="related-thumb-wrap">
                  <img v-if="v.thumbnail" :src="v.thumbnail" class="related-thumb" loading="lazy" />
                  <div v-else class="related-thumb-placeholder" />
                  <div class="related-play-overlay">
                    <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20"><path d="M8 5v14l11-7z"/></svg>
                  </div>
                </div>
                <p class="related-card-title">{{ v.title }}</p>
              </div>
            </div>
          </div>
          <Ad300x250 v-else />
        </aside>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import VideoPlayer from '../components/VideoPlayer.vue'
import VideoCard from '../components/VideoCard.vue'
import NativeBanner1 from '../components/ads/NativeBanner_1.vue'
import Ad300x250 from '../components/ads/300x250_1.vue'
import { apiFetch } from '../utils/apiFetch'

interface TaxRef { slug: string; name?: string; label?: string }
interface VideoResult {
  url: string
  type: 'mp4' | 'hls' | 'iframe'
  mirrors?: string[]
  title?: string
  categories?: TaxRef[]
  actors?: TaxRef[]
  tags?: TaxRef[]
}

interface PlaylistItem {
  id: string
  title: string
  thumbnail: string
  site?: string
}

interface RelatedItem {
  id: string
  title: string
  thumbnail: string
  site: string
}

const route  = useRoute()
const router = useRouter()

const watchId = route.params.id as string
const stored  = sessionStorage.getItem(`vid:${watchId}`)

const LEGACY_SITE: Record<string, string> = {
  '3xchina': 'channel2',
}
function normalizeChannel(s: string) { return LEGACY_SITE[s] ?? s }

const meta = stored ? (() => {
  const m = JSON.parse(stored) as { videoId: string; title: string; thumbnail: string; site: string }
  m.site = normalizeChannel(m.site)
  return m
})() : null

const title = meta?.title ?? 'Video'

// Playlist for auto-play next
const playlistRaw   = sessionStorage.getItem('playlist')
const playlistIndex = parseInt(sessionStorage.getItem('playlist_index') ?? '-1', 10)
const playlist      = playlistRaw ? JSON.parse(playlistRaw) as PlaylistItem[] : []

const nextVideo = computed<PlaylistItem | null>(() =>
  playlist.length > 0 && playlistIndex >= 0 && playlistIndex + 1 < playlist.length
    ? playlist[playlistIndex + 1]
    : null
)
const hasNextVideo = computed(() => nextVideo.value !== null)

function goFilter(kind: 'tag', slug: string) {
  router.push({ path: '/', query: { site: meta?.site ?? 'channel2', [kind]: slug } })
}

function openActress(slug: string) {
  const s = meta?.site ?? 'channel2'
  router.push(s === 'channel2' ? `/actor/${slug}` : `/actor/${slug}?site=channel1`)
}

const fetching          = ref(false)
const fetchError        = ref('')
const result            = ref<VideoResult | null>(null)
const activeUrl         = ref('')
const activeMirrorIndex = ref(0)
const autoRetrying      = ref(false)

const hasTaxonomy = computed(() => {
  const r = result.value
  return !!r && !!(r.categories?.length || r.actors?.length || r.tags?.length)
})

// Share
const shareUrl     = ref('')
const shareCopied  = ref(false)

async function shareVideo() {
  const res  = await apiFetch('/api/share', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ video_id: meta?.videoId ?? '', title, thumbnail: meta?.thumbnail ?? '', site: meta?.site ?? '' }),
  })
  const data = await res.json()
  if (data.code) {
    const link = `${location.origin}/s/${data.code}`
    shareUrl.value = link
    await navigator.clipboard.writeText(link).catch(() => {})
    shareCopied.value = true
    setTimeout(() => { shareCopied.value = false }, 3000)
  }
}

// Related videos
const relatedVideos = ref<RelatedItem[]>([])

async function fetchRelated() {
  if (!meta?.site) return
  try {
    const res  = await apiFetch(`/api/related?site=${encodeURIComponent(meta.site)}&exclude=${encodeURIComponent(meta.videoId ?? '')}`)
    const data = await res.json()
    if (res.ok) relatedVideos.value = data as RelatedItem[]
  } catch { /* ignore */ }
}

function navigateRelated(v: RelatedItem) {
  const newId = Math.random().toString(36).slice(2, 10)
  sessionStorage.setItem(`vid:${newId}`, JSON.stringify({ videoId: v.id, title: v.title, thumbnail: v.thumbnail, site: normalizeChannel(v.site) }))
  sessionStorage.setItem('playlist', JSON.stringify(relatedVideos.value.map(x => ({ id: x.id, title: x.title, thumbnail: x.thumbnail, site: x.site }))))
  sessionStorage.setItem('playlist_index', String(relatedVideos.value.findIndex(x => x.id === v.id)))
  router.push(`/watch/${newId}`)
}

function switchMirror(i: number) {
  if (!result.value) return
  activeMirrorIndex.value = i
  activeUrl.value = result.value.mirrors?.[i] ?? result.value.url
  autoRetrying.value = false
}

function onPlayerError() {
  if (!result.value) return
  const mirrors = result.value.mirrors ?? []
  const next = activeMirrorIndex.value + 1
  if (next < mirrors.length) {
    autoRetrying.value = true
    activeMirrorIndex.value = next
    activeUrl.value = mirrors[next]
    setTimeout(() => { autoRetrying.value = false }, 3000)
  }
}

function goToNext() {
  if (!nextVideo.value) return
  const next = nextVideo.value
  const newId = Math.random().toString(36).slice(2, 10)
  sessionStorage.setItem(`vid:${newId}`, JSON.stringify({
    videoId: next.id,
    title: next.title,
    thumbnail: next.thumbnail,
    site: next.site ?? meta?.site ?? '',
  }))
  sessionStorage.setItem('playlist_index', String(playlistIndex + 1))
  router.push(`/watch/${newId}`)
}

function onPlayerEnded() {
  if (hasNextVideo.value) goToNext()
}

function saveToHistory() {
  if (!meta?.videoId) return
  const key = 'watch_history'
  const raw = localStorage.getItem(key)
  let list: PlaylistItem[] = raw ? JSON.parse(raw) : []
  list = list.filter(x => x.id !== meta.videoId)
  list.unshift({ id: meta.videoId, title, thumbnail: meta?.thumbnail ?? '', site: meta?.site ?? '' })
  if (list.length > 50) list = list.slice(0, 50)
  localStorage.setItem(key, JSON.stringify(list))
}

async function fetchVideo() {
  if (!meta?.videoId) return

  fetching.value        = true
  fetchError.value      = ''
  result.value          = null
  activeUrl.value       = ''
  activeMirrorIndex.value = 0
  autoRetrying.value    = false

  try {
    const res  = await apiFetch(`/api/video-url?id=${encodeURIComponent(meta!.videoId)}`)
    const data = await res.json()
    if (!res.ok) throw new Error(data.error ?? 'Failed to fetch video')
    result.value    = data as VideoResult
    activeUrl.value = data.url
    saveToHistory()
    fetchRelated()
  } catch (e) {
    fetchError.value = e instanceof Error ? e.message : 'Unknown error'
  } finally {
    fetching.value = false
  }
}

onMounted(fetchVideo)
</script>

<style scoped>
.watch-app { min-height: 100vh; display: flex; flex-direction: column; background: #0f172a; }

/* Header */
.header { background: rgba(15,23,42,0.95); border-bottom: 1px solid #1e293b; position: sticky; top: 0; z-index: 50; }
.header-inner { max-width: 1400px; margin: 0 auto; padding: 0 20px; height: 56px; display: flex; align-items: center; gap: 20px; }
.back-btn { display: flex; align-items: center; gap: 6px; padding: 6px 14px; border-radius: 6px; border: 1px solid #334155; background: transparent; color: #94a3b8; font-size: .875rem; cursor: pointer; transition: all .15s; }
.back-btn:hover { color: #f1f5f9; border-color: #6366f1; }
.logo { display: flex; align-items: center; gap: 8px; font-size: .95rem; font-weight: 700; color: #f1f5f9; text-decoration: none; }

/* Main layout */
.main { max-width: 1400px; margin: 0 auto; padding: 20px 20px 80px; width: 100%; display: flex; flex-direction: column; gap: 20px; }

/* Watch layout */
.watch-layout { display: grid; grid-template-columns: 1fr 300px; gap: 24px; align-items: start; }

.player-col { display: flex; flex-direction: column; gap: 0; }

.player-wrap { background: #000; border-radius: 10px; overflow: hidden; }

.video-meta { padding: 16px 0 20px; display: flex; flex-direction: column; gap: 14px; }
.video-title { margin: 0; font-size: 1.1rem; font-weight: 600; color: #f1f5f9; line-height: 1.4; }

.retry-notice { margin: 0; font-size: .8rem; color: #fbbf24; background: rgba(251,191,36,.1); border: 1px solid rgba(251,191,36,.2); padding: 6px 12px; border-radius: 6px; }

.taxonomy { display: flex; flex-direction: column; gap: 8px; padding: 12px 14px; background: #0f172a; border: 1px solid #1e293b; border-radius: 8px; }
.tax-line { margin: 0; display: flex; align-items: center; flex-wrap: wrap; gap: 6px; font-size: .82rem; }
.tax-key { color: #64748b; font-weight: 600; text-transform: uppercase; letter-spacing: .04em; font-size: .72rem; }
.tax-val { color: #cbd5e1; }
.tax-chip { padding: 3px 10px; border-radius: 999px; border: 1px solid #334155; background: #1e293b; color: #94a3b8; font-size: .78rem; cursor: pointer; transition: all .15s; }
.tax-chip:hover { border-color: #6366f1; color: #f1f5f9; }
.tax-chip.actress { color: #818cf8; border-color: rgba(99,102,241,.4); background: rgba(99,102,241,.12); }
.tax-chip.actress:hover { background: rgba(99,102,241,.25); }

.meta-actions { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }

.mirrors { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; flex: 1; }
.mirrors-label { font-size: .75rem; color: #64748b; text-transform: uppercase; letter-spacing: .05em; }
.mirror-btn { padding: 5px 14px; border-radius: 5px; border: 1px solid #334155; background: #1e293b; color: #94a3b8; font-size: .8rem; cursor: pointer; transition: all .15s; }
.mirror-btn:hover { border-color: #6366f1; color: #f1f5f9; }
.mirror-btn.active { background: #6366f1; border-color: #6366f1; color: #fff; }

.share-btn { display: flex; align-items: center; gap: 5px; padding: 5px 14px; border-radius: 6px; border: 1px solid #334155; background: transparent; color: #94a3b8; font-size: .8rem; cursor: pointer; white-space: nowrap; transition: all .15s; flex-shrink: 0; }
.share-btn:hover { border-color: #6366f1; color: #f1f5f9; }

/* Related videos */
.related-section { display: flex; flex-direction: column; gap: 10px; }
.related-title { margin: 0; font-size: .8rem; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: .05em; }
.related-list { display: flex; flex-direction: column; gap: 10px; }
.related-card { display: flex; gap: 10px; cursor: pointer; border-radius: 6px; padding: 4px; transition: background .15s; }
.related-card:hover { background: #1e293b; }
.related-thumb-wrap { position: relative; width: 100px; min-width: 100px; height: 56px; border-radius: 5px; overflow: hidden; background: #1e293b; }
.related-thumb { width: 100%; height: 100%; object-fit: cover; }
.related-thumb-placeholder { width: 100%; height: 100%; background: #334155; }
.related-play-overlay { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,.45); opacity: 0; transition: opacity .2s; color: #fff; }
.related-card:hover .related-play-overlay { opacity: 1; }
.related-card-title { margin: 0; font-size: .78rem; color: #cbd5e1; line-height: 1.35; display: -webkit-box; -webkit-line-clamp: 3; -webkit-box-orient: vertical; overflow: hidden; }

.autoplay-row { display: flex; align-items: center; gap: 10px; padding: 8px 14px; background: #1e293b; border-radius: 8px; border: 1px solid #334155; flex-wrap: wrap; }
.autoplay-label { font-size: .75rem; color: #64748b; text-transform: uppercase; letter-spacing: .05em; white-space: nowrap; }
.autoplay-title { font-size: .85rem; color: #cbd5e1; flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.autoplay-skip { padding: 4px 12px; border-radius: 5px; border: 1px solid #6366f1; background: transparent; color: #6366f1; font-size: .78rem; cursor: pointer; white-space: nowrap; transition: all .15s; }
.autoplay-skip:hover { background: #6366f1; color: #fff; }

.sidebar { display: flex; flex-direction: column; gap: 16px; position: sticky; top: 76px; }

/* State boxes */
.state-box { display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 300px; gap: 16px; background: #1e293b; border-radius: 10px; color: #94a3b8; }
.error-box { color: #f87171; }
.spinner-ring { width: 40px; height: 40px; border: 3px solid rgba(99,102,241,.3); border-top-color: #6366f1; border-radius: 50%; animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.retry-btn { padding: 8px 20px; border-radius: 6px; border: 1px solid #6366f1; background: transparent; color: #6366f1; cursor: pointer; font-size: .875rem; transition: all .15s; }
.retry-btn:hover { background: #6366f1; color: #fff; }

/* Mobile */
@media (max-width: 768px) {
  .watch-layout { grid-template-columns: 1fr; }
  .sidebar { position: static; }
  .main { padding: 12px 12px 80px; }
}

@media (max-width: 480px) {
  .header-inner { padding: 0 12px; gap: 12px; }
  .video-title { font-size: .95rem; }
}
</style>
