# fileneko config import

[English](fileneko-config-import.md) | **简体中文**

从文件导入配置

## 概要（Synopsis）

导入配置文件。这将会完全替换现有的所有配置。可以针对已存在或不存在的数据库使用。

如果针对尚不存在的数据库使用，将自动生成密钥；否则将保留数据库中的现有密钥。

文件路径必须为 JSON 或 YAML 文件。

```
fileneko config import <path> [flags]
```

## 选项（Options）

```
  -h, --help   import 子命令帮助信息
```

## 继承自父命令的选项

```
  -c, --config string     配置文件路径
  -d, --database string   数据库文件路径（默认 "./filebrowser.db"）
```

## 参考（See Also）

* [fileneko config](fileneko-config_CN.md)	 - 配置管理工具
