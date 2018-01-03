package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"reflect"

	"text/tabwriter"

	"io/ioutil"

	"bytes"

	"github.com/imdario/mergo"
	"github.com/pd/apiauth"
	"github.com/urfave/cli"
)

type SystemConfiguration struct {
	Login     string
	Secret    string
	Endpoint  string
	ProjectID string
}

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
	fmt.Printf("Saving to config file at %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to open file: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(originalConfig)
}

func readConfigFile(file string) (config SystemConfiguration, err error) {
	f, err := os.Open(file)
	defer f.Close()

	if os.IsNotExist(err) {
		return config, nil
	}
	if err == nil {
		err = json.NewDecoder(f).Decode(&config)
	}
	return
}

func setConfig(key string, value string) {
	file, err := configFile()
	if err != nil {
		log.Fatal(err)
	}
	switch key {
	case "Endpoint":
		saveConfigFile(file, SystemConfiguration{Endpoint: value})
	case "Login":
		saveConfigFile(file, SystemConfiguration{Login: value})
	case "Secret":
		saveConfigFile(file, SystemConfiguration{Secret: value})
	case "ProjectID":
		saveConfigFile(file, SystemConfiguration{ProjectID: value})
	default:
		fmt.Println("Unknow configuration")
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
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Println("Current configuration:\n")
	for i := 0; i < numberOfElements; i++ {
		f := reflect.TypeOf(&currentConfig).Elem().FieldByIndex([]int{i})
		v := reflect.Indirect(reflect.ValueOf(&currentConfig)).FieldByName(f.Name)
		fmt.Fprintf(w, "%v:\t%v\n", f.Name, v)
	}
	w.Flush()
}

type Property struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

func (p Property) String() string {
	return p.Value
}

type Card struct {
	Name       string     `xml:"name"`
	Type       string     `xml:"card_type>name"`
	Number     int        `xml:"number"`
	Properties []Property `xml:"properties>property"`
}

type CardsResource struct {
	Cards []Card `xml:"card"`
}

func getMingleCFD() (cfd string, err error) {
	file, _ := configFile()
	config, _ := readConfigFile(file)
	req, _ := http.NewRequest("GET", fmt.Sprintf("%v/api/v2/projects/%v/cards.xml", config.Endpoint, config.ProjectID), nil)
	req.Header.Set("Date", apiauth.Date())
	err = apiauth.Sign(req, config.Login, config.Secret)
	if err != nil {
		log.Fatal(err)
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var resource CardsResource
	err = xml.Unmarshal(body, &resource)
	if err != nil {
		log.Fatal(err)
	}
	cfd = printCSV(resource.Cards)
	return
}

func printCSV(cards []Card) (cfd string) {
	var buffer bytes.Buffer
	buffer.WriteString("Number;Name;Type;Status;Moved to Backlog on;Moved to In Analysis on;Moved to Ready for Dev on;Moved to In Dev on;Moved to Ready for Prod on;Moved to Done on\n")
	for i := range cards {
		card := cards[i]
		buffer.WriteString(fmt.Sprintf("%v;%v;%v", card.Number, card.Name, card.Type))
		for i := range card.Properties {
			if card.Properties[i].Name == "Status" {
				buffer.WriteString(fmt.Sprintf(";%v", strings.Trim(card.Properties[i].Value, " ")))
			}
		}
		for i := range card.Properties {
			if strings.HasPrefix(card.Properties[i].Name, "Moved to") {
				buffer.WriteString(fmt.Sprintf(";%v", strings.Trim(card.Properties[i].Value, " ")))
			}
		}
		buffer.WriteString(fmt.Sprint("\n"))
	}
	return buffer.String()
}

func updateSpreadsheetCFD(cfd string) (err error) {
	return errors.New("Not Implemented")
}

func main() {

	app := cli.NewApp()
	app.Name = "Mingle api client"
	app.Usage = "Get your project data from mingle"
	app.Version = "0.0.3"
	app.Commands = []cli.Command{
		{
			Name:  "cfd",
			Usage: "Get cfd info from mingle",
			Action: func(c *cli.Context) error {
				cfd, err := getMingleCFD()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(cfd)
				return nil
			},
		},
		{
			Name:  "config",
			Usage: "Helps you set configuration values",
			Description: "This application needs your user data to access your mingle server." +
				" To do so, you need to set your security and instance attributes. Currently we need you to set:\n\n" +
				"\t\t\tLogin - This is your login name at mingle. 'access_key_id' at your .csv file downloaded from " +
				"HMAC Auth Key tab on your profile.\n\n" +
				"\t\t\tSecret - This is the secret key. 'secret_access_key' at your .csv file downloaded from " +
				"HMAC Auth Key tab on your profile.\n\n" +
				"\t\t\tEndpoint - Url of your mingle instance. Usually in the form \"https://instance_name.company_name.com\"\n\n" +
				"\t\t\tProjectID - The project identifier that you gave when you created your project. You can find " +
				"this information at Project admin -> Project Settings -> Basic information.",
			ArgsUsage: "Login your_value_here",
			Action: func(c *cli.Context) error {
				if len(c.Args()) < 2 {
					printCurrentConfig()
				} else {
					setConfig(c.Args().Get(0), c.Args().Get(1))
					printCurrentConfig()
				}
				return nil
			},
		},
	}
	app.Run(os.Args)
}
