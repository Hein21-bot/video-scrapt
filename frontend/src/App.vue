<template>
  <SideRail side="left" :ad-key="SIDE_LEFT_KEY" />
  <SideRail side="right" :ad-key="SIDE_RIGHT_KEY" />
  <RouterView :key="$route.path" />
  <Ad320x50 />
  <!--
    Popunder1 disabled: it opens a new tab on the visitor's first click anywhere
    on the page (including a video card), not just on the ad itself.
    To turn back on: uncomment this line and the import below.
    <Popunder1 />
  -->
  <SocialBar1 />
</template>

<script setup lang="ts">
import Ad320x50 from './components/ads/320x50_1.vue'
import SideRail from './components/ads/SideRail.vue'
// import Popunder1 from './components/ads/Popunder_1.vue'
import SocialBar1 from './components/ads/SocialBar_1.vue'
import { SIDE_LEFT_KEY, SIDE_RIGHT_KEY } from './components/ads/adKeys'
import { useKeepAlive } from './utils/keepAlive'

useKeepAlive()
</script>
