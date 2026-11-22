# fileneko cmds rm

[English](fileneko-cmds-rm.md) | **简体中文**

从事件钩子中移除命令

## 概要（Synopsis）

从事件钩子中移除命令。提供的索引值与运行 'cmds ls' 时打印的序号相同。请注意，在每次添加或删除之后，命令的索引序号都会改变，因此在连续删除命令时请格外小心。

您还可以指定可选参数 (index_end)，以便批量删除从 'index' 到 'index_end' 之间的所有命令（包含 index_end）。

```
fileneko cmds rm <event> <index> [index_end] [flags]
```

## 选项（Options）

```
  -h, --help   rm 子命令帮助信息
```

## 继承自父命令的选项

```
  -c, --config string     配置文件路径
  -d, --database string   数据库文件路径（默认 "./filebrowser.db"）
```

## 参考（See Also）

* [fileneko cmds](fileneko-cmds_CN.md)	 - 命令运行器管理工具
