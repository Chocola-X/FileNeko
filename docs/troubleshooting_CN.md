# 故障排除

[English](troubleshooting.md) | **简体中文**

## 会话超时（Session Timeout）

默认情况下，用户会话在 **2 小时** 后过期。如果您在较慢的网络连接上上传超大文件，可能需要增加此超时时间，以防止在上传过程中会话失效。您可以使用 `tokenExpirationTime` 设置来配置会话超时时间。

您可以在运行时通过启动参数 `--tokenExpirationTime`、环境变量 `FB_TOKEN_EXPIRATION_TIME` 或在配置文件中进行配置。如果您希望将此设置持久化到数据库中，请使用 [`fileneko config set`](cli/fileneko-config-set_CN.md)。

有效的时间格式包括 `"2h"`、`"30m"`、`"24h"`，或者组合格式如 `"2h30m"`。
