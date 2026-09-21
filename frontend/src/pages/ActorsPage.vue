<template>
  <div class="actors-app">
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
      <div class="chan-toggle">
        <button :class="{ active: site === 'channel2' }" @click="setSite('channel2')">Chinese AV</button>
        <button :class="{ active: site === 'channel1' }" @click="setSite('channel1')">Myanmar</button>
      </div>

      <div class="page-head">
        <h1 class="page-title">{{ t('actors.title') }} <span class="count">{{ filtered.length }}</span></h1>
        <div class="search-wrap">
          <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
            <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35"/>
          </svg>
          <input v-model="q" class="search-input" :placeholder="t('actors.searchPlaceholder')" />
          <button v-if="q" class="search-clear" @click="q = ''">✕</button>
        </div>
      </div>

      <div v-if="loading" class="actor-grid">
        <div v-for="i in 24" :key="i" class="skeleton">
          <div class="skeleton-thumb" /><div class="skeleton-line" />
        </div>
      </div>

      <div v-else-if="filtered.length === 0" class="empty-state">
        <p>{{ t('actors.noMatch', { q }) }}</p>
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
import HeaderTools from '../components/HeaderTools.vue'
import { t } from '../utils/i18n'

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
.actors-app { min-height: 100vh; display: flex; flex-direction: column; background: var(--bg); }

.header { background: rgb(var(--bg-rgb) / 0.95); border-bottom: 1px solid var(--surface); position: sticky; top: 0; z-index: 50; }
.header-inner { max-width: 1400px; margin: 0 auto; padding: 0 20px; height: 56px; display: flex; align-items: center; gap: 20px; }
.back-btn { display: flex; align-items: center; gap: 6px; padding: 6px 14px; border-radius: 6px; border: 1px solid var(--border); background: transparent; color: var(--text-3); font-size: .875rem; cursor: pointer; transition: all .15s; }
.back-btn:hover { color: var(--text); border-color: var(--accent); }
.logo { display: flex; align-items: center; gap: 8px; font-size: .95rem; font-weight: 700; color: var(--text); text-decoration: none; }

.main { max-width: 1400px; width: 100%; margin: 0 auto; padding: 20px; flex: 1; }

.chan-toggle { display: flex; gap: 6px; margin-bottom: 16px; }
.chan-toggle button {
  padding: 6px 16px; border-radius: 999px; border: 1px solid var(--border);
  background: var(--surface); color: var(--text-3); font-size: .82rem; cursor: pointer; transition: all .15s;
}
.chan-toggle button:hover { color: var(--text); }
.chan-toggle button.active { background: var(--accent); border-color: var(--accent); color: var(--on-accent); }

.page-head { display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap; margin-bottom: 20px; }
.page-title { margin: 0; font-size: 1.25rem; font-weight: 700; color: var(--text); }
.count { color: var(--text-4); font-size: .9rem; font-weight: 500; margin-left: 6px; }

.search-wrap { display: flex; align-items: center; gap: 8px; background: var(--surface); border: 1px solid var(--border); border-radius: 8px; padding: 0 14px; height: 38px; width: 280px; max-width: 100%; transition: border-color .15s; }
.search-wrap:focus-within { border-color: var(--accent); }
.search-icon { color: var(--text-4); flex-shrink: 0; }
.search-input { background: transparent; border: none; outline: none; color: var(--text); font-size: .875rem; flex: 1; min-width: 0; }
.search-clear { background: none; border: none; color: var(--text-4); cursor: pointer; font-size: .9rem; }

.actor-grid { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 16px 14px; }
.actor-card { display: flex; flex-direction: column; gap: 8px; background: none; border: none; padding: 0; cursor: pointer; text-align: center; }
.actor-thumb-wrap { position: relative; aspect-ratio: 3/4; border-radius: 8px; overflow: hidden; background: var(--surface); }
.actor-thumb { width: 100%; height: 100%; object-fit: cover; transition: transform .2s; }
.actor-card:hover .actor-thumb { transform: scale(1.05); }
.actor-thumb-ph { width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; font-size: 2rem; font-weight: 700; color: var(--border-strong); }
.actor-count { position: absolute; bottom: 6px; right: 6px; background: rgb(var(--bg-rgb) / .85); color: var(--text-2); font-size: .72rem; font-weight: 600; padding: 2px 7px; border-radius: 999px; }
.actor-name { margin: 0; font-size: .82rem; color: var(--text-2); font-weight: 500; line-height: 1.3; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.actor-card:hover .actor-name { color: var(--accent-text); }

.skeleton { display: flex; flex-direction: column; gap: 8px; }
.skeleton-thumb { aspect-ratio: 3/4; border-radius: 8px; background: linear-gradient(90deg,var(--surface) 25%,var(--surface-2) 50%,var(--surface) 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.skeleton-line { height: 10px; width: 70%; margin: 0 auto; border-radius: 4px; background: linear-gradient(90deg,var(--surface) 25%,var(--surface-2) 50%,var(--surface) 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
@keyframes shimmer { 0%{background-position:200% 0} 100%{background-position:-200% 0} }

.empty-state { display: flex; align-items: center; justify-content: center; min-height: 200px; color: var(--text-3); }

@media (max-width: 1024px) { .actor-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); } }
@media (max-width: 720px)  { .actor-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px 10px; } }
@media (max-width: 480px)  { .actor-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 400px) {
  .logo-text { display: none; }
  .header-inner { padding: 0 12px; gap: 10px; }
}
</style>
