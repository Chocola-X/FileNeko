# Deploying FileNeko

**English** | [简体中文](deployment_CN.md)

FileNeko manages files accessible to its server process. Treat it as a sensitive service rather than a public static-file website. It is a personal project with best-effort maintenance, not a security-audited hosting platform.

## Recommended baseline

- Run `fileneko` as an unprivileged system user, with access only to intended file directories and its database/configuration.
- Do not expose the application directly to the public internet. Put it behind a reverse proxy with HTTPS and suitable access restrictions or additional authentication.
- Keep **command execution disabled** unless you need it and accept the risk; see [Command execution](command-execution.md).
- Back up your file data, database, and settings before upgrades. Avoid publicly exposing the database or sharing sensitive logs.
- Review [SECURITY.md](../SECURITY.md) and the [security audit](security-audit-2026-11-15.md). The audit is not a guarantee of safety.

The repository does **not currently publish official FileNeko Docker images**. Use [standalone executables](installation.md) or a source build; upstream File Browser images are not equivalent.

## Self-registration

Self-registration is configurable under **Settings → Global Settings**, or through `fileneko config set --signup`. Newly registered accounts inherit the configured user defaults and scope. **Review these defaults before enabling signup**: otherwise users may be assigned access broader than intended.

FileNeko provides the `createUserDir` option to create a per-user directory. Regardless of that option, confirm the effective default scope, user permissions, and filesystem ownership for your installation. Disable signup if you do not need public registration.

## Login protection and reverse proxies

Unlike older File Browser documentation, FileNeko has **built-in temporary login blocking** for failed password attempts, covering web login and WebDAV Basic authentication. The documented defaults are five failures in five minutes followed by a five-minute block. These settings can be administered under **Settings → Global Settings → Fail2Ban / Login Protection**. See [WebDAV](webdav.md#preventing-password-brute-force).

By default, the application uses the direct connecting IP address. When operating behind a reverse proxy, configure trusted proxy addresses **carefully** and ensure the proxy overwrites client-supplied forwarding headers. Otherwise users behind a proxy may share a block budget, while trusting arbitrary forwarding headers can make IP-based limits ineffective. Refer to [reverse-proxy.md](reverse-proxy.md) for the current Nginx example.

The built-in limiter is process-local and does not replace proxy-level protection or appropriate authentication. Multiple application instances do not automatically share the in-memory block list.

### Optional external fail2ban

You can still use an external fail2ban installation for additional host-level filtering, but it is **not a prerequisite** for FileNeko's built-in login protection. If you configure it, verify log formats and source IP handling on your actual deployment rather than copying an older upstream File Browser regular expression untested.
