# fileneko config init

[English](fileneko-config-init.md) | **简体中文**

初始化新数据库

## 概要（Synopsis）

初始化用于 FileNeko 的新数据库。此处的所有选项后续均可使用 'fileneko config set' 命令进行修改。与用户相关的参数将在创建新用户且未显式指定选项时作为默认值生效。

```
fileneko config init [flags]
```

## 选项（Options）

```
      --aceEditorTheme string            Ace 代码编辑器语法高亮主题
  -a, --address string                   服务监听地址（默认 "127.0.0.1"）
      --auth.command string              auth.method=hook 时执行的外部认证命令
      --auth.header string               auth.method=proxy 时的 HTTP 身份请求头
      --auth.logoutPage string           自定义登出页面 URL
      --auth.method string               身份认证类型（默认 "json"）
      --auth.trustedProxies strings      受信任的代理 IP 或 CIDR 列表（默认：仅回环地址）
  -b, --baseURL string                   基础路径（BaseURL）
      --branding.color string            设置主题颜色
      --branding.disableExternal         禁用 GitHub 等外部链接
      --branding.disableUsedPercentage   禁用已用磁盘空间百分比统计图
      --branding.files string            包含图片和自定义样式的品牌资源目录路径
      --branding.hideVersion             在所有网页上隐藏软件版本号
      --branding.name string             使用此名称替换实例名称 'FileNeko'
      --branding.theme string            设置主题
  -t, --cert string                      TLS 证书路径
      --commands strings                 用户可执行的命令列表
      --createUserDir                    自动为新用户生成主目录
      --dateFormat                       日期格式（true 为绝对时间，false 为相对时间）
      --dirMode string                   新建目录的权限位（默认 "0o750"）
      --disableExec                      禁用命令运行器功能（默认 true）
      --disableImageResolutionCalc       禁用读取图片文件计算分辨率
      --disablePreviewResize             禁用调整图片预览尺寸
      --disableThumbnails                禁用图片缩略图
      --disableTypeDetectionByHeader     禁用读取文件头检测文件类型
      --enableWebDAV                     在 /dav/ 启用 WebDAV 支持
      --fileMode string                  新建文件的权限位（默认 "0o640"）
      --followExternalSymlinks           跟随目标超出用户作用域的符号链接（不安全）
  -h, --help                             init 命令帮助信息
      --hideDotfiles                     在文件列表中隐藏点文件（隐藏文件）
      --hideLoginButton                  在公开页面上隐藏登录按钮
  -k, --key string                       TLS 私钥路径
      --locale string                    用户界面语言（默认 "en"）
      --lockPassword                     锁定密码（禁止用户自行修改）
  -l, --log string                       日志输出目标（默认 "stdout"）
      --minimumPasswordLength uint       新用户的最低密码长度（默认 12）
      --perm.admin                       管理员权限
      --perm.create                      创建文件/目录权限（默认 true）
      --perm.delete                      删除文件/目录权限（默认 true）
      --perm.download                    下载文件权限（默认 true）
      --perm.execute                     命令执行权限（默认 true）
      --perm.modify                      修改/覆盖文件权限（默认 true）
      --perm.rename                      重命名文件/目录权限（默认 true）
      --perm.share                       分享文件/目录权限（默认 true）
  -p, --port string                      服务监听端口（默认 "8080"）
      --recaptcha.host string            设置 ReCAPTCHA 服务器地址，国内可设为 recaptcha.net（默认 "https://www.google.com"）
      --recaptcha.key string             ReCaptcha 站点密钥（Site Key）
      --recaptcha.secret string          ReCaptcha 私钥（Secret Key）
      --redirectAfterCopyMove            复制/移动后自动跳转到目标目录
  -r, --root string                      相对路径的前缀根目录（默认 "."）
      --scope string                     用户文件作用域目录（默认 "."）
      --shell string                     其他命令追加执行的 Shell 命令
  -s, --signup                           允许用户公开自助注册
      --singleClick                      仅使用单击打开文件/目录
      --socket string                    监听的 Unix Socket（不能与 address、port、cert、key 同时使用）
      --sorting.asc                      升序排序
      --sorting.by string                排序方式（name、size 或 modified，默认 "name"）
      --tokenExpirationTime string       用户会话超时时间（默认 "2h"）
      --trustedProxies strings           客户端 IP 与 HTTPS 请求头的受信反向代理 IP 或 CIDR 列表（默认：无）
      --tus.chunkSize uint               TUS 上传分块大小（默认 10485760 字节 / 10MB）
      --tus.retryCount uint16            TUS 上传重试次数（默认 5）
      --viewMode string                  文件视图模式（默认 "list"）
```

## 继承自父命令的选项

```
  -c, --config string     配置文件路径
  -d, --database string   数据库文件路径（默认 "./filebrowser.db"）
```

## 参考（See Also）

* [fileneko config](fileneko-config_CN.md)	 - 配置管理工具
