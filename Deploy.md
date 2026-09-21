# Deploy — NightMM (free hosting)

Step-by-step. Everything here is **$0/month**.

```
┌─────────────────┐     ┌──────────────────────┐     ┌─────────────────────┐
│ Cloudflare Pages │──▶ │  Render (Docker web) │──▶ │ MongoDB Atlas  M0   │
│ frontend/  (SPA) │/api │  backend/cmd/api     │     │  free 512 MB        │
└─────────────────┘     └──────────────────────┘     └─────────────────────┘
        public site            public read-only API          database

        ┌─────────── stays on your Mac, never deployed ───────────┐
        │  backend/cmd/admin  ·  admin/  (panel)  ·  bridge/ (TG) │
        └────────────────────────────────────────────────────────┘
```

**Why this split:** the deployed half has zero admin code and never proxies video
bytes (videos play straight from the CDN, thumbnails are 302 redirects), so the
Render free tier is genuinely enough. You add/manage videos from your Mac.

---

## 0. One-time: install the CLIs (optional but easier)

```bash
brew install mongodb/brew/mongodb-database-tools   # mongodump / mongorestore
npm  install -g wrangler                           # Cloudflare (optional)
```

You also need accounts on: **GitHub**, **MongoDB Atlas**, **Render**, **Cloudflare**.

---

## 1. Database — MongoDB Atlas (free M0)

1. Go to <https://cloud.mongodb.com> → sign up → **Create** a cluster.
2. Pick **M0 (Free)**, any provider, a region close to your Render region
   (use `Singapore` for both Render and Atlas → pick AWS `ap-southeast-1`; closest to Thailand/Myanmar).
3. **Security → Database Access → Add New Database User**
   - username: `nightmm`  ·  password: *(generate, copy it)*
   - role: **Read and write to any database**
4. **Security → Network Access → Add IP Address → Allow access from anywhere**
   (`0.0.0.0/0`). Render's outbound IP isn't fixed on the free tier, so this is
   required.
5. **Database → Connect → Drivers** → copy the connection string. It looks like:
   ```
   mongodb+srv://nightmm:<password>@cluster0.xxxxx.mongodb.net/?retryWrites=true&w=majority
   ```
   Replace `<password>` with the real password. Keep this — it's your `MONGO_URI`.
   The app uses database name `videoscraper` (from `MONGO_DB`), you don't need to
   create it — it appears on first write.

---

## 2. Move your data to Atlas

You currently have ~1367 Chinese-AV videos + your Myanmar videos in local Mongo.

### Option A — copy what you have (keeps everything)

```bash
# dump local
mongodump --uri="mongodb://localhost:27017" --db=videoscraper --out=/tmp/nmm-dump

# restore into Atlas
mongorestore --uri="mongodb+srv://nightmm:<password>@cluster0.xxxxx.mongodb.net" \
  --nsInclude="videoscraper.*" /tmp/nmm-dump
```

### Option B — start fresh on Atlas

Skip the dump. After step 4 below, point your **local admin** at Atlas and run a
sync from the admin panel (`Scrape` page → run). Chinese-AV re-scrapes from
`3xchina.page`; re-add Myanmar videos from Telegram.

Verify either way:
```bash
mongosh "mongodb+srv://nightmm:<password>@cluster0.xxxxx.mongodb.net/videoscraper" \
  --eval 'db.listings.countDocuments()'
```

---

## 3. Put the project on GitHub

The project is not a git repo yet.

```bash
cd "/Users/heinminhtet/Desktop/video scrapt"
git init
git add .
git status          # <-- CHECK: no .env, no bridge/config.ini, no *.session listed
git commit -m "Initial commit"
```

`.gitignore` already excludes every secret file (`**/.env`, `bridge/config.ini`,
`bridge/*.session`, caches, `node_modules`). If `git status` shows any of those,
stop and fix `.gitignore` before pushing.

