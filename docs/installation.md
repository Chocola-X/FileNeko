# Installing FileNeko

FileNeko is maintained independently of the archived File Browser project. Download it from [Chocola-X/FileNeko](https://github.com/Chocola-X/FileNeko), **not** from File Browser's upstream install scripts, Homebrew formulae or container images.

## Executables

Choose a platform archive from:

- [FileNeko Releases](https://github.com/Chocola-X/FileNeko/releases), **when a release has been published**, or
- A successful **Continuous Integration** run under [GitHub Actions](https://github.com/Chocola-X/FileNeko/actions), in its **Artifacts** section. These are development artifacts, are retained for a limited time, and may require a GitHub sign-in.

The CI workflow packages the compiled backend with its embedded frontend. Extract the platform archive and run `fileneko` (`fileneko.exe` on Windows). Linux archives preserve the executable permission.

Example on Linux or macOS:

```sh
./fileneko -r /path/to/your/files
```

Use a directory that the FileNeko process is permitted to access. The application has read/write access according to the operating-system user running it: do not run it as root unless there is a well-understood reason.

To build your own executable, follow [CONTRIBUTING.md](../CONTRIBUTING.md). Go and Node.js are build-time requirements, not runtime dependencies of the standalone executable. Platform requirements and the build matrix are described there.

## Docker status

**Official FileNeko Docker images are not currently built or published.** The upstream File Browser Docker images run the upstream project, not FileNeko. Although Dockerfiles and a Compose example exist in this repository, they should not be treated as a supported FileNeko deployment or a working image download instruction.

## First boot

On the first run, FileNeko initializes its database and prints the service address and a generated password for the `admin` account in the terminal output. Save that password: it is only shown once.

If you lose it, stop the running server and use the `fileneko users update` command against the existing database to reset the password; **do not delete the database**. See the [command-line reference](cli/fileneko.md) for options, including [`config init`](cli/fileneko-config-init.md) and [`config set`](cli/fileneko-config-set.md).

For network-exposed installations, read [Deployment](deployment.md) and [Security](../SECURITY.md) before opening access.

## Compatibility with File Browser

The executable is renamed to `fileneko`, while older `FB_` environment variables, `filebrowser.db` and `.filebrowser.*` configuration conventions are retained where needed. Back up the database and configuration before upgrading an existing File Browser deployment; compatibility does not replace an upgrade test.
