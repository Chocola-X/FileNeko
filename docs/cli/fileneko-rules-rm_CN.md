# fileneko rules rm

[English](fileneko-rules-rm.md) | **简体中文**

移除全局规则或用户规则

## 概要（Synopsis）

移除全局规则或用户专用规则。提供的索引值与运行 'rules ls' 时打印的序号相同。请注意，在每次添加或删除之后，规则的索引序号都会改变，因此在连续删除时请格外小心。

您还可以指定可选参数 (index_end)，以便批量删除从 'index' 到 'index_end' 之间的所有规则（包含 index_end）。

```
fileneko rules rm <index> [index_end] [flags]
```

## 选项（Options）

```
  -h, --help         rm 子命令帮助信息
      --index uint   要移除的规则索引
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