Create an **empty private** repo on GitHub (no README), then:
```bash
git remote add origin git@github.com:<you>/nightmm.git
git branch -M main
git push -u origin main
```

---

## 4. Backend — Render (free Docker web service)

The repo already has `render.yaml` (a Blueprint) and `backend/Dockerfile`.

1. <https://dashboard.render.com> → **New → Blueprint**.
2. Connect the GitHub repo. Render finds `render.yaml` and proposes a service
   named **nightmm-api**.
3. It will ask for the env vars marked `sync: false`. Fill them:

   | Key             | Value                                                        |
   |-----------------|-------------------------------------------------------------|
   | `MONGO_URI`     | the Atlas `mongodb+srv://…` string from step 1              |
   | `MONGO_DB`      | `videoscraper` (already set in the blueprint)               |
   | `API_TOKEN`     | `<your-API_TOKEN>` (or a new random hex — see note) |
   | `ADMIN_PASSWORD`| any password (only used if you ever run admin against Render — you won't; set anything) |
   | `JWT_SECRET`    | leave — the blueprint auto-generates it                     |
   | `CORS_ORIGINS`  | leave blank for now, set it in step 6                       |

   > **Do NOT set** `BRIDGE_URL` or `ADMIN_PORT` — the public API ignores them and
   > `BRIDGE_URL` blank is what keeps the Telegram code dormant in production.

4. **Apply / Create**. First build takes ~3–5 min.
5. When live, open `https://nightmm-api.onrender.com/health` → should show
   `{"ok":true}`. Copy that base URL.

**About `API_TOKEN`:** it's a soft gate that keeps random bots off the API, not a
real secret. With the Cloudflare setup below, the browser never sees it — only
Render and the Cloudflare Function know it. If you want a fresh one:
```bash
openssl rand -hex 24
```
Use the same value in Render (`API_TOKEN`) and Cloudflare (`API_TOKEN`, step 5).

---

## 5. Frontend — Cloudflare Pages (free)

The repo already has the two pieces Pages needs:
- `frontend/functions/api/[[path]].js` — proxies `/api/*` to Render and injects
  the API token server-side.
- `frontend/public/_redirects` — SPA fallback so deep links (`/watch/...`) work.

1. <https://dash.cloudflare.com> → **Workers & Pages → Create → Pages →
   Connect to Git** → pick the repo.
2. Build settings:

   | Field                    | Value            |
   |--------------------------|------------------|
   | Production branch        | `main`           |
   | Framework preset         | `None` (or Vue)  |
   | **Root directory**       | `frontend`       |
   | Build command            | `npm run build`  |
   | Build output directory   | `dist`           |

3. **Environment variables** (Settings → Environment variables → Production
   *and* Preview):

   | Key           | Value                                             |
   |---------------|---------------------------------------------------|
   | `API_ORIGIN`  | `https://nightmm-api.onrender.com` (your Render URL, no trailing slash) |
   | `API_TOKEN`   | same value as Render's `API_TOKEN`                |
   | `NODE_VERSION`| `20`                                              |

   > `API_ORIGIN` and `API_TOKEN` are read by the Pages **Function** at runtime —
   > they are never bundled into the browser JS. You do **not** need
   > `VITE_API_TOKEN` here.

4. **Save and Deploy.** You get a URL like `https://nightmm.pages.dev`.
5. Test: open it, the Chinese-AV grid should load, a video should play, and
   `https://nightmm.pages.dev/watch/anything` should not 404.

> **Build fails on type errors?** `npm run build` runs `vue-tsc` first. If a type
> error blocks the deploy, change the build command to just `vite build` in the
> Pages settings.

---

## 6. Connect the two — CORS

Back in **Render → nightmm-api → Environment**:

```
CORS_ORIGINS = https://nightmm.pages.dev
```

Add your custom domain too if you set one, comma-separated:
```
CORS_ORIGINS = https://nightmm.pages.dev,https://nightmm.com
```

Save → Render redeploys automatically. Done — the public site is live.

---

## 7. Keep Render awake (free tier sleeps after 15 min idle)

1. <https://uptimerobot.com> → free account → **Add New Monitor**.
2. Type: **HTTP(s)** · URL: `https://nightmm-api.onrender.com/health` ·
   Interval: **5 minutes**.

First visitor after a cold start waits ~30–50 s; with this, the service almost
never sleeps.

---

## 8. Run the admin side locally (against Atlas)

Your Mac is now the control room. One-time: point local config at Atlas.

**`backend/.env`** — change one line:
```
MONGO_URI=mongodb+srv://nightmm:<password>@cluster0.xxxxx.mongodb.net
```
(leave `MONGO_DB=videoscraper`, keep `API_TOKEN` matching Render, keep
`BRIDGE_URL=http://127.0.0.1:9001`)

**`bridge/config.ini`** — `[site]` section:
```
url = http://localhost:8081       # keep — the bridge talks to your LOCAL admin API
```

Then whenever you want to add/manage videos:

```bash
# terminal 1 — admin API (has sync + bridge proxy)
cd backend && go run ./cmd/admin

# terminal 2 — admin panel UI  → http://localhost:5174/admin
cd admin && npm run dev

# terminal 3 — Telegram bridge (only when adding Myanmar videos)
cd bridge && .venv/bin/python server.py
```

Add a video → it writes to Atlas → it's live on the public site within seconds.
When you're done, close all three. The public site keeps running.

> The public API also runs locally if you want to preview exactly what visitors
> see: `cd backend && PORT=8090 go run ./cmd/api`, then
> `cd frontend && API_PROXY=http://localhost:8090 npm run dev`.
> (Port 8090 because 8080 is used by another project on your machine.)

---

## 9. Keepalive for Myanmar videos (StreamWish/VidHide GC inactive files)

Roughly once a month, with the bridge config pointing at your **Render** URL
temporarily — or just run it against local admin while `cmd/admin` is up:

```bash
cd bridge && .venv/bin/python keepalive.py
```

It pings every Myanmar video so the free hosts don't delete them (~120-day
inactivity limit). Chinese-AV needs no keepalive — it re-scrapes.

