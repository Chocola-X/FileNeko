# Nginx 反向代理与真实客户端 IP

FileNeko 支持由 Nginx 终止 TLS、管理域名，再通过 HTTP 转发到后端。
配置受信代理后，网页登录、WebDAV Basic、Fail2Ban、短期密码限流和错误日志统一使用真实客户端 IP；登录、续期、退出及密码分享 Cookie 会根据可信的外部 HTTPS 信息设置 `Secure`。

## FileNeko 配置

默认不信任任何代理，不接受客户端 IP 或协议转发头。本机 Nginx 通过 `127.0.0.1` 连接后端时：

```sh
fileneko --address 127.0.0.1 --port 8080 --trustedProxies=127.0.0.1/32
```

也可在启动前保存到数据库（请使用实际数据库路径，停止占用它的服务后修改）：

```sh
fileneko -d /path/to/filebrowser.db config set --trustedProxies=127.0.0.1/32
```

启动参数、环境变量和配置文件可覆盖数据库配置。清空配置可用 `config set --trustedProxies=`，或本次启动传入 `--trustedProxies=`。

环境变量使用逗号分隔的 IP/CIDR：

```sh
FB_TRUSTED_PROXIES='127.0.0.1/32,::1/128' fileneko
```

运行时 JSON 配置文件的对应项：

```json
{
  "trustedProxies": ["127.0.0.1/32", "::1/128"]
}
```

完整 `config export` / `config import` 文件中，这一项位于 `server.trustedProxies`。非法 IP/CIDR 会在保存、导入或启动时被拒绝。

容器或独立主机上的 Nginx 应配置它连接后端时实际使用的源地址，优先选择单个 IP 或最小必要 CIDR。不要把所有客户端网段或 `0.0.0.0/0`、`::/0` 当作代理网段。Unix socket 没有可验证的 TCP 对端 IP，本机制不接受其转发头；需要此功能时请使用受限的本机 TCP 入口。

`trustedProxies` 与 `auth.trustedProxies` 各自独立：前者只允许传递客户端 IP/HTTPS 元数据，后者用于代理传递已认证用户名。Nginx 只负责 TLS 和域名时，保持默认 JSON 用户名密码认证即可，无需启用代理身份认证。

## Nginx 配置示例

下例同时代理网页和 WebDAV 到同机后端；证书路径和域名替换为实际值。两个 `map` 均位于 `http` 上下文：前者保留原始 Host（包含端口）并提供空值回落，后者兼容 Command Runner 的 WebSocket 连接。WebDAV 需先在 FileNeko 的全局设置中开启。

```nginx
map $http_host $fileneko_host {
    default $http_host;
    ''      $host;
}

map $http_upgrade $fileneko_connection {
    default upgrade;
    ''      close;
}

server {
    listen 443 ssl;
    server_name files.example.com;

    ssl_certificate     /path/to/fullchain.pem;
    ssl_certificate_key /path/to/privkey.pem;

    # 按实际上传需求调整；0 表示不在 Nginx 限制请求体大小。
    client_max_body_size 1g;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $fileneko_host;

        # 保留 WebDAV 的 Basic 认证和其他原始请求头。
        proxy_pass_request_headers on;
        proxy_set_header Authorization $http_authorization;

        # 单层代理覆盖客户端传入的值，不保留其自报地址或协议。
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $fileneko_connection;

        # 上传直接转发，避免先在 Nginx 完整缓存请求体。
        proxy_request_buffering off;
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;

        # 保留后端的认证挑战和限流响应。
        proxy_intercept_errors off;
    }
}
```

