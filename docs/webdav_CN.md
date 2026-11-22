# WebDAV

[English](webdav.md) | **简体中文**

WebDAV 在同一个 HTTP 服务上提供 `/dav/` 文件访问入口，默认关闭。

管理员可在“设置 → 全局设置 → WebDAV”开启或关闭，并查看完整连接地址。
保存时复核当前密码；保存成功后立即对新请求生效，无需重启。关闭后 `/dav` 与 `/dav/…` 均返回 404，包括使用已有登录令牌的请求。设置保存在数据库中，重启后保留。

启动时开启：

```sh
fileneko --enableWebDAV
```

也可以保存到现有数据库配置中，然后重启服务：

```sh
fileneko config set --enableWebDAV
```

命令行可用 `--enableWebDAV=false` 关闭。显式启动参数、环境变量及配置文件中的 `enableWebDAV` 在启动时覆盖数据库值；运行期间仍可以从设置页更改。如果始终设置启动参数，下次启动将再次采用该参数。未显式配置时兼容原有数据库中的 WebDAV 开关。

在 WebDAV 客户端填写 `https://你的域名/dav/`，使用 FileNeko 的用户名和密码。
如果设置了 `--baseURL /files`，地址为 `https://你的域名/files/dav/`。
Basic 认证通过 HTTPS 传输；普通账号密码认证（JSON auth）可用。
启用验证码、代理认证或无认证模式时不接受 Basic 登录；支持自定义请求头的客户端可以使用现有登录接口取得的 `X-Auth` token，过期后需重新登录。

## 防止密码暴力破解

“设置 → 全局设置 → Fail2Ban · 登录保护”可管理临时封禁。默认开启：**5 分钟内累计失败 5 次，封禁 5 分钟**。可调整失败次数（1–100）、统计时间（1–1440 分钟）、封禁时长（1–10080 分钟），也可关闭临时封禁。

网页密码登录和 WebDAV Basic 认证使用同一计数器，不能通过切换入口绕过；使用 FileNeko 原有用户名和密码，不需要独立 WebDAV 密码。成功请求不计入失败，也不会抹去统计时间内已经发生的失败。封禁期间即使密码正确也返回 `429`，`Retry-After` 表示剩余秒数；被拒绝的请求不会延长封禁，到期自动恢复。网页登录显示等待时间。

封禁默认按实际连接来源 IP 计算，忽略未经验证的 `X-Forwarded-For`、`X-Real-IP`，IPv6 合并为 /64。Nginx 后部署时配置 `--trustedProxies=127.0.0.1/32` 并由 Nginx 覆盖转发头，即可按真实客户端分别计数；未配置时共用代理地址的预算。配置和多层代理规则见 [Nginx 部署说明](reverse-proxy_CN.md)。已有登录会话不会因为失败密码尝试而被撤销，管理员仍可调整设置。独立的 bcrypt 并发限制与短期限流在关闭 Fail2Ban 后仍保留。

计数和封禁保存在进程内存中，重启后清空；多个实例不共享，多实例应在受信任反向代理上配合集群级来源限流。

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

当前限制：锁保存在当前进程内存中，重启后失效，不适用于多实例共享锁；最多同时保留 4096 个锁（含请求期间的临时锁），已到期的锁按访问回收。
不持久化自定义 WebDAV 属性（PROPPATCH），文件操作暂不触发 Command Runner 的 before/after hooks。
尚未逐一验证 Windows Explorer、macOS Finder 等系统客户端的兼容性。

## 本轮独立核查（2026-11-15）

设置与认证功能提交：`d0b4c073`。在本地代码及所使用的 `golang.org/x/net/webdav` 实现上独立检查请求分发、密码认证、权限、文件类型、作用域、缓存和锁处理；以下问题均先用回归测试复现，再修复并分别提交。

| 确认的问题 | 修复 | 提交与回归证据 |
| --- | --- | --- |
| WebDAV 的 HTML/SVG/JS 下载缺少脚本隔离，可作为应用同源内容使用 | 附件下载、禁止脚本与默认资源的 CSP、sandbox、nosniff；文件内容保持原样 | `2866093c`；`http/webdav_content_test.go` |
| 特殊文件直接打开，管道可阻塞，设备可产生无界数据 | 打开前检查类型，只允许普通文件和目录；列表中不提供特殊文件 | `c73e6a3f`；`http/webdav_special_test.go` |
| 同一文件的符号链接别名可绕过原路径上的锁 | 根据解析后的物理路径统一锁键，兼顾大小写不敏感文件系统 | `324cca29`；`http/webdav_alias_lock_test.go` |
| 底层锁令牌为时间戳递增值；解锁未校验作用域，续期先修改再检查权限 | 外部令牌改用密码学随机值，内部值不可直接授权；解锁和续期先校验作用域、当前路径规则及别名指向，限制映射大小并回收到期条目 | `85e44059`；`http/webdav_lock_scope_test.go` |
| tagged If 可引用另一文件的有效令牌来绕过当前写入目标的锁 | 锁条件绑定实际操作的源路径；COPY 仅锁定其写入目标 | `766cb102`；`http/webdav_if_lock_test.go` |

同时验证管理员设置持久化与即时启停、普通账户不能修改开关、非法封禁参数不落库、网页登录与 WebDAV 共用失败计数、成功认证不触发封禁、IPv6 /64 合并、修改端口/转发头不能绕过、到期自动恢复、并发失败计数、长时间封禁不被提前回收。对应 `http/security_settings_test.go`、`auth/password_limiter_test.go`、`settings/fail2ban_test.go`。

既有 WebDAV 回归还覆盖 baseURL、中文文件名、下载/上传/复制/移动/删除、锁续期、缩略图失效、递归拒绝规则、根目录保护、作用域外符号链接拒绝、撤销权限立即生效，以及验证码/代理/无认证模式不能绕过密码认证限制。上述测试不代表所有系统挂载客户端或生产代理部署均已验证。

验证结果：

- `go test -race ./...`：全部通过。
- `go vet ./...`：通过。
- 前端 `pnpm run lint`、`pnpm run typecheck`、`pnpm run test`（5 个文件、29 项）、`pnpm run build`：全部通过。生产构建仍有原有资源分包、运行时自定义资源及 Browserslist 数据更新提示。
- `GOOS=windows GOARCH=amd64 go test -c ./http -o /tmp/neko-webdav-http-windows.test.exe`：交叉编译通过，未在 Windows 上执行。
- `go build -o /tmp/neko-webdav-filebrowser .`：通过。使用当前构建、临时数据库和文件目录、本机随机端口及 `/files` baseURL 运行真实服务：验证即时启停、Basic 与已有会话均受关闭开关限制、重启保留设置、复用账号、PROPFIND/GET/PUT/MOVE/COPY/DELETE、随机锁令牌、拒绝无令牌覆盖、授权覆盖与解锁、混合登录与 WebDAV 失败触发同一 5 分钟封禁、剩余等待时间、已有管理员会话仍可调整保护设置。全部通过；测试服务和临时数据已清理。
- `git diff --check`：通过。
