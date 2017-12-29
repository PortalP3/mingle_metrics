package main

import (
	"errors"
	"log"
	"os"

	"github.com/urfave/cli"
)

func getMingleCFD() (cfd string, err error) {
	return "", errors.New("Not Implemented")
}

func updateSpreadsheetCFD(cfd string) (err error) {
	return errors.New("Not Implemented")
}

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
				cfd, err := getMingleCFD()
				if err != nil {
					log.Fatal(err)
				}
				err = updateSpreadsheetCFD(cfd)
				if err != nil {
					log.Fatal(err)
				}
				return nil
			},
		},
	}
	app.Run(os.Args)
}
