# Keep-alive Worker

A tiny Cloudflare Worker whose Cron Trigger pings the Render API every 13 minutes
so the free service never goes to sleep (Render sleeps after 15 min idle, and a
cold start takes 30-50 s). It runs 24/7 — unlike the in-app ping in
`frontend/src/utils/keepAlive.ts`, which only works while a tab is open.

Cloudflare's free plan includes Cron Triggers.

## Deploy (one time)

```bash
cd keepalive-worker
npm install
npx wrangler login            # opens the browser once
npx wrangler deploy
```

Optional — also keep the database busy (calls `/api/channels`, a real DB query):

```bash
npx wrangler secret put API_TOKEN     # paste the same API_TOKEN Render uses
```

## Pin the videos uploaded from Telegram (so they are never deleted)

The free video hosts (StreamWish / VidHide) delete files that see no activity for
months. Every Monday 03:00 UTC the Worker tells the API to "pin" every video you
added from Telegram: it fetches the first bytes of each video's stream, exactly like a
player, so the host counts it as watched. It covers every manually-added video in
every channel, including ones you add later — nothing to configure per video.

Set it up once, using the **same** random value in both places:

```bash
openssl rand -hex 24                       # make a key, copy it

# 1) Render → your service → Environment → add  KEEPALIVE_KEY = <the key>
# 2) the Worker:
cd keepalive-worker
npx wrangler secret put KEEPALIVE_KEY      # paste the same key
npx wrangler deploy
```

Run it right now instead of waiting for Monday: open
`https://chitnya-keepalive.<your-subdomain>.workers.dev/?pin=1` in a browser.
Without `KEEPALIVE_KEY` on Render the endpoint is switched off (404).

## Check it's firing

```bash
npx wrangler tail             # live logs: "keepalive ... /health -> 200" every 13 min
```

Or open the Worker's URL in a browser — that pings once and shows the result.
Dashboard: **Workers & Pages → chitnya-keepalive → Settings → Triggers**, and **Logs**.

## Change the schedule / URL

Edit `crons` or `HEALTH_URL` in `wrangler.toml`, then `npx wrangler deploy`.

## Remove it

```bash
npx wrangler delete
```
