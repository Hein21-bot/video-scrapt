<template>
  <div v-if="KEY_320x50" class="ad-sticky-wrap">
    <iframe ref="adFrame" width="320" height="50" frameborder="0" scrolling="no" style="border:none;display:block;"></iframe>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { AD_HOST, KEY_320x50 } from './adKeys'

const adFrame = ref<HTMLIFrameElement | null>(null)

const html = `<!DOCTYPE html><html><head>
<script>atOptions={'key':'${KEY_320x50}','format':'iframe','height':50,'width':320,'params':{}}<\/script>
<script data-cfasync="false" src="${AD_HOST}/${KEY_320x50}/invoke.js"><\/script>
</head><body style="margin:0;padding:0;overflow:hidden;"></body></html>`

onMounted(() => {
  if (adFrame.value) adFrame.value.srcdoc = html
})
</script>

<style scoped>
.ad-sticky-wrap {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 50;
  display: flex;
  justify-content: center;
  background: rgb(var(--bg-rgb) / 0.9);
  backdrop-filter: blur(4px);
}
</style>
