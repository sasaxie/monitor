package config

import (
	"github.com/BurntSushi/toml"
	"log"
)

const configFilePath = "conf/monitor.toml"

var MonitorConfig Config

type Config struct {
	Node     Node
	InfluxDB InfluxDB
}

type Node struct {
	DataFile string
}

type InfluxDB struct {
	Address  string
	Username string
	Password string
	Database string
}

func init() {
	if _, err := toml.DecodeFile(
		configFilePath,
		&MonitorConfig); err != nil {
		log.Fatal(err)
	}
}
