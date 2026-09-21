<template>
  <div></div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { ENABLE_POPUNDER, POPUNDER_SRC } from './adKeys'

const SCRIPT_ID  = 'popunder-1'

onMounted(() => {
  if (!ENABLE_POPUNDER || document.getElementById(SCRIPT_ID)) return
  const script = document.createElement('script')
  script.id    = SCRIPT_ID
  script.async = true
  script.setAttribute('data-cfasync', 'false')
  script.src   = POPUNDER_SRC
  document.head.appendChild(script)
})

onUnmounted(() => {
  document.getElementById(SCRIPT_ID)?.remove()
})
</script>
