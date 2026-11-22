# FileNeko Licensing and Attribution

**English** | [简体中文](licensing_CN.md)

FileNeko is an independently maintained personal derivative project created by **Chocola-X**, hosted at <https://github.com/Chocola-X/FileNeko>.

## Project License

Modifications and derivative works in FileNeko are licensed under the **GNU Affero General Public License version 3 only** (SPDX identifier: `AGPL-3.0-only`). The complete license text is available in the root [LICENSE](../LICENSE). This is not a dual license with Apache 2.0, nor does it grant permission under "AGPL version 3 or any later version".

The copyright notice `Copyright (C) 2026 Chocola-X` applies to original modifications in this project and does not claim ownership over upstream code or contributions from others. Other contributors retain their respective copyrights.

## Upstream Attribution

This project is derived from [File Browser](https://github.com/filebrowser/filebrowser), which is licensed under the Apache License 2.0. The full text of the original license is preserved in [LICENSES/Apache-2.0.txt](../LICENSES/Apache-2.0.txt), while upstream origin, copyrights, and modifications made by this project are recorded in [NOTICE](../NOTICE). All original copyright, patent, trademark, and attribution notices must continue to be preserved; licensing of this derivative work does not retroactively alter the licensing of independent upstream code.

## Distribution and Network Use

When distributing modified versions of this project, you must comply with AGPLv3 requirements regarding licensing, change notices, and source code availability, while preserving upstream and third-party notices. Provided source code must correspond to the running or distributed version, including source files, frontend code, dependency manifests, and build scripts necessary for building.

Section 13 of the AGPLv3 requires that modified versions interacting with users remotely over a network provide those users with an opportunity to receive the corresponding source code free of charge. Author attribution, source code, and license links are displayed exclusively in the authenticated sidebar for file management and settings; they are not hidden by the "Disable external links" option. The login page, public share pages, standalone error pages, and unauthenticated guest mode do not display these notices.

Official binary builds set `version.SourceURL` via Go link flags to reference the specific commit archive of this repository. If you further modify or fork this project, publish your corresponding source code and update this entry at build time to prevent users from obtaining upstream code that diverges from the running version. When building directly from an untracked working copy, the default entry points to the repository; corresponding complete source code must still be provided prior to distribution.

## Third-Party Assets

Go and npm dependencies, fonts, and other assets carrying independent licenses remain governed by their respective licenses; the project's AGPL terms do not override or replace them. Dependencies and exact versions are recorded in `go.mod`, `go.sum`, `frontend/package.json`, and `frontend/pnpm-lock.yaml`.

Font documentation and licenses are located in [frontend/src/assets/fonts](../frontend/src/assets/fonts/README.md). SIL OFL texts for Fredoka, JetBrains Mono, M PLUS Rounded 1c, and Dancing Script remain unchanged and are included in release packages. Assets such as CuteRound and Roboto are not automatically re-licensed by this fork; redistributing them requires verifying their respective licenses.

Binary release distributions include `LICENSE`, `NOTICE`, `LICENSES/`, and font licenses. Although Dockerfiles and container configurations are present in the repository, **no official FileNeko container images are currently published**; paths referenced in legacy container configurations must not be assumed to reflect published images.

References: [Apache-2.0 Section 4](https://www.apache.org/licenses/LICENSE-2.0#redistribution), [AGPLv3 Section 13](https://www.gnu.org/licenses/agpl-3.0.html#section13).
