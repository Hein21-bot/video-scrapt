#!/usr/bin/env python3
"""
One-off: add a single Telegram video from the command line.

    python tgadd.py "https://t.me/somechannel/7661" [--title "..."]

For day-to-day use, run  server.py  instead and add links from the admin panel.
"""
import argparse
import asyncio

from common import load_config
from core import process_link


async def run(args):
    cfg = load_config()
    print(f"processing {args.link}")
    res = await process_link(cfg, args.link, args.title, channel=args.channel,
                             category=args.category, actors=args.actress, tags=args.tags,
                             on_progress=lambda s: print(f"  {s}"))
    print(f"done ✓  {res['title']}  ({res['size']})")
    print(f"  primary: {res['primary']}")
    if res["mirror"]:
        print(f"  mirror : {res['mirror']}")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("link", help="Telegram message link, e.g. https://t.me/chan/7661")
    ap.add_argument("--title", default="", help="override title (else uses the caption)")
    ap.add_argument("--channel", default="", help="channel key, e.g. channel1 (default: Myanmar)")
    ap.add_argument("--category", default="", help="japanese / english / chinese")
    ap.add_argument("--actress", default="", help="comma-separated, optional")
    ap.add_argument("--tags", default="", help="comma-separated, optional")
    asyncio.run(run(ap.parse_args()))


if __name__ == "__main__":
    main()
