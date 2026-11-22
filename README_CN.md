# FileNeko

<p align="center">
  <img src="./branding/banner.png" width="550" alt="FileNeko" />
</p>

<p align="center">轻量级、自托管、带有二次元动漫风格界面的文件管理器。</p>

<p align="center">
  <a href="README.md">English</a> | <strong>简体中文</strong>
</p>

FileNeko 是由 [Chocola-X](https://github.com/Chocola-X) 维护的 [File Browser](https://github.com/filebrowser/filebrowser) 个人衍生分支。它将 File Browser 的 Web 文件管理功能与明亮、魔法图书馆风格的界面、可自定义的 CSS 主题、WebDAV 支持以及额外的安全补强结合在一起。

<img width="1920" height="1080" alt="demo1" src="https://github.com/user-attachments/assets/d1ad0f63-b51f-4d95-a334-b5a7ed721c3c" />

<img width="1920" height="1080" alt="demo2" src="https://github.com/user-attachments/assets/97f91a0b-7d4a-4e50-991f-c74a4fd6fd8c" />

<img width="1920" height="1080" alt="demo3" src="https://github.com/user-attachments/assets/1b2754de-bb32-4032-8c84-920530b297a5" />

<img width="1920" height="1080" alt="demo4" src="https://github.com/user-attachments/assets/4fde99ba-5dfe-4896-96ad-caf65a4aced3" />

<img width="1920" height="1080" alt="demo5" src="https://github.com/user-attachments/assets/957d72b8-b2c7-4d6d-80f5-a99fb2c583d3" />

<img width="1920" height="1080" alt="demo6" src="https://github.com/user-attachments/assets/d27a391a-9845-47ca-8b54-985d2ddef465" />

<img width="1920" height="1080" alt="demo7" src="https://github.com/user-attachments/assets/35741353-cc36-46cd-a25e-be77c73541e6" />

<img width="1920" height="1080" alt="demo8" src="https://github.com/user-attachments/assets/8ec6b2e4-1134-4a8e-9605-a48b444e4c4e" />


上游 File Browser 仓库已于 2026 年 9 月 1 日归档。FileNeko 独立进行维护；上游的二进制文件和容器镜像**不包含** FileNeko 的改动。

## 功能特性

- 在浏览器中管理文件：上传、下载、预览、编辑、整理与分享。
- 二次元风格界面，支持自定义品牌与[公开的 CSS 主题接口](docs/theme-api_CN.md)。
- 多用户支持，具备灵活的文件访问权限与独立用户主目录。
- 可选的 WebDAV 访问（默认关闭）；详见 [WebDAV 文档](docs/webdav_CN.md)。
- 额外的安全修复与访问控制强化；详见[安全核查记录](docs/security-audit-2026-11-15_CN.md)。
- 为项目 CI 构建的目标平台提供自包含的独立可执行文件；前端资源已内置于二进制程序中。

## 下载与运行

FileNeko 目前**尚未发布官方容器镜像**。**请勿使用已归档上游项目的 Docker 镜像或安装脚本**来安装 FileNeko。

- **开发构建（CI 产物）：** 打开 [GitHub Actions](https://github.com/Chocola-X/FileNeko/actions)，选择最近一次成功的 **Continuous Integration** 工作流运行，在 **Artifacts** 区域下载对应平台的归档包。构建产物保留期有限，GitHub 可能需要登录才能下载。
- **正式发布版：** 请查看 [FileNeko Releases](https://github.com/Chocola-X/FileNeko/releases)（若已发布）。
- **源码构建：** 详见 [CONTRIBUTING_CN.md](CONTRIBUTING_CN.md)。

解压 Linux 归档包后的运行示例：

```sh
./fileneko -r /path/to/your/files
```

Windows 下的可执行文件为 `fileneko.exe`。初次启动时，请在终端输出中查看生成的管理员密码。有关配置选项、升级注意事项及其他平台说明，请阅读[安装指南](docs/installation_CN.md)。

## 项目文档

请从[文档首页](docs/README_CN.md)开始阅读，包括[安装指南](docs/installation_CN.md)、[部署与安全](docs/deployment_CN.md)、[外观定制](docs/customization_CN.md)以及 [WebDAV 说明](docs/webdav_CN.md)。

原有的 File Browser `FB_` 环境变量、`filebrowser.db` 数据库以及部分 `.filebrowser.*` 配置文件名称继续保留以保证兼容性。可执行程序本身命名为 `fileneko` / `fileneko.exe`。

## 安全性

FileNeko 是一个个人项目，按“尽力而为（best-effort）”原则维护。在没有合适保护措施的情况下，切勿将其直接暴露在公网中。建议使用终止 TLS 的反向代理、限制网络与账户访问权限、以最低必要文件系统权限运行服务进程，并在充分理解风险前保持命令执行功能处于关闭状态。

[本地安全核查记录](docs/security-audit-2026-11-15_CN.md)记录了已审查的问题和修复；但这**并不构成**软件绝对无漏洞的保证。如需报告安全问题，请遵循 [SECURITY_CN.md](SECURITY_CN.md)。

## 上游与许可协议

本项目是 [File Browser](https://github.com/filebrowser/filebrowser) 的非官方衍生项目，并非其官方继任者。FileNeko 的修改部分及衍生作品整体采用 [GNU AGPLv3-only](LICENSE) 许可协议发布；原 File Browser 及第三方的版权和许可声明继续保留。详见 [NOTICE](NOTICE)、[LICENSES](LICENSES) 及[许可证与来源说明](docs/licensing_CN.md)。
