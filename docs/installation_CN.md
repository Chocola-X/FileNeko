# 安装 FileNeko

[English](installation.md) | **简体中文**

FileNeko 独立于已归档的 File Browser 项目进行维护。请从 [Chocola-X/FileNeko](https://github.com/Chocola-X/FileNeko) 下载，**切勿**使用 File Browser 上游的安装脚本、Homebrew 配方或容器镜像。

## 可执行程序

请从以下位置选择对应平台的归档包：

- [FileNeko Releases](https://github.com/Chocola-X/FileNeko/releases)（**当发布了正式版本时**），或者
- [GitHub Actions](https://github.com/Chocola-X/FileNeko/actions) 中成功运行的 **Continuous Integration** 工作流的 **Artifacts** 区域。这些属于开发阶段构件，保留时间有限，且可能需要登录 GitHub 才能下载。

CI 工作流会将编译后的后端与内置前端打包在一起。解压对应平台的归档包后直接运行 `fileneko`（Windows 下为 `fileneko.exe`）。Linux 归档包保留了可执行权限。

在 Linux 或 macOS 上的运行示例：

```sh
./fileneko -r /path/to/your/files
```

请指定 FileNeko 进程有权访问的目录。应用程序将继承运行它的操作系统用户的读写权限：除非有明确充分的理由，否则请勿以 root 身份运行。

如需自行构建可执行文件，请参考 [CONTRIBUTING_CN.md](../CONTRIBUTING_CN.md)。Go 和 Node.js 是编译时依赖，独立可执行文件在运行时并不依赖它们。平台要求和构建矩阵在贡献指南中有详细说明。

## Docker 现状

**官方 FileNeko Docker 镜像目前未进行构建和发布。** 上游 File Browser 的 Docker 镜像运行的是上游旧版项目，并非 FileNeko。虽然本仓库中保留了 Dockerfile 和 Compose 示例，但它们不应被视为受支持的 FileNeko 部署方式或可用的镜像下载指引。

## 首次启动

首次运行时，FileNeko 会初始化数据库，并在终端输出中打印服务监听地址和为 `admin` 账户自动生成的初始密码。请务必保存好该密码：它只会在终端显示一次。

如果丢失了密码，请先停止正在运行的服务，然后使用 `fileneko users update` 命令针对现有数据库重置密码；**切勿直接删除数据库**。参数选项详见[命令行参考](cli/fileneko_CN.md)，包括 [`config init`](cli/fileneko-config-init_CN.md) 和 [`config set`](cli/fileneko-config-set_CN.md)。

对于暴露在网络中的部署，在对外开放访问之前请务必阅读[部署指南](deployment_CN.md)和[安全政策](../SECURITY_CN.md)。

## 与 File Browser 的兼容性

可执行文件已重命名为 `fileneko`，但在需要的地方依然保留了旧版 `FB_` 环境变量、`filebrowser.db` 和 `.filebrowser.*` 配置文件命名约定。在升级现有的 File Browser 部署前，请先备份数据库和配置文件；具有兼容性并不等同于可以省略升级测试。
