package main // import "src.techknowlogick.com/shiori"

import (
	"os"
	"runtime"
	"strings"

	gap "github.com/muesli/go-app-paths"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"src.techknowlogick.com/shiori/cmd"
	"src.techknowlogick.com/shiori/cmd/serve"
)

var (
	Version = "0.0.0"
	Tags    = ""
)

func main() {
	app := cli.NewApp()
	app.Name = "shiori"
	app.Usage = "Simple command-line bookmark manager built with Go"
	app.Version = Version + formatBuiltWith(Tags)
	app.Commands = []cli.Command{
		cmd.CmdAccount,
		cmd.CmdAdd,
		cmd.CmdDelete,
		cmd.CmdExport,
		cmd.CmdImport,
		cmd.CmdOpen,
		cmd.CmdPocket,
		cmd.CmdPrint,
		cmd.CmdSearch,
		serve.CmdServe,
		cmd.CmdUpdate,
	}
	globalFlags := []cli.Flag{
		cli.StringFlag{
			Name:   "db-type",
			Value:  "sqlite3",
			Usage:  "Type of database to use",
			EnvVar: "SHIORI_DBTYPE",
		},
		cli.StringFlag{
			Name:   "db-dsn",
			Value:  "shiori.db",
			Usage:  "database connection string",
			EnvVar: "SHIORI_DSN",
		},
		cli.StringFlag{
			Name:   "data-dir",
			Value:  getDataDir(),
			Usage:  "directory to store all files",
			EnvVar: "SHIORI_DIR, ENV_SHIORI_DIR",
		},
		cli.BoolFlag{
			Name:   "show-sql-log",
			Usage:  "Log SQL quries to command line",
			Hidden: true,
			EnvVar: "SHIORI_SHOW_SQL",
		},
	}
	app.Flags = append(app.Flags, globalFlags...)
	app.Before = func(c *cli.Context) error {
		// ensure data dir is created
		return os.MkdirAll(c.GlobalString("data-dir"), os.ModePerm)
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Errorf("%s: %v", os.Args, err)
	}
}

func getDataDir() string {
	// Try to use platform specific app path
	scope := gap.NewScope(gap.User, "shiori")
	dataDirs, err := scope.DataDirs()
	if err == nil && len(dataDirs) > 0 {
		return dataDirs[0]
	}

	// When all else fails, use current working directory
	return "."
}

func formatBuiltWith(Tags string) string {
	if len(Tags) == 0 {
		return " built with " + runtime.Version()
	}

	return " built with " + runtime.Version() + " : " + strings.Replace(Tags, " ", ", ", -1)
}
