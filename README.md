# FileNeko

<p align="center">
  <img src="./branding/banner.png" width="550" alt="FileNeko"/>
</p>

FileNeko provides a file managing interface within a specified directory, with a bright anime theme inspired by a magical library. Upload, delete, preview and edit your files through a self-hosted web interface, with user management and configurable WebDAV access.

This fork is based on File Browser, whose upstream repository was archived on 2026-09-01. Upstream releases and images do not include the changes in FileNeko.

## Security

Published upstream advisories are listed under [security advisories](https://github.com/filebrowser/filebrowser/security/advisories). The [local security audit](docs/security-audit-2026-11-15.md) records the fixes and checks made in this fork, including session revocation and command handling. This audit does not guarantee the absence of other vulnerabilities. [SECURITY.md](SECURITY.md) retains the upstream reporting policy for historical reference.

When deploying FileNeko:

- **Do not expose it directly to the internet.** Put it behind a reverse proxy that terminates TLS and performs its own authentication.
- **Keep the command runner disabled.** It is off by default, so leave it off. See [#5199](https://github.com/filebrowser/filebrowser/issues/5199) and [`docs/command-execution.md`](docs/command-execution.md).
- **Run it unprivileged, inside a container**, with only the directory you intend to serve mounted into it.

## Documentation

Documentation on how to install, configure, and build this project lives in [`docs`](docs) in this repository.

[CONTRIBUTING.md](CONTRIBUTING.md) documents how to build and develop FileNeko. Executable names, configuration paths and Go module imports retain their upstream identifiers for compatibility.

## License

[Apache License 2.0](LICENSE) © File Browser Contributors
