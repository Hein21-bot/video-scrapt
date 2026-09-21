<template>
  <SideRail side="left" :ad-key="SIDE_LEFT_KEY" />
  <SideRail side="right" :ad-key="SIDE_RIGHT_KEY" />
  <RouterView :key="$route.path" />
  <Ad320x50 />
</template>

<script setup lang="ts">
import Ad320x50 from './components/ads/320x50_1.vue'
import SideRail from './components/ads/SideRail.vue'
import { SIDE_LEFT_KEY, SIDE_RIGHT_KEY } from './components/ads/adKeys'
import { useKeepAlive } from './utils/keepAlive'

useKeepAlive()
</script>
