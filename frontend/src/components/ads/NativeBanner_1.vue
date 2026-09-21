<template>
  <div v-if="isOwner" :id="`container-${NATIVE_KEY}`" class="native-wrap"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { NATIVE_KEY, NATIVE_SRC } from './adKeys'

let activeInstance: symbol | null = null

const SCRIPT_ID  = 'native-banner-1'

const id      = Symbol()
const isOwner = ref(false)

onMounted(() => {
  if (activeInstance !== null) return
  activeInstance = id
  isOwner.value  = true

  document.getElementById(SCRIPT_ID)?.remove()
  const script = document.createElement('script')
  script.id    = SCRIPT_ID
  script.async = true
  script.setAttribute('data-cfasync', 'false')
  script.src   = NATIVE_SRC
  document.head.appendChild(script)
})

onUnmounted(() => {
  if (activeInstance === id) {
    activeInstance = null
    isOwner.value  = false
    document.getElementById(SCRIPT_ID)?.remove()
  }
})
</script>

<style scoped>
.native-wrap { width: 100%; min-height: 50px; }

/* The network renders a 2x2 grid of large tiles (~480px tall on a phone), which
   pushes the videos off-screen. Scale the whole widget down on phones instead of
   clipping it, so no tile is cut in half. Raise/lower the zoom to taste. */
@media (max-width: 640px) {
  .native-wrap { zoom: 0.6; }
}
</style>
