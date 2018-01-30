package main

import (
	"encoding/json"
	"encoding/xml"
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

	"github.com/PortalP3/mingle_metrics/config"
	"github.com/imdario/mergo"
	"github.com/pd/apiauth"
	"github.com/urfave/cli"
	"math"
)

func configFile() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to find current user configuration: %v", err)
	}
	configDir := filepath.Join(usr.HomeDir, ".mingle_metrics")
	os.MkdirAll(configDir, 0700)
	return filepath.Join(configDir, url.QueryEscape("config.json"))
}

func saveConfigFile(file string, newConfig config.SystemConfiguration) {
	originalConfig, err := config.Load(file)
	if os.IsNotExist(err) {
		originalConfig = config.SystemConfiguration{}
		err = nil
	}
	if err != nil {
		log.Fatal(err)
	}
	mergo.MergeWithOverwrite(&originalConfig, newConfig)
	fmt.Printf("Saving to config file at %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to open file: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(originalConfig)
}

func setConfig(key string, value string, file string) {
	switch key {
	case "Endpoint":
		saveConfigFile(file, config.SystemConfiguration{Endpoint: value})
	case "Login":
		saveConfigFile(file, config.SystemConfiguration{Login: value})
	case "Secret":
		saveConfigFile(file, config.SystemConfiguration{Secret: value})
	case "ProjectID":
		saveConfigFile(file, config.SystemConfiguration{ProjectID: value})
	default:
		fmt.Println("Unknow configuration")
	}
}

func printCurrentConfig(currentConfig config.SystemConfiguration) {
	numberOfElements := reflect.TypeOf(&currentConfig).Elem().NumField()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Println("Current configuration:")
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
	currentConfig, err := config.Load(configFile())
	if err != nil {
		return "", err
	}
	page := 1
	var resource CardsResource
	DoRequest(currentConfig, page, &resource)
	var lastCardNumber int
	const MAX_PAGE_SIZE = 25
	for math.Mod(float64(len(resource.Cards)), MAX_PAGE_SIZE) == 0 && lastCardNumber < resource.Cards[len(resource.Cards)-1].Number {
		page++
		lastCardNumber = resource.Cards[len(resource.Cards)-1].Number
		DoRequest(currentConfig, page, &resource)
	}
	return printCSV(resource.Cards), nil
}

func DoRequest(currentConfig config.SystemConfiguration, page int, lastPage *CardsResource) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%v/api/v2/projects/%v/cards.xml?page=%v&sort=number&order=ASC", currentConfig.Endpoint, currentConfig.ProjectID, page), nil)
	req.Header.Set("Date", apiauth.Date())
	err := apiauth.Sign(req, currentConfig.Login, currentConfig.Secret)
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
	err = xml.Unmarshal(body, lastPage)
	if err != nil {
		log.Fatal(err)
	}
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

var version = "master"

func main() {

	app := cli.NewApp()
	app.Name = "Mingle api client"
	app.Usage = "Get your project data from mingle"
	app.Version = version
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
				file := configFile()
				currentConfig, err := config.Load(file)
				if os.IsNotExist(err) {
					printCurrentConfig(config.SystemConfiguration{})
				}
				if err != nil {
					log.Fatal(err)
				}
				if len(c.Args()) < 2 {
					printCurrentConfig(currentConfig)
				} else {
					setConfig(c.Args().Get(0), c.Args().Get(1), file)
					currentConfig, _ := config.Load(file)
					printCurrentConfig(currentConfig)
				}
				return nil
			},
		},
	}
	app.Run(os.Args)
}
