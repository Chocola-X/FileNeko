# WebDAV

**English** | [简体中文](webdav_CN.md)

WebDAV provides file access over `/dav/` on the same HTTP service and is disabled by default.

Administrators can enable or disable WebDAV and view the complete connection URL under **Settings → Global Settings → WebDAV**.
Changing this setting requires re-entering the current administrator password; upon saving, it takes effect immediately for new requests without restarting the server. When disabled, requests to `/dav` and `/dav/...` return 404, including requests using existing authentication tokens. The setting is stored in the database and persists across restarts.

Enable at startup:

```sh
fileneko --enableWebDAV
```

Or persist the setting to the database and restart:

```sh
fileneko config set --enableWebDAV
```

You can disable it on the command line with `--enableWebDAV=false`. Command-line flags, environment variables, and config files override database values at startup; the setting can still be toggled dynamically from the web interface during runtime. If startup flags are always provided, subsequent restarts will re-apply those flags. When unconfigured, existing database WebDAV settings are preserved for backward compatibility.

In your WebDAV client, enter `https://your-domain/dav/` along with your FileNeko username and password.
If `--baseURL /files` is configured, the URL is `https://your-domain/files/dav/`.
Basic authentication is transmitted over HTTPS; standard username/password authentication (JSON auth) is supported.
Basic authentication is rejected when captcha, proxy authentication, or no-authentication mode is enabled. Clients supporting custom HTTP headers can provide an `X-Auth` token obtained from the web login API; re-authentication is required once the token expires.

## Preventing password brute force

Temporary login blocking can be configured under **Settings → Global Settings → Fail2Ban / Login Protection**. It is enabled by default: **5 failed attempts within 5 minutes trigger a 5-minute block**. You can adjust the allowed failure count (1–100), detection window (1–1440 minutes), ban duration (1–10080 minutes), or disable temporary blocking altogether.

Web password logins and WebDAV Basic authentication share the same failure counter, preventing attackers from bypassing limits by switching endpoints. They use FileNeko's standard user credentials; separate WebDAV passwords are not required. Successful requests are not counted as failures, nor do they clear failures already logged within the detection window. During a block, requests return `429 Too Many Requests` even if the password is correct, with `Retry-After` indicating the remaining seconds. Blocked attempts do not extend the ban duration; access is restored automatically once the ban expires. Web login displays the remaining waiting time.

Blocks are tracked by default using the direct connecting client IP address, ignoring untrusted `X-Forwarded-For` and `X-Real-IP` headers. IPv6 addresses are aggregated by `/64` prefix. When deploying behind Nginx, configure `--trustedProxies=127.0.0.1/32` and have Nginx overwrite client forwarding headers to track client IPs individually; without this setting, all proxied clients share the proxy's rate-limit budget. Configuration details and multi-tier proxy rules are described in [Nginx reverse proxy deployment](reverse-proxy.md). Existing sessions are not revoked due to failed password attempts, and administrators can still adjust settings. Process-level bcrypt concurrency limits and short-term rate limiting remain active even if Fail2Ban is disabled.

Attempt counters and active blocks reside in process memory and are cleared on restart; they are not synchronized across multiple instances. For multi-instance deployments, complementary cluster-level rate limiting should be configured on trusted reverse proxies.

For example, to list a directory using curl (which will prompt for your password):

```sh
curl --user admin -X PROPFIND -H 'Depth: 1' https://your-domain/dav/
```

Supported HTTP methods include `OPTIONS`, `PROPFIND`, `GET`, `HEAD`, `PUT`, `MKCOL`, `COPY`, `MOVE`, `DELETE`, `LOCK`, and `UNLOCK`.

Clients can send `X-OC-Mtime` with a Unix timestamp in seconds when uploading a file with `PUT` or creating a directory with `MKCOL`. The server returns `X-OC-Mtime: accepted` when it saves the requested timestamp.
Uploading children updates the parent directory's modification time. After uploading the children, clients can restore the directory timestamp by setting `DAV:lastmodified` to the original Unix timestamp with `PROPPATCH`. This requires Modify permission and respects WebDAV locks.
The mounted root corresponds to the user account's scope, reusing the user's filesystem isolation, external symlink policy, global rules, user rules, and hidden dotfile settings.

| Operation | Required Permission |
| --- | --- |
| Directory listing, metadata read | Passes path rules |
| Download | Download |
| Upload new file, create directory | Create |
| Overwrite existing file | Modify |
| Copy | Download, Create |
| Move, rename | Rename |
| Delete, or overwrite existing target during copy/move | Delete |
| Lock, unlock | Create or Modify; creating a locked file also requires Create |

