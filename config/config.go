package config

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server     string
	Database   string
	Sessionkey string
}

// Read and parse the configuration file
func (c *Config) Read() {
	if _, err := toml.DecodeFile("config.toml", &c); err != nil {
		log.Fatal(err)
	}
}
