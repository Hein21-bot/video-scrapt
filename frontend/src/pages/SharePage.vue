<template>
  <div class="share-page">
    <div class="share-box">
      <div v-if="loading" class="share-state">
        <div class="spinner-ring" />
        <p>{{ t('share.loading') }}</p>
      </div>
      <div v-else-if="err" class="share-state error">
        <p>{{ err }}</p>
        <RouterLink to="/" class="home-link">{{ t('share.goHome') }}</RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { t } from '../utils/i18n'
import { onMounted, ref } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { apiFetch } from '../utils/apiFetch'

const route  = useRoute()
const router = useRouter()

const loading = ref(true)
const err     = ref('')

onMounted(async () => {
  const code = route.params.code as string
  try {
    const res  = await apiFetch(`/api/s/${code}`)
    const data = await res.json()
    if (!res.ok) throw new Error(data.error ?? t('share.notFound'))

    const id = Math.random().toString(36).slice(2, 10)
    sessionStorage.setItem(`vid:${id}`, JSON.stringify({
      videoId:   data.video_id,
      title:     data.title,
      thumbnail: data.thumbnail,
      site:      data.site,
    }))
    router.replace(`/watch/${id}`)
  } catch (e) {
    err.value     = e instanceof Error ? e.message : 'Invalid share link'
    loading.value = false
  }
})
</script>

<style scoped>
.share-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
}
.share-box {
  text-align: center;
  color: var(--text-3);
}
.share-state { display: flex; flex-direction: column; align-items: center; gap: 16px; }
.share-state.error { color: #f87171; }
.spinner-ring { width: 40px; height: 40px; border: 3px solid rgb(var(--accent-rgb) / .3); border-top-color: var(--accent); border-radius: 50%; animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.home-link { color: var(--accent); text-decoration: none; font-size: .875rem; }
.home-link:hover { text-decoration: underline; }
</style>
