<template>
  <button class="lang-toggle" :title="t('lang.switch')" :aria-label="t('lang.switch')" @click="toggleLang">
    <span :class="{ on: lang === 'my' }">မြန်မာ</span><i>/</i><span :class="{ on: lang === 'en' }">EN</span>
  </button>
</template>

<script setup lang="ts">
import { lang, toggleLang, t } from '../utils/i18n'
</script>

<style scoped>
.lang-toggle {
  flex-shrink: 0;
  height: 36px;
  padding: 0 10px;
  display: inline-flex; align-items: center; gap: 4px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text-4);
  font-size: .75rem; font-weight: 600;
  cursor: pointer;
  transition: border-color .15s, background .15s;
}
.lang-toggle:hover { border-color: var(--accent); background: var(--surface-2); }
.lang-toggle i { font-style: normal; opacity: .5; }
.lang-toggle .on { color: var(--accent); }
/* Phones: show only the active language to save header space. */
@media (max-width: 520px) {
  .lang-toggle { padding: 0 8px; }
  .lang-toggle i, .lang-toggle span:not(.on) { display: none; }
}
</style>
