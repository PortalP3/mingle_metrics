package config

import (
	"encoding/json"
	"log"
	"os"
)

type SystemConfiguration struct {
	Login     string
	Secret    string
	Endpoint  string
	ProjectID string
}

func Load(file string) (config SystemConfiguration, err error) {
	f, err := os.Open(file)
	defer f.Close()

	if os.IsNotExist(err) {
		log.Fatal("Configuration not set. Use \"config\" command to configure the application.")
	}

	if err == nil {
		err = json.NewDecoder(f).Decode(&config)
	}
	return
}