Deleting, moving, or copying directories pre-checks descendant path rules to prevent bypassing restrictions by operating on parent folders.
Deleting or moving the root directory is prohibited, as is copying or moving between directories that are ancestors of one another.

Current limitations: Locks are stored in process memory and reset upon restart; they cannot be shared across multiple server instances. A maximum of 4,096 concurrent locks are maintained (including transient locks held during requests); expired locks are reclaimed upon subsequent access.
`PROPPATCH` can update the filesystem modification time. Custom properties are not persisted. File operations do not currently trigger Command Runner before/after hooks.
Compatibility with operating system native clients such as Windows Explorer and macOS Finder has not been exhaustively validated.

## Independent audit records (2026-11-15)

Settings and authentication functionality commit: `d0b4c073`. Request dispatching, password authentication, permissions, file types, scopes, caching, and lock handling were independently audited against the local codebase and `golang.org/x/net/webdav`. Each of the following issues was first reproduced with regression tests before being fixed and committed.

| Confirmed Issue | Fix | Commit and Regression Evidence |
| --- | --- | --- |
| WebDAV HTML/SVG/JS downloads lacked script isolation and could execute in the application's origin | Enforce attachment downloads; CSP restricting scripts and default sources; sandbox; nosniff; preserving raw file content | `2866093c`; `http/webdav_content_test.go` |
| Special files opened directly; named pipes could block indefinitely, device files could stream unbounded data | Validate file type before opening; restrict to regular files and directories; omit special files from listings | `c73e6a3f`; `http/webdav_special_test.go` |
| Symlink aliases pointing to the same file could bypass locks on the target path | Canonicalize lock keys based on resolved physical paths, accounting for case-insensitive filesystems | `324cca29`; `http/webdav_alias_lock_test.go` |
| Underlying lock tokens were incrementing timestamps; unlock did not verify scope, renewal modified locks prior to permission checks | Replace external tokens with cryptographic random values; prevent internal values from direct authorization; verify scope, path rules, and alias targets before unlock/renewal; bound map size and reclaim expired entries | `85e44059`; `http/webdav_lock_scope_test.go` |
| Tagged `If` headers could reference a valid token from another file to bypass locks on the current write target | Bind lock conditions to the actual source path of the operation; `COPY` locks only its write target | `766cb102`; `http/webdav_if_lock_test.go` |

Additional tests verify administrator setting persistence and dynamic toggling, rejection of non-admin changes, rejection of invalid ban parameters, unified failure counters across web login and WebDAV, immunity of successful logins against bans, IPv6 `/64` aggregation, resilience against port/header spoofing, automatic unbanning upon expiry, concurrent failure counting, and preservation of long bans against early eviction. See `http/security_settings_test.go`, `auth/password_limiter_test.go`, and `settings/fail2ban_test.go`.

Existing WebDAV regression tests also cover `baseURL`, non-ASCII (Chinese) filenames, download/upload/copy/move/delete, lock renewal, thumbnail invalidation, recursive deny rules, root protection, out-of-scope symlink denial, immediate effect of permission revocations, and inability of captcha/proxy/no-auth modes to bypass password limits. These tests do not imply exhaustive validation of every OS mount client or production proxy deployment.

Verification results:

- `go test -race ./...`: all passed.
- `go vet ./...`: passed.
- Frontend `pnpm run lint`, `pnpm run typecheck`, `pnpm run test` (5 files, 29 tests), `pnpm run build`: all passed. Production build continues to report expected chunking and Browserslist notices.
- `GOOS=windows GOARCH=amd64 go test -c ./http -o /tmp/neko-webdav-http-windows.test.exe`: cross-compilation passed (not executed on Windows).
- `go build -o /tmp/neko-webdav-filebrowser .`: passed. End-to-end verification with temporary database and file tree, random local port, and `/files` baseURL confirmed immediate toggling, enforcement of disable switch on Basic and existing sessions, persistence across restarts, account reuse, PROPFIND/GET/PUT/MOVE/COPY/DELETE, random lock tokens, rejection of unauthenticated overwrites, authorized overwrites and unlocks, shared 5-minute ban triggered by mixed web and WebDAV failures, remaining wait time, and admin access to settings during bans. All temporary test services and data were cleaned up.
- `git diff --check`: passed.
