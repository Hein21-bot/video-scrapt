import { onMounted, onBeforeUnmount } from 'vue'
import { apiFetch } from './apiFetch'

// Render's free service sleeps after ~15 min idle (30-50 s cold start). While the
// site is open in a tab, ping a cheap API endpoint (it also touches the database)
// so it stays warm. Only in production builds — pointless in local dev.
//
// Only covers time when someone is browsing; the Cloudflare Worker in
// keepalive-worker/ keeps it awake 24/7.
const INTERVAL_MS = 13 * 60 * 1000

export function useKeepAlive() {
  if (!import.meta.env.PROD) return

  let timer: ReturnType<typeof setInterval> | undefined
  const ping = () => { apiFetch('/api/channels').catch(() => {}) }
  const onVisible = () => { if (document.visibilityState === 'visible') ping() }

  onMounted(() => {
    timer = setInterval(ping, INTERVAL_MS)
    document.addEventListener('visibilitychange', onVisible)
  })
  onBeforeUnmount(() => {
    if (timer) clearInterval(timer)
    document.removeEventListener('visibilitychange', onVisible)
  })
}
