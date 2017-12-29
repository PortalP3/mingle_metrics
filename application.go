package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "Mingle api client"
	app.Usage = "Get project data from mingle"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		{
			Name:  "cfd",
			Usage: "Get cfd info from mingle",
			Action: func(c *cli.Context) error {
				fmt.Println("not implemented!")
				return nil
			},
		},
	}
	app.Run(os.Args)
}
