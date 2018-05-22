package config

import (
  "io/ioutil"
  "log"

  "github.com/BurntSushi/toml"
)

type GutenConfig struct {
  Email emailConfig
  Storage storageConfig
  Db dbConfig
}

type emailConfig struct {
  SendgridApiKey string `toml:"sendgrid_api_key"`
  FromEmail string `toml:"from_email"`
}

type storageConfig struct {
  ProjectId string `toml:"project_id"`
  BucketName string `toml:"bucket_name"`
}

type dbConfig struct {
  ConnectionStr string `toml:"connection_str"`
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
