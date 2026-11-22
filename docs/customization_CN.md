# 外观定制

[English](customization.md) | **简体中文**

您可以自定义 FileNeko 实例的样式、品牌与图标，为其增添个性化色彩。

## 自定义品牌

您可以对 FileNeko 进行品牌定制，包括以下内容：

- **名称（Name）**：显示在浏览器标签页标题、登录页面等位置的实例名称。
- **禁用外部链接（Disable External Links）**：禁用除文档之外的所有外部链接。
- **隐藏版本号（Hide Version）**：在所有网页（包括公开分享页面）上隐藏软件版本号。版本号也会从页面的初始化引导数据中移除。保存此项更改会重新加载当前页面；其他已打开的标签页在刷新后生效。这不会影响 CLI 的 `version` 命令。
- **禁用已用容量百分比（Disable Used Percentage）**：在侧栏中禁用磁盘空间占用百分比显示。
- **品牌资源目录（Branding Folder）**：该目录可包含两个核心项：
  - `custom.css`：应用于所有用户的全局自定义样式表。
  - `img`：用于替换应用程序[默认图标与图样](https://github.com/filebrowser/filebrowser/tree/master/frontend/public/img)的图片目录。

管理员可以在**设置 → 全局设置**中进行配置。您也可以直接使用 [CLI](cli/fileneko-config-set_CN.md) 更新配置：

```sh
fileneko config set --branding.name "My Name" \
  --branding.files "/abs/path/to/my/dir" \
  --branding.disableExternal
```

使用 `fileneko config set --branding.hideVersion` 可以在网页端隐藏版本号，使用 `fileneko config set --branding.hideVersion=false` 则重新显示。

> [!NOTE] 
>
> 如果使用 Docker，需要为容器挂载 `branding` 目录所在的卷，以确保容器内部能够读取该目录。

### 自定义图标

要替换默认的 Logo 图标和 Favicon 网页图标，您需要在品牌目录下创建 `img` 目录。该目录结构需仿照[默认图标目录结构](https://github.com/filebrowser/filebrowser/tree/master/frontend/public/img)：

```
img/
  logo.svg
  icons/
    favicon.ico
    favicon.svg
    (...)
```

请注意，相同的 Favicon 存在多种不同尺寸的版本。如需彻底替换，应为每种尺寸都提供对应版本。您可以使用 [Real Favicon Generator](https://realfavicongenerator.net/) 基于您的原始图像一键生成这些文件。

> [!NOTE]
>
> 浏览器会对图标进行缓存，因此您可能无法立即看到更改。您可以通过清除浏览器缓存来解决此问题。

## 此分支的主题定制

自带主题使用正常的 CSS 优先级，不再使用 `!important`。外观保持原来的亮色主题；管理员仍可通过品牌资源目录中的独立 `custom.css` 定制所有用户的外观，无需修改前端源码。

可以从 [custom.css 示例](personalized-examples/custom.css) 开始，按需删除不需要的规则。完整的壁纸、Logo、字体及 90% 白色卡片配置见 [个性化示例](personalized-examples/README_CN.md)。资源目录示例：

```text
branding/
  custom.css
  img/
    logo.svg
    background.webp
    login.webp
    fonts/
      XiaolaiMono-Regular.woff2
```

`img/logo.svg` 同时用于登录页和登录后的左上角 Logo。CSS 中使用 `url("./img/background.webp")` 这样的相对路径，在配置 `/library` 等 base URL 后也能正确加载。自定义字体请放在本地，通过 `@font-face` 引入，并保留 `--font-ui` 的系统字体回落；不要覆盖 Material Icons 的图标字体。旧版主题中依赖组件哈希属性、旧 DOM 结构或固定布局的选择器需要适配，不能直接整份照搬。

优先使用 [CSS 主题接口 v1](theme-api_CN.md) 中的变量和固定类名。未设置的变量
保留默认效果。主题接口列出了受兼容承诺保护的名称；普通内部选择器仍可按需覆盖。

卡片有外层背景和内层装饰面，设置透明效果时需要同时调整两层。不要对整张卡片设置 `opacity`，否则文字、图片和按钮也会变透明。文件卡片的 `--card-main` 等配色变量在各卡片上定义；引用这些变量的 `color-mix()` 应写在卡片规则中，避免在 `:root` 提前解析。

更细的调整可以在 `custom.css` 中使用普通选择器。它在内置样式之后加载，同等优先级时自定义规则生效；遇到更具体的默认选择器时，优先使用主题变量。数字输入组件的适配规则以 `#app` 为作用域，其库内部类名不属于主题接口。

修改品牌资源不需要重新编译程序；浏览器可能缓存资源一天，调试时禁用缓存或强制刷新。首次设置或变更品牌资源目录后，重新加载页面；通过 CLI 修改数据库配置前请先停止使用该数据库的服务。

### 内置样式维护与验证

样式层顺序为 `vendor → base → components`，在 HTML 的最前部声明，避免 Vite 的组件 CSS 输出顺序改变层顺序。库样式放在 `vendor`，旧布局与基础控件放在 `base`，Vue 组件样式放在 `components`。主题文件保持未分层的普通规则，用于适配运行时插入的第三方样式；品牌 `custom.css` 最后加载，也应使用未分层的普通规则。

新增样式应修改负责该元素的规则或主题变量，不要追加 `!important`。文件列表列宽由 CSS Grid 决定，不再由 JavaScript 修改样式表。`pnpm run test` 会检查项目自有 CSS、Vue 组件和内联样式中没有强制优先级，并验证 HTML 中的层声明位置。第三方依赖的内部样式不在清理范围。

视觉回归脚本使用临时数据库和文件，不连接现有部署。需要 Chromium、FFmpeg，以及 Python 的 `websocket-client` 和 `Pillow`：

```sh
python3 frontend/scripts/theme-regression.py /absolute/path/to/old-filebrowser /tmp/theme-before
python3 frontend/scripts/theme-regression.py /absolute/path/to/new-filebrowser /tmp/theme-after
python3 frontend/scripts/compare-theme.py /tmp/theme-before /tmp/theme-after
```

脚本覆盖桌面、手机端的登录、设置、数字控件、六类弹窗、搜索结果、三种文件视图、多选操作、侧边栏、图片、音频、视频、PDF、EPUB 和 CSV 预览、文本编辑及 Markdown 预览。截图前固定时间、容量文本与进度条，清除鼠标悬浮状态和内容随临时分享地址变化的二维码画布；仍保留二维码的尺寸与外框供比较。比较结果包含截图像素差异与元素计算样式差异，不替代对新增交互的人工检查。
