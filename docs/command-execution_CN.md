# 命令执行

[English](command-execution.md) | **简体中文**

> [!CAUTION]
>
> 出于已知安全风险的考虑，自 v2.33.8 版本起，所有新旧安装默认均已禁用**钩子运行器（Hook Runner）**和**交互式终端（Interactive Shell）**功能。仅当您充分了解其中涉及的所有安全风险时，才应开启此功能。最新信息请参阅 Issue [#5199](https://github.com/filebrowser/filebrowser/issues/5199)。

## 钩子运行器（Hook Runner）

钩子运行器允许您在特定事件发生之前或之后执行自定义 Shell 命令。目前支持以下事件：

* 复制（Copy）
* 重命名（Rename）
* 上传（Upload）
* 删除（Delete）
* 保存（Save）

在执行这些钩子命令时，系统会注入以下环境变量供脚本使用：

* `FILE`：发生变更文件的绝对路径。
* `SCOPE`：当前用户的作用域路径。
* `TRIGGER`：触发事件的名称。
* `USERNAME`：触发操作的用户名称。
* `DESTINATION`：目标的绝对路径（仅用于**复制**和**重命名**事件）。

您可以通过命令行界面管理这些钩子命令（可通过 `--help` 参数查看具体用法）：

```bash
fileneko cmds add before_copy "echo $FILE"
fileneko cmds rm before_copy 0
fileneko cmds ls
```

或者通过网页端的**设置** → **全局设置**进行可视化管理。

## 交互式终端（Interactive Shell）

在 FileNeko 网页端，您可以点击右上角的 `< >` 图标打开屏幕下方的命令终端窗口。该功能可通过环境变量 `FB_DISABLE_EXEC=false` 或启动参数 `--disable-exec=false` 开启。

默认情况下命令列表为空，无可用命令。必须针对用户逐一启用（包括 Admin 管理员用户）。

您可以在**设置** → **用户管理** →（编辑用户）→ **命令（Commands）**中为现有用户添加可用命令；若要**对之后创建的所有新用户生效**，可在**设置** → **全局设置**中进行设置。

> [!NOTE]
> 
> 如果使用反向代理管理器，请务必为 FileNeko 代理启用 WebSocket 支持。

> [!NOTE]
> 
> 如果使用 Docker，且希望添加基础镜像中不存在的新命令，您需要基于 `ghcr.io/chocola-x/fileneko` 构建自定义 Docker 镜像。例如添加 7z：
> 
> ```docker
> FROM ghcr.io/chocola-x/fileneko
> RUN sudo apt install p7zip-full
> ```

## 本分支的安全变更

钩子变量现在**仅通过环境变量传递**。服务端不再将变量内容直接展开到可执行参数或 Shell 源码中。例如，应编写读取环境变量的脚本，或将 POSIX Shell 钩子配置为：
`sh -c 'printf "%s\n" "$FILE"'`。在脚本中请给变量加上引号，切勿将其内容作为源码传递给 `eval` 或其他解释器。直接配置为 `echo $FILE` 的钩子现在将接收到字面量字符串 `$FILE`。

交互式命令直接作为独立的可执行文件与参数列表执行。全局 Shell 设置仅适用于管理员配置的钩子，不会用来解释交互式命令的输入内容。

交互式命令执行权限现在**严格限制为具有 Execute 权限的管理员用户**，并且仍需要服务端显式开启命令执行开关。向非管理员用户授予的 Execute 权限不再允许执行服务器命令。这彻底关闭了受限作用域提权逃逸以及子进程白名单绕过的通道，避免将当前工作目录或白名单误当作真正的操作系统安全沙箱。启用此功能的管理员将直接拥有服务端进程的操作系统权限；作用域与路径访问规则无法限制这些命令。对于必须受限于普通用户作用域的操作，请务必使用文件 API 完成。
