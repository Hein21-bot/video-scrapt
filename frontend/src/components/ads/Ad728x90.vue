<template>
  <div v-if="show && KEY_728x90" class="ad-728-wrap">
    <iframe ref="frame" width="728" height="90" frameborder="0" scrolling="no" style="border:none;display:block;" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { AD_HOST, KEY_728x90 } from './adKeys'

// 728px is wider than any phone, so this is desktop/tablet only. It is mounted (and the
// ad loaded) only when the screen is wide enough to actually show it — ad networks
// penalise ads that load while hidden.
const mq = window.matchMedia('(min-width: 800px)')
const show = ref(mq.matches)
const onChange = (e: MediaQueryListEvent) => { show.value = e.matches }
const frame = ref<HTMLIFrameElement | null>(null)

function load() {
  if (!frame.value) return
  frame.value.srcdoc = `<!DOCTYPE html><html><head>
<script>atOptions={'key':'${KEY_728x90}','format':'iframe','height':90,'width':728,'params':{}}<\/script>
<script data-cfasync="false" src="${AD_HOST}/${KEY_728x90}/invoke.js"><\/script>
</head><body style="margin:0;padding:0;overflow:hidden;"></body></html>`
}

onMounted(() => {
  mq.addEventListener('change', onChange)
  if (show.value) nextTick(load)
})
watch(show, async v => { if (v) { await nextTick(); load() } })
onBeforeUnmount(() => mq.removeEventListener('change', onChange))
</script>

<style scoped>
.ad-728-wrap { display: flex; justify-content: center; width: 100%; overflow: hidden; }
</style>
