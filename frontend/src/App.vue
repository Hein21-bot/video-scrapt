<template>
  <RouterView :key="$route.path" />
  <Ad320x50 />
</template>

<script setup lang="ts">
import Ad320x50 from './components/ads/320x50_1.vue'
</script>
