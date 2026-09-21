<template>
  <div class="ad-300-wrap">
    <iframe ref="frame" width="300" height="250" frameborder="0" scrolling="no" style="border:none;display:block;" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { AD_HOST, KEY_300x250 } from './adKeys'

const frame = ref<HTMLIFrameElement | null>(null)

onMounted(() => {
  if (!frame.value || !KEY_300x250) return
  frame.value.srcdoc = `<!DOCTYPE html><html><head>
<script>atOptions={'key':'${KEY_300x250}','format':'iframe','height':250,'width':300,'params':{}}<\/script>
<script data-cfasync="false" src="${AD_HOST}/${KEY_300x250}/invoke.js"><\/script>
</head><body style="margin:0;padding:0;overflow:hidden;"></body></html>`
})
</script>

<style scoped>
.ad-300-wrap {
  display: flex;
  justify-content: center;
  width: 100%;
  overflow: hidden;
}
</style>
