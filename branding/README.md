# FileNeko branding

The mascot is shared by the application icons and the horizontal banner. The
banner pairs it with a blue `File` and pink `Neko` wordmark in the bundled Fredoka
Bold font, with a white outline and solid shadow matching the application theme.

`banner.svg` uses vector glyph outlines; it does not load fonts or other external
assets. `banner.png` is the matching transparent raster export.

To regenerate both banners after changing the mascot or wordmark, install Python
`fonttools`, `woff2_decompress` and `rsvg-convert`, then run:

```sh
python3 branding/render-banner.py
```

The font and its license are included in `frontend/src/assets/fonts/`.
