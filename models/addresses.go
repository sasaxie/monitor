package models

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const FilePath = "/data/monitor/conf/servers_prod2.json"

var Sc = new(ServerConfig)

func init() {
	Sc.Load(FilePath)
}

type ServerConfig struct {
	Addresses []*Address `json:"addresses"`
}

type Address struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
	Type string `json:"type"`
}

func (s *ServerConfig) Load(filePath string) {
	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		log.Fatalln("Initialization server config error: ", err.Error())
	}

	r := bufio.NewReader(file)

	data, err := ioutil.ReadAll(r)

	if err != nil {
		log.Fatalln("Initialization server config error: ", err.Error())
	}

	err = json.Unmarshal(data, s)

	if err != nil {
		log.Fatalln("Initialization server config error: ", err.Error())
	}
}

func (s *ServerConfig) String() string {

	b, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		log.Fatalln("Server config String() error:", err.Error())
	}

	return string(b)
}
