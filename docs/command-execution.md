# Command Execution

**English** | [简体中文](command-execution_CN.md)

> [!CAUTION]
>
> The **hook runner** and **interactive shell** functionalities have been disabled for all existent and new installations by default from version v2.33.8 and onwards, due to continuous and known security vulnerabilities. You should only use this feature if you are aware of all of the security risks involved. For more up to date information, consult issue [#5199](https://github.com/filebrowser/filebrowser/issues/5199).

## Hook Runner

The hook runner is a feature that enables you to execute any shell command you want before or after a certain event. Right now, these are the events:

* Copy
* Rename
* Upload
* Delete
* Save

Also, during the execution of the commands set for those hooks, there will be some environment variables available to help you perform your commands:

* `FILE` with the full absolute path to the changed file.
* `SCOPE` with the path to user's scope.
* `TRIGGER` with the name of the event.
* `USERNAME` with the user's username.
* `DESTINATION` with the absolute path to the destination. Only used for **copy** and **rename.**

At this moment, you can edit the commands via the command line interface, using the following commands \(please check the flag `--help` to know more about them\):

```bash
fileneko cmds add before_copy "echo $FILE"
fileneko cmds rm before_copy 0
fileneko cmds ls
```

Or you can use the web interface to manage them via **Settings** → **Global Settings**.

## Interactive Shell

Within FileNeko you can toggle the shell (`< >` icon at the top right) and this will open a shell command window at the bottom of the screen. This functionality can be turned on using the environment variable `FB_DISABLE_EXEC=false` or the flag `--disable-exec=false`.

By default no commands are available as the command list is empty. To enable commands these need to either be done on a per-user basis (including for the Admin user).

You can do this by adding them in Settings > User Management > (edit user) > Commands or to *apply to all new users created from that point forward* they can be set in Settings > Global Settings

> [!NOTE]
> 
> If using a proxy manager then remember to enable websockets support for the FileNeko proxy

> [!NOTE]
> 
> If using Docker and you want to add a new command that is not in the base image then you will need to build a custom Docker image using `ghcr.io/chocola-x/fileneko` as a base image.  For example to add 7z:
> 
> ```docker
> FROM ghcr.io/chocola-x/fileneko
> RUN sudo apt install p7zip-full
> ```

## Local fork security changes

Hook variables are passed only as environment variables. They are no longer
expanded by the server into executable arguments or shell source. For example,
use a script that reads its environment, or configure a POSIX shell hook as
`sh -c 'printf "%s\n" "$FILE"'`. Quote variable expansions in scripts and do not
pass their contents to `eval` or another interpreter as source code. A direct
hook such as `echo $FILE` now receives the literal argument `$FILE`.

Interactive commands are executed directly as an executable and an argument
list. The global shell setting applies only to administrator-configured hooks;
it does not interpret interactive command input.

Interactive execution is now restricted to administrators with Execute
permission, and still requires the explicit server execution opt-in. Existing
Execute grants to non-administrators no longer permit command execution. This
closes delegated command scope escapes and subprocess allowlist bypasses without
pretending that a working directory or an executable allowlist is an OS sandbox.
An administrator who enables this feature receives the server process's OS
privileges; scope and path rules do not confine those commands. Use file APIs for
operations that must respect a restricted user's scope.
