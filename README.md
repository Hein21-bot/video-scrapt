# NightMM — video aggregator

Split into a **public** half (deployed) and an **admin** half (local only).

```
backend/
  cmd/api/     → public, read-only API        (deployed)      :8080
  cmd/admin/   → admin API + auto-sync + bridge (local only)  :8081
  (shared packages: config, db, models, scraper, handlers, middleware, router)

frontend/      → user site  (Home / Watch / Actors / Share)   (deployed)  :5173
admin/         → admin panel                                  (local)     :5174
bridge/        → Telegram → video-host → site  (local Python)              :9001
```

The public API has **no admin routes at all** (`/api/admin/*` → 404). The admin
API serves the admin routes *and* the public routes, so the admin panel and the
bridge only need one base URL (`:8081`).

## Run locally

```bash
# MongoDB
brew services start mongodb-community

# public API
cd backend && go run ./cmd/api            # :8080

# admin API (also runs the 6-hourly auto-sync)
cd backend && go run ./cmd/admin          # :8081

# user site
cd frontend && npm install && npm run dev # :5173  → proxies /api to :8080

# admin panel
cd admin && npm install && npm run dev    # :5174  → proxies /api to :8081
```

Admin login password = `ADMIN_PASSWORD` in `backend/.env` (default `admin123`).

For the Telegram bridge see [bridge/README.md](bridge/README.md).

## Deploy (public half only)

| part | where |
|---|---|
| `backend` (Dockerfile builds `cmd/api`) | Render Web Service — env: `MONGO_URI` (Atlas), `JWT_SECRET`, `API_TOKEN`, `CORS_ORIGINS` (your site origin) |
| `frontend` | Render Static Site — `npm run build`, publish `dist`, rewrite `/api/*` → backend URL |
| MongoDB | Atlas M0 |

`admin/` and `backend/cmd/admin` are **never deployed** — run them on your machine
against the same Atlas database. Keep `BRIDGE_URL` unset in production.
