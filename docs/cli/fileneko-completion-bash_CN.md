# fileneko completion bash

[English](fileneko-completion-bash.md) | **简体中文**

生成 Bash Shell 自动补全脚本

## 概要（Synopsis）

生成适用于 Bash Shell 的自动补全脚本。

此脚本依赖 'bash-completion' 软件包。如果尚未安装，您可以通过操作系统的包管理器进行安装。

在当前 Shell 会话中加载补全：

	source <(fileneko completion bash)

如需在每个新会话中自动加载补全，请执行一次：

### Linux:

	fileneko completion bash > /etc/bash_completion.d/fileneko

### macOS:

	fileneko completion bash > $(brew --prefix)/etc/bash_completion.d/fileneko

您需要重新启动 Shell 才能使配置生效。

```
fileneko completion bash
```

## 选项（Options）

```
  -h, --help              bash 补全帮助信息
      --no-descriptions   禁用补全描述信息
```

## 继承自父命令的选项

```
  -c, --config string     配置文件路径
  -d, --database string   数据库文件路径（默认 "./filebrowser.db"）
```

## 参考（See Also）

* [fileneko completion](fileneko-completion_CN.md)	 - 为指定的 Shell 生成自动补全脚本
