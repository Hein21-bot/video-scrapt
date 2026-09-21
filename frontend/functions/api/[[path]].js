// Cloudflare Pages Function — proxies /api/* to the Render backend so the
// frontend keeps using relative URLs (apiFetch, <img src="/api/thumb/..">) and
// the API token never reaches the browser.
//
// Pages project → Settings → Environment variables (Production + Preview):
//   API_ORIGIN  = https://<your-service>.onrender.com
//   API_TOKEN   = <same value as the backend's API_TOKEN>
export async function onRequest({ request, env }) {
  if (!env.API_ORIGIN) return new Response('API_ORIGIN not configured', { status: 500 })

  const url = new URL(request.url)
  const headers = new Headers(request.headers)
  if (env.API_TOKEN) headers.set('X-Api-Token', env.API_TOKEN)

  const resp = await fetch(env.API_ORIGIN.replace(/\/$/, '') + url.pathname + url.search, {
    method: request.method,
    headers,
    body: ['GET', 'HEAD'].includes(request.method) ? undefined : request.body,
    redirect: 'manual', // let /api/thumb's 302 reach the browser
  })

  return new Response(resp.body, {
    status: resp.status,
    statusText: resp.statusText,
    headers: resp.headers,
  })
}
