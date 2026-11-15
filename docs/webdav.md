# WebDAV

WebDAV 在同一个 HTTP 服务上提供 `/dav/` 文件访问入口，默认关闭。

管理员可在“设置 → 全局设置 → WebDAV”开启或关闭，并查看完整连接地址。
保存时复核当前密码；保存成功后立即对新请求生效，无需重启。关闭后 `/dav` 与 `/dav/…` 均返回 404，包括使用已有登录令牌的请求。设置保存在数据库中，重启后保留。

启动时开启：

```sh
filebrowser --enableWebDAV
```

也可以保存到现有数据库配置中，然后重启服务：

```sh
filebrowser config set --enableWebDAV
```

命令行可用 `--enableWebDAV=false` 关闭。显式启动参数、环境变量及配置文件中的 `enableWebDAV` 在启动时覆盖数据库值；运行期间仍可以从设置页更改。如果始终设置启动参数，下次启动将再次采用该参数。未显式配置时兼容原有数据库中的 WebDAV 开关。

在 WebDAV 客户端填写 `https://你的域名/dav/`，使用 File Browser 的用户名和密码。
如果设置了 `--baseURL /files`，地址为 `https://你的域名/files/dav/`。
Basic 认证通过 HTTPS 传输；普通账号密码认证（JSON auth）可用。
启用验证码、代理认证或无认证模式时不接受 Basic 登录；支持自定义请求头的客户端可以使用现有登录接口取得的 `X-Auth` token，过期后需重新登录。

## 防止密码暴力破解

“设置 → 全局设置 → Fail2Ban · 登录保护”可管理临时封禁。默认开启：**5 分钟内累计失败 5 次，封禁 5 分钟**。可调整失败次数（1–100）、统计时间（1–1440 分钟）、封禁时长（1–10080 分钟），也可关闭临时封禁。

网页密码登录和 WebDAV Basic 认证使用同一计数器，不能通过切换入口绕过；使用 File Browser 原有用户名和密码，不需要独立 WebDAV 密码。成功请求不计入失败，也不会抹去统计时间内已经发生的失败。封禁期间即使密码正确也返回 `429`，`Retry-After` 表示剩余秒数；被拒绝的请求不会延长封禁，到期自动恢复。网页登录显示等待时间。

封禁按实际连接来源 IP 计算，忽略未经验证的 `X-Forwarded-For`、`X-Real-IP`，IPv6 合并为 /64。反向代理后的客户端共用代理地址的预算。已有登录会话不会因为失败密码尝试而被撤销，管理员仍可调整设置。独立的 bcrypt 并发限制与短期限流在关闭 Fail2Ban 后仍保留。

计数和封禁保存在进程内存中，重启后清空；多个实例不共享。多实例或需要识别代理后客户端地址时，应在受信任反向代理上配合来源限流。

例如列出目录（curl 将提示输入密码）：

```sh
curl --user admin -X PROPFIND -H 'Depth: 1' https://你的域名/dav/
```

支持 OPTIONS、PROPFIND、GET、HEAD、PUT、MKCOL、COPY、MOVE、DELETE、LOCK 和 UNLOCK。
挂载后的根目录是该账号的 scope；复用用户的文件系统隔离、外部符号链接策略、全局规则、用户规则和隐藏文件设置。

| 操作 | 所需权限 |
| --- | --- |
| 列目录、读取元数据 | 通过路径规则 |
| 下载 | Download |
| 上传新文件、建目录 | Create |
| 覆盖已有文件 | Modify |
| 复制 | Download、Create |
| 移动、重命名 | Rename |
| 删除、复制或移动时替换已有目标 | Delete |
| 加锁、解锁 | Create 或 Modify；新建锁定文件还需 Create |

删除和移动目录、复制目录时会预先检查后代路径规则，避免通过操作父目录绕过限制。
禁止删除、移动根目录，以及在互为祖先的目录间执行复制或移动。

当前限制：锁保存在当前进程内存中，重启后失效，不适用于多实例共享锁。
不持久化自定义 WebDAV 属性（PROPPATCH），文件操作暂不触发 Command Runner 的 before/after hooks。
尚未逐一验证 Windows Explorer、macOS Finder 等系统客户端的兼容性。
