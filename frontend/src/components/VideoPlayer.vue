<template>
  <div class="player-wrapper">
    <iframe
      v-if="type === 'iframe'"
      :src="url"
      class="video-el"
      allowfullscreen
      allow="autoplay; fullscreen; encrypted-media"
      referrerpolicy="origin"
      scrolling="no"
    />
    <video v-else ref="videoEl" controls playsinline referrerpolicy="origin" class="video-el" />
    <p v-if="hlsError" class="player-error">{{ hlsError }}</p>
    <div v-if="qualityLevels.length > 1" class="quality-bar">
      <button
        :class="['q-btn', { active: selectedLevel === -1 }]"
        @click="setQuality(-1)"
      >Auto</button>
      <button
        v-for="(l, i) in qualityLevels"
        :key="i"
        :class="['q-btn', { active: selectedLevel === i }]"
        @click="setQuality(i)"
      >{{ l.height }}p</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import Hls from 'hls.js'

const props = defineProps<{
  url: string
  type: 'mp4' | 'hls' | 'iframe'
}>()

const emit = defineEmits<{
  (e: 'error'): void
  (e: 'ended'): void
}>()

const videoEl      = ref<HTMLVideoElement | null>(null)
const hlsError     = ref('')
const qualityLevels = ref<{ height: number }[]>([])
const selectedLevel = ref(-1)
let hlsInstance: Hls | null = null

function destroyHls() {
  if (hlsInstance) {
    hlsInstance.destroy()
    hlsInstance = null
  }
}

function setQuality(level: number) {
  selectedLevel.value = level
  if (hlsInstance) hlsInstance.currentLevel = level
}

function loadVideo() {
  if (props.type === 'iframe') { destroyHls(); return } // host-hosted player
  const video = videoEl.value
  if (!video || !props.url) return

  destroyHls()
  hlsError.value      = ''
  qualityLevels.value = []
  selectedLevel.value = -1
  video.pause()
  video.removeAttribute('src')
  video.load()

  if (props.type === 'hls') {
    // Prefer hls.js wherever MSE is available (Chrome/Firefox/Edge). Chromium's
    // canPlayType() answers "maybe" for HLS but its native playback is unreliable,
    // so native HLS is only a fallback for browsers without MSE (Safari/iOS).
    if (Hls.isSupported()) {
      hlsInstance = new Hls({
        enableWorker: true,
        xhrSetup(xhr) {
          xhr.withCredentials = false
        },
      })
      hlsInstance.loadSource(props.url)
      hlsInstance.attachMedia(video)
      hlsInstance.on(Hls.Events.MANIFEST_PARSED, (_evt, data) => {
        qualityLevels.value = data.levels.map(l => ({ height: l.height }))
        video.play().catch(() => {})
      })
      hlsInstance.on(Hls.Events.ERROR, (_evt, data) => {
        if (data.fatal) {
          hlsError.value = `Playback error: ${data.details}`
          destroyHls()
          emit('error')
        }
      })
    } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = props.url
      video.play().catch(() => {})
    } else {
      hlsError.value = 'HLS is not supported in this browser.'
      emit('error')
    }
  } else {
    video.src = props.url
    video.addEventListener('error', () => emit('error'), { once: true })
    video.play().catch(() => {})
  }
}

function onVideoEnded() {
  emit('ended')
}

onMounted(() => {
  loadVideo()
  videoEl.value?.addEventListener('ended', onVideoEnded)
})
watch(() => [props.url, props.type] as const, loadVideo)
onBeforeUnmount(() => {
  destroyHls()
  videoEl.value?.removeEventListener('ended', onVideoEnded)
})
</script>

<style scoped>
.player-wrapper {
  width: 100%;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}
.video-el {
  width: 100%;
  display: block;
  max-height: 75vh;
}
iframe.video-el {
  aspect-ratio: 16 / 9;
  height: auto;
  border: 0;
}
.player-error {
  color: #f87171;
  padding: 8px 12px;
  font-size: 0.875rem;
}

.quality-bar {
  display: flex;
  gap: 6px;
  padding: 8px 10px;
  background: #0f172a;
  flex-wrap: wrap;
}
.q-btn {
  padding: 3px 10px;
  border-radius: 4px;
  border: 1px solid #334155;
  background: transparent;
  color: #94a3b8;
  font-size: .75rem;
  cursor: pointer;
  transition: all .15s;
}
.q-btn:hover { border-color: #6366f1; color: #f1f5f9; }
.q-btn.active { background: #6366f1; border-color: #6366f1; color: #fff; }
</style>
