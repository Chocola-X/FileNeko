# FileNeko 文档

[English](README.md) | **简体中文**

FileNeko 是一个自托管的网页端文件管理器，是已归档的 [File Browser](https://github.com/filebrowser/filebrowser) 项目的独立维护个人分支。有关项目概述和可用下载，请参阅[仓库主说明](../README_CN.md)。

## 用户指南

- [安装与首次运行](installation_CN.md) — 二进制程序、CI 构件与初始登录
- [部署指南](deployment_CN.md) — 访问限制、反向代理与登录保护
- [外观定制](customization_CN.md) 与 [CSS 主题接口](theme-api_CN.md)
- [身份认证](authentication_CN.md)
- [WebDAV](webdav_CN.md)
- [命令执行](command-execution_CN.md) — 默认禁用；安全注意事项
- [文件夹信息统计](folder-statistics_CN.md) — 流式递归统计大小与文件数量
- [反向代理配置](reverse-proxy_CN.md) — Nginx 真实客户端 IP 与受信代理配置
- [故障排除](troubleshooting_CN.md)
- [命令行参考](cli/fileneko_CN.md)

## 项目信息

- [构建与贡献指南](../CONTRIBUTING_CN.md)
- [安全政策](../SECURITY_CN.md) 与 [安全核查记录](security-audit-2026-11-15_CN.md)
- [许可证与来源说明](licensing_CN.md)
- [更新日志](../CHANGELOG.md)

**发布状态：** FileNeko 目前未发布官方 Docker 镜像。仓库中虽保留了继承自上游的 Docker 相关配置，但推荐的安装方式是使用 FileNeko 可执行归档包或从源码编译。上游 File Browser 的镜像不包含此分支的修改。
