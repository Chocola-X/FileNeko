# FileNeko 品牌资源

[English](README.md) | **简体中文**

吉祥物图标同时用于应用程序图标和横向横幅（Banner）。横幅将其与内置 Fredoka Bold 字体的蓝色 `File` 和粉色 `Neko` 文字组合在一起，搭配白色描边和与应用主题匹配的实心阴影。

`banner.svg` 采用矢量字形轮廓，无需加载外部字体或其他资源。`banner.png` 是对应的透明位图导出文件。

如果修改了吉祥物或品牌文字后需要重新生成横幅，请先安装 Python 的 `fonttools`、`woff2_decompress` 与 `rsvg-convert`，然后运行：

```sh
python3 branding/render-banner.py
```

字体及其许可证位于 `frontend/src/assets/fonts/`。
