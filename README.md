# FileNeko

<p align="center">
  <img src="./branding/banner.png" width="550" alt="FileNeko" />
</p>

<p align="center">A lightweight, self-hosted file manager with an anime-inspired interface.</p>

<p align="center">
  <strong>English</strong> | <a href="README_CN.md">简体中文</a>
</p>

FileNeko is a personal fork of [File Browser](https://github.com/filebrowser/filebrowser), maintained by [Chocola-X](https://github.com/Chocola-X). It combines File Browser's web-based file management with a bright, magical-library-inspired interface, customizable CSS themes, WebDAV support and additional security work.

<img width="1920" height="1080" alt="demo1" src="https://github.com/user-attachments/assets/d1ad0f63-b51f-4d95-a334-b5a7ed721c3c" />

<img width="1920" height="1080" alt="demo2" src="https://github.com/user-attachments/assets/97f91a0b-7d4a-4e50-991f-c74a4fd6fd8c" />

<img width="1920" height="1080" alt="demo3" src="https://github.com/user-attachments/assets/1b2754de-bb32-4032-8c84-920530b297a5" />

<img width="1920" height="1080" alt="demo4" src="https://github.com/user-attachments/assets/4fde99ba-5dfe-4896-96ad-caf65a4aced3" />

<img width="1920" height="1080" alt="demo5" src="https://github.com/user-attachments/assets/957d72b8-b2c7-4d6d-80f5-a99fb2c583d3" />

<img width="1920" height="1080" alt="demo6" src="https://github.com/user-attachments/assets/d27a391a-9845-47ca-8b54-985d2ddef465" />

<img width="1920" height="1080" alt="demo7" src="https://github.com/user-attachments/assets/35741353-cc36-46cd-a25e-be77c73541e6" />

<img width="1920" height="1080" alt="demo8" src="https://github.com/user-attachments/assets/8ec6b2e4-1134-4a8e-9605-a48b444e4c4e" />


The upstream File Browser repository was archived on September 1, 2026. FileNeko is independently maintained; upstream binaries and container images **do not** contain FileNeko's changes.

## Features

- Manage files in a browser: upload, download, preview, edit, organize and share.
- Anime-inspired interface with optional custom branding and a [documented CSS theme API](docs/theme-api.md).
- Multiple users, configurable file access and user directories.
- Optional WebDAV access (disabled by default); see [WebDAV documentation](docs/webdav.md).
- Additional security fixes and access-control hardening; see the [security audit](docs/security-audit-2026-11-15.md).
- Self-contained executables for the platforms built by the project's CI; the frontend is embedded in the executable.

## Download and run

There is not yet an official FileNeko container image. **Do not use the archived upstream project's Docker images or installation scripts** to install FileNeko.

- **Development builds:** open [GitHub Actions](https://github.com/Chocola-X/FileNeko/actions), select a successful **Continuous Integration** run, and download the archive for your platform from **Artifacts**. Artifacts may expire and GitHub may require sign-in.
- **Tagged releases:** check [FileNeko Releases](https://github.com/Chocola-X/FileNeko/releases) for published release archives, if available.
- **Source build:** see [CONTRIBUTING.md](CONTRIBUTING.md).

After extracting a Linux archive, for example:

```sh
./fileneko -r /path/to/your/files
```

The Windows executable is `fileneko.exe`. On first launch, check the terminal output for the generated administrator password. For configuration, upgrade considerations and other platforms, read the [installation guide](docs/installation.md).

## Documentation

Install `ffmpeg` on the server to enable video thumbnails in the icon and
gallery views and posters in the player. Keep thumbnails enabled, make FFmpeg
available on the FileNeko service's `PATH`, and restart the service. An FFmpeg
build with the `fd` input protocol supports MP4 files whose metadata is at the
end. Use `--cacheDir` to reuse generated thumbnails. Videos keep their default
icon if a frame cannot be extracted.

Start with the [documentation index](docs/README.md), including [installation](docs/installation.md), [deployment and security](docs/deployment.md), [customization](docs/customization.md) and [WebDAV](docs/webdav.md).

Existing File Browser `FB_` environment variables, `filebrowser.db` databases and some `.filebrowser.*` configuration names are preserved for compatibility. The executable itself is named `fileneko` / `fileneko.exe`.

## Security

FileNeko is a personal project and is maintained on a best-effort basis. Do not expose the application directly to the public internet without appropriate protection. Prefer a TLS-terminating reverse proxy, restrict network and account access, run FileNeko with the least filesystem privileges needed, and leave command execution disabled unless you understand the risk.

The [local audit](docs/security-audit-2026-11-15.md) documents reviewed issues and fixes; it is **not** a guarantee that the application is vulnerability-free. Report issues following [SECURITY.md](SECURITY.md).

## Upstream and license

This is an unofficial derivative of [File Browser](https://github.com/filebrowser/filebrowser), not an official successor. FileNeko modifications and the derivative work are distributed under [GNU AGPLv3-only](LICENSE); original File Browser and third-party copyrights and license notices are retained. See [NOTICE](NOTICE), [LICENSES](LICENSES) and [licensing details](docs/licensing.md).
