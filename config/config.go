package config

import (
	"io/ioutil"

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

func LoadConfig() (conf GutenConfig, err error) {
	blob, err := ioutil.ReadFile("cfg/app.toml")
	if err != nil {
		return
	}

	_, err = toml.Decode(string(blob), &conf)
	if err != nil {
		return
	}

	return
}
