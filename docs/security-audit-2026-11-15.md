# File Browser 上游安全核查记录

核查日期：2026-11-15。先提交界面改造 `91cdbbfd`，再以本地归档分支为基线核查和修复。上游最终版本为 2.63.23，公告不能仅按发布时间判断本地是否仍受影响。

核查来源：[上游全部公开安全公告](https://github.com/filebrowser/filebrowser/security/advisories)、GitHub 公开 issue 正文及公开 issue 搜索结果。使用 GitHub REST API 取得全部 62 条公告，并检查安全相关 issue；没有访问私有报告或向上游发送消息。

结果：19 条公告对应本轮修复、权限边界收紧或补强；43 条已有对应措施，未重复套用旧补丁。此外修复 #5213 的敏感操作复核缺口，以及最终回归发现的会话轮换撤销、文件包装层锁与分享过期清理问题。下表的“已有措施”指对这份公告所述问题的核查结论，不是软件不存在其他漏洞的保证。

## 逐条公告

| 公告 | 上游问题 | 本地结论、提交与证据 |
| --- | --- | --- |
| [GHSA-39cx-23x9-5c8p](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-39cx-23x9-5c8p) | Command WebSocket buffers unbounded messages before checking permissions | 本轮处理 `40a96ced`：升级 WebSocket 前验证权限，限制消息大小和等待时间；`http/commands_test.go`。 |
| [GHSA-448h-jr2h-3vhp](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-448h-jr2h-3vhp) | Subtitle conversion reads whole files into memory, allowing memory exhaustion | 本轮处理 `ba8e6c4a`：字幕转换输入限制为 10 MiB，检查实际读取量；`http/subtitle_test.go`。 |
| [GHSA-c4fr-5f24-4wrj](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-c4fr-5f24-4wrj) | Upload failure-cleanup recursively deletes directories, bypassing Perm.Delete and deny rules | 本轮处理 `2e7889b2`：拒绝向已有目录覆盖上传，移除失败后的递归删除；`http/upload_cleanup_test.go`。 |
| [GHSA-7w29-q235-57m9](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-7w29-q235-57m9) | Symlink aliases inside a user scope bypass path deny rules | 本轮处理 `75362bda`：对别名和解析后的目标同时应用路径规则，包括新建目标与公开分享；`http/rules_symlink_test.go`。 |
| [GHSA-8q5j-8wcr-8v2v](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-8q5j-8wcr-8v2v) | Archive downloads hang forever on named pipes | 本轮处理 `6fc5455a`：归档排除特殊文件，单文件下载拒绝管道、设备、套接字；`http/raw_special_test.go`。 |
| [GHSA-r6pg-pg54-rcr5](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-r6pg-pg54-rcr5) | Deleting another user's shared file leaves their public share link behind | 本轮处理 `014c2c9f`：按各分享者作用域解析物理路径，删除同一对象的分享，保留其他作用域的同名分享；含 WebDAV；`http/share_cleanup_test.go`。 |
| [GHSA-m8v4-4w34-rrvf](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-m8v4-4w34-rrvf) | Renaming a shared file leaves the public share link behind | 本轮处理 `94da0540`：重命名前撤销源路径及其后代分享；`http/share_rename_test.go`。 |
| [GHSA-4r8p-gqj2-mwgm](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-4r8p-gqj2-mwgm) | Concurrent TUS uploads write past the declared Upload-Length | 本轮处理 `a269cc10 / ab79ef10`：TUS 请求互斥及系统文件锁，锁内重新检查大小与偏移；穿透 afero 包装获取描述符；`http/tus_concurrent_test.go`、`http/upload_lock_unix_test.go`。 |
| [GHSA-v3jv-rmh2-635j](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-v3jv-rmh2-635j) | Expired JWTs accepted under proxy auth with a custom logout page | 原代码已有措施：原有代理身份复核已存在；本轮会话改造进一步统一拒绝过期 JWT；`http/auth_test.go`。 |
| [GHSA-77x8-73f4-5485](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-77x8-73f4-5485) | Recursive copy, rename and delete ignore rules denying descendants | 原代码已有措施：递归操作调用 checkDescendants；`http/rules_recursive_test.go`。 |
| [GHSA-7whw-q6gh-xr59](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-7whw-q6gh-xr59) | File checksum disclosure via /api/resources endpoint bypassing Perm.Download check | 原代码已有措施：校验和入口检查 Download；`http/resource_checksum_test.go`。 |
| [GHSA-m9f5-2232-frp6](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-m9f5-2232-frp6) | Out-of-scope file deletion via symlink-following delete in TUS upload-cache eviction | 原代码已有措施：上传缓存驱逐使用用户作用域内的删除回调；`http/upload_cache_memory_test.go`。 |
| [GHSA-fgm5-pw99-w2p7](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-fgm5-pw99-w2p7) | Access rule bypass via case-variant and Windows-separator paths | 原代码已有措施：规范化分隔符、路径和大小写匹配；`http/rules_path_test.go`、`rules/rules_test.go`。 |
| [GHSA-576v-w77m-gr84](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-576v-w77m-gr84) | Case-folded signup usernames share one home directory on case-insensitive filesystems | 原代码已有措施：派生目录保存前用不区分大小写的作用域碰撞检查；`users/storage.go`、`storage/bolt/users_test.go`。 |
| [GHSA-j7jh-37pf-mf8h](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-j7jh-37pf-mf8h) | Proxy and hook auto-provisioning ignore createUserDir and grant the server root scope | 原代码已有措施：代理和认证钩子调用 CreateUserHome，遵守 createUserDir；`auth/proxy_test.go`、`settings/dir_test.go`。 |
| [GHSA-ffv3-7h97-993q](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-ffv3-7h97-993q) | TUS uploads ignore the declared Upload-Length, allowing disk exhaustion by authenticated users | 原代码已有措施：TUS 按声明长度限制每块写入，超长回滚；本轮另修并发缺口；`http/tus_upload_length_test.go`。 |
| [GHSA-833g-cqhp-h72j](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-833g-cqhp-h72j) | Share API exposes the password hash and bypass token | 原代码已有措施：shareResponse 不输出密码哈希和永久 token；`http/share_test.go`。 |
| [GHSA-83xp-526h-j3ww](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-83xp-526h-j3ww) | Archive builder turns backslash filenames into path traversal (zip-slip) | 原代码已有措施：归档名中反斜杠替换为惰性字符并检查安全路径；`http/raw_test.go`。 |
| [GHSA-7rc3-g7h6-22m7](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-7rc3-g7h6-22m7) | Colliding username normalization gives two users the same home directory | 原代码已有措施：SaveProvisioned 将碰撞检查和保存串行化；本轮注册进一步拒绝已有目录；`http/auth_test.go`、`users/storage_test.go`。 |
| [GHSA-6759-996p-gpj6](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-6759-996p-gpj6) | Self-signup users inherit the server root as their scope | 本轮处理 `183d45e6`：上游只作启动警告，本地仍存在；公开注册改为始终新建独立家目录，拒绝接管已有目录或符号链接；`http/signup_scope_test.go`。 |
| [GHSA-8wc8-hf36-mjh9](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-8wc8-hf36-mjh9) | ScopedFs follows a dangling symlink on write, letting a scoped user create files outside their scope | 原代码已有措施：ScopedFs 检查悬空链接实际写入目标；`files/file_test.go`、`http/resource_test.go`。 |
| [GHSA-pp88-jhwj-5qh5](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-pp88-jhwj-5qh5) | Trailing-slash delete leaves a stale public share behind | 原代码已有措施：规范化删除路径并按分隔符处理分享前缀；`storage/bolt/share_test.go`。 |
| [GHSA-fmm7-x4gx-8jhr](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-fmm7-x4gx-8jhr) | Out-of-scope file deletion by a Create-only scoped user via symlink-following RemoveAll in upload failure-cleanup | 原代码已有措施：ScopedFs.RemoveAll 经过作用域检查；本轮另删除直接上传的递归失败清理；`files/file_test.go`。 |
| [GHSA-gxjx-7m74-hcq8](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-gxjx-7m74-hcq8) | Path traversal in download-as-zip/tar via Windows-style backslash separators in stored filenames | 原代码已有措施：查询路径及归档项名称规范化、阻止生成越界条目；`http/raw_test.go`。 |
| [GHSA-3q2p-72cj-682c](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-3q2p-72cj-682c) | Improper Access Control Occurs via Pre-Created Public Share for a Non-existent Path | 原代码已有措施：sharePostHandler 在创建分享前 Stat 目标，拒绝不存在路径；`http/share.go`。 |
| [GHSA-239w-m3h6-ch8v](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-239w-m3h6-ch8v) | Symlinked directories let scoped users and public-share recipients read and write files outside their scope | 原代码已有措施：用户和公开分享使用受符号链接约束的 ScopedFs；`http/public_symlink_test.go`、`files/file_test.go`。 |
| [GHSA-v7vv-5wj2-gfcj](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-v7vv-5wj2-gfcj) | Password reset does not invalidate existing JWT sessions | 本轮处理 `d1bc2116`：令牌绑定当前密码的 HMAC 指纹，修改密码立即失效，客户端重新登录；`http/auth_password_test.go`。 |
| [GHSA-8c9q-7855-wfxq](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-8c9q-7855-wfxq) | Command Execution Allowlist Bypass via Shell Metacharacter Injection | 本轮处理 `6bb0ef54`：交互输入直接作为可执行程序参数，不再交给配置的 shell；`http/commands_test.go`。 |
| [GHSA-5ww9-jg6q-38r7](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-5ww9-jg6q-38r7) | Cross-user unauthorized share-link deletion via unbounded prefix match in DeleteWithPathPrefix | 原代码已有措施：分享前缀检查绑定用户及路径分隔符；本轮物理路径清理保留其他作用域同名对象；`storage/bolt/share_test.go`、`http/share_cleanup_test.go`。 |
| [GHSA-j9jx-hp4c-ghhh](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-j9jx-hp4c-ghhh) | Incorrect access control in public directory shares via rule path rebasing | 原代码已有措施：公开分享重设根目录时保留 checkerPrefix，规则仍对应原始作用域；`http/public_test.go`。 |
| [GHSA-jvpw-637p-h3pw](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-jvpw-637p-h3pw) | Command Injection via Hook Runner | 本轮处理 `3e5a57db`：取消将文件名和用户名插入钩子脚本源码，改用环境变量；`runner/runner_test.go`。 |
| [GHSA-m93h-4hw7-5qcm](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-m93h-4hw7-5qcm) | Command Injection via Authentication Hook Shell Substitution (Pre-Authentication RCE) | 原代码已有措施：认证钩子用环境变量传递凭据，没有 shell 源码替换；`auth/hook.go`。 |
| [GHSA-7526-j432-6ppp](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-7526-j432-6ppp) | Proxy auth auto-provisioned users inherit Execute permission and Commands | 原代码已有措施：代理创建用户清除 Admin/Execute/Commands；`auth/proxy_test.go`。 |
| [GHSA-67cg-cpj7-qgc9](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-67cg-cpj7-qgc9) | Text file content disclosure via /api/resources endpoint bypassing Perm.Download check | 原代码已有措施：resources 的 Content 读取受 Download 控制，编码分支同样检查；`http/resource.go`。 |
| [GHSA-5q48-q4fm-g3m6](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-5q48-q4fm-g3m6) | Access rule bypass via HasPrefix without trailing separator in path matching | 原代码已有措施：路径前缀匹配要求分隔符边界；`rules/rules.go`、`rules/rules_test.go`。 |
| [GHSA-v9w4-gm2x-6rvf](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-v9w4-gm2x-6rvf) | Share links remain accessible after Share/Download permissions are revoked | 原代码已有措施：每次公开分享读取重新取用户并检查 Share 和 Download；`http/public_test.go`。 |
| [GHSA-xfqj-3vmx-63wv](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-xfqj-3vmx-63wv) | Stored Cross-Site Scripting via text/template branding injection | 原代码已有措施：品牌模板使用 html/template；`http/static.go`。 |
| [GHSA-5vpr-4fgw-f69h](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-5vpr-4fgw-f69h) | Stored Cross-Site Scripting via crafted EPUB file | 原代码已有措施：EPUB 不启用 allowScriptedContent，epubjs 默认 false；核对本地库 iframe sandbox 实现及 Preview.vue。 |
| [GHSA-x8jc-jvqm-pm3f](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-x8jc-jvqm-pm3f) | Signup Grants Execution Permissions When Default Permissions Includes Execution | 原代码已有措施：注册用户强制清除 Execute 和 Commands；`http/auth.go`。 |
| [GHSA-ffx7-75gc-jg7c](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-ffx7-75gc-jg7c) | TUS Negative Upload-Length Fires Post-Upload Hooks Prematurely | 原代码已有措施：TUS 注册前拒绝负长度，完成钩子只在合法块完成后执行；`http/tus_handlers.go`。 |
| [GHSA-5gg9-5g7w-hm73](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-5gg9-5g7w-hm73) | Signup Grants Admin When Default Permissions Include Admin | 原代码已有措施：注册用户强制清除 Admin；`http/auth.go`。 |
| [GHSA-9f3r-2vgw-m8xp](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-9f3r-2vgw-m8xp) | Access Rule Bypass via Path Traversal in Copy/Rename Destination Parameter | 原代码已有措施：复制/移动目标规范化后检查规则；`http/resource.go`、`http/rules_path_test.go`。 |
| [GHSA-xqp3-jq6g-x3qm](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-xqp3-jq6g-x3qm) | Authentication Bypass via Proxy Auth Header Forgery | 本轮处理 `84041e1c`：仅信任配置中的实际连接来源；默认回环地址；不信任 X-Forwarded-For/X-Real-IP；`auth/proxy_test.go`。 |
| [GHSA-79pf-vx4x-7jmm](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-79pf-vx4x-7jmm) | TUS Delete Endpoint Bypasses Delete Permission Check | 原代码已有措施：TUS 删除入口检查 Delete；`http/tus_handlers.go`。 |
| [GHSA-68j5-4m99-w9w9](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-68j5-4m99-w9w9) | Authorization Policy Bypass in Public Share Download Flow | 原代码已有措施：公开下载统一经过 withHashFile 的权限、路径和规则检查；`http/public_test.go`。 |
| [GHSA-mr74-928f-rw69](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-mr74-928f-rw69) | Path Traversal in Public Share Links Exposes Files Outside Shared Directory | 原代码已有措施：公开路径 slashClean 后在分享子文件系统内解析；`http/public.go`、`http/public_symlink_test.go`。 |
| [GHSA-hxw8-4h9j-hq2r](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-hxw8-4h9j-hq2r) | Authentication Bypass in User Password Update | 原代码已有措施：密码更新要求当前密码；本轮补上字段列表省略和大小写分支；`http/users.go`、`http/reauth_test.go`。 |
| [GHSA-4mh3-h929-w968](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-4mh3-h929-w968) | Path-Based Access Control Bypass via Multiple Leading Slashes in URL | 原代码已有措施：请求路径统一规范化后才进入文件处理；`http/utils.go`、`http/utils_test.go`。 |
| [GHSA-43mm-m3h2-3prc](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-43mm-m3h2-3prc) | Username Enumeration via Timing Attack in /api/login | 原代码已有措施：不存在的用户名仍进行 dummyHash bcrypt 校验；`auth/json.go`。 |
| [GHSA-6jqf-mv7m-3q7p](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-6jqf-mv7m-3q7p) | Risk of HTTP Request/Response smuggling through vulnerable dependency | 原代码已有措施：go.mod 中 x/net 为 v0.57.0，超出该旧公告受影响范围；本项为依赖版本核对，不代表完整依赖漏洞扫描。 |
| [GHSA-6cqf-cfhv-659g](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-6cqf-cfhv-659g) | Insecure Direct Object Reference (IDOR) in Share Deletion Function | 原代码已有措施：删除分享核对 UserID 或管理员权限；`http/share.go`。 |
| [GHSA-w5fm-68j4-fpc4](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-w5fm-68j4-fpc4) | DoS Vulnerability on Public Login API | 原代码已有措施：登录/注册请求体已有 1 MiB 限制；本轮另补认证尝试限制；`http/auth.go`。 |
| [GHSA-7xwp-2cpp-p8r7](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-7xwp-2cpp-p8r7) | Insecure JWT Handling Allows Session Replay Attacks after Logout | 本轮处理 `0c718764 / b3550827`：数据库持久化可撤销会话；退出撤销整个续期会话族，旧签名令牌不能重放；`http/auth_session_test.go`、`storage/bolt/sessions_test.go`。 |
| [GHSA-7xqm-7738-642x](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-7xqm-7738-642x) | Uncontrolled Memory Consumption Due to Oversized File Processing | 本轮处理 `1d862200`：原版已限制静态文本大小；本轮补强读取期间文件增长的缺口，并消除编码下载的第二次无界读取；`files/content_limit_test.go`。 |
| [GHSA-rmwh-g367-mj4x](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-rmwh-g367-mj4x) | Sensitive Data Transferred in URL | 原代码已有措施：普通认证不再从 URL 提取 JWT；只接受头及 GET Cookie；本轮亦移除分享 token URL；`http/auth.go`。 |
| [GHSA-3v48-283x-f2w4](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-3v48-283x-f2w4) | Password Protection of Links Bypassable | 本轮处理 `cfd496f8`：永久 URL token 不再授予分享访问；密码验证后用最长一小时、绑定该分享的 HttpOnly Cookie；`http/share_access_test.go`。 |
| [GHSA-cm2r-rg7r-p7gg](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-cm2r-rg7r-p7gg) | Insecure Password Handling | 本轮处理 `29e47551`：密码策略和随机初始密码原已存在；补上登录和 WebDAV 共用的失败预算、校验并发上限与有界来源缓存；`http/login_limit_test.go`、`auth/password_limiter_test.go`。 |
| [GHSA-3q2w-42mv-cph4](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-3q2w-42mv-cph4) | Shell Commands Can Spawn Other Commands | 本轮处理 `d7d6573e`：收紧权限边界：普通账户的 Execute 授权不能再获取服务器命令执行，交互命令仅管理员可用；未宣称任意程序可被白名单安全沙箱化。 |
| [GHSA-hc8f-m8g5-8362](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-hc8f-m8g5-8362) | Command Execution not Limited to Scope | 本轮处理 `d7d6573e`：同上：关闭向受限用户委派服务器 OS 权限的路径；管理员显式启用后仍拥有服务进程的 OS 权限。 |
| [GHSA-jj2r-455p-5gvf](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-jj2r-455p-5gvf) | Insecure File Permissions | 原代码已有措施：文件、目录默认模式为 0640/0750，数据库初始化采用受限模式；`settings/settings.go`、`cmd/utils.go`。 |
| [GHSA-w7qc-6grj-w7r8](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-w7qc-6grj-w7r8) | Command Execution Allowlist Bypass | 原代码已有措施：命令名使用 slices.Contains 精确匹配；另按本轮命令边界调整，仅管理员可以执行。 |
| [GHSA-4wx8-5gm2-2j97](https://github.com/filebrowser/filebrowser/security/advisories/GHSA-4wx8-5gm2-2j97) | Stored Cross-Site Scripting | 原代码已有措施：原始文件响应设置 script-src none 与 nosniff；`http/raw.go`。 |

## 相关 issue

- [#5199](https://github.com/filebrowser/filebrowser/issues/5199)：交互命令与钩子。已分别处理 WebSocket、shell 元字符、钩子源码插入及非管理员服务器命令权限；管理员命令仍具有服务进程 OS 权限，未提供跨平台任意程序沙箱。
- [#5216](https://github.com/filebrowser/filebrowser/issues/5216)：使用随机、持久化、可撤销的会话标识；续期原子替换；退出撤销会话族；密码变化使旧令牌失效。JWT 保留为前端元数据兼容外壳，不再单靠 JWT 签名授权。过期记录按访问及新登录回收，每用户最多 64 个登录会话。
- [#5213](https://github.com/filebrowser/filebrowser/issues/5213)：`f0901a20`。用户更新省略 which 或改变敏感字段大小写时也必须复核当前密码；全局设置写入同样复核。`http/reauth_test.go` 覆盖拒绝和允许场景。
- [#5239](https://github.com/filebrowser/filebrowser/issues/5239)：密码分享不再通过永久 URL token 授权；浏览器下载使用有期限的专属 Cookie。
- [#5294](https://github.com/filebrowser/filebrowser/issues/5294)：普通静态大文本原有 10 MiB 分类上限；实际读取限制及字幕单独修复，未将旧报告一概认定为完全复现。
- [#5834](https://github.com/filebrowser/filebrowser/issues/5834)：负 Upload-Length 已被拒绝，未重复修复。
- [#1195](https://github.com/filebrowser/filebrowser/issues/1195)、[#1761](https://github.com/filebrowser/filebrowser/issues/1761)、[#2570](https://github.com/filebrowser/filebrowser/issues/2570)、[#2042](https://github.com/filebrowser/filebrowser/issues/2042)：旧用户权限大小写绕过、其他用户信息泄露、原始文件脚本执行及 Cookie 写操作问题已有对应边界；同时核对了当前用户字段限制、敏感操作复核、响应 CSP 和非 GET Cookie 不授权。
- [#1805](https://github.com/filebrowser/filebrowser/issues/1805)：部署不得把数据库和凭据文件放入不可信用户可读写的文件作用域。没有本地生产部署配置可据以确认该条件；本轮数据库只存会话标识摘要，获得 JWT 签名密钥也不能凭空建立对应用户的有效会话。不要把这一点当作允许暴露数据库的理由。
- [#3900](https://github.com/filebrowser/filebrowser/issues/3900)：受信任代理自动创建用户是代理认证的既有行为；本轮限制代理来源，未把公开 signup 开关误当作代理自动配置策略。
- [#5971](https://github.com/filebrowser/filebrowser/issues/5971)：公开内容只说明已提交私有报告，没有细节可供验证，未声称修复未公开问题。
- 其他检索命中中的 2FA、审计日志、配额、OAuth/LDAP 等功能建议、无复现细节报告及与安全无关条目，没有当作已确认漏洞新增补丁。

## 升级与兼容性

- 所有实例需同步升级；旧登录令牌失效，用户重新登录。JWT 仍用于前端兼容，服务器每次另查可撤销会话。
- 非管理员的交互 Execute 授权不再生效。交互命令不使用全局 shell；钩子变量通过环境传递，不再由服务器展开进参数/脚本。
- 代理认证默认仅信任回环地址。远程/容器代理配置 `--auth.trustedProxies=IP/CIDR`，并替换客户端身份头。详见 [认证说明](authentication.md)。
- 密码分享旧 token 下载 URL 失效；输入密码后可在同一浏览器中下载，Cookie 最长有效一小时。公开无密码分享维持公开行为。
- 公开注册始终新建独立家目录。旧账户作用域及旧文件权限没有自动迁移，避免擅自移动或修改已有数据；部署管理员需审查此前注册的根作用域账户。
- TUS 使用作用域文件的系统锁；Linux 测试验证了真实包装文件上的锁。Windows 只验证交叉编译，未在 Windows 上运行；不支持系统锁的平台拒绝该锁操作。共享存储必须支持系统文件锁语义。
- 密码尝试限制按实际来源计算，IPv6 按 /64；代理后的用户共享失败预算；限制是单实例的，多实例应在受信代理补充集群级限制。

## 验证

- `go test -race ./...`：全部通过，覆盖权限、符号链接规则、分享撤销、并发上传、密码与会话、注册隔离、认证限制等回归场景。
- `go vet ./...`、`go build -o /tmp/neko-security-filebrowser .`：通过。
- `frontend` 中 `pnpm run lint`：通过；`pnpm run test`：5 个文件、29 项测试全部通过。
- `frontend` 中 `pnpm run build`：应用 TypeScript 检查及 Vite 生产构建通过。构建仍提示现有资源分包和 Browserslist 更新等警告，没有构建错误。
- `GOOS=windows GOARCH=amd64 go test -c ./http -o /tmp/neko-security-http-windows.test.exe`：交叉编译通过；没有在 Windows 上执行测试。
- 真实服务验证：使用当前构建、临时数据库和文件目录、独立本机端口运行。验证登录 Cookie 为 HttpOnly、续期使旧令牌失效、使用刚轮换的旧令牌退出能撤销后继会话、密码分享 Cookie 下载成功且复制 URL/token 不能授权，以及全局设置必须复核当前密码。全部通过，服务及临时数据随后清理。
- `git diff --check`：通过。

核查范围是上述公开公告、可验证的公开 issue 及相关本地代码；没有执行完整依赖漏洞扫描，也没有验证未公开的私有报告或所有平台、生产部署配置。

后续独立核查与功能扩展：新增可配置的登录/WebDAV 共享临时封禁、WebDAV 设置页即时启停，并修复 WebDAV 文件响应隔离、特殊文件、符号链接锁、随机锁令牌与作用域校验、tagged If 锁绕过问题。每项复现、修复提交及验证见 [WebDAV 核查记录](webdav.md#本轮独立核查2026-11-15)。
