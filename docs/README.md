# FileNeko documentation

**English** | [简体中文](README_CN.md)

FileNeko is a self-hosted browser-based file manager and an independently maintained personal fork of the archived [File Browser](https://github.com/filebrowser/filebrowser) project. For the project overview and available downloads, see the [repository README](../README.md).

## User guides

- [Installation and first run](installation.md) — binaries, CI artifacts and initial login
- [Deployment](deployment.md) — access restrictions, reverse proxy and login protection
- [Customization](customization.md) and [CSS theme API](theme-api.md)
- [Authentication](authentication.md)
- [WebDAV](webdav.md)
- [Command execution](command-execution.md) — disabled by default; security considerations
- [Folder statistics](folder-statistics.md) — streaming recursive size and file counts
- [Reverse proxy setup](reverse-proxy.md) — Nginx real IP and trusted proxy configuration
- [Troubleshooting](troubleshooting.md)
- [Command-line reference](cli/fileneko.md)

## Project information

- [Building and contributing](../CONTRIBUTING.md)
- [Security policy](../SECURITY.md) and [security audit](security-audit-2026-11-15.md)
- [Licensing and attribution](licensing.md)
- [Changelog](../CHANGELOG.md)

**Release status:** FileNeko does not currently publish official Docker images. The repository contains inherited Docker-related configuration, but the recommended install paths are FileNeko executable archives and building from source. The upstream File Browser images do not include this fork.
