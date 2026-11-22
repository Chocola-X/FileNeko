# 构建与参与 FileNeko 项目

[English](CONTRIBUTING.md) | **简体中文**

FileNeko 是已归档的 File Browser 项目的独立维护个人分支。我们欢迎在“尽力而为”基础上的社区贡献。FileNeko 的修改部分在 AGPL-3.0-only 协议下分发，同时保留原作者与第三方的版权及许可声明；详见[许可证说明](docs/licensing_CN.md)。

## 项目与前置要求

项目后端使用 Go 编写；前端为 `frontend/` 目录下的 Vue.js 项目。构建工具包括匹配的 Go 工具链、Node.js、pnpm 以及 [Task](https://taskfile.dev/)。构建前请检查 `go.mod`、`frontend/package.json` 以及当前 CI 配置中所要求的版本。

```sh
git clone https://github.com/Chocola-X/FileNeko.git
cd FileNeko
task build
```

`task build` 会依次构建前端并随后构建后端。后端编译时禁用了 CGO，并以通用基础架构 AMD64 (`GOAMD64=v1`) 与 ARM64 (`GOARM64=v8.0`) 为目标。以此方式生成的 Linux 二进制文件不依赖系统的 glibc 或 musl 动态库（但仍需满足内核版本要求）。

生成的可执行文件命名为 `fileneko`（Windows 下为 `fileneko.exe`）。现有的 `FB_` 环境变量、`filebrowser.db` 数据库、`.filebrowser.*` 配置文件约定以及 Go 模块导入路径均保留上游标识以保证兼容性，请勿仅为品牌重命名而更改这些标识。

前端构建会将压缩后的 JavaScript 资源打包到 Go 内置的 HTTP 静态处理程序中。如需生成可分发的可执行文件，请使用 `task build`，而不是假设单独执行 `go build` 会自动重新生成前端静态资源。

## CI 构件与发布

GitHub Actions 的 **Continuous Integration** 工作流会先编译共享的前端产物，然后通过多平台 GoReleaser 快照矩阵（目前覆盖 21 个目标平台）打包各平台的二进制文件。生成的 `linux-amd64-fileneko.tar.gz` / `windows-amd64-fileneko.zip` 风格归档包会发布在成功运行的工作流 **Artifacts** 区域，并附带各平台的校验和。这些构件保留期有限（目前记录为 30 天）；GitHub 可能需要登录才能下载。

CI 快照构件属于**开发构建**，并非正式发布版。其二进制文件已包含前端资源，并已通过打包与静态编译检查。具体的验证状态以对应提交在 GitHub Actions 中显示的结果为准。

仓库中同样包含针对 GitHub Releases 发布标记版本归档包的配置。在宣传发布版本可用之前，请确认对应的 GitHub Actions 发布工作流已启用且已成功执行。

**Docker 现状：** FileNeko 目前既不编译也不发布官方 Docker 镜像。仓库中保留的 Dockerfile 和 Compose 配置并不构成受支持的安装说明。文档补丁中的 GoReleaser 配置已禁用了 Docker 构建发布。上游 File Browser 的镜像并不包含 FileNeko。

本项目不在上游 Homebrew、Docker Hub 或其他上游分发渠道发布软件包。

如需运行矩阵辅助工具自身，请执行：

```sh
go test ./.github/ci
go run ./.github/ci matrix
```

### 旧版操作系统兼容性

- **CentOS 7 x86-64：** 静态编译的 Linux AMD64 构建避免了对旧 glibc 的依赖；但这并不代表在 CentOS 7 上进行了完整测试。
- **Windows 7/8：** 当前 Go 工具链和项目依赖已不再支持；如需支持需进行专门移植。
- **macOS：** 请检查当前 Go 工具链的最低操作系统要求；现有 Go 1.26 构建以 macOS 12 或更高版本为目标。

参考 [Go 最低系统要求](https://go.dev/wiki/MinimumRequirements) 与 [Go 平台移植说明](https://go.dev/doc/go1.26#ports)。

## 本地开发

### 前端开发

```sh
cd frontend
pnpm install
pnpm run build  # 构建生产环境前端静态资源
pnpm run dev    # 启动前端开发服务器
```

进行前端交互式开发时，请根据终端输出提示访问前端开发服务器。

### 后端开发

```sh
go mod download
go build -o fileneko .
go run .
```

如需生成包含最新内置前端资源的最终可执行文件，请在仓库根目录下优先使用 `task build`。

## 命令行文档

项目文档以 Markdown 格式保存在 [`docs`](docs) 目录下。[`docs/cli`](docs/cli) 下的命令行参考文档可从应用程序命令定义自动重新生成：

```sh
task docs:cli:generate
```

## 创建新版本

`task release` 任务会重新生成 CLI 文档并调用 pnpm 运行 `commit-and-tag-version`。该任务可以在本地创建版本提交和 Git 标签；**但这本身并不代表已发布到远端 GitHub Release**。请单独检查 Git 标签、推送状态、对应的 CI 工作流以及发布的资源。请勿发布或宣传容器镜像发布。

发布归档包中包含本项目与上游的版权声明。编译到二进制文件中的源码链接指向对应的 FileNeko 提交；请确保根据 AGPL 条款提供匹配的完整源码。详见[许可证说明](docs/licensing_CN.md)。

## 界面翻译

原有的 Transifex 集成已不再作为 FileNeko 活跃的翻译提交渠道。如需贡献或改进翻译，请直接编辑 [`frontend/src/i18n`](frontend/src/i18n) 中的语言文件并提交 Pull Request。

## 身份认证提供者扩展

如需新增身份认证提供者，请实现 [Auther 接口](auth/auth.go)，并接入 CLI 配置解析器与存储认证后端，同时补充针对性的单元测试。当前实际接口与调用方式请参考 `auth/`、`cmd/config.go` 和 `storage/bolt/auth.go`。
