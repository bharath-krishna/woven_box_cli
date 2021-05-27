package main

import (
	"github.com/urfave/cli/v2"
)

func getCommands() []*cli.Command {
	commands := []*cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list files",
			Action:  listFilesAction,
		},
		{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "delete files",
			Action:  deleteFileAction,
		},
		{
			Name:    "upload",
			Aliases: []string{"u"},
			Usage:   "upload files",
			Action:  uploadFileAction,
		},
		{
			Name:    "login",
			Aliases: []string{"lo"},
			Usage:   "upload files",
			Action:  loginAction,
		},
	}
	return commands
}
