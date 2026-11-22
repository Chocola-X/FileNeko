# fileneko completion zsh

[English](fileneko-completion-zsh.md) | **简体中文**

生成 Zsh Shell 自动补全脚本

## 概要（Synopsis）

生成适用于 Zsh Shell 的自动补全脚本。

如果您的环境中尚未启用 Shell 补全功能，需要先启用它。您可以执行一次以下命令：

	echo "autoload -U compinit; compinit" >> ~/.zshrc

在当前 Shell 会话中加载补全：

	source <(fileneko completion zsh)

如需在每个新会话中自动加载补全，请执行一次：

### Linux:

	fileneko completion zsh > "${fpath[1]}/_fileneko"

### macOS:

	fileneko completion zsh > $(brew --prefix)/share/zsh/site-functions/_fileneko

您需要重新启动 Shell 才能使配置生效。

```
fileneko completion zsh [flags]
```

## 选项（Options）

```
  -h, --help              zsh 补全帮助信息
      --no-descriptions   禁用补全描述信息
```

## 继承自父命令的选项

```
  -c, --config string     配置文件路径
  -d, --database string   数据库文件路径（默认 "./filebrowser.db"）
```

## 参考（See Also）

* [fileneko completion](fileneko-completion_CN.md)	 - 为指定的 Shell 生成自动补全脚本
