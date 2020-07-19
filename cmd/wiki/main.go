package main

import (
	"fmt"
	"log"
	"os"

	"net/http"
	_ "net/http/pprof"

	_ "github.com/mkevac/debugcharts"
	"github.com/sebnyberg/wikirel/cmd/wiki/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	if err := NewApp().Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

const pkgName string = "wikirel"

func NewApp() *cli.App {
	app := &cli.App{
		Name:     pkgName,
		HelpName: pkgName,
		Usage:    "wiki commands",
		Commands: []*cli.Command{
			cmd.Parse(),
		},
	}

	return app
}
