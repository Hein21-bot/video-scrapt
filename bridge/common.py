"""Shared helpers for the Telegram -> video-host -> site bridge."""
import configparser
import os
import sys

import requests

HERE = os.path.dirname(os.path.abspath(__file__))


def load_config():
    path = os.path.join(HERE, "config.ini")
    if not os.path.exists(path):
        sys.exit("config.ini not found — copy config.example.ini to config.ini and fill it in")
    cfg = configparser.ConfigParser()
    cfg.read(path)
    return cfg


# ── video host upload (XFileSharing-style API: StreamWish / VidHide / …) ────────

def host_upload(file_path, api_key, api_base, embed_host):
    """Upload a local file, return its embed URL (https://<host>/e/<code>)."""
    r = requests.get(f"{api_base}/upload/server", params={"key": api_key}, timeout=30)
    r.raise_for_status()
    server = r.json().get("result")
    if not server:
        raise RuntimeError(f"upload/server gave no result: {r.text[:200]}")

    with open(file_path, "rb") as fh:
        r = requests.post(
            server,
            data={"key": api_key},
            files={"file": (os.path.basename(file_path), fh, "video/mp4")},
            timeout=None,
        )
    r.raise_for_status()
    data = r.json()
    files = data.get("files") or []
    if not files or files[0].get("filecode") in (None, ""):
        raise RuntimeError(f"upload failed: {r.text[:300]}")
    code = files[0]["filecode"]
    return f"{embed_host.rstrip('/')}/e/{code}"


# ── site admin API ────────────────────────────────────────────────────────────

def site_token(site_url, password):
    r = requests.post(f"{site_url.rstrip('/')}/api/admin/login",
                      json={"password": password}, timeout=15)
    r.raise_for_status()
    return r.json()["token"]


def site_add_video(site_url, token, title, video_url, mirror_url="", thumbnail="",
                   channel="", category="", actors="", tags=""):
    r = requests.post(
        f"{site_url.rstrip('/')}/api/admin/videos",
        headers={"Authorization": f"Bearer {token}"},
        json={"title": title, "video_url": video_url, "mirror_url": mirror_url,
              "thumbnail": thumbnail, "channel": channel, "category": category,
              "actors": actors, "tags": tags},
        timeout=15,
    )
    if not r.ok:
        raise RuntimeError(f"site add failed {r.status_code}: {r.text[:300]}")
    return r.json()
