# FileNeko CSS 主题接口 v1

[English](theme-api.md) | **简体中文**

在 **全局设置 → 品牌 → 品牌资源目录路径** 设置目录，把主题样式保存为目录下的
`custom.css`。图片、Logo 和字体放在 `img/`，在 CSS 中用
`url("./img/文件名")` 引用。根路径和子路径部署都支持这种写法。
可复制 [示例主题](personalized-examples/custom.css) 开始修改。配置和缓存说明见
[自定义外观](customization_CN.md)。

## 公开变量

在 `:root` 设置全局变量，或在表中的组件类上设置变量以缩小影响范围。
未设置的变量沿用内置主题。背景值可使用颜色、渐变或图片；只接受其中一种类型的变量在表中另有说明。

| 变量 | 作用 |
| --- | --- |
| `--font-ui`、`--font-heading` | 界面字体与品牌及标题字体。标题字体默认继承界面字体。 |
| `--font-mono`、`--font-signature` | 编辑器及等宽信息；英文文件后缀、Folder 和签名字体。不要把这些字体应用到图标字形上。 |
| `--page-background-color`、`--page-background-image` | 全局页面底色和背景图层，也用于启动加载页。分别接受颜色和 `url(...)`、渐变或 `none`。 |
| `--page-background-size`、`--page-background-position`、`--page-background-repeat` | 全局及启动页的背景铺放方式。单张壁纸建议设为 `cover`、`center`、`no-repeat`。 |
| `--login-background-image`、`--login-background-size`、`--login-background-position`、`--login-background-repeat` | 登录页独立背景及铺放方式。背景变量接受 `url(...)`、渐变或颜色。 |
| `--header-background`、`--sidebar-background`、`--login-panel-background` | 页头、侧栏及登录卡片背景。 |
| `--file-card-background`、`--file-card-inner-background`、`--file-card-frame-color` | 文件卡片外层背景、内层背景、内层虚线颜色。列表模式可能简化装饰。 |
| `--settings-card-background`、`--dialog-background`、`--loading-card-background` | 设置卡片、通用弹窗和加载卡片的底层背景；标题条、按钮区和输入控件可有独立填色。 |
| `--card-background` | 公开分享卡片和状态卡片（例如错误页、空目录）的背景，可在页面或卡片范围内覆盖。 |
| `--login-frame-opacity`、`--sidebar-frame-opacity`、`--settings-frame-opacity` | 对应卡片的虚线装饰透明度，取值 0 到 1。 |
| `--dialog-backdrop` | 弹窗和移动端覆盖菜单的遮罩背景。 |
| `--ink-navy`、`--ink-slate`、`--ink-muted` | 常用正文、次要文字和弱提示色。 |
| `--cobalt-main`、`--cobalt-dark`、`--cobalt-light`、`--cobalt-surface` | 蓝色色板：主色、描边、浅色和底色。 |
| `--pink-main`、`--pink-dark`、`--pink-light`、`--pink-surface` | 粉色色板，同上。 |
| `--cyan-main`、`--cyan-dark`、`--cyan-light`、`--cyan-surface` | 青色色板，同上。 |
| `--purple-main`、`--purple-dark`、`--purple-light`、`--purple-surface` | 紫色色板，同上。 |

色板影响使用它们的组件，例如按钮、边框、阴影和文件类型配色，并不会重写
所有插画和组件的专用颜色。若启用了深色模式，希望覆盖其专用配色时，可在
`:root.dark` 中设置相应变量。自定义字体请保留回落字体，并使用拥有相应使用权的资源。

使用本地壁纸时，请按示例在 `.fn-page-background` 和 `.fn-page--login` 上
直接写 `background-image: url("./img/文件名")`。这样路径始终相对于
`custom.css`。图片 URL 若放进变量、再由内置规则读取，可能相对于内置样式或页面解析；
在变量中使用图片时应填写包含部署子路径的完整站内路径。

## 固定选择器

以下类名会原样出现在页面中。可以对类本身设置变量或写普通 CSS 规则，
例如 `.fn-page--share { --card-background: #fff4fa; }`。
同一部位若已有更具体的内置规则，优先设置对应变量；自定义样式在内置样式之后加载，
同等优先级时自定义规则生效。

| 选择器 | 对应部位 |
| --- | --- |
| `.fn-page-background` | 全局页面和启动加载屏的壁纸层。 |
| `.fn-page--login`、`.fn-page--files`、`.fn-page--settings`、`.fn-page--share` | 登录、文件管理、设置、公开分享页面容器。 |
| `.fn-login-panel`、`.fn-logo`、`.fn-wordmark` | 登录表单卡片、登录与页头 Logo 图片、品牌文字。 |
| `.fn-header`、`.fn-sidebar`、`.fn-storage` | 页头、导航侧栏、侧栏存储空间卡片。 |
| `.fn-file-card`、`.fn-file-card__surface`、`.fn-file-card__suffix` | 文件卡片、内层装饰面、后缀或 Folder 文字。按内容和视图模式显示。 |
| `.fn-settings-card`、`.fn-share-card` | 设置的各管理卡片、公开分享页的卡片。 |
| `.fn-dialog`、`.fn-dialog-title`、`.fn-backdrop` | 通用弹窗容器、标题栏、弹窗及移动菜单的遮罩。可在 `.fn-dialog` 上设置 `--dialog-background`。 |
| `.fn-preview`、`.fn-editor`、`.fn-error` | 文件预览、编辑器、错误页面或错误区域的外壳。预览文档和第三方控件内部不属于这一接口。 |
| `.fn-loader` | 启动和页面内的加载卡片。 |

## 兼容约定

`html[data-fn-theme-api="1"]` 标记当前主题接口版本，不需要写进普通 v1 主题。
同一接口版本内可新增变量和类；已公开的名称、作用和适用部位须继续兼容。
如果内部实现改名，应保留旧类或变量入口。确实需要移除或改变公开含义时，
必须先提供迁移说明并明确更新接口版本。CI 的主题契约检查会校验 v1 名称仍由页面和样式使用。

公开接口不依赖 Vite 资源文件名、Vue 的 `data-v-*`、元素标签或 DOM 层级。
这些内部细节仍可用于当前版本的临时调整，但不在兼容承诺内。
接口保证可定位的区域和变量用途，允许新版主题演进，包括默认配色、间距和布局变化。
