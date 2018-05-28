package config

import (
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"
)

type GutenConfig struct {
	Main    appConfig
	Email   emailConfig
	Storage storageConfig
	Db      dbConfig
}

type appConfig struct {
	UrlBase  string `toml:"url_base"`
	HashSalt string `toml:"hash_salt"`
}

type emailConfig struct {
	SendgridApiKey string `toml:"sendgrid_api_key"`
	FromEmail      string `toml:"from_email"`
}

type storageConfig struct {
	ProjectId  string `toml:"project_id"`
	BucketName string `toml:"bucket_name"`
}

type dbConfig struct {
	User           string `toml:"user"`
	Password       string `toml:"password"`
	ConnectionName string `toml:"connection_name"`
	DbName         string `toml:"db_name"`
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
