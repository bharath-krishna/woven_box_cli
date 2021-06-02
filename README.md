# Woven Box
## Command Line Application

This is a cli tool for intercting with woven box application [http://woven-box.bharathk.in](http://woven-box.bharathk.in)

Written in GoLang using [urfave vli (v2)](https://github.com/urfave/cli) package.

Download the woctl tool from [http://woven-box.bharathk.in/woctl](http://woven-box.bharathk.in/woctl)

## How to use
Download woctl provide executable permission, and move to system's bin directory
```
> curl http://woven-box.bharathk.in/woctl -o woctl
> chmod +x woctl
> mv woctl /usr/local/bin
> woctl help
```
You will see 

```bash
NAME:
   Woven Box - Store files securely and access from anywhere

USAGE:
   woctl [global options] command [command options] [arguments...]

COMMANDS:
   list, l    list files
   delete, d  delete files
   upload, u  upload files
   login, lo  Login to Woven Box
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```


### Commands
Field | Value
--- | ---
login, lo | login to woven box app with the username and password
list, l | List the files stored in woven box
delete, d | delete the uploaded file
upload, u | upload a file, the path needs to be preset in the current directory
help | show help message
---
### Flags
Field | Value
--- | ---
--help, -h | show help message
---

## Logging in and getting authentication token
Authenticate using username and password by running `woctl login` command. Once successfully authenticated a token in json format containing access token and refresh token will be downloaded and stored in `~/.woven_box_authn_token.json`.

This JWT token is used in every calls to woven box APIs [http://api.bharathk.in/apidocs](http://api.bharathk.in/apidocs) to authenticate and fetch files list, upload and delete files.

Once the token expires and failed to authenticate user needs to recreate the token by running `woctl login`. (Refreshing of access token feature is not implemented in this release).


## Listing files
Use `woctl list` or `woctl l` command to list the current files stored in woven box.
```
> woctl list
```
You will see something like this

```bash
+---+-------------------------------------+
| # | FILE NAME                           |
+---+-------------------------------------+
| 1 | IMG-20210529-WA0005.jpg             |
| 2 | 1622382903266422663446722082430.jpg |
| 3 | commands.go                         |
| 4 | main.go                             |
+---+-------------------------------------+
```
## Uploading files
Use `woctl upload` or `woctl u` command to upload files preset int he current directory. (Uploading multiple files and Uploading with relative path and absolute path is not implemented in this release).

```
> woctl upload <file in local dir>
```

## Deleting
Use `woctl delete` or `woctl d` command to delete files from woven box.

**_NOTE:_**  This command will permanently delete the files from storage and can not be recoverable.

Enjoy weaving with woctl.....

### END
