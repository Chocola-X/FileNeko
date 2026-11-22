# Nginx Reverse Proxy and Real Client IP

**English** | [简体中文](reverse-proxy_CN.md)

FileNeko supports terminating TLS and managing domains via Nginx while proxying HTTP traffic to the backend.
When trusted proxies are configured, web logins, WebDAV Basic auth, Fail2Ban, short-term password rate limiting, and error logs consistently use the real client IP. Furthermore, cookies for login, renewal, logout, and password-protected shares set the `Secure` flag based on trusted external HTTPS status.

## FileNeko Configuration

By default, FileNeko trusts no proxies and ignores forwarded client IP and scheme headers. When local Nginx connects to the backend over `127.0.0.1`:

```sh
fileneko --address 127.0.0.1 --port 8080 --trustedProxies=127.0.0.1/32
```

You can also persist this to the database before launching (stop the running service before modifying):

```sh
fileneko -d /path/to/filebrowser.db config set --trustedProxies=127.0.0.1/32
```

CLI flags, environment variables, and configuration files override database settings at startup. To clear the setting, pass `--trustedProxies=` or execute `config set --trustedProxies=`.

The environment variable accepts a comma-separated list of IP/CIDR values:

```sh
FB_TRUSTED_PROXIES='127.0.0.1/32,::1/128' fileneko
```

In the runtime JSON configuration file:

```json
{
  "trustedProxies": ["127.0.0.1/32", "::1/128"]
}
```

In `config export` / `config import` dumps, this option is located under `server.trustedProxies`. Invalid IP or CIDR values are rejected at startup, configuration import, or when saving.

For Nginx running in a container or on a separate host, configure the source IP address it actually uses when connecting to the backend, preferring single IPs or minimal CIDR ranges. Never configure entire client subnets or `0.0.0.0/0`, `::/0` as trusted proxies. Unix domain sockets lack a verifiable peer TCP IP and forwarded headers over them are not accepted; use a loopback TCP socket if this functionality is required.

`trustedProxies` and `auth.trustedProxies` are completely separate settings: the former governs client IP/HTTPS metadata forwarding, while the latter is used by proxy authentication to receive pre-authenticated usernames. When Nginx only handles TLS and domain routing, standard JSON username/password authentication remains active and proxy authentication should remain disabled.

## Nginx Configuration Example

The following example proxies both the web interface and WebDAV to a local backend; replace certificate paths and domain names with your actual values. Both `map` blocks reside in the `http` context: the first preserves the original `Host` (including port) with fallback to `$host`, while the second supports WebSocket connections used by Command Runner. Note that WebDAV must first be enabled in FileNeko Global Settings.

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

    # Adjust according to your upload needs; 0 disables request body size limits in Nginx.
    client_max_body_size 1g;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $fileneko_host;

        # Forward WebDAV Basic authentication and other original headers.
        proxy_pass_request_headers on;
        proxy_set_header Authorization $http_authorization;

        # Single-tier proxy overwrites client headers, ignoring untrusted client values.
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;

        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $fileneko_connection;

        # Stream uploads directly without buffering the entire body in Nginx first.
        proxy_request_buffering off;
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;

        # Pass backend authentication challenges and rate-limit responses through.
        proxy_intercept_errors off;
    }
}
```

In a single-tier reverse proxy setup, overwriting `X-Forwarded-For` with `$remote_addr` is recommended. Nginx's `$proxy_add_x_forwarded_for` appends the client address to existing incoming headers and should only be used when managing a multi-tier proxy chain. See the [Nginx proxy_set_header documentation](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_set_header) and [variable documentation](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#var_proxy_add_x_forwarded_for).

If using `--baseURL /files`, Nginx must preserve the `/files` prefix, and the client WebDAV URL becomes `https://files.example.com/files/dav/`. The whole-site proxy configuration shown above preserves original request paths. Adjust timeouts and body size limits to match expected upload sizes.

### WebDAV Authentication and Headers

WebDAV clients transmit `Authorization: Basic ...` headers, which FileNeko validates against its internal user credentials. Nginx's `proxy_pass_request_headers` is enabled by default and forwards this header unless overridden; explicitly listing `Authorization` in the configuration clarifies that client credentials pass directly to the backend without needing FileNeko passwords stored in Nginx. See the [Nginx proxy_pass_request_headers documentation](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_pass_request_headers).

Do not configure Nginx `auth_basic` / `auth_basic_user_file` for WebDAV. Two separate Basic auth challenges cannot share a single `Authorization` header. If an external Basic auth gate already exists, provide a separate entrypoint for FileNeko WebDAV. For restrictions when captcha or other auth modes are enabled, see [WebDAV documentation](webdav.md).

