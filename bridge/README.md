# bridge/ — Telegram → video host → site

Runs **locally**, alongside the Go backend. Not deployed with the app.

```
Admin panel  "📥 Add from Telegram"  (link + title)
   → POST /api/admin/bridge/add   → Go backend proxies to →   bridge/server.py (127.0.0.1:9001)
                                                                    │  queue job
                                                                    ▼
                                          download (Telethon)  →  upload (StreamWish + VidHide)
                                                                    │
                                          POST /api/admin/videos  ←─┘   → appears in "Myanmar" channel
```

Admin panel shows a live job table (queued → processing → done / failed).

## Setup (once)

```bash
cd bridge
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
cp config.example.ini config.ini      # then fill it in
```

`config.ini`:

| key | where |
|---|---|
| `telegram.api_id` / `api_hash` | https://my.telegram.org → API development tools |
| `streamwish.api_key` | streamwish.com → Account → API |
| `vidhide.api_key` | vidhide.com → Account → API (optional mirror; blank = skip) |
| `site.url` / `admin_password` / `api_token` | your backend's values |

Use a **secondary** Telegram account (automated downloading breaks Telegram ToS).
It must be a **member** of the source channels (private channels included).

## Run

```bash
# 1. start the bridge (first run asks for your phone + login code, once)
cd bridge && source .venv/bin/activate && python server.py

# 2. point the backend at it
echo 'BRIDGE_URL=http://127.0.0.1:9001' >> backend/.env
#    then restart the backend

# 3. open the admin panel → "📥 Add from Telegram" → paste a t.me link → Queue
```

CLI alternative (no server / no admin panel):

```bash
python tgadd.py "https://t.me/somechannel/7661" --title "..."
```

## Keep-alive (weekly cron)

Free hosts delete files with no views for ~120 days. Ping them:

```
0 4 * * 1  cd "/path/video scrapt/bridge" && .venv/bin/python keepalive.py >> keepalive.log 2>&1
```

## Notes

- The site stores the **embed link**; on playback the backend extracts the m3u8
  and falls back to an `<iframe>` if the host changed its player.
- Big videos (1 GB+) download to a temp dir and are deleted after upload —
  keep ~2× the largest file free on disk.
- `BRIDGE_URL` unset (production / no bridge running) ⇒ the admin section hides,
  everything else works normally.
