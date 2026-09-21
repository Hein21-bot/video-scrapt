<template>
  <div class="card" @click="$emit('select')" @mouseenter="$emit('hover')">
    <div class="thumb-wrap">
      <img
        v-if="thumbnail && !imgError"
        :src="thumbnail"
        :alt="title"
        loading="lazy"
        class="thumb"
        @error="imgError = true"
      />
      <div v-if="!thumbnail || imgError" class="thumb-placeholder">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32" height="32">
          <path d="M15.75 10.5l4.72-4.72a.75.75 0 011.28.53v11.38a.75.75 0 01-1.28.53l-4.72-4.72M4.5 18.75h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25h-9A2.25 2.25 0 002.25 7.5v9A2.25 2.25 0 004.5 18.75z"/>
        </svg>
      </div>
      <span v-if="isNew" class="new-badge">{{ t('card.new') }}</span>
      <div class="play-overlay">
        <div class="play-icon">
          <svg viewBox="0 0 24 24" fill="currentColor" width="36" height="36">
            <path d="M8 5v14l11-7z"/>
          </svg>
        </div>
      </div>
    </div>
    <p class="card-title">{{ title }}</p>
  </div>
</template>

<script setup lang="ts">
import { t } from '../utils/i18n'
import { ref } from 'vue'

defineProps<{
  title: string
  thumbnail: string
  isNew?: boolean
}>()

defineEmits<{ select: []; hover: [] }>()

const imgError = ref(false)
</script>

<style scoped>
.card {
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.thumb-wrap {
  position: relative;
  aspect-ratio: 16 / 9;
  background: var(--surface);
  border-radius: 6px;
  overflow: hidden;
}

.thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 0.3s ease;
}

.thumb-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--surface), var(--border));
  color: var(--border-strong);
}

.play-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  opacity: 0;
  transition: opacity 0.2s;
}

.play-icon {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: rgb(var(--accent-rgb) / 0.9);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--on-accent);
  transform: scale(0.85);
  transition: transform 0.2s;
}

.card:hover .play-overlay {
  opacity: 1;
}

.card:hover .play-icon {
  transform: scale(1);
}

.card:hover .thumb {
  transform: scale(1.04);
}

.new-badge {
  position: absolute;
  top: 6px;
  left: 6px;
  background: var(--accent);
  color: var(--on-accent);
  font-size: .65rem;
  font-weight: 700;
  letter-spacing: .06em;
  padding: 2px 6px;
  border-radius: 4px;
  z-index: 2;
}

.card-title {
  margin: 0;
  font-size: 0.82rem;
  color: var(--text-2);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
