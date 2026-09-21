#!/usr/bin/env python3
"""
Weekly keep-alive: fetch a few bytes of every Myanmar-channel video so the
free video hosts count it as "active" and don't garbage-collect the file.

    python keepalive.py

Run from cron, e.g.:  0 4 * * 1  cd /path/bridge && python keepalive.py
"""
import sys
import requests

from common import load_config


def main():
    cfg = load_config()
    base = cfg["site"]["url"].rstrip("/")
    tok = cfg["site"]["api_token"]
    h = {"X-Api-Token": tok}

    # collect every video id in the manual channel
    ids, page = [], 1
    while True:
        r = requests.get(f"{base}/api/videos", params={"site": "channel1", "page": page}, headers=h, timeout=15)
        r.raise_for_status()
        items = r.json().get("items", [])
        if not items:
            break
        ids += [v["id"] for v in items]
        page += 1

    print(f"{len(ids)} videos")
    ok = fail = 0
    for vid in ids:
        try:
            r = requests.get(f"{base}/api/video-url", params={"id": vid}, headers=h, timeout=25)
            r.raise_for_status()
            data = r.json()
            urls = data.get("mirrors") or ([data["url"]] if data.get("url") else [])
            for u in urls:
                if data.get("type") == "iframe":
                    requests.get(u, headers={"Referer": u}, timeout=15)
                else:
                    requests.get(u, headers={"Range": "bytes=0-2047", "Referer": base + "/"}, timeout=15)
            ok += 1
        except Exception as e:
            fail += 1
            print(f"  {vid}: {e}")
    print(f"pinged {ok} ok / {fail} fail")
    sys.exit(1 if fail and not ok else 0)


if __name__ == "__main__":
    main()
