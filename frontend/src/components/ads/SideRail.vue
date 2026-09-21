<template>
  <aside v-if="wide && (adKey || showPlaceholder)" :class="['rail', `rail--${side}`]">
    <iframe v-if="adKey" ref="frame" :width="WIDTH" :height="HEIGHT" frameborder="0" scrolling="no" class="rail-frame" />
    <div v-else class="rail-placeholder">Ad {{ WIDTH }}×{{ HEIGHT }}<br />{{ side }} side<br /><small>(dev only)</small></div>
  </aside>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { SIDE_WIDTH, SIDE_HEIGHT } from './adKeys'

const props = defineProps<{ side: 'left' | 'right'; adKey: string }>()

const WIDTH = SIDE_WIDTH
const HEIGHT = SIDE_HEIGHT
const showPlaceholder = import.meta.env.DEV
const frame = ref<HTMLIFrameElement | null>(null)

// Only mount (and so only load the ad) when the slot is actually visible —
// ad networks penalise ads that load while hidden.
const mq = window.matchMedia('(min-width: 1912px)')
const wide = ref(mq.matches)
const onChange = (e: MediaQueryListEvent) => { wide.value = e.matches }

function load() {
  if (!frame.value || !props.adKey) return
  // An isolated iframe per rail, so two rails never fight over the network's global atOptions.
  frame.value.srcdoc = `<!DOCTYPE html><html><head>
<script>atOptions={'key':'${props.adKey}','format':'iframe','height':${HEIGHT},'width':${WIDTH},'params':{}}<\/script>
<script data-cfasync="false" src="https://www.highperformanceformat.com/${props.adKey}/invoke.js"><\/script>
</head><body style="margin:0;padding:0;overflow:hidden;"></body></html>`
}

onMounted(() => {
  mq.addEventListener('change', onChange)
  if (wide.value) nextTick(load)
})
watch(wide, async v => { if (v) { await nextTick(); load() } })
onBeforeUnmount(() => mq.removeEventListener('change', onChange))
</script>

<style scoped>
/* Hidden unless the screen is wide enough: on >= 1912px the content column is
   narrowed to 1280px (see style.css), leaving 320px each side for a 300px ad. */
.rail { display: none; }

@media (min-width: 1912px) {
  .rail {
    display: block;
    position: fixed;
    top: 96px;
    width: 300px;
    z-index: 5;
  }
  .rail--left  { right: calc(50% + 656px); }
  .rail--right { left:  calc(50% + 656px); }
}

.rail-frame { display: block; border: none; }
.rail-placeholder {
  width: 300px; height: 250px;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  text-align: center; font-size: .8rem; line-height: 1.6;
  color: var(--text-4);
  border: 1px dashed var(--border-strong); border-radius: 8px;
}
</style>
