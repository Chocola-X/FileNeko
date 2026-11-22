# fileneko completion powershell

[English](fileneko-completion-powershell.md) | **简体中文**

生成 PowerShell 自动补全脚本

## 概要（Synopsis）

生成适用于 PowerShell 的自动补全脚本。

在当前 Shell 会话中加载补全：

	fileneko completion powershell | Out-String | Invoke-Expression

如需在每个新会话中自动加载补全，请将上述命令的输出添加到您的 PowerShell 配置文件（profile）中。

```
fileneko completion powershell [flags]
```

## 选项（Options）

```
  -h, --help              powershell 补全帮助信息
      --no-descriptions   禁用补全描述信息
```

## 继承自父命令的选项

```
  -c, --config string     配置文件路径
  -d, --database string   数据库文件路径（默认 "./filebrowser.db"）
```

## 参考（See Also）

* [fileneko completion](fileneko-completion_CN.md)	 - 为指定的 Shell 生成自动补全脚本
