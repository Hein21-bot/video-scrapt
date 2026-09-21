<template>
  <div class="actors-app">
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
      <div class="chan-toggle">
        <button :class="{ active: site === 'channel2' }" @click="setSite('channel2')">Chinese AV</button>
        <button :class="{ active: site === 'channel1' }" @click="setSite('channel1')">Myanmar</button>
      </div>

      <div class="page-head">
        <h1 class="page-title">Actresses <span class="count">{{ filtered.length }}</span></h1>
        <div class="search-wrap">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
            <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
          </svg>
          <input v-model="q" class="search-input" placeholder="Actress ရှာရန်…" />
          <button v-if="q" class="search-clear" @click="q = ''">✕</button>
        </div>
      </div>

      <div v-if="loading" class="actor-grid">
        <div v-for="i in 24" :key="i" class="skeleton">
          <div class="skeleton-thumb" /><div class="skeleton-line" />
        </div>
      </div>

      <div v-else-if="filtered.length === 0" class="empty-state">
        <p>"{{ q }}" အတွက် actress မတွေ့ပါ</p>
      </div>

      <div v-else class="actor-grid">
        <button
          v-for="a in filtered" :key="a.slug"
          class="actor-card"
          @click="openActor(a.slug)"
        >
          <div class="actor-thumb-wrap">
            <img v-if="a.thumb && !broken[a.slug]" :src="a.thumb" class="actor-thumb" loading="lazy" @error="broken[a.slug] = true" />
            <div v-else class="actor-thumb-ph">{{ a.name.charAt(0) }}</div>
            <span class="actor-count">{{ a.count }}</span>
          </div>
          <p class="actor-name">{{ a.name }}</p>
        </button>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, reactive } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { apiFetch } from '../utils/apiFetch'

interface Actor { slug: string; name: string; count: number; thumb: string }

const route   = useRoute()
const router  = useRouter()
const actors  = ref<Actor[]>([])
const loading = ref(true)
const q       = ref('')
const broken  = reactive<Record<string, boolean>>({})

const site = ref(route.query.site === 'channel1' ? 'channel1' : 'channel2')

function setSite(s: string) {
  if (site.value === s) return
  site.value = s
  q.value = ''
  router.replace({ query: s === 'channel2' ? {} : { site: s } })
}

const filtered = computed(() => {
  const s = q.value.trim().toLowerCase()
  if (!s) return actors.value
  return actors.value.filter(a => a.name.toLowerCase().includes(s) || a.slug.includes(s.replace(/\s+/g, '-')))
})

function openActor(slug: string) {
  router.push(site.value === 'channel2' ? `/actor/${slug}` : `/actor/${slug}?site=channel1`)
}

async function loadActors() {
  loading.value = true
  try {
    const res = await apiFetch(`/api/actors?site=${site.value}`)
    actors.value = res.ok ? await res.json() : []
  } catch {
    actors.value = []
  } finally {
    loading.value = false
  }
}

watch(site, loadActors)
onMounted(loadActors)
</script>

<style scoped>
.actors-app { min-height: 100vh; display: flex; flex-direction: column; background: #0f172a; }

.header { background: rgba(15,23,42,0.95); border-bottom: 1px solid #1e293b; position: sticky; top: 0; z-index: 50; }
.header-inner { max-width: 1400px; margin: 0 auto; padding: 0 20px; height: 56px; display: flex; align-items: center; gap: 20px; }
.back-btn { display: flex; align-items: center; gap: 6px; padding: 6px 14px; border-radius: 6px; border: 1px solid #334155; background: transparent; color: #94a3b8; font-size: .875rem; cursor: pointer; transition: all .15s; }
.back-btn:hover { color: #f1f5f9; border-color: #6366f1; }
.logo { display: flex; align-items: center; gap: 8px; font-size: .95rem; font-weight: 700; color: #f1f5f9; text-decoration: none; }

.main { max-width: 1400px; width: 100%; margin: 0 auto; padding: 20px; flex: 1; }

.chan-toggle { display: flex; gap: 6px; margin-bottom: 16px; }
.chan-toggle button {
  padding: 6px 16px; border-radius: 999px; border: 1px solid #334155;
  background: #1e293b; color: #94a3b8; font-size: .82rem; cursor: pointer; transition: all .15s;
}
.chan-toggle button:hover { color: #f1f5f9; }
.chan-toggle button.active { background: #6366f1; border-color: #6366f1; color: #fff; }

.page-head { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; margin-bottom: 20px; }
.page-title { margin: 0; font-size: 1.25rem; font-weight: 700; color: #f1f5f9; }
.count { color: #64748b; font-size: .9rem; font-weight: 500; margin-left: 6px; }

.search-wrap { display: flex; align-items: center; gap: 8px; background: #1e293b; border: 1px solid #334155; border-radius: 8px; padding: 0 14px; height: 38px; width: 280px; max-width: 100%; transition: border-color .15s; }
.search-wrap:focus-within { border-color: #6366f1; }
.search-icon { color: #64748b; flex-shrink: 0; }
.search-input { background: transparent; border: none; outline: none; color: #f1f5f9; font-size: .875rem; flex: 1; min-width: 0; }
.search-clear { background: none; border: none; color: #64748b; cursor: pointer; font-size: .9rem; }

.actor-grid { display: grid; grid-template-columns: repeat(6, 1fr); gap: 16px 14px; }
.actor-card { display: flex; flex-direction: column; gap: 8px; background: none; border: none; padding: 0; cursor: pointer; text-align: center; }
.actor-thumb-wrap { position: relative; aspect-ratio: 3/4; border-radius: 8px; overflow: hidden; background: #1e293b; }
.actor-thumb { width: 100%; height: 100%; object-fit: cover; transition: transform .2s; }
.actor-card:hover .actor-thumb { transform: scale(1.05); }
.actor-thumb-ph { width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; font-size: 2rem; font-weight: 700; color: #475569; }
.actor-count { position: absolute; bottom: 6px; right: 6px; background: rgba(15,23,42,.85); color: #cbd5e1; font-size: .72rem; font-weight: 600; padding: 2px 7px; border-radius: 999px; }
.actor-name { margin: 0; font-size: .82rem; color: #cbd5e1; font-weight: 500; line-height: 1.3; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.actor-card:hover .actor-name { color: #818cf8; }

.skeleton { display: flex; flex-direction: column; gap: 8px; }
.skeleton-thumb { aspect-ratio: 3/4; border-radius: 8px; background: linear-gradient(90deg,#1e293b 25%,#273549 50%,#1e293b 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.skeleton-line { height: 10px; width: 70%; margin: 0 auto; border-radius: 4px; background: linear-gradient(90deg,#1e293b 25%,#273549 50%,#1e293b 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { 0%{background-position:200% 0} 100%{background-position:-200% 0} }

.empty-state { display: flex; align-items: center; justify-content: center; min-height: 200px; color: #94a3b8; }

@media (max-width: 1024px) { .actor-grid { grid-template-columns: repeat(4, 1fr); } }
@media (max-width: 720px)  { .actor-grid { grid-template-columns: repeat(3, 1fr); gap: 12px 10px; } }
@media (max-width: 480px)  { .actor-grid { grid-template-columns: repeat(2, 1fr); } }
</style>
