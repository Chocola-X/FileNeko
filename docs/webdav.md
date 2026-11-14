# WebDAV

WebDAV 在同一个 HTTP 服务上提供 `/dav/` 文件访问入口，默认关闭。

启动时开启：

```sh
filebrowser --enableWebDAV
```

也可以保存到现有数据库配置中，然后重启服务：

```sh
filebrowser config set --enableWebDAV
```

在 WebDAV 客户端填写 `https://你的域名/dav/`，使用 File Browser 的用户名和密码。
如果设置了 `--baseURL /files`，地址为 `https://你的域名/files/dav/`。
Basic 认证通过 HTTPS 传输；普通账号密码认证（JSON auth）可用。
启用验证码、代理认证或无认证模式时不接受 Basic 登录；支持自定义请求头的客户端可以使用现有登录接口取得的 `X-Auth` token，过期后需重新登录。

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