单层代理直接覆盖 `X-Forwarded-For` 即可；Nginx 的 `$proxy_add_x_forwarded_for` 会保留已有头并追加来源地址，仅在明确管理多层代理链时使用。参见 [Nginx 官方请求头说明](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_set_header) 与 [变量说明](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#var_proxy_add_x_forwarded_for)。

若使用 `--baseURL /files`，Nginx 必须保留 `/files` 路径，客户端 WebDAV 地址为 `https://files.example.com/files/dav/`。上面的整站转发方式保留原始路径。请求体大小、超时等按实际上传需求配置。

### WebDAV 认证与请求头

WebDAV 客户端发送的是 `Authorization: Basic …`，由 FileNeko 使用自己的用户名和密码校验。Nginx 的 `proxy_pass_request_headers` 默认开启，原示例在没有其他配置干预时就会转发该请求头；上面显式写出 `Authorization` 是为了明确该 location 使用客户端提交的认证头，无需在 Nginx 中填写、生成或保存 FileNeko 密码。参见 [Nginx 原始请求头转发说明](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_pass_request_headers)。

不需要为 WebDAV 额外配置 Nginx 的 `auth_basic` / `auth_basic_user_file`。两套独立的 Basic 登录无法在同一个 `Authorization` 头里同时提供各自的凭据；已有外层 Basic 登录时，应为 WebDAV 提供由 FileNeko 自行认证的入口。启用验证码或使用其他认证模式时，Basic 登录限制见 [WebDAV 说明](webdav.md)。

`Depth`、`Destination`、`Overwrite`、`If`、`Lock-Token` 等 WebDAV 请求头由上述配置原样转发，不需要逐个添加 `proxy_set_header`。COPY/MOVE 的绝对 `Destination` 必须包含正确的外部主机名、端口和 baseURL；保留原始 Host 可避免使用 `https://files.example.com:8443/` 等非标准端口时校验失败。TLS 在 Nginx 解密后，也无需把 `Destination` 中的 `https://` 改成后端的 `http://`。

若后续单独添加 `/dav/` 的 `location`，注意 `proxy_set_header` 仅在当前层完全没有该指令时才继承上一层的整组规则，不能只添加认证头后假定其他头规则仍会继承。参见 [Nginx 请求头继承规则](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_set_header)。

验证认证与目录访问（curl 会提示输入 FileNeko 密码）：

```sh
curl --user admin -i -X PROPFIND -H 'Depth: 0' https://files.example.com/dav/
```

正常返回 `207 Multi-Status`；未提供或提供错误凭据时返回 `401` 和 `WWW-Authenticate: Basic …`，触发封禁时返回 `429` 和 `Retry-After`。使用 baseURL 或非标准端口时，替换为对应的完整 WebDAV 地址。

## 信任边界与解析规则

- 先检查底层连接的对端 IP，仅对端命中 `trustedProxies` 才解析转发头；不修改 `RemoteAddr`，避免破坏代理身份认证的信任边界。
- 优先读取 `X-Forwarded-For`，从右向左检查可信链，遇到第一个非受信 IP 停止；其左边的客户端自报前缀不参与判断。支持多个同名头、IPv4、IPv6、IPv4-mapped IPv6，最多追踪 32 跳。
- 没有 `X-Forwarded-For` 时，允许单值 `X-Real-IP`。权威链尾无效、空地址、地址带端口/zone、单值 IP 头重复或配置错误时回退到底层对端，不猜测其他头中的地址。不会跳过私网地址：局域网用户也需要独立计数。
- 仅对受信对端接受单值 `X-Forwarded-Proto: https`；不接受逗号分隔的协议链，且不会将实际 TLS 连接降级。多层代理中，最终连接后端的代理必须清理并准确传递外部协议。
- 不从 `Forwarded`、`X-Forwarded-Host` 或其他头推断身份、主机名或信任关系。后端入口只向受信代理开放；同机 TCP 示例监听回环地址。
- Fail2Ban 仍为单进程状态、IPv6 /64 共享、重启后清空；该配置解决客户端来源识别，不提供多实例之间的封禁同步。

## 示例复核（2026-11-16）

- 使用本文配置、本机 Nginx 1.30.1、临时自签名证书与独立 FileNeko 数据库验证，`nginx -t` 通过。显式转发 Authorization 与使用默认请求头转发时，均正确返回 Basic 认证挑战，并允许认证后的 PROPFIND；普通入口和 `/files` baseURL 均通过。
- 超过 2 MiB 的 WebDAV 上传、下载及带外部 HTTPS 端口的 COPY/MOVE/DELETE 均通过。将 Host 改回 `$host` 时复现 COPY 返回 400；保留原始 Host 后正常。登录 Cookie 带 Secure，伪造的来源/协议头被覆盖，被封禁来源返回 429 与 Retry-After，另一真实来源仍可访问。所有临时服务与数据已清理。

## 验证记录（2026-11-15）

- `requestinfo/proxy_test.go`：默认/非受信来源忽略转发头、单层/多层/多头代理链、伪造前缀、异常尾部、私网 IP、IPv6、IPv4-mapped IPv6、最多 32 跳、模糊协议值与实际 TLS 不降级，以及保持原始连接元数据。
- `cmd/trusted_proxy_test.go`：CLI 设置/清空、环境变量逗号列表、持久化、运行时覆盖和无效配置拒绝；`http/trusted_proxy_test.go`：登录与 WebDAV 共用真实来源封禁、不同客户端隔离、直接连接伪造头无效、错误日志一致、登录/续期/退出/分享 Cookie 的 HTTPS 属性。
- `go test -race ./...`、`go vet ./...`：全部通过。Windows HTTP 测试交叉编译通过，未在 Windows 执行。
- 前端 lint、29 项测试、应用类型检查及生产构建通过；构建保留既有分包、运行时资源和 Browserslist 更新提示。
- 使用当前构建、本机 Nginx 1.30.1、临时自签名证书、独立配置/数据库/文件目录及随机端口进行真实 TLS 反代验证。客户端 `127.0.0.2` 的网页登录与 WebDAV 混合错误触发同一封禁，`127.0.0.3` 仍可访问；客户端注入的 IP/协议头由 Nginx 覆盖，不能绕过封禁；不受信直接连接不能伪装 IP 或 HTTPS；日志记录被封禁客户端的真实 IP；登录和密码分享 Cookie 为 Secure；无环境变量覆盖的重启保留代理配置，无效 CLI 配置被拒绝。全部通过，临时服务和数据随后清理。
- `git diff --check`：通过。
