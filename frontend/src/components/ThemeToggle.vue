<template>
  <button class="theme-toggle" :title="theme === 'dark' ? t('theme.toLight') : t('theme.toDark')" :aria-label="theme === 'dark' ? t('theme.toLight') : t('theme.toDark')" @click="toggleTheme">
    <svg v-if="theme === 'dark'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
      <circle cx="12" cy="12" r="4"/>
      <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/>
    </svg>
    <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
      <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z"/>
    </svg>
  </button>
</template>

<script setup lang="ts">
import { theme, toggleTheme } from '../utils/theme'
import { t } from '../utils/i18n'
</script>

<style scoped>
.theme-toggle {
  flex-shrink: 0;
  width: 36px; height: 36px;
  display: inline-flex; align-items: center; justify-content: center;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--accent);
  cursor: pointer;
  transition: border-color .15s, background .15s;
}
.theme-toggle:hover { border-color: var(--accent); background: var(--surface-2); }
</style>
