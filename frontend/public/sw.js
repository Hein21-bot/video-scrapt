// Bump this whenever the caching strategy changes; old caches are deleted on activate.
const CACHE = 'vs-v4'

self.addEventListener('install', () => {
  self.skipWaiting()
})

self.addEventListener('activate', e => {
  e.waitUntil((async () => {
    const stale = (await caches.keys()).filter(k => k !== CACHE)
    await Promise.all(stale.map(k => caches.delete(k)))
    await self.clients.claim()
    // A phone that had the old worker is showing the old build; reload it once so it
    // picks up the current one. (Skipped on a first-ever install: nothing was stale.)
    if (stale.length) {
      const wins = await self.clients.matchAll({ type: 'window' })
      wins.forEach(c => c.navigate(c.url))
    }
  })())
})

self.addEventListener('fetch', e => {
  const req = e.request
  const url = new URL(req.url)

  // Only handle our own GETs. Ads, video hosts and the API always go straight to the network.
  if (req.method !== 'GET' || url.origin !== self.location.origin) return
  if (url.pathname.startsWith('/api/') || url.pathname.startsWith('/s/')) return

  // Hashed build files (/assets/index-abc123.js) never change → cache-first is safe and fast.
  if (url.pathname.startsWith('/assets/')) {
    e.respondWith(
      caches.match(req).then(hit => hit || fetch(req).then(resp => {
        if (resp.ok) { const copy = resp.clone(); caches.open(CACHE).then(c => c.put(req, copy)) }
        return resp
      }))
    )
    return
  }

  // Pages, manifest, icons: network-first so an update is picked up immediately;
  // fall back to the saved copy only when offline.
  e.respondWith(
    fetch(req).then(resp => {
      if (resp.ok) { const copy = resp.clone(); caches.open(CACHE).then(c => c.put(req, copy)) }
      return resp
    }).catch(() => caches.match(req).then(hit => hit || caches.match('/')))
  )
})
