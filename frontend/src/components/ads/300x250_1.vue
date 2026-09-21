<template>
  <div class="ad-300-wrap">
    <div ref="adContainer"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const adContainer = ref<HTMLDivElement | null>(null)

onMounted(() => {
  if (!adContainer.value) return

  const cfg  = document.createElement('script')
  cfg.text   = `atOptions={'key':'abe392b1e82cb06276cfba7fd8c1eef2','format':'iframe','height':250,'width':300,'params':{}};`
  adContainer.value.appendChild(cfg)

  const invoke = document.createElement('script')
  invoke.src   = 'https://www.highperformanceformat.com/abe392b1e82cb06276cfba7fd8c1eef2/invoke.js'
  adContainer.value.appendChild(invoke)
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
