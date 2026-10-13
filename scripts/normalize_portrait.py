"""Normalize a portrait for compact, deterministic document embedding."""
from pathlib import Path
import sys

from PIL import Image, ImageOps

source = Path(sys.argv[1])
target = Path(sys.argv[2])
target.parent.mkdir(parents=True, exist_ok=True)
with Image.open(source) as image:
    normalized = ImageOps.fit(image.convert("RGB"), (420, 420), method=Image.Resampling.LANCZOS)
    normalized.save(target, format="JPEG", quality=84, optimize=True, progressive=True)
