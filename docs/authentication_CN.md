# 身份认证

[English](authentication.md) | **简体中文**

FileNeko 支持三种身份认证方式，每种方式都有其特定的功能和规范。关于如何添加新的身份认证方式，请参阅[构建 FileNeko](../CONTRIBUTING_CN.md#身份认证提供者扩展)。

## JSON 认证（默认）

我们称之为 JSON 认证，它是系统的默认认证方式，在未作修改时即为启用状态。如之前更改过认证方式，可通过以下命令恢复为 JSON 认证：

```sh
fileneko config set --auth.method=json
```

该方式还支持在登录过程中启用 **reCAPTCHA** 人机验证：

```sh
fileneko config set --auth.method=json \
  --recaptcha.key site-key \
  --recaptcha.secret private-key
```

默认情况下，系统使用 [Google reCAPTCHA](https://developers.google.com/recaptcha/docs/display) 服务。在中国大陆地区或希望使用其他服务商时，可通过以下命令更改服务器地址：

```sh
fileneko config set --recaptcha.host https://recaptcha.net
```

其中 `https://recaptcha.net` 可以替换为您需要的任意兼容服务地址。

## 代理请求头认证（Proxy Header）

如果您希望通过反向代理完成用户登录验证，可以使用 `proxy` 认证方式。在此模式下，反向代理必须在 HTTP 请求头中传递已登录用户的用户名：

```sh
fileneko config set --auth.method=proxy --auth.header=X-My-Header
```

其中 `X-My-Header` 是反向代理传递用户名的 HTTP 请求头名称。

> [!WARNING]
> 
> 本分支仅接受来自 `auth.trustedProxies` 所配置的网络对端（默认为回环地址）传递的身份头。请确保未受信任的客户端无法直接访问后端端口，并由代理统一完成身份验证及覆盖客户端提交的身份头。此项配置与用于客户端真实 IP/HTTPS 元数据的 `trustedProxies` 配置相互独立；详见[反向代理部署说明](reverse-proxy_CN.md)。

## 钩子认证（Hook Authentication）

FileNeko 中的钩子认证允许将用户认证委托给外部脚本或可执行程序。FileNeko 不在内部验证凭证，而是将用户名和密码传递给管理员定义的自定义命令。该命令通过环境变量接收凭据，并返回键值对以告知系统应当放行认证、阻止登录还是交回系统内部处理。

钩子脚本的输出可以控制用户权限、作用域（Scope）、界面语言等属性，是一种高度灵活且可扩展的认证机制。

> [!WARNING]
>
> 用户提交的用户名和密码由攻击者完全可控，并通过 `USERNAME` 和 `PASSWORD` 环境变量传递给钩子命令。FileNeko 直接执行目标程序而非通过 Shell 调用，因此变量值本身不会直接触发代码注入。但是，您的脚本必须将这些值视为不受信任的数据：务必在脚本中将其加上引号（`"$USERNAME"`、`"$PASSWORD"`），切勿将未加引号的值传递给 Shell、`eval`、`bash -c`、命令替换或反引号。如果脚本通过 Shell 计算这些不受信任的值，任何登录请求都可能演变成远程代码执行。

例如，以下命令在 Windows 上将认证委托给 PowerShell 脚本（您也可以配置任意命令，如 Python、Node.js 脚本等）：

```sh
fileneko config set --auth.method=hook --auth.command="powershell.exe -File C:\route\to\your\script\auth.ps1"
```

以下是示例 `auth.ps1` 脚本的内容：

```powershell
param()

# 从环境变量中获取 FileNeko 传递的凭据
$username = $env:USERNAME
$password = $env:PASSWORD

# 用户名字典（仅供测试演示）
$users = @{
    "admin" = "kideW48v7-SdE*"
    "test"  = "2sDd3-etrytñK"
}

# 检查字典中是否存在该用户并验证密码
if ($users.ContainsKey($username) -and $users[$username] -eq $password) {

    # 认证成功
    Write-Output "hook.action=auth"        # 钩子动作（"auth" 表示成功通过认证）

    Write-Output "user.perm.admin=true"    # 设置管理员权限（赋予所有权限）
    # 也可以单独定义各项权限：
    Write-Output "user.perm.execute=true"
    Write-Output "user.perm.create=true"
    Write-Output "user.perm.rename=true"
    Write-Output "user.perm.modify=true"
    Write-Output "user.perm.delete=true"
    Write-Output "user.perm.share=true"
    Write-Output "user.perm.download=true"

    Write-Output "user.locale=zh-cn"       # 设置语言
    Write-Output "user.viewMode=list"      # 设置视图模式
    Write-Output "user.scope=/"            # 设置文件作用域
    Write-Output "user.singleClick=true"   # 设置单机打开
    Write-Output "user.hideDotfiles=false" # 设置隐藏点文件

} else {
    # 阻止认证
    Write-Output "hook.action=block"
}
```

### 钩子输出格式

钩子认证脚本必须输出一系列键值对，每行一组，采用以下格式：

```
key=value
```

FileNeko 读取这些行并执行相应的认证操作与用户属性配置。

#### 必填字段

钩子必须输出以下动作之一：

| 键 | 描述 |
|---|---|
| `hook.action=auth` | 允许用户认证。FileNeko 会根据需要自动创建或更新该用户。 |
| `hook.action=block` | 拒绝认证。登录请求失败。 |
| `hook.action=pass` | 将认证委托给 FileNeko 内部的密码验证逻辑。 |

在绝大多数自定义认证流程中，通常使用 `auth` 或 `block`。

认证成功时的输出示例：

```sh
hook.action=auth
```

#### 可选用户字段

当返回 `hook.action=auth` 时，钩子还可以定义额外的用户属性。这些字段会覆盖 FileNeko 的默认设置，实现对认证用户的完全自定义：

1. 权限（Permissions）
```
user.perm.admin=true
user.perm.execute=true
user.perm.create=true
user.perm.rename=true
user.perm.modify=true
user.perm.delete=true
user.perm.share=true
user.perm.download=true
```
> 设置 `user.perm.admin=true` 会自动赋予该用户所有权限。

2. 界面与行为配置
```
user.locale=zh-cn
user.viewMode=list
user.singleClick=true
user.hideDotfiles=false
```

3. 用户文件作用域（Scope）
```
user.scope=/
```

## 无认证模式（No Authentication）

针对希望在私有局域网等受信任环境中便捷使用 FileNeko 的场景，我们提供了免认证模式。启用此模式后，系统将默认以 **ID 为 1** 的用户登录，创建更多用户将不再生效。

```sh
fileneko config set --auth.method=noauth
```

### 受信任的代理对端（本分支安全特性）

代理传递的身份头仅在直接连接来自受信任的代理对端时才被接受。未配置信任列表的向后兼容配置现在默认仅信任回环地址（Loopback）。对于部署在其他机器或容器中的反向代理，请配置其精确地址：

```sh
fileneko config set --auth.method=proxy --auth.header=X-Remote-User --auth.trustedProxies=10.0.0.4/32
```

该配置支持 IPv6 地址与 CIDR，也可通过代理认证配置中的 `trustedProxies` 数组进行导入。转发头不能用于建立信任关系。代理服务器必须替换而非保留客户端提交的身份头，并对每次转发的请求进行严格认证。应将后端端口严格限制为仅向该代理开放，切勿信任整个客户端网段或 `0.0.0.0/0`。

### 密码保护分享链接（本分支安全特性）

旧版的 `?token=...` 密码绕过方式已不再被分享 API 接受或返回。在网页中输入分享密码后，当前浏览器会获得一个最长一小时有效（且绝不超过分享链接自身有效期）的 `HttpOnly`、`SameSite=Strict` 专属 Cookie。下载、媒体和归档请求均使用该 Cookie；复制下载链接不会将密码授权状态泄露给其他浏览器。现有的受密码保护链接通过再次输入密码依然正常可用。无密码的公开分享保留原有的公开链接行为。

### 自助注册隔离（本分支安全特性）

无论 `createUserDir` 和默认用户 scope 如何配置，公开注册始终会在 `userHomeBasePath`（默认为 `/users`）下新建独立的专属私有目录。用户无法通过将注册用户名设置为已有目录或符号链接的名称来接管该目录。旧账户不会被自动迁移；管理员应审查先前版本创建的根作用域账户。代理认证和钩子认证的自动置备继续遵循其显式配置的作用域策略。

### 密码防爆破限制（本分支安全特性）

JSON 密码登录与 WebDAV Basic 认证共享同一个失败尝试配额：每个连接来源拥有 10 次突发校验配额，每 6 秒补充 1 次。成功的认证会返还额度。系统最多同时执行 8 个并发密码哈希校验，且跟踪的来源缓存是有界的。此外，可配置的 Fail2Ban 默认在 5 分钟内累计 5 次失败后触发 5 分钟封禁。IPv6 地址按 `/64` 前缀合并共享配额。被拒绝的请求返回带有 `Retry-After` 头的 HTTP 429。仅当直接对端明确在服务端层面的 `trustedProxies` 列表中时，转发头才会影响计数来源；WebDAV 和错误日志使用完全相同的来源解析器。在没有配置该列表的情况下，反代背后的所有客户端将共用该配额。详见 [Nginx 部署说明](reverse-proxy_CN.md) 和 [WebDAV 保护设置](webdav_CN.md)。限流保护为各服务端进程本地状态。

### 会话迁移与持久化（本分支安全特性）

请将所有集群实例同时升级。原有的无状态签名令牌将被拒绝，用户需要重新登录。JWT 继续作为前端用户元数据的兼容外壳保留；授权还额外需要在数据库中持久化验证随机且可撤销的会话标识符。登出会撤销整个会话家族（包含并发续期生成的新会话）。续期操作以原子方式消耗旧会话标识符；密码变更会立即让旧令牌全部失效；过期记录在后续访问和新登录时自动回收清理。每个用户最多保留 64 个活跃登录会话。认证 Cookie 设置为 `HttpOnly` 和 `SameSite=Strict`；在后端使用 TLS 或受信反向代理传递单一 `X-Forwarded-Proto: https` 时会自动添加 `Secure` 标记。非受信的转发头将被忽略。当 Nginx 终止 TLS 时，所需信任和请求头配置请参阅[反向代理部署说明](reverse-proxy_CN.md)。
