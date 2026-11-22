# fileneko rules

[English](fileneko-rules.md) | **简体中文**

访问规则管理工具

## 概要（Synopsis）

在各个子命令中，您至少可以使用两个参数："username" 和 "id"。您必须且只能设置其中一个，或者两个都不设置。如果设置了其中一个，命令将应用于指定用户；否则将应用于全局规则集。

## 选项（Options）

```
  -h, --help              rules 命令帮助信息
  -i, --id uint           规则适用的用户 ID
  -u, --username string   规则适用的用户名
```

## 继承自父命令的选项

```
  -c, --config string     配置文件路径
  -d, --database string   数据库文件路径（默认 "./filebrowser.db"）
```

## 参考（See Also）

* [fileneko](fileneko_CN.md)	 - FileNeko：精致二次元风格的自托管网页文件管理器
* [fileneko rules add](fileneko-rules-add_CN.md)	 - 添加全局规则或指定用户规则
* [fileneko rules ls](fileneko-rules-ls_CN.md)	 - 列出全局规则或用户专用规则
* [fileneko rules rm](fileneko-rules-rm_CN.md)	 - 移除全局规则或用户规则
