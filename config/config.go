package config

import (
	"encoding/json"
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

	if err == nil {
		err = json.NewDecoder(f).Decode(&config)
	}
	return
}
