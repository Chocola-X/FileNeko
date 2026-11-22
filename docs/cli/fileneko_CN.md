# fileneko

[English](fileneko.md) | **简体中文**

FileNeko：精致二次元风格的自托管网页文件管理器

## 概要（Synopsis）

FileNeko CLI 允许您创建用于 FileNeko 的数据库，并在无需访问 Web 界面即可管理用户及所有系统配置。

如果您从未运行过 FileNeko，需要先为其准备一个数据库。无需担心：您不需要配置独立的数据库服务器。我们使用 Bolt DB 单文件数据库，完全由程序自行管理。

对于此命令，除指定配置文件的 "--config" 外，所有选项均可通过环境变量设置。环境变量以 "FB_" 为前缀，后接大写蛇形命名（UPPER_SNAKE_CASE）的参数名。例如，"--disablePreviewResize" 对应的环境变量为 FB_DISABLE_PREVIEW_RESIZE。

如果未指定 "--config"，FileNeko 将按以下目录顺序寻找名为 .filebrowser.{json, toml, yaml, yml} 的配置文件：

- ./
- $HOME/
- /etc/filebrowser/

**注意：** 只有下方列出的选项可以通过配置文件或环境变量进行设置。其他高级配置项仅保存在数据库中，必须通过 "config set" 或 "config import" 命令设置。

配置项的优先级顺序如下：

- 命令行参数（Flags）
- 环境变量（Environment variables）
- 配置文件（Configuration file）
- 数据库存储值（Database values）
- 默认值（Defaults）

此外，如果数据库文件尚不存在，FileNeko 将进入快速设置模式（Quick setup），自动引导创建新数据库，并使用 "username" 和 "password" 参数提供的凭据创建初始管理员账户。

```
fileneko [flags]
```

## 选项（Options）

```
  -a, --address string                 服务监听地址（默认 "127.0.0.1"）
  -b, --baseURL string                 基础路径（BaseURL）
      --cacheDir string                文件缓存目录（留空为禁用）
  -t, --cert string                    TLS 证书路径
  -c, --config string                  配置文件路径
  -d, --database string                数据库文件路径（默认 "./filebrowser.db"）
      --disableExec                    禁用命令运行器功能（默认 true）
      --disableImageResolutionCalc     禁用读取图片文件计算分辨率
      --disablePreviewResize           禁用调整图片预览尺寸
      --disableThumbnails              禁用图片缩略图
      --disableTypeDetectionByHeader   禁用读取文件头检测文件类型
      --enableWebDAV                   在 /dav/ 启用 WebDAV 支持
      --followExternalSymlinks         跟随目标超出用户作用域的符号链接（不安全）
  -h, --help                           fileneko 命令帮助信息
      --imageProcessors int            图片处理并发协程数（默认 4）
  -k, --key string                     TLS 私钥路径
  -l, --log string                     日志输出目标（默认 "stdout"）
      --noauth                         快速配置时使用免认证模式（noauth）
      --password string                快速配置时初始用户的哈希密码
  -p, --port string                    服务监听端口（默认 "8080"）
      --redisCacheUrl string           Redis 缓存地址（用于多实例部署），例如 redis://user:pass@host:port
  -r, --root string                    相对路径的前缀根目录（默认 "."）
      --socket string                  监听的 Unix Socket（不能与 address、port、cert、key 同时使用）
      --socketPerm uint32              Unix Socket 文件权限位（默认 438 即 0666）
      --tokenExpirationTime string     用户会话超时时间（默认 "2h"）
      --trustedProxies strings         客户端 IP 与 HTTPS 请求头的受信反向代理 IP 或 CIDR 列表（默认：无）
      --username string                快速配置时初始用户的用户名（默认 "admin"）
```

## 参考（See Also）

* [fileneko cmds](fileneko-cmds_CN.md)	 - 命令运行器管理工具
* [fileneko completion](fileneko-completion_CN.md)	 - 为指定的 Shell 生成自动补全脚本
* [fileneko config](fileneko-config_CN.md)	 - 配置管理工具
* [fileneko hash](fileneko-hash_CN.md)	 - 计算密码的哈希值
* [fileneko rules](fileneko-rules_CN.md)	 - 访问规则管理工具
* [fileneko users](fileneko-users_CN.md)	 - 用户管理工具
* [fileneko version](fileneko-version_CN.md)	 - 打印版本号
