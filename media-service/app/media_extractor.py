"""
Media Extractor — logic inti diadaptasi dari riset yt-dlp sebelumnya.
Hanya fokus ekstraksi metadata & stream URL, TIDAK ada download permanen,
TIDAK ada auth sendiri (trust boundary dijaga di level network/gRPC caller).
"""

import yt_dlp
from enum import Enum


class AudioQuality(str, Enum):
    LOW = "128"
    MEDIUM = "192"
    HIGH = "320"


AUDIO_QUALITY_MAP = {
    1: AudioQuality.LOW,     # AUDIO_QUALITY_LOW
    2: AudioQuality.MEDIUM,  # AUDIO_QUALITY_MEDIUM
    3: AudioQuality.HIGH,    # AUDIO_QUALITY_HIGH
    0: AudioQuality.MEDIUM,  # AUDIO_QUALITY_UNSPECIFIED -> default medium
}


class MediaExtractionError(Exception):
    """Dilempar saat yt-dlp gagal ekstrak info/stream."""
    pass


def get_media_info(youtube_url: str) -> dict:
    """
    Ambil metadata video YouTube (judul, channel, durasi, thumbnail).
    TIDAK mendownload apapun — cukup untuk disimpan permanen di Music Service (Postgres).
    """
    ydl_opts = {
        "quiet": True,
        "no_warnings": True,
        "noplaylist": True,
    }
    try:
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(youtube_url, download=False)
            video_id = info.get("id")
            return {
                "video_id": video_id or "",
                "title": info.get("title") or "",
                "channel": info.get("channel") or info.get("uploader") or "",
                "duration_seconds": int(info.get("duration") or 0),
                "thumbnail_url": info.get("thumbnail")
                or (f"https://i.ytimg.com/vi/{video_id}/hqdefault.jpg" if video_id else ""),
            }
    except yt_dlp.DownloadError as e:
        raise MediaExtractionError(f"URL video tidak valid atau tidak tersedia: {e}")
    except Exception as e:
        raise MediaExtractionError(f"Terjadi kesalahan saat mengambil info: {e}")


def get_audio_stream_url(video_id: str, quality: AudioQuality = AudioQuality.MEDIUM) -> str:
    """
    Ambil URL stream audio LANGSUNG dari YouTube (signed URL, expired dalam beberapa jam).
    PENTING: URL ini TIDAK PERNAH disimpan ke database manapun — hanya diteruskan
    sekali pakai ke caller (Music Service -> Gateway -> Client).
    """
    youtube_url = f"https://www.youtube.com/watch?v={video_id}"
    ydl_opts = {
        "format": f"bestaudio[abr<={quality.value}]/bestaudio",
        "quiet": True,
        "no_warnings": True,
        "noplaylist": True,
    }
    try:
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(youtube_url, download=False)
            return info["url"]
    except yt_dlp.DownloadError as e:
        raise MediaExtractionError(f"URL video tidak valid atau tidak tersedia: {e}")
    except Exception as e:
        raise MediaExtractionError(f"Terjadi kesalahan: {e}")

def search_youtube(query: str, limit: int = 10) -> list[dict]:
    """
    Cari video di YouTube via yt-dlp (ytsearch) TANPA download.
    Kembalikan daftar kandidat ringan untuk ditampilkan ke user.
    """
    if limit < 1 or limit > 50:
        limit = 10

    ydl_opts = {
        "quiet": True,
        "no_warnings": True,
        "noplaylist": True,
        "extract_flat": True,  # cepat: tidak resolve tiap video penuh, cukup metadata ringkas
        "skip_download": True,
    }
    try:
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(f"ytsearch{limit}:{query}", download=False)
    except Exception as e:
        raise MediaExtractionError(f"Gagal mencari di YouTube: {e}")

    entries = info.get("entries") or []
    results = []
    for entry in entries:
        if not entry:
            continue
        video_id = entry.get("id") or ""
        results.append({
            "video_id": video_id,
            "title": entry.get("title") or "",
            "channel": entry.get("channel") or entry.get("uploader") or "",
            "duration_seconds": int(entry.get("duration") or 0),
            "thumbnail_url": entry.get("thumbnail")
            or (f"https://i.ytimg.com/vi/{video_id}/hqdefault.jpg" if video_id else ""),
        })
    return results
