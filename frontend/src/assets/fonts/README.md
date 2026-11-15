# 本地主题字体

应用的字体全部随前端构建发布，无需访问 Google Fonts。开发入口和生产入口都不引入字体外链；Vite 为字体生成带内容哈希的静态资源地址，沿用 File Browser 的 `StaticURL` 配置和 CSS 相对路径，支持子路径部署。

| 用途                                 | 本地字体                             | 实际字重                                  |
| ------------------------------------ | ------------------------------------ | ----------------------------------------- |
| 英文及拉丁文字                       | Fredoka                              | 可变字体 300–700                          |
| 简体中文                             | 原有 CuteRound（迷你简粗圆）         | 单一 400 字重，其他字重由浏览器匹配或合成 |
| 日文、假名及部分汉字回落             | M PLUS Rounded 1c                    | 独立的 400、700 字重                      |
| 代码、终端及数字装饰                 | JetBrains Mono                       | 可变字体 100–800                          |
| 文件卡片英文签名                     | 原有 FileSignature（Dancing Script） | 可变字体 400–700                          |
| 其他拉丁、西里尔、希腊及越南文字回落 | 原有 Roboto 分区字体                 | 400、500、700                             |

统一字体栈位于 `src/css/_variables.css`。界面与标题优先使用圆润字体；缺少字形时，依次使用其他本地字体、操作系统中日韩字体和通用无衬线字体。代码与终端使用单独的等宽字体栈；文件夹的 `Folder` 与文件后缀使用独立手写字体栈，不随界面语言切换。Material Icons 继续由 npm 依赖提供本地图标字体。

所有文字字体使用 `font-display: swap`：字体下载完成前先显示回落字体，字体不可用时仍能阅读和操作。只声明字体文件真实提供的字重，不把静态字体声明为可变字体。未打包的罕见汉字或其他文字由系统字体补全；具体外观取决于设备安装的字体。

新增字体取自 [Google Fonts 官方源码仓库](https://github.com/google/fonts)，于 2026-11-15 下载，用 `woff2_compress` 转为完整 WOFF2，保留字形覆盖范围：

- [Fredoka](https://github.com/google/fonts/tree/main/ofl/fredoka)：`Fredoka[wdth,wght].ttf`，附 `fredoka-OFL.txt`。
- [JetBrains Mono](https://github.com/google/fonts/tree/main/ofl/jetbrainsmono)：`JetBrainsMono[wght].ttf`，附 `jetbrains-mono-OFL.txt`。
- [M PLUS Rounded 1c](https://github.com/google/fonts/tree/main/ofl/mplusrounded1c)：`MPLUSRounded1c-Regular.ttf`、`MPLUSRounded1c-Bold.ttf`。字体内部声明 SIL Open Font License 1.1；`mplus-rounded-OFL.txt` 保留其版权声明及该许可证全文。

原有 Dancing Script 的许可证为 `dancing-script-OFL.txt`。本次保留已有 CuteRound、Dancing Script 和 Roboto 文件；新增的圆体和等宽字体不需要在安装或构建时联网下载。

编辑器的 Ace 模式、主题、补全片段和校验 Worker 也由 `src/utils/editor-assets.ts` 从已安装的 `ace-builds` 打包为本地静态资源，按需加载，不再请求 CDN。可选验证码服务及用户主动访问的外部链接不属于字体或编辑器资源。
