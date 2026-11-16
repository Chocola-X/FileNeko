# Building FileNeko

FileNeko is a fork of the archived File Browser project. This document describes building and developing FileNeko from source.

## Project Structure

The backend side of the application is written in [Go](https://golang.org/), while the frontend (located on a subdirectory of the same name) is written in [Vue.js](https://vuejs.org/). Due to the tight coupling required by some features, basic knowledge of both Go and Vue.js is recommended.

* Learn Go: [https://github.com/golang/go/wiki/Learn](https://github.com/golang/go/wiki/Learn)
* Learn Vue.js: [https://vuejs.org/guide/introduction.html](https://vuejs.org/guide/introduction.html)

We encourage you to use git to manage your fork. To clone the main repository, just run:

```bash
git clone https://github.com/Chocola-X/Neko-Filebrowser
```

We use [Taskfile](https://taskfile.dev/) to manage the different processes (building, releasing, etc) automatically.

## Build

You can fully build the project in order to produce a binary by running:

```bash
task build
```

The backend task disables CGO and uses baseline CPU targets (`GOAMD64=v1`,
`GOARM64=v8.0`). On Linux this produces a static executable without a glibc or
musl dependency. Running a plain `go build` follows your environment instead.

## CI builds and downloads

Every branch/tag push and pull request runs the CI checks. You can also start
**Continuous Integration** manually from the repository's **Actions** page.
The **Build and Package** job uses the same GoReleaser configuration as releases,
in snapshot mode, and uploads the original `linux-amd64-filebrowser.tar.gz` /
`windows-amd64-filebrowser.zip` style packages and checksums as a
`filebrowser-ci-<commit>` artifact. Download it from the completed workflow run's
**Artifacts** section within 30 days. Snapshot builds do not publish releases,
Docker images or Homebrew changes. The other CI checks must also pass before a
build is considered validated.

The embedded frontend is already included in each executable. Extract the
appropriate archive; Linux packages preserve executable permissions inside the
tarball. Node.js and Go are needed to build the application, not to run it.

### Older operating systems

- **CentOS 7 x86-64:** the static Linux amd64 build avoids its old glibc 2.17.
  Its standard 3.10 kernel meets Go's Linux runtime minimum (3.2 for these
  targets). This is a compatibility target, not a claim of testing on a CentOS 7
  VM; a container still shares its host's kernel.
- **Windows 7/8:** unsupported by the current build. Go 1.21 and later require
  Windows 10 / Windows Server 2016 or newer; disabling CGO does not remove the
  runtime's Windows API requirements. Go 1.20 was the last release supporting
  Windows 7, but this project and dependencies such as bbolt and `golang.org/x`
  require Go 1.25, and the source uses newer standard library APIs. A Windows 7
  port needs a separate source/dependency/runtime compatibility effort, rather
  than simply changing the CI Go version.
- **macOS:** the Go 1.26 builds require macOS 12 or newer.

References: [Go minimum requirements](https://go.dev/wiki/MinimumRequirements),
[Go 1.21 Windows requirements](https://go.dev/doc/go1.21#windows), and
[Go 1.26 ports](https://go.dev/doc/go1.26#ports).

## Development

For development, there are a few things to have in mind.

### Frontend

We use [Node.js](https://nodejs.org/en/) on the frontend to manage the build process. Prepare the frontend environment:

```bash
# From the root of the repo, go to frontend/
cd frontend

# Install the dependencies
pnpm install
```

If you just want to develop the backend, you can create a static build of the frontend:

```bash
pnpm run build
```

If you want to develop the frontend, start a development server which watches for changes:

```bash
pnpm run dev
```

Please note that you need to access FileNeko's interface through the development server of the frontend.

### Backend

First prepare the backend environment by downloading all required dependencies:

```bash
go mod download
```

You can now build or run FileNeko as any other Go project:

```bash
# Build
go build

# Run
go run .
```

## Documentation

Documentation lives in [`docs`](docs) as plain Markdown and is no longer built into a site. The command line reference in [`docs/cli`](docs/cli) is generated from the commands themselves:

```bash
task docs:cli:generate
```

## Release

To make a release, just run:

```bash
task release
```

## Translations

The Transifex integration stopped on 2026-09-01 and translations submitted there no longer reach this repository. Locale files live in [`frontend/src/i18n`](frontend/src/i18n) and can be edited directly.

## Authentication Provider

To build a new authentication provider, you need to implement the [Auther interface](https://github.com/filebrowser/filebrowser/blob/master/auth/auth.go), whose method will be called on the login page after the user has submitted their login data.

```go
// Auther is the authentication interface.
type Auther interface {
    // Auth is called to authenticate a request.
    Auth(r *http.Request, s *users.Storage, root string) (*users.User, error)
}
```

After implementing the interface you should:

1. Add it to [`auth` directory](https://github.com/filebrowser/filebrowser/blob/master/auth).
2. Add it to the [configuration parser](https://github.com/filebrowser/filebrowser/blob/master/cmd/config.go) for the CLI.
3. Add it to the [`authBackend.Get`](https://github.com/filebrowser/filebrowser/blob/master/storage/bolt/auth.go).

If you need to add more flags, please update the function `addConfigFlags`.

