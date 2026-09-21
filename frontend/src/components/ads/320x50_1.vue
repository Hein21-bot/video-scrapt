<template>
  <div class="ad-sticky-wrap">
    <iframe ref="adFrame" width="320" height="50" frameborder="0" scrolling="no" style="border:none;display:block;"></iframe>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const adFrame = ref<HTMLIFrameElement | null>(null)

const html = `<!DOCTYPE html><html><head>
<script>atOptions={'key':'f6cf14a18159c586f8f45e899a8002af','format':'iframe','height':50,'width':320,'params':{}}<\/script>
<script data-cfasync="false" src="https://www.highperformanceformat.com/f6cf14a18159c586f8f45e899a8002af/invoke.js"><\/script>
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
  background: rgba(15, 23, 42, 0.9);
  backdrop-filter: blur(4px);
}
</style>
