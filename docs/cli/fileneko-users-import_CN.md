# fileneko users import

[English](fileneko-users-import.md) | **简体中文**

从文件导入用户列表

## 概要（Synopsis）

从文件导入用户列表。文件路径必须为 JSON 或 YAML 文件。您可以使用此命令向现有安装中批量导入新用户。为此，只需在文件列表中不要指定用户的 ID 或将其设置为 0 即可。

```
fileneko users import <path> [flags]
```

## 选项（Options）

```
  -h, --help        import 子命令帮助信息
      --overwrite   覆盖具有相同 ID/用户名组合的用户
      --replace     替换整个用户列表（完全覆盖）
```

## 继承自父命令的选项

```
  -c, --config string     配置文件路径
  -d, --database string   数据库文件路径（默认 "./filebrowser.db"）
```

## 参考（See Also）

* [fileneko users](fileneko-users_CN.md)	 - 用户管理工具
