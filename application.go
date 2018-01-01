package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"reflect"

	"text/tabwriter"

	"github.com/imdario/mergo"
	"github.com/urfave/cli"
)

type SystemConfiguration struct {
	MingleLogin          string
	MingleSecretAcessKey string
	MingleAPIEndpoint    string
	MingleProjectId      string
}

var defaultClient = &http.Client{}

func configFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(usr.HomeDir, ".mingle_metrics")
	os.MkdirAll(configDir, 0700)
	return filepath.Join(configDir,
		url.QueryEscape("config.json")), nil
}

func saveConfigFile(file string, config SystemConfiguration) {
	originalConfig, err := readConfigFile(file)
	if err != nil {
		log.Fatal(err)
	}
	mergo.MergeWithOverwrite(&originalConfig, config)
	fmt.Printf("Saving config file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to open file: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(originalConfig)
}

func readConfigFile(file string) (SystemConfiguration, error) {
	f, err := os.Open(file)
	if err != nil {
		return SystemConfiguration{}, err
	}
	var config SystemConfiguration
	err = json.NewDecoder(f).Decode(&config)
	defer f.Close()
	return config, err
}

func setConfig(key string, value string) {
	file, err := configFile()
	if err != nil {
		log.Fatal(err)
	}
	switch key {
	case "endpoint":
		saveConfigFile(file, SystemConfiguration{MingleAPIEndpoint: value})
	case "login":
		saveConfigFile(file, SystemConfiguration{MingleLogin: value})
	case "secret":
		saveConfigFile(file, SystemConfiguration{MingleSecretAcessKey: value})
	case "project":
		saveConfigFile(file, SystemConfiguration{MingleProjectId: value})
	default:
		fmt.Printf("Unknow configuration")
	}
}

func printCurrentConfig() {
	file, err := configFile()
	if err != nil {
		log.Fatal(err)
	}
	currentConfig, err := readConfigFile(file)
	if err != nil {
		log.Fatal(err)
	}
	numberOfElements := reflect.TypeOf(&currentConfig).Elem().NumField()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
	for i := 0; i < numberOfElements; i++ {
		f := reflect.TypeOf(&currentConfig).Elem().FieldByIndex([]int{i})
		v := reflect.Indirect(reflect.ValueOf(&currentConfig)).FieldByName(f.Name)
		fmt.Fprintf(w, "%v:\t%v\n", f.Name, v)
	}
	w.Flush()
}

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
	app.Version = "0.0.2"
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
		{
			Name:  "config",
			Usage: "Config the application",
			Action: func(c *cli.Context) error {
				firstArg := c.Args().Get(0)
				if firstArg == "set" {
					setConfig(c.Args().Get(1), c.Args().Get(2))
					printCurrentConfig()
				} else {
					printCurrentConfig()
				}
				return nil
			},
		},
	}
	app.Run(os.Args)
}