---

## 10. Redeploying after code changes

```bash
git add . && git commit -m "…" && git push
```

- Render rebuilds `backend/` automatically on push to `main`.
- Cloudflare Pages rebuilds `frontend/` automatically on push to `main`.
- `admin/` and `bridge/` are local only — nothing to deploy, just `git pull` on
  other machines.

---

## Env var cheat-sheet

| Where            | Keys |
|------------------|------|
| **Render** (public API) | `MONGO_URI`, `MONGO_DB`, `JWT_SECRET` (auto), `API_TOKEN`, `ADMIN_PASSWORD`, `CORS_ORIGINS` |
| **Cloudflare Pages**    | `API_ORIGIN`, `API_TOKEN`, `NODE_VERSION` |
| **Local `backend/.env`**| `MONGO_URI` (Atlas), `MONGO_DB`, `ADMIN_PASSWORD`, `JWT_SECRET`, `API_TOKEN`, `PORT`, `ADMIN_PORT`, `BRIDGE_URL` |
| **Local `bridge/config.ini`** | telegram api_id/api_hash/session, streamwish + vidhide keys, `[site] url` = `http://localhost:8081` |

## Free-tier limits you're living within

| Service         | Limit | Fits because |
|-----------------|-------|--------------|
| Render free web | sleeps after 15 min idle, 512 MB RAM, 750 h/mo | UptimeRobot ping; Go binary uses ~15 MB; one service = 730 h |
| MongoDB Atlas M0| 512 MB storage, shared CPU | metadata only, no video/images stored — ~1 KB/doc, 1400 docs ≈ 2 MB |
| Cloudflare Pages| 500 builds/mo, unlimited bandwidth/requests | static assets + a tiny proxy function |
| StreamWish/VidHide | delete files inactive ~120 days | monthly `keepalive.py` |
