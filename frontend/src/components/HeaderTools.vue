<template>
  <div class="header-tools">
    <LangToggle />
    <ThemeToggle />
  </div>
</template>

<script setup lang="ts">
import LangToggle from './LangToggle.vue'
import ThemeToggle from './ThemeToggle.vue'
</script>

<style scoped>
.header-tools { margin-left: auto; display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
@media (max-width: 520px) { .header-tools { gap: 6px; } }
</style>
