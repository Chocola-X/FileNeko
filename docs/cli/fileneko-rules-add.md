# fileneko rules add

**English** | [简体中文](fileneko-rules-add_CN.md)

Add a global rule or user rule

## Synopsis

Add a global rule or user rule.

```
fileneko rules add <path|expression> [flags]
```

## Options

```
  -a, --allow   indicates this is an allow rule
  -h, --help    help for add
  -r, --regex   indicates this is a regex rule
```

## Options inherited from parent commands

```
  -c, --config string     config file path
  -d, --database string   database path (default "./filebrowser.db")
  -i, --id uint           id of user to which the rules apply
  -u, --username string   username of user to which the rules apply
```

## See Also

* [fileneko rules](fileneko-rules.md)	 - Rules management utility

