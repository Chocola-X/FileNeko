# Installing FileNeko

FileNeko is a personal project maintained by Chocola-X. Use packages from
[Chocola-X/FileNeko](https://github.com/Chocola-X/FileNeko), rather than upstream
File Browser installers or images, to get this fork's changes.

## Binary

Download the archive for your system from a completed **Actions** run's
**Artifacts** list, or from the project's [Releases](https://github.com/Chocola-X/FileNeko/releases)
when a release is available. Each platform is listed separately, along with a
checksum file. Extract the archive and keep its `LICENSE`, `NOTICE` and
`LICENSES/` files when redistributing it. The executable name remains
`filebrowser` (`filebrowser.exe` on Windows).

```sh
./filebrowser -r /path/to/your/files
```

To build from source, follow [CONTRIBUTING.md](../CONTRIBUTING.md). Operating system
requirements and CI download details are also documented there. Upstream
Homebrew packages and download scripts install File Browser, not FileNeko.

## Docker

The version-tag release workflow publishes a [minimal image](#bare-alpine-image) and an [S6 Overlay image](#s6-overlay-image) to `ghcr.io/chocola-x/fileneko`. These commands require the corresponding release image to have been published and made accessible.

### Bare Alpine Image

```sh
docker run \
    -v filebrowser_data:/srv \
    -v filebrowser_database:/database \
    -v filebrowser_config:/config \
    -p 8080:80 \
    ghcr.io/chocola-x/fileneko
```

Where `filebrowser_data`, `filebrowser_database` and `filebrowser_config` are Docker [volumes](https://docs.docker.com/engine/storage/volumes/), where the data, database and configuration will be stored, respectively. The default configuration and database will be automatically initialized.

The default user that runs FileNeko inside the container has UID 1000 and GID 1000. If, for one reason or another, you want to run the Docker container with a different user, please consult Docker's [user documentation](https://docs.docker.com/engine/containers/run/#user).

> [!NOTE]
>
> When using [bind mounts](https://docs.docker.com/engine/storage/bind-mounts/), that is, when you mount a path on the host in the container, you must manually ensure that they have the correct **permissions**. Docker does not do this automatically for you. The host directories must be readable and writable by the user running inside the container. You can use the [`chown`](https://linux.die.net/man/1/chown) command to change the owner of those paths.

FileNeko is now up and running. Read the ["First Boot"](#first-boot) section for more information.

### S6 Overlay Image

The `s6` image is based on LinuxServer and leverages the [s6-overlay](https://github.com/just-containers/s6-overlay) system for a standard, highly customizable image. It should be used as follows:

```shell
docker run \
    -v /path/to/srv:/srv \
    -v /path/to/database:/database \
    -v /path/to/config:/config \
    -e PUID=$(id -u) \
    -e PGID=$(id -g) \
    -p 8080:80 \
    ghcr.io/chocola-x/fileneko:s6
```

Where:

- `/path/to/srv` contains the files root directory for FileNeko
- `/path/to/config` contains a `settings.json` file
- `/path/to/database` contains a `filebrowser.db` file

Both `settings.json` and `filebrowser.db` will automatically be initialized if they don't exist.

FileNeko is now up and running. Read the ["First Boot"](#first-boot) section for more information.

## First Boot

Your instance is now up and running. FileNeko will automatically bootstrap a database, in which the configuration and the users are stored. You can find the address in which your instance is running, as well as the randomly generated password for the user `admin`, in the console logs.

> [!WARNING]
>
> The automatically generated password for the user `admin` is only displayed once. If you lose it, reset the password using the `filebrowser users update` CLI after stopping the server; keep your existing database.

Although this is the fastest way to bootstrap an instance, we recommend you to take a look at other possible options, by checking [`config init`](cli/filebrowser-config-init.md) and [`config set`](cli/filebrowser-config-set.md), to make the installation as safe and customized as it can be.

If your goal is to have a public-facing deployment, we recommend taking a look at the [deployment](deployment.md) page for more information on how you can secure your installation.
