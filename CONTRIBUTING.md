# Building and contributing to FileNeko

FileNeko is an independently maintained personal fork of the archived File Browser project. Contributions are welcome on a best-effort basis. FileNeko changes are distributed under AGPL-3.0-only, while original copyright and third-party notices are preserved; see [licensing details](docs/licensing.md).

## Project and prerequisites

The backend is written in Go; the frontend in Vue.js under `frontend/`. Build tools include a suitable Go toolchain, Node.js, pnpm and [Task](https://taskfile.dev/). Check the versions required by `go.mod`, `frontend/package.json` and the current CI before building.

```sh
git clone https://github.com/Chocola-X/FileNeko.git
cd FileNeko
task build
```

`task build` builds the frontend and then the backend. The backend build disables CGO and targets baseline AMD64 (`GOAMD64=v1`) and ARM64 (`GOARM64=v8.0`). Linux binaries produced this way do not depend on glibc or musl, although they still depend on a suitable Linux kernel.

The executable is named `fileneko` (`fileneko.exe` on Windows). The existing `FB_` environment variables, `filebrowser.db`, `.filebrowser.*` configuration conventions and Go module imports retain upstream identifiers for compatibility. Do not rename these identifiers merely for branding.

The frontend build stores compressed JavaScript resources for the Go embedded HTTP handler. Use `task build` for a distributable executable instead of assuming that a plain `go build` will regenerate frontend resources.

## CI artifacts and releases

The **Continuous Integration** GitHub Actions workflow builds a shared frontend and packages platform-specific binaries through a multi-platform GoReleaser snapshot matrix (currently 21 targets). The resulting `linux-amd64-fileneko.tar.gz` / `windows-amd64-fileneko.zip` style archives are published under the successful workflow run's **Artifacts** section, with per-platform checksums. Artifacts are retained for a limited time (currently documented as 30 days); GitHub may require sign-in for download.

CI snapshot artifacts are **development builds**, not official releases. Binaries contain the frontend and are checked for packaging and static-build requirements. The exact validation status is the status shown by GitHub Actions for the selected commit.

Release configuration also exists for tagged binary archives published through GitHub Releases. Verify that the corresponding GitHub Actions release workflow is enabled and has succeeded before advertising a release as available.

**Docker status:** FileNeko does not currently build or publish official Docker images. Inherited Dockerfiles and Compose configuration are not supported installation instructions. The GoReleaser configuration in the documentation patch disables Docker publication. Upstream File Browser images are not FileNeko.

The project does not publish upstream Homebrew packages, Docker Hub images or other upstream distribution channels.

For the matrix helper itself, run:

```sh
go test ./.github/ci
go run ./.github/ci matrix
```

### Older operating systems

- **CentOS 7 x86-64:** static Linux AMD64 builds avoid the old glibc requirement; this is not a claim of successful testing on CentOS 7.
- **Windows 7/8:** unsupported by the current Go toolchain and project dependencies; a separate port would be needed.
- **macOS:** check the current Go toolchain's minimum OS requirements; the existing Go 1.26 build targets macOS 12 or newer.

See the [Go minimum requirements](https://go.dev/wiki/MinimumRequirements) and [Go ports](https://go.dev/doc/go1.26#ports).

## Development

### Frontend

```sh
cd frontend
pnpm install
pnpm run build  # prepare production frontend assets
pnpm run dev    # run the frontend development server
```

For interactive frontend development, access the application through the frontend development server as described by its output.

### Backend

```sh
go mod download
go build -o fileneko .
go run .
```

For an end-user binary with freshly built embedded assets, prefer the `task build` sequence from the repository root.

## Command-line documentation

Project documentation is stored as Markdown under [`docs`](docs). The CLI reference under [`docs/cli`](docs/cli) can be regenerated from the application's commands:

```sh
task docs:cli:generate
```

## Creating a release

The `task release` task regenerates CLI documentation and invokes `commit-and-tag-version` through pnpm. It can create a local version commit and tag; **that is not itself proof that a remote GitHub Release has been published**. Check the tag, push behavior, corresponding workflow and published assets separately. Do not advertise a container release.

Release archives include the project and upstream notices. The source link embedded into built binaries refers to the corresponding FileNeko commit; keep matching source available under the AGPL terms. See [licensing details](docs/licensing.md).

## Translations

The original Transifex integration is no longer an active translation submission channel for FileNeko. To propose translations, edit the locale files in [`frontend/src/i18n`](frontend/src/i18n) and open a pull request.

## Authentication providers

New authentication providers implement the [Auther interface](auth/auth.go), integrate with the CLI configuration parser and the storage authentication backend, and should include focused tests. See `auth/`, `cmd/config.go` and `storage/bolt/auth.go` for the actual current interfaces and call sites.
