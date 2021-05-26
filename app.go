package main

import (
	"github.com/urfave/cli/v2"
)

func getNewApp() *cli.App {
	app := &cli.App{
		Name:     "Woven Box",
		Usage:    "Store files securely and access from anywhere",
		Commands: getCommands(),
	}

	return app
}
