# Local Theme Fonts

**English** | [简体中文](README_CN.md)

All application fonts are bundled and distributed with the frontend build, requiring no requests to Google Fonts. Neither development nor production builds link to external font services; Vite generates content-hashed static asset URLs that respect FileNeko's `StaticURL` configuration and CSS relative paths, supporting subpath deployments out of the box.

| Purpose | Local Font | Actual Weight |
| --- | --- | --- |
| English and Latin characters | Fredoka | Variable font 300–700 |
| Simplified Chinese | CuteRound (Mini Jian Cu Yuan) | Single 400 weight; other weights matched or synthesized by browser |
| Japanese, Kana, and CJK fallback | M PLUS Rounded 1c | Independent 400 and 700 weights |
| Code, terminal, and numeric accents | JetBrains Mono | Variable font 100–800 |
| File card English signature | FileSignature (Dancing Script) | Variable font 400–700 |
| Fallbacks for Latin, Cyrillic, Greek, and Vietnamese | Roboto subset fonts | 400, 500, 700 |

The unified font stack is defined in `src/css/_variables.css`. UI and headings prioritize rounded typography; when glyphs are missing, fallback proceeds sequentially through other local fonts, operating system CJK fonts, and generic sans-serif fonts. Code and terminal views use a dedicated monospace font stack; folder `Folder` badges and file suffixes use an independent handwritten font stack that does not change with UI locale. Material Icons continues to be provided locally by npm dependencies.

All text fonts use `font-display: swap`: fallback fonts render immediately until font downloads finish, ensuring the application remains readable and functional even if fonts fail to load. Declarations reflect the actual weights provided by font files; static weights are not misrepresented as variable fonts. Unbundled rare Chinese characters or other scripts are rendered using system fonts; exact rendering depends on the fonts installed on the user's device.

New fonts were sourced from the [official Google Fonts repository](https://github.com/google/fonts) on 2026-11-15 and compressed to full WOFF2 files using `woff2_compress`, preserving complete glyph coverage:

- [Fredoka](https://github.com/google/fonts/tree/main/ofl/fredoka): `Fredoka[wdth,wght].ttf`, accompanied by `fredoka-OFL.txt`.
- [JetBrains Mono](https://github.com/google/fonts/tree/main/ofl/jetbrainsmono): `JetBrainsMono[wght].ttf`, accompanied by `jetbrains-mono-OFL.txt`.
- [M PLUS Rounded 1c](https://github.com/google/fonts/tree/main/ofl/mplusrounded1c): `MPLUSRounded1c-Regular.ttf`, `MPLUSRounded1c-Bold.ttf`. Internal font metadata declares SIL Open Font License 1.1; `mplus-rounded-OFL.txt` preserves copyright notices and the full license text.

Dancing Script's license is preserved in `dancing-script-OFL.txt`. Pre-existing CuteRound, Dancing Script, and Roboto files are retained; newly added rounded and monospace fonts do not require network downloads during installation or building.

Ace editor modes, themes, snippets, and validation workers are similarly bundled as local static assets from installed `ace-builds` by `src/utils/editor-assets.ts`, loading on demand without CDN requests. Optional captcha services and external links followed intentionally by users are separate from font and editor assets.
