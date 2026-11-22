# Personalized Wallpaper and Xiaolai Font Example

**English** | [简体中文](README_CN.md)

This example can be used directly as a branding directory for FileNeko without recompiling the application. It provides separate wallpapers for the login screen and file manager, custom logos, and a local Xiaolai Mono font, while retaining the built-in blue/pink palette and layout.

## Visual Previews

![Login Screen](previews/login-desktop.webp)

![File Manager](previews/files-desktop.webp)

Previews are generated using the bundled assets and CI builds; the `previews/` directory is for documentation purposes only and does not need to be deployed.

## Directory Structure

```text
personalized-examples/
  README.md
  custom.css
  img/
    login.webp                  # Login screen wallpaper
    background.webp             # File manager and general background
    logo.svg                    # Automatically replaces login and header logos
    logo.png                    # Fallback raster logo (not loaded by default)
    fonts/
      XiaolaiMono-Regular.woff2 # Xiaolai monospace font
```

At runtime, only `custom.css` and the `img/` directory are required. `README.md` is not loaded by the application. Asset licensing and redistribution rights must be verified with the original creators of the images and fonts; the project's software license does not extend to these media assets.

## Configuration

1. Copy this directory to your server (e.g., `/srv/fileneko/branding`), preserving the folder structure, and ensure the user running FileNeko has read permissions.
2. Log in as an administrator, navigate to **Settings → Global Settings → Branding → Branding directory path**, and enter `/srv/fileneko/branding`. This must be an absolute path on the server filesystem, not a web URL or the `img/` subdirectory.
3. Save and refresh the page. The instance name can be configured separately; the sidebar footer continues to display FileNeko. Default favicon icons remain unchanged; this example replaces only the login and header logos.

Alternatively, stop the service and configure the database directly via CLI:

```sh
fileneko -d /path/to/filebrowser.db config set \
  --branding.files /srv/fileneko/branding
```

Then restart the service with your regular parameters. Do not write to the database using the CLI while the server process is running. When running inside a container, provide the path mounted inside the container.

## Styling Details

- The login page loads `img/login.webp`; the file manager, settings, and share pages load `img/background.webp`. A 70% opacity white overlay is applied over the file manager background (allowing 30% of the wallpaper to show through) to prevent the wallpaper from conflicting with UI elements; settings and share pages share this effect, while the login screen wallpaper remains unobstructed. You can tune the two `70%` values inside `linear-gradient()` on `.fn-page-background`. Wallpapers are centered and cover the viewport; aspect ratio differences may crop edges, which can be adjusted via `background-position` in CSS.
- The header, sidebar, login card, file cards, settings cards, dialogs, loaders, and status cards use `rgb(255 255 255 / 90%)`: 90% white opacity with 10% background translucency. The inner surface of file cards is transparent to avoid stacking white layers.
- Opacity is adjusted on background colors only rather than setting container-level `opacity`, keeping text, icons, thumbnails, and buttons crisp and readable. Input boxes, buttons, accent bars, and storage indicators retain their built-in colors.
- `@font-face` loads fonts locally to set UI, heading, monospace, and signature typography variables without interfering with Material Icons. The font file is ~11.3 MiB and may take a moment on first load; `font-display: swap` ensures fallback fonts display immediately.
- Asset URLs in CSS use relative paths, supporting both root and subpath deployments like `/library`. Do not move these relative URLs into variables read by built-in stylesheets.

For detailed interface documentation, see [Theme API v1](../theme-api.md); for other branding options, see [Customization](../customization.md).

## Troubleshooting and Verification

Hard-refresh the browser after making changes, or disable caching in developer tools. Branding assets may be cached for up to a day; other users will also need to refresh.

In the network tab, verify that `custom.css`, the two WebP files, `img/logo.svg`, and the font file load successfully. In subpath deployments, requests should preserve the prefix. If HTML or 404 responses are returned, check the directory path, file casing, and server read permissions. If only fallback fonts appear, check font requests and browser console font errors.

To restore default styling, clear the branding directory path in settings, save, and refresh; there is no need to delete the asset files.

## Verification Records

Verified using `temp/linux-amd64-fileneko/fileneko` (`36101f27`) and Chromium 133 on a temporary database and test directory:

- Theme, background, logo, and fonts load correctly under both root and `/library` subpaths.
- Browser confirmed headings compute to XiaolaiMono.
- Login card, header, and file cards compute background color to `rgba(255, 255, 255, 0.9)`.
- Verified desktop (1440×1000) and mobile (390×844 / 390×667) screenshots for login and file listing pages.

Previews represent light mode; background positioning can be tailored to match wallpaper compositions across different aspect ratios.
