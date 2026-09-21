const CACHE = 'vs-v2'

self.addEventListener('install', e => {
  self.skipWaiting()
  e.waitUntil(
    caches.open(CACHE).then(c => c.addAll(['/', '/index.html']))
  )
})

self.addEventListener('activate', e => {
  e.waitUntil(
    caches.keys().then(keys =>
      Promise.all(keys.filter(k => k !== CACHE).map(k => caches.delete(k)))
    )
  )
})

self.addEventListener('fetch', e => {
  const url = new URL(e.request.url)
  // Never cache API calls or streams
  if (url.pathname.startsWith('/api/') || url.pathname.startsWith('/s/')) return

  e.respondWith(
    caches.match(e.request).then(cached => cached || fetch(e.request).then(resp => {
      if (resp.ok && e.request.method === 'GET') {
        const clone = resp.clone()
        caches.open(CACHE).then(c => c.put(e.request, clone))
      }
      return resp
    }))
  )
})
