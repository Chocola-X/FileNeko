# fileneko users add

[English](fileneko-users-add.md) | **简体中文**

创建新用户

## 概要（Synopsis）

创建新用户并将其添加到数据库中。

```
fileneko users add <username> <password> [flags]
```

## 选项（Options）

```
      --aceEditorTheme string   Ace 代码编辑器语法高亮主题
      --commands strings        用户可执行的命令列表
      --dateFormat              日期格式（true 为绝对时间，false 为相对时间）
  -h, --help                    add 子命令帮助信息
      --hideDotfiles            在文件列表中隐藏点文件（隐藏文件）
      --locale string           用户界面语言（默认 "en"）
      --lockPassword            锁定密码（禁止用户自行修改）
      --perm.admin              管理员权限
      --perm.create             创建文件/目录权限（默认 true）
      --perm.delete             删除文件/目录权限（默认 true）
      --perm.download           下载文件权限（默认 true）
      --perm.execute            命令执行权限（默认 true）
      --perm.modify             修改/覆盖文件权限（默认 true）
      --perm.rename             重命名文件/目录权限（默认 true）
      --perm.share              分享文件/目录权限（默认 true）
      --redirectAfterCopyMove   复制/移动后自动跳转到目标目录
      --scope string            用户文件作用域目录（默认 "."）
      --singleClick             仅使用单击打开文件/目录
      --sorting.asc             升序排序
      --sorting.by string       排序方式（name、size 或 modified，默认 "name"）
      --viewMode string         文件视图模式（默认 "list"）
```

## 继承自父命令的选项

```
  -c, --config string     配置文件路径
  -d, --database string   数据库文件路径（默认 "./filebrowser.db"）
```

## 参考（See Also）

* [fileneko users](fileneko-users_CN.md)	 - 用户管理工具
