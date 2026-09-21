<template>
  <div></div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'

const SCRIPT_ID  = 'popunder-1'
const SCRIPT_SRC = 'https://pl29751081.effectivecpmnetwork.com/52/57/17/52571775184d099cc6151ae631218f5a.js'

onMounted(() => {
  if (document.getElementById(SCRIPT_ID)) return
  const script = document.createElement('script')
  script.id    = SCRIPT_ID
  script.async = true
  script.setAttribute('data-cfasync', 'false')
  script.src   = SCRIPT_SRC
  document.head.appendChild(script)
})

onUnmounted(() => {
  document.getElementById(SCRIPT_ID)?.remove()
})
</script>
