#!/usr/bin/env python3
"""Render the FileNeko banner using the bundled Fredoka font, without font links.

Requires fonttools, woff2_decompress and rsvg-convert. Run from any working directory.
The generated SVG contains glyph outlines so viewers need no installed font.
"""

import copy
from pathlib import Path
import subprocess
import tempfile
import xml.etree.ElementTree as ET

from fontTools.pens.boundsPen import BoundsPen
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen
from fontTools.ttLib import TTFont
from fontTools.varLib.instancer import instantiateVariableFont


HERE = Path(__file__).resolve().parent
SVG = "http://www.w3.org/2000/svg"
ET.register_namespace("", SVG)


def element(tag, **attrs):
    return ET.Element(f"{{{SVG}}}{tag}", {key.replace("_", "-"): str(value) for key, value in attrs.items()})


def main():
    with tempfile.TemporaryDirectory(prefix="fileneko-wordmark-") as folder:
        source = Path(folder) / "fredoka.woff2"
        source.write_bytes((HERE.parent / "frontend/src/assets/fonts/fredoka.woff2").read_bytes())
        subprocess.run(["woff2_decompress", str(source)], check=True)
        with TTFont(source.with_suffix(".ttf")) as variable:
            font = instantiateVariableFont(variable, {"wght": 700}, inplace=False)
    glyphs = font.getGlyphSet()
    cmap = font.getBestCmap()
    paths = []
    bounds = BoundsPen(glyphs)
    advance = 0
    for word, color, shadow in (("File", "#4E8DF5", "#1E4EB8"), ("Neko", "#FF6B9D", "#C93B6B")):
        pen = SVGPathPen(glyphs)
        for char in word:
            glyph = glyphs[cmap[ord(char)]]
            transform = (1, 0, 0, -1, advance, 0)
            glyph.draw(TransformPen(pen, transform))
            glyph.draw(TransformPen(bounds, transform))
            advance += glyph.width
        paths.append((pen.getCommands(), color, shadow))

    x_min, y_min, x_max, y_max = bounds.bounds
    scale = min(1770 / (x_max - x_min), 290 / (y_max - y_min))
    left = 620 - x_min * scale
    top = (516 - (y_max - y_min) * scale) / 2 - y_min * scale
    root = element("svg", width=2486, height=516, viewBox="0 0 2486 516", role="img", aria_labelledby="banner-title")
    title = element("title", id="banner-title")
    title.text = "FileNeko"
    root.append(title)
    mascot = copy.deepcopy(ET.parse(HERE / "logo.svg").getroot())
    mascot.set("x", "8")
    mascot.set("y", "8")
    mascot.set("width", "500")
    mascot.set("height", "500")
    root.append(mascot)

    for offset, outlined in (("translate(5 7)", False), ("", True)):
        group = element("g", transform=f"{offset} translate({left:.3f} {top:.3f}) scale({scale:.6f})")
        for path, color, shadow in paths:
            group.append(element("path", d=path, fill=color if outlined else shadow,
                                 stroke="#FFFFFF" if outlined else shadow,
                                 stroke_width=10, stroke_linejoin="round",
                                 paint_order="stroke fill", vector_effect="non-scaling-stroke"))
        root.append(group)

    ET.indent(root)
    ET.ElementTree(root).write(HERE / "banner.svg", encoding="utf-8", xml_declaration=True)
    subprocess.run(["rsvg-convert", "-o", str(HERE / "banner.png"), str(HERE / "banner.svg")], check=True)


if __name__ == "__main__":
    main()
