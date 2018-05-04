package config

import (
  "log"
  "io/ioutil"

  "github.com/BurntSushi/toml"
)

type GutenConfig struct {
  Email emailConfig
}

type emailConfig struct {
  SendgridApiKey string `toml:"sendgrid_api_key"`
}

func LoadConfig() GutenConfig {
	blob, err := ioutil.ReadFile("cfg/app.toml")
	if err != nil {
		log.Println(err)
  }

  var conf GutenConfig
  if _, err := toml.Decode(string(blob), &conf); err != nil {
    log.Println(err)
  }

	return conf
}
