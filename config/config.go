package config

import (
	"encoding/json"
	"log"
	"os"
)

// Define the config struct
type config struct {
	Port                string   `json:"port"`
	HealthCheckInterval string   `json:"healthCheckInterval"`
	Servers             []string `json:"servers"`
}

// Define the config loader function
func LoadConfig() config {
	var config config
	res, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error while read config file : %s", err.Error())
	}
	err = json.Unmarshal(res, &config)
	if err != nil {
		log.Fatalf("Error while convert json to struct: %s", err.Error())
	}
	return config
}
