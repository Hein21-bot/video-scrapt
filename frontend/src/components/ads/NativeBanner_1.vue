<template>
  <div v-if="isOwner" id="container-5157180e710ec7725ab3e794109d781b" class="native-wrap"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

let activeInstance: symbol | null = null

const SCRIPT_ID  = 'native-banner-1'
const SCRIPT_SRC = 'https://pl29751079.effectivecpmnetwork.com/5157180e710ec7725ab3e794109d781b/invoke.js'

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
  script.src   = SCRIPT_SRC
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
</style>
