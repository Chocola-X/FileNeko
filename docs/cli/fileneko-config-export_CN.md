# fileneko config export

[English](fileneko-config-export.md) | **简体中文**

将配置导出到文件

## 概要（Synopsis）

将配置导出到文件。文件路径必须为 JSON 或 YAML 文件格式。导出的配置可以修改后通过 'config import' 命令重新导入。

```
fileneko config export <path> [flags]
```

## 选项（Options）

```
  -h, --help   export 子命令帮助信息
```

## 继承自父命令的选项

```
  -c, --config string     配置文件路径
  -d, --database string   数据库文件路径（默认 "./filebrowser.db"）
```

## 参考（See Also）

* [fileneko config](fileneko-config_CN.md)	 - 配置管理工具
