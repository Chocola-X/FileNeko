# FileNeko 许可证与来源

FileNeko 是由 **Chocola-X** 维护的个人衍生项目，项目地址为
<https://github.com/Chocola-X/FileNeko>。

## 项目许可

FileNeko 的修改及衍生作品整体采用 **GNU Affero General Public License version 3 only**，SPDX 标识为 `AGPL-3.0-only`。完整许可证位于根目录 [LICENSE](../LICENSE)。这不是 Apache / AGPL 二选一的双重许可，也不额外授予“AGPL 第 3 版或任意后续版本”的许可。

版权声明 `Copyright (C) 2026 Chocola-X` 适用于本项目的原创修改，不表示维护者拥有上游或其他贡献者的全部代码版权。其他贡献者保留各自的版权。

## 保留上游信息

本项目基于 [File Browser](https://github.com/filebrowser/filebrowser)，原项目采用 Apache License 2.0。原许可证全文原样保存在 [LICENSES/Apache-2.0.txt](../LICENSES/Apache-2.0.txt)，上游来源、版权和本项目的修改说明保存在 [NOTICE](../NOTICE)。原有版权、专利、商标和归属声明应继续保留；衍生项目的许可不会追溯改变独立上游代码的原有授权。

## 分发与网络使用

分发本项目的修改版本时，应遵循 AGPLv3 对许可、修改说明和对应源码的要求，并继续保留上游及第三方声明。提供的源码需要对应实际运行或分发的版本，并包括构建所需的源文件、前端、依赖清单和构建脚本。

AGPLv3 第 13 条规定，修改后的程序支持远程网络交互时，应向与其交互的用户提供免费取得该版本对应源码的机会。界面中的作者署名、源码和许可提示仅在登录后的文件管理及设置页面侧栏显示，不随“禁用外部链接”设置隐藏。登录页、公开分享页、独立错误页及无认证访客模式不显示这些提示。

当前二进制构建通过 Go 链接变量 `version.SourceURL` 指向此仓库的具体提交归档。若继续修改或迁移本项目，请公开自己的对应源码并更新该入口，不要让用户获得与运行版本不符的上游代码。直接从未跟踪的工作区构建时，默认入口可能是项目仓库；分发之前仍需提供对应的完整源码。

## 第三方内容

Go / npm 依赖、字体及其他附有独立许可声明的内容继续使用各自的许可证，项目的 AGPL 声明不会覆盖或删除这些许可。依赖及版本由 `go.mod`、`go.sum`、`frontend/package.json` 和 `frontend/pnpm-lock.yaml` 记录。

字体说明和已有许可证位于 [frontend/src/assets/fonts](../frontend/src/assets/fonts/README.md)。其中 Fredoka、JetBrains Mono、M PLUS Rounded 1c 和 Dancing Script 的 SIL OFL 文本保持不变，并随发布包附带。原有 CuteRound、Roboto 等资源不因本次切换自动获得新的授权；再分发这些资源时仍需核查其各自的授权来源。

二进制发行包配置包含 `LICENSE`、`NOTICE`、`LICENSES/` 和已有字体许可。仓库中虽然保留了 Dockerfile 等文件，但**目前并未提供官方 FileNeko 容器镜像**，因此不能把旧容器配置中的路径作为已发布镜像的保证。

参考：[Apache-2.0 第 4 条](https://www.apache.org/licenses/LICENSE-2.0#redistribution)、[AGPLv3 第 13 条](https://www.gnu.org/licenses/agpl-3.0.html#section13)。
