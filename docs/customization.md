# Customization

You can customize the styles, branding and icons of your FileNeko instance in order to give it a personal touch.

## Custom Branding

You can customize FileNeko to use your own branding. This includes the following:

- **Name**: the name of the instance that shows up on the tab title, login pages, and some other places.
- **Disable External Links**: disables all external links, except to the documentation.
- **Disable Used Percentage**: disables the disk usage information on the sidebar.
- **Branding Folder**: directory which can contain two items:
  - `custom.css`, containing a global stylesheet to apply to all users.
  - `img`, a directory which can replace all the [default logotypes](https://github.com/filebrowser/filebrowser/tree/master/frontend/public/img) from the application.

This can be configured by the administrator user, under **Settings → Global Settings**. You can also update the configuration directly using the [CLI](cli/fileneko-config-set.md):

```sh
fileneko config set --branding.name "My Name" \
  --branding.files "/abs/path/to/my/dir" \
  --branding.disableExternal
```

> [!NOTE] 
>
> If you are using Docker, you need to mount a volume with the `branding` directory in order for it to be accessible from within the container.

### Custom Icons

To replace the default logotype and favicons, you need to create an `img` directory under the branding directory. The structure of this directory must mimic the one from the [default logotypes](https://github.com/filebrowser/filebrowser/tree/master/frontend/public/img):

```
img/
  logo.svg
  icons/
    favicon.ico
    favicon.svg
    (...)
```

Note that there are different versions of the same favicon in multiple sizes. To replace all of them, you need to add versions for all of them. You can use the [Real Favicon Generator](https://realfavicongenerator.net/) to generate these for you from your base image. 

> [!NOTE]
>
> The icons are cached by the browser, so you may not see your changes immediately. You can address this by clearing your browser's cache.

## 此分支的主题定制

自带主题使用正常的 CSS 优先级，不再使用 `!important`。外观保持原来的亮色主题；管理员仍可通过品牌资源目录中的独立 `custom.css` 定制所有用户的外观，无需修改前端源码。

可以从 [custom.css 示例](examples/custom.css) 开始，按需删除不需要的规则。资源目录示例：

```text
branding/
  custom.css
  img/
    logo.svg
    background.jpg
    login.jpg
    JCuYuan.woff2
```

`img/logo.svg` 同时用于登录页和登录后的左上角 Logo。CSS 中使用 `url("./img/background.jpg")` 这样的相对路径，在配置 `/library` 等 base URL 后也能正确加载。自定义字体请放在本地，通过 `@font-face` 引入，并保留 `--font-ui` 的系统字体回落；不要覆盖 Material Icons 的图标字体。旧版主题中依赖组件哈希属性、旧 DOM 结构或固定布局的选择器需要适配，不能直接整份照搬。

优先通过 `:root` 中的变量配置外观，未设置的变量保留默认效果：

| 变量 | 用途 |
| --- | --- |
| `--font-ui`、`--font-heading`、`--font-mono`、`--font-signature` | 界面、标题、代码、英文后缀装饰字体 |
| `--page-background-color`、`--page-background-image`、`--page-background-size` | 文件管理界面的页面背景 |
| `--login-background-image`、`--login-background-size` | 登录页背景 |
| `--login-panel-background`、`--sidebar-background` | 登录卡片、侧边栏背景，支持半透明颜色 |
| `--file-card-background`、`--file-card-inner-background` | 文件卡片外层和内层背景 |
| `--file-card-frame-color` | 文件卡片内层装饰边框颜色 |
| `--login-frame-opacity`、`--sidebar-frame-opacity`、`--settings-frame-opacity` | 登录、侧边栏、设置卡片装饰框透明度 |

卡片有外层背景和内层装饰面，设置透明效果时需要同时调整两层。不要对整张卡片设置 `opacity`，否则文字、图片和按钮也会变透明。文件卡片的 `--card-main` 等配色变量在各卡片上定义；引用这些变量的 `color-mix()` 应写在卡片规则中，避免在 `:root` 提前解析。

更细的调整可以在 `custom.css` 中使用普通选择器。它在内置样式之后加载，同等优先级时自定义规则生效；遇到更具体的默认选择器时，使用对应选择器或上述变量即可。示例也演示了设置卡片、弹窗和卡片装饰的普通样式覆盖。数字输入组件的适配规则以 `#app` 为作用域，覆盖它时也需要保留此作用域。

修改品牌资源不需要重新编译程序；浏览器可能缓存资源一天，调试时禁用缓存或强制刷新。首次设置或变更品牌资源目录后，重新加载页面；通过 CLI 修改数据库配置前请先停止使用该数据库的服务。

### 内置样式维护与验证

样式层顺序为 `vendor → base → components`，在 HTML 的最前部声明，避免 Vite 的组件 CSS 输出顺序改变层顺序。库样式放在 `vendor`，旧布局与基础控件放在 `base`，Vue 组件样式放在 `components`。三个主题文件保持未分层的普通规则，用于适配运行时插入的第三方样式；品牌 `custom.css` 最后加载，也应使用未分层的普通规则。

新增样式应修改负责该元素的规则或主题变量，不要追加 `!important`。文件列表列宽由 CSS Grid 决定，不再由 JavaScript 修改样式表。`pnpm run test` 会检查项目自有 CSS、Vue 组件和内联样式中没有强制优先级，并验证 HTML 中的层声明位置。第三方依赖的内部样式不在清理范围。

视觉回归脚本使用临时数据库和文件，不连接现有部署。需要 Chromium、FFmpeg，以及 Python 的 `websocket-client` 和 `Pillow`：

```sh
python3 frontend/scripts/theme-regression.py /absolute/path/to/old-filebrowser /tmp/theme-before
python3 frontend/scripts/theme-regression.py /absolute/path/to/new-filebrowser /tmp/theme-after
python3 frontend/scripts/compare-theme.py /tmp/theme-before /tmp/theme-after
```

脚本覆盖桌面、手机端的登录、设置、数字控件、六类弹窗、搜索结果、三种文件视图、多选操作、侧边栏、图片、音频、视频、PDF、EPUB 和 CSV 预览、文本编辑及 Markdown 预览。截图前固定时间、容量文本与进度条，清除鼠标悬浮状态和内容随临时分享地址变化的二维码画布；仍保留二维码的尺寸与外框供比较。比较结果包含截图像素差异与元素计算样式差异，不替代对新增交互的人工检查。