Headers such as `Depth`, `Destination`, `Overwrite`, `If`, and `Lock-Token` are forwarded as-is by the configuration above without needing individual `proxy_set_header` directives. For `COPY` and `MOVE`, absolute `Destination` URLs must match the external hostname, port, and baseURL; preserving the original `Host` avoids validation failures when non-standard ports (such as `https://files.example.com:8443/`) are used. Furthermore, after TLS termination in Nginx, there is no need to rewrite `https://` in `Destination` to `http://`.

If adding a separate `/dav/` location block, note that `proxy_set_header` directives in a child block cause all inherited `proxy_set_header` directives from parent blocks to be discarded unless none are specified in that block. See the [Nginx header inheritance rules](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_set_header).

To test authentication and directory listing (curl will prompt for password):

```sh
curl --user admin -i -X PROPFIND -H 'Depth: 0' https://files.example.com/dav/
```

Expected response is `207 Multi-Status`; missing or incorrect credentials return `401 Unauthorized` with `WWW-Authenticate: Basic ...`, and blocked sources return `429 Too Many Requests` with `Retry-After`. When using a baseURL or custom port, substitute the full WebDAV URL accordingly.

## Trust Boundaries and Resolution Rules

- FileNeko checks the peer IP of the direct network connection. Forwarded headers are parsed only if the direct peer matches `trustedProxies`. `RemoteAddr` itself is never modified, preserving the security boundary for proxy authentication.
- `X-Forwarded-For` is prioritized. The proxy chain is evaluated from right to left across trusted IPs, stopping at the first untrusted IP; any untrusted prefix claimed by the client is ignored. Multiple header occurrences, IPv4, IPv6, and IPv4-mapped IPv6 are supported up to 32 hops.
- When `X-Forwarded-For` is absent, a single-value `X-Real-IP` header is accepted. If the header contains invalid formats, empty addresses, port/zone suffixes, duplicate single-value headers, or configuration errors, resolution falls back safely to the direct connecting peer without guessing from other headers. Private IP addresses are not skipped: LAN clients receive distinct rate-limiting counters.
- Single-value `X-Forwarded-Proto: https` is accepted only from trusted peers; comma-separated protocol chains are rejected, and physical TLS connections are never downgraded. In multi-tier proxy environments, the proxy directly connecting to FileNeko must clean and accurately propagate the protocol.
- Identity, hostname, or trust relationships are never inferred from `Forwarded`, `X-Forwarded-Host`, or other headers. Backend ports should only be accessible to trusted proxies; local TCP setups should bind to loopback addresses.
- Fail2Ban remains process-local, groups IPv6 by `/64`, and clears on restart; this setup resolves client source attribution but does not synchronize bans across multiple instances.

## Verification Records (2026-11-16)

- Verified with this configuration, local Nginx 1.30.1, temporary self-signed certificates, and a standalone FileNeko database; `nginx -t` passed. Both explicit `Authorization` forwarding and default header passthrough correctly returned Basic authentication challenges and permitted authenticated PROPFIND queries; both root paths and `/files` baseURL passed.
- Uploads exceeding 2 MiB, downloads, and COPY/MOVE/DELETE operations with custom external HTTPS ports all passed. Reverting `Host` to `$host` reproduced COPY returning HTTP 400; preserving the original `Host` resolved the issue. Login cookies carried `Secure`, spoofed IP/protocol headers were overwritten, banned sources received 429 with `Retry-After`, and other real sources retained access. All temporary services and data were cleaned up.

## Verification Records (2026-11-15)

- `requestinfo/proxy_test.go`: default/untrusted sources ignore forwarded headers; single/multi-tier/multi-header proxy chains; spoofed prefixes; malformed tails; private IPs; IPv6; IPv4-mapped IPv6; up to 32 hops; ambiguous protocol values and preservation of real TLS; retention of original connection metadata.
- `cmd/trusted_proxy_test.go`: CLI configuration and clearing, environment variable comma lists, persistence, runtime overrides, and rejection of invalid values; `http/trusted_proxy_test.go`: shared real-client bans across web login and WebDAV, isolation between different clients, immunity against direct-connection header spoofing, consistent error logging, and HTTPS cookie attributes for login/renewal/logout/shares.
- `go test -race ./...` and `go vet ./...`: all passed. Windows HTTP tests cross-compiled (not executed on Windows).
- Frontend linting, 29 tests, type checks, and production builds passed; chunking and Browserslist notices preserved.
- End-to-end TLS reverse proxy verification conducted using current build, local Nginx 1.30.1, temporary self-signed certificate, isolated config/database/files, and random port. Mixed login/WebDAV errors from client `127.0.0.2` triggered a shared ban, while `127.0.0.3` remained accessible; injected client headers were overwritten by Nginx and could not bypass bans; direct untrusted connections could not spoof IP or HTTPS; logs recorded the real client IP of banned requests; login and password share cookies carried `Secure`; restarts without environment overrides preserved proxy config; invalid CLI values were rejected. All passed; temporary services and data cleaned up.
- `git diff --check`: passed.
