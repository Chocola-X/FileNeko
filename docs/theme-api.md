# FileNeko CSS Theme API v1

**English** | [简体中文](theme-api_CN.md)

Set the directory path under **Global Settings → Branding → Branding directory path**, and save your theme stylesheet as `custom.css` within that directory. Place images, logos, and fonts in `img/` and reference them in CSS using `url("./img/filename")`. This relative path syntax works seamlessly for both root and subpath deployments.
You can copy the [example theme](personalized-examples/custom.css) to get started. For configuration and caching details, see [Customization](customization.md).

## Public Variables

Set global variables on `:root`, or scope them to specific component classes in the table below to limit their impact.
Unset variables retain their built-in theme defaults. Background variables accept colors, gradients, or image URLs; variables restricted to a single type are noted in the table.

| Variable | Description |
| --- | --- |
| `--font-ui`, `--font-heading` | Interface typography, branding, and heading fonts. Heading fonts inherit interface fonts by default. |
| `--font-mono`, `--font-signature` | Editor and monospace information; English file suffixes, Folder badges, and signature font. Do not apply these fonts to icon glyphs. |
| `--page-background-color`, `--page-background-image` | Global page base color and background layer, also applied to startup loading screen. Accepts colors, `url(...)`, gradients, or `none`. |
| `--page-background-size`, `--page-background-position`, `--page-background-repeat` | Background tiling and positioning for global and startup pages. Single wallpaper setups typically use `cover`, `center`, `no-repeat`. |
| `--login-background-image`, `--login-background-size`, `--login-background-position`, `--login-background-repeat` | Standalone background and positioning for the login page. Background accepts `url(...)`, gradients, or colors. |
| `--header-background`, `--sidebar-background`, `--login-panel-background` | Background colors for the header bar, navigation sidebar, and login card. |
| `--file-card-background`, `--file-card-inner-background`, `--file-card-frame-color` | File card outer background, inner decorative surface, and inner dashed frame color. List view may simplify decoration. |
| `--settings-card-background`, `--dialog-background`, `--loading-card-background` | Base background for settings cards, modal dialogs, and loading cards; header bars, button areas, and inputs may have independent fills. |
| `--card-background` | Background for public share cards and status cards (e.g., error pages, empty directories); can be overridden per page or card. |
| `--login-frame-opacity`, `--sidebar-frame-opacity`, `--settings-frame-opacity` | Dashed border opacity for corresponding cards, ranging from 0 to 1. |
| `--dialog-backdrop` | Backdrop overlay for modal dialogs and mobile overlay menus. |
| `--ink-navy`, `--ink-slate`, `--ink-muted` | Typography colors for primary body text, secondary text, and subtle hints. |
| `--cobalt-main`, `--cobalt-dark`, `--cobalt-light`, `--cobalt-surface` | Blue palette: primary, border, light, and surface tint. |
| `--pink-main`, `--pink-dark`, `--pink-light`, `--pink-surface` | Pink palette: primary, border, light, and surface tint. |
| `--cyan-main`, `--cyan-dark`, `--cyan-light`, `--cyan-surface` | Cyan palette: primary, border, light, and surface tint. |
| `--purple-main`, `--purple-dark`, `--purple-light`, `--purple-surface` | Purple palette: primary, border, light, and surface tint. |

Palettes affect components utilizing them (e.g., buttons, borders, shadows, and file type highlights); they do not override dedicated illustration or specialized component colors. To override colors when dark mode is enabled, define variables under `:root.dark`. When using custom fonts, retain fallback fonts and ensure you hold appropriate usage and redistribution rights.

When using local wallpapers, set `background-image: url("./img/filename")` directly on `.fn-page-background` and `.fn-page--login` as demonstrated in the example. This ensures paths resolve relative to `custom.css`. Passing relative image URLs through variables read by built-in stylesheets may resolve them relative to the page or base styles; if using images inside variables, specify the full site path including any deployment subpath.

## Stable Selectors

The following class names are guaranteed to appear on their corresponding elements. You can set variables on these classes or apply standard CSS rules, for example: `.fn-page--share { --card-background: #fff4fa; }`.
If a more specific built-in rule already targets an element, prefer adjusting its designated variable; custom stylesheets load after built-in styles and take precedence at equal specificity.

| Selector | Targeted Element / Area |
| --- | --- |
| `.fn-page-background` | Global page and startup loading screen wallpaper layer. |
| `.fn-page--login`, `.fn-page--files`, `.fn-page--settings`, `.fn-page--share` | Page containers for login, file manager, settings, and public share pages. |
| `.fn-login-panel`, `.fn-logo`, `.fn-wordmark` | Login form card, login/header logo image, and branding text. |
| `.fn-header`, `.fn-sidebar`, `.fn-storage` | Header bar, navigation sidebar, and sidebar storage indicator card. |
| `.fn-file-card`, `.fn-file-card__surface`, `.fn-file-card__suffix` | File cards, inner decorative surface, and suffix/Folder badge text. |
| `.fn-settings-card`, `.fn-share-card` | Management cards in settings, public share page cards. |
| `.fn-dialog`, `.fn-dialog-title`, `.fn-backdrop` | Modal dialog container, dialog title bar, and backdrop overlay for dialogs and mobile menus. Set `--dialog-background` on `.fn-dialog`. |
| `.fn-preview`, `.fn-editor`, `.fn-error` | Outer shell for file previews, text editor, and error pages/regions. Previewed documents and third-party widgets are outside this scope. |
| `.fn-loader` | Startup and in-page loading cards. |

## Compatibility Contract

`html[data-fn-theme-api="1"]` denotes the current Theme API version; standard v1 themes do not need to declare this attribute.
New variables and classes may be added within the same API version. Public names, purposes, and applicable elements will remain backward-compatible.
If internal implementations are renamed, legacy classes and variable entries will be preserved. If a public name must be removed or its behavior altered, migration guidelines and an explicit API version increment will precede the change. CI theme contract tests verify that v1 names remain active in templates and stylesheets.

The public interface does not rely on Vite asset hashes, Vue scoped `data-v-*` attributes, element tag names, or DOM nesting depths.
While those internal details may still be used for temporary adjustments in current builds, they are not covered by compatibility commitments.
The API guarantees stable selector targets and variable roles, allowing the default theme to evolve with updated palettes, spacing, and layouts.
