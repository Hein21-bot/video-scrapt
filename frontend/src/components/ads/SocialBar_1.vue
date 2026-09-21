<template>
  <div></div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { ENABLE_SOCIAL_BAR, SOCIAL_BAR_SRC } from './adKeys'

const SCRIPT_ID = 'social-bar-1'

// One script for the whole site; it injects its own floating bar.
onMounted(() => {
  if (!ENABLE_SOCIAL_BAR || document.getElementById(SCRIPT_ID)) return
  const script = document.createElement('script')
  script.id    = SCRIPT_ID
  script.async = true
  script.setAttribute('data-cfasync', 'false')
  script.src   = SOCIAL_BAR_SRC
  document.body.appendChild(script)
})
</script>
