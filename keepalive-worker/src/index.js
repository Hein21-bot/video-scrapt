// Cloudflare Worker — keeps the free Render API (and the Atlas database) awake.
// A Cron Trigger (see wrangler.toml) calls scheduled() every 13 minutes. Render's
// free plan sleeps after 15 minutes without traffic, so this runs 24/7 whether or
// not anyone has the site open.
//
//   HEALTH_URL     (var)    the API's /health — wakes the Render service
//   API_TOKEN      (secret) optional; if set, also calls /api/channels, which runs a
//                           real database query so Atlas sees activity too
//   KEEPALIVE_KEY  (secret) optional; if set, the weekly cron also "pins" every video
//                           uploaded from Telegram (see below). Must equal KEEPALIVE_KEY on Render.
//
// Two crons (wrangler.toml):
//   */13 * * * *   ping — keeps Render awake
//   0 3 * * 1      pin  — Mondays 03:00 UTC: the API fetches the first bytes of every
//                  manually-added video's stream so the free video hosts (StreamWish /
//                  VidHide) count it as watched and don't delete it for inactivity.

const PIN_CRON = '0 3 * * 1'

export default {
  async scheduled(event, env, ctx) {
    ctx.waitUntil(event.cron === PIN_CRON ? pin(env) : ping(env))
  },

  // Visiting the Worker's URL in a browser also triggers a ping (handy for testing).
  // Add ?pin=1 to run the weekly video pin right now.
  async fetch(req, env) {
    if (!env.HEALTH_URL) return new Response('HEALTH_URL not set', { status: 500 })
    const doPin = new URL(req.url).searchParams.get('pin') === '1'
    const lines = doPin ? await pin(env) : await ping(env)
    return new Response(lines.join('\n'), { status: 200 })
  },
}

async function ping(env) {
  const out = []
  const health = env.HEALTH_URL
  if (!health) return ['HEALTH_URL not set']

  out.push(await hit(health, {}))

  if (env.API_TOKEN) {
    const api = new URL('/api/channels', health).toString()
    out.push(await hit(api, { 'X-Api-Token': env.API_TOKEN }))
  }
  return out
}

// Wake the API first (a sleeping Render service takes ~50 s), then start the pin.
async function pin(env) {
  const out = await ping(env)
  if (!env.KEEPALIVE_KEY) return [...out, 'pin skipped: KEEPALIVE_KEY secret not set']
  const url = new URL('/keepalive', env.HEALTH_URL).toString()
  out.push(await hit(url, { 'X-Keepalive-Key': env.KEEPALIVE_KEY }, 'POST'))
  return out
}

async function hit(url, headers, method = 'GET') {
  try {
    const r = await fetch(url, { method, headers: { 'user-agent': 'chitnya-keepalive', ...headers } })
    const line = `keepalive ${method} ${url} -> ${r.status}`
    console.log(line)
    return line
  } catch (e) {
    const line = `keepalive ${url} failed: ${e.message}`
    console.error(line)
    return line
  }
}
