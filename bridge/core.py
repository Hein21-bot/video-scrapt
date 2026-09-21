"""The actual work: Telegram message link -> download -> host upload -> site."""
import os
import re
import tempfile
import time

from telethon import TelegramClient
from telethon.tl.types import PeerChannel

from common import host_upload, site_token, site_add_video

HERE = os.path.dirname(os.path.abspath(__file__))


def parse_link(link):
    """Return (chat, msg_id). chat is a username str or an int peer id."""
    link = link.strip().split("?")[0].rstrip("/")
    m = re.search(r"t\.me/c/(\d+)/(\d+)$", link)
    if m:
        return int("-100" + m.group(1)), int(m.group(2))
    m = re.search(r"t\.me/([A-Za-z0-9_]+)/(\d+)$", link)
    if m:
        return m.group(1), int(m.group(2))
    raise ValueError(f"cannot parse Telegram link: {link}")


def human(n):
    for unit in ("B", "KB", "MB", "GB"):
        if n < 1024:
            return f"{n:.1f}{unit}"
        n /= 1024
    return f"{n:.1f}TB"


async def get_client(cfg):
    client = TelegramClient(
        os.path.join(HERE, cfg["telegram"]["session"]),
        int(cfg["telegram"]["api_id"]),
        cfg["telegram"]["api_hash"],
    )
    await client.start()
    return client


async def process_link(cfg, link, title="", channel="", category="", actors="", tags="",
                       client=None, on_progress=None):
    """Download the video at `link`, upload it, register it on the site.
    Returns a dict describing the result. Raises on hard failure."""
    chat, msg_id = parse_link(link)

    own_client = client is None
    if own_client:
        client = await get_client(cfg)

    try:
        entity = await client.get_entity(PeerChannel(chat) if isinstance(chat, int) else chat)
        msg = await client.get_messages(entity, ids=msg_id)
        if not msg or not msg.media:
            raise RuntimeError("message has no media")
        is_video = msg.video or (msg.document and (msg.document.mime_type or "").startswith("video"))
        if not is_video:
            raise RuntimeError("message media is not a video")

        cap = (msg.message or "").strip().splitlines()
        final_title = title.strip() or (cap[0].strip() if cap else "") or f"Video {msg_id}"

        tmp = tempfile.mkdtemp(prefix="tgbridge_")
        last = [0.0]

        def prog(cur, total):
            now = time.time()
            if now - last[0] > 1 or cur == total:
                last[0] = now
                if on_progress:
                    on_progress(f"download {human(cur)}/{human(total)}")

        path = await client.download_media(msg, file=tmp, progress_callback=prog)
        if not path or not os.path.exists(path):
            raise RuntimeError("download failed")
        size = os.path.getsize(path)
    finally:
        if own_client:
            await client.disconnect()

    try:
        sw = cfg["streamwish"]
        if not sw.get("api_key"):
            raise RuntimeError("streamwish api_key missing in config.ini")
        if on_progress:
            on_progress("uploading to StreamWish")
        primary = host_upload(path, sw["api_key"], sw["api_base"], sw["embed_host"])

        mirror = ""
        vh = cfg["vidhide"]
        if vh.get("api_key"):
            if on_progress:
                on_progress("uploading to VidHide")
            try:
                mirror = host_upload(path, vh["api_key"], vh["api_base"], vh["embed_host"])
            except Exception:
                mirror = ""  # mirror is best-effort

        site = cfg["site"]
        token = site_token(site["url"], site["admin_password"])
        site_add_video(site["url"], token, final_title, primary, mirror,
                       channel=channel, category=category, actors=actors, tags=tags)
    finally:
        try:
            os.remove(path)
            os.rmdir(os.path.dirname(path))
        except OSError:
            pass

    return {"title": final_title, "size": human(size), "primary": primary, "mirror": mirror}
