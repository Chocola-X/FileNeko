# fileneko completion fish

**English** | [简体中文](fileneko-completion-fish_CN.md)

Generate the autocompletion script for fish

## Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	fileneko completion fish | source

To load completions for every new session, execute once:

	fileneko completion fish > ~/.config/fish/completions/fileneko.fish

You will need to start a new shell for this setup to take effect.


```
fileneko completion fish [flags]
```

## Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
```

## See Also

* [fileneko completion](fileneko-completion.md)	 - Generate the autocompletion script for the specified shell

