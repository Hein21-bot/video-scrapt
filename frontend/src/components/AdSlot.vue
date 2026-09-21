<template>
  <div :class="['ad-slot', `ad-slot--${size}`]">
    <span class="ad-label">{{ t('ad.label') }}</span>
    <!-- Paste your ad network code (AdSense, etc.) inside this component's slot -->
    <slot />
  </div>
</template>

<script setup lang="ts">
import { t } from '../utils/i18n'
defineProps<{
  /**
   * leaderboard  → 728×90  (full-width horizontal bar)
   * rectangle    → 300×250 (medium rectangle, sidebar)
   * halfpage     → 300×600 (half page, sidebar)
   */
  size: 'leaderboard' | 'rectangle' | 'halfpage'
}>()
</script>

<style scoped>
.ad-slot {
  background: var(--surface);
  border: 1px dashed var(--border);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--border-strong);
  font-size: .75rem;
  text-transform: uppercase;
  letter-spacing: .08em;
  flex-shrink: 0;
}
.ad-label { pointer-events: none; user-select: none; }

.ad-slot--leaderboard { width: 100%; height: 90px; }
.ad-slot--rectangle   { width: 100%; height: 250px; }
.ad-slot--halfpage    { width: 100%; height: 600px; }

@media (max-width: 768px) {
  .ad-slot--halfpage { height: 250px; }
}
</style>
