# Project Structure — what each folder is for

This project has **4 top-level folders**. Each one is a separate program with a
separate job. They talk to each other over HTTP, not by sharing code directly.

```
┌──────────────┐        ┌──────────────┐        ┌──────────────────┐
│   frontend/  │──/api─▶│   backend/   │◀──────▶│     MongoDB       │
│  (user site) │        │ (cmd/api)    │        │ (video metadata)   │
└──────────────┘        └──────────────┘        └──────────────────┘
   deployed                 deployed                  deployed
   (Cloudflare)             (Render)                  (Atlas)

┌──────────────┐        ┌──────────────┐        ┌──────────────┐
│    admin/    │──/api─▶│   backend/   │──────▶ │    bridge/    │
│ (admin panel)│        │ (cmd/admin)  │        │ (Telegram bot) │
└──────────────┘        └──────────────┘        └──────────────┘
   your Mac only          your Mac only            your Mac only
```

The big idea: **anything a visitor can reach is deployed to the internet.
Anything only you touch stays on your Mac.** That split is why hosting this
for free actually works — the public half is tiny and never touches Telegram,
your password, or upload keys.

---

## `backend/` — the Go server (the brain)

This is the only folder written in Go. It talks to MongoDB, scrapes video
sites, and answers every `/api/...` request. It is actually **two separate
programs** that share the same code, started differently:

```
backend/
  cmd/api/     → the PUBLIC program   — what real visitors hit    (deployed to Render)
  cmd/admin/   → the ADMIN program    — what only you use          (your Mac only)
  config/      → reads .env / environment variables
  db/          → connects to MongoDB, defines the collections
  models/      → the shape of a "video", "category", "channel", etc.
  scraper/     → the code that reads other websites (3xchina, muskuduu) and
                 figures out the real, playable video link
  handlers/    → one file per feature — this is where most logic lives:
      video.go       → video list, search, actors, "get me the play URL"
      categories.go  → the Myanmar sub-tabs (Japanese/English/Chinese/…)
      channels.go    → the main tabs (Chinese AV / Myanmar / Muskuduu / any you add)
      admin.go       → login, stats, add/delete a video
      sync.go        → runs the scraper on a schedule, writes results to MongoDB
      tgbridge.go    → forwards "add from Telegram" requests to bridge/
      share.go       → the shareable /s/<code> links
      thumb.go       → thumbnail redirects
      stream.go      → video redirects
  router/      → wires all of the above into two different sets of routes:
                 SetupPublic() for cmd/api, SetupAdmin() for cmd/admin
  middleware/  → the login-token checks
```

**Why two programs instead of one?** So the version deployed to the internet
(`cmd/api`) physically cannot serve `/api/admin/*` — there's no code path for
it, not just a password check. If someone finds your public site's address,
there is nothing admin-shaped to find.

**Nothing in `backend/` ever downloads or re-serves a video file.** It only
figures out the real video URL and hands it to the browser, which then talks
directly to the video's actual host (3xchina's CDN, muskuduu's CDN, etc). This
is why your server bandwidth stays close to zero even with a lot of visitors.

---

## `frontend/` — the user site (what visitors see)

A Vue 3 app. This is the **only folder deployed publicly as a website**
(to Cloudflare Pages). It has zero admin code in it at all — not hidden,
not disabled, just not present in the files that get built and shipped.

```
frontend/src/pages/
  HomePage.vue        → the main grid: channel tabs, sub-tabs, search, filters
  WatchPage.vue        → the video player page
  ActorsPage.vue       → the actress grid
  ActorVideosPage.vue → one actress's videos
  SharePage.vue        → opens a video from a shared /s/<code> link
```

It never talks to MongoDB directly — every piece of data comes from calling
`backend`'s public API (`cmd/api`) over `/api/...`.

---

## `admin/` — the admin panel (what only you see)

Also a Vue 3 app, but a **completely separate project** from `frontend/` —
different `package.json`, different build, runs on a different port. It is
**never deployed**; you only ever open it as `localhost` on your own Mac.

```
admin/src/pages/
  AdminLogin.vue      → password screen
  AdminDashboard.vue  → video counts per channel
  AdminVideos.vue      → video list + delete + "add video from Telegram" form
  AdminChannels.vue    → create/rename/delete main tabs (Chinese AV, Myanmar, …)
  AdminCategories.vue  → create/rename/delete Myanmar's sub-tabs
  AdminScrape.vue      → manually trigger a scrape of 3xchina / muskuduu
```

This talks to `backend`'s **admin** program (`cmd/admin`), which is the only
one that understands `/api/admin/*` routes.

---

## `bridge/` — the Telegram → video-host pipeline (a separate Python program)

The only folder **not** written in Go or Vue — it's plain Python, and it's the
only piece that talks to Telegram. It runs as its own small local web service.

```
bridge/
  server.py          → the local web service (127.0.0.1:9001) the admin panel
                        talks to; keeps a job queue so uploads happen in the
                        background while you keep using the admin panel
  core.py             → the actual work: download the video from Telegram,
                        upload it to a video host, tell the site about it
  common.py           → shared helpers (config loading, the upload API call,
                        the "add this video" API call to backend)
  tgadd.py            → do the same thing from the command line, one video
                        at a time, no server needed
  keepalive.py        → pings every Myanmar video occasionally so the free
                        video hosts don't delete them for being "inactive"
  config.ini          → your real Telegram/API keys (never committed to git)
  config.example.ini  → a blank template showing what config.ini needs
```

Flow when you use "Add from Telegram" in the admin panel:

```
admin panel  →  backend (cmd/admin)  →  bridge (server.py)
                                              │
                                    downloads video from Telegram
                                              │
                                  uploads it to a video host (gets a link)
                                              │
                              tells backend to save that link as a new video
```

This only runs on your Mac. It's what actually holds your Telegram login
session and video-host upload keys, which is exactly why it's never deployed.

---

## Quick reference: what's public vs. private

| Folder | Deployed? | Talks to |
|---|---|---|
| `backend/cmd/api` | ✅ Render | MongoDB, 3xchina.page, muskuduu.com |
| `frontend/` | ✅ Cloudflare Pages | `backend/cmd/api` only |
| `backend/cmd/admin` | ❌ your Mac only | MongoDB, `bridge/` |
| `admin/` | ❌ your Mac only | `backend/cmd/admin` only |
| `bridge/` | ❌ your Mac only | Telegram, video hosts, `backend/cmd/admin` |

- **`keepalive-worker/`** — a tiny Cloudflare Worker (runs free, 24/7). Every 13 minutes it
  pings the Render API so it never sleeps, and every Monday it "pins" every video uploaded
  from Telegram so the free video hosts don't delete them for inactivity.

Two more files worth knowing about:
- **`run.sh`** (project root) — starts/stops all 5 pieces at once for local use.
- **`Deploy.md`** (project root) — the step-by-step guide for putting the
  public half (`backend/cmd/api` + `frontend/`) on the internet for free.
