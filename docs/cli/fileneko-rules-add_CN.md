# fileneko rules add

[English](fileneko-rules-add.md) | **简体中文**

添加全局规则或指定用户规则

## 概要（Synopsis）

添加全局规则或指定用户规则。

```
fileneko rules add <path|expression> [flags]
```

## 选项（Options）

```
  -a, --allow   指示该规则为允许规则（Allow）
  -h, --help    add 子命令帮助信息
  -r, --regex   指示该规则为正则表达式规则（Regex）
```

## 继承自父命令的选项

```
  -c, --config string     配置文件路径
  -d, --database string   数据库文件路径（默认 "./filebrowser.db"）
  -i, --id uint           规则适用的用户 ID
  -u, --username string   规则适用的用户名
```

## 参考（See Also）

* [fileneko rules](fileneko-rules_CN.md)	 - 访问规则管理工具
