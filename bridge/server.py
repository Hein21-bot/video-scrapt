#!/usr/bin/env python3
"""
Local bridge service — the admin panel's "Add from Telegram" talks to this
(via the Go backend proxy). Runs on 127.0.0.1 only.

    python server.py

Then set  BRIDGE_URL=http://127.0.0.1:9001  in the backend's .env and restart it.
"""
import asyncio
import itertools
import threading
import time
from queue import Queue

from flask import Flask, jsonify, request

from common import load_config
from core import get_client, process_link

cfg = load_config()
HOST = cfg["bridge"].get("host", "127.0.0.1") if cfg.has_section("bridge") else "127.0.0.1"
PORT = int(cfg["bridge"].get("port", "9001")) if cfg.has_section("bridge") else 9001

app = Flask(__name__)

_jobs = {}                 # id -> job dict
_ids = itertools.count(1)
_queue: "Queue" = Queue()
_lock = threading.Lock()


def _job(**kw):
    with _lock:
        jid = next(_ids)
        _jobs[jid] = {"id": jid, "status": "queued", "message": "", "result": None,
                      "created": time.time(), **kw}
    return jid


def _set(jid, **kw):
    with _lock:
        if jid in _jobs:
            _jobs[jid].update(kw)


def worker():
    """Single background thread with its own asyncio loop + one Telegram client."""
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    client = None
    try:
        client = loop.run_until_complete(get_client(cfg))
        print(f"[bridge] telegram client ready")
    except Exception as e:
        print(f"[bridge] WARNING: telegram login failed ({e}); jobs will fail until fixed")

    while True:
        jid = _queue.get()
        job = _jobs.get(jid)
        if not job:
            _queue.task_done()
            continue
        _set(jid, status="processing", message="starting")
        try:
            if client is None:  # retry login lazily
                client = loop.run_until_complete(get_client(cfg))
            res = loop.run_until_complete(process_link(
                cfg, job["link"], job.get("title", ""),
                channel=job.get("channel", ""), category=job.get("category", ""),
                actors=job.get("actors", ""), tags=job.get("tags", ""), client=client,
                on_progress=lambda s, j=jid: _set(j, message=s),
            ))
            _set(jid, status="done", message=f"{res['title']} ({res['size']})", result=res)
        except Exception as e:
            _set(jid, status="failed", message=str(e))
        finally:
            _queue.task_done()


@app.get("/health")
def health():
    return jsonify(ok=True)


@app.post("/add")
def add():
    data = request.get_json(force=True, silent=True) or {}
    link = (data.get("link") or "").strip()
    if "t.me/" not in link:
        return jsonify(error="a t.me/ message link is required"), 400
    jid = _job(
        link=link,
        title=(data.get("title") or "").strip(),
        channel=(data.get("channel") or "").strip(),
        category=(data.get("category") or "").strip(),
        actors=(data.get("actors") or "").strip(),
        tags=(data.get("tags") or "").strip(),
    )
    _queue.put(jid)
    return jsonify(id=jid, status="queued")


@app.get("/jobs")
def jobs():
    with _lock:
        items = sorted(_jobs.values(), key=lambda j: j["id"], reverse=True)[:50]
    return jsonify(jobs=items)


if __name__ == "__main__":
    threading.Thread(target=worker, daemon=True).start()
    app.run(host=HOST, port=PORT, threaded=True)
