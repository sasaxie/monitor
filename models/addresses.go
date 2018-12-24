package models

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/util"
	"io/ioutil"
	"log"
	"os"
)

var NodeList = new(Nodes)

func init() {
	NodeList.Load(
		fmt.Sprintf("%s/conf/%s",
			util.GetCurrentDirectory(),
			config.MonitorConfig.Node.DataFile))
}

type Nodes struct {
	Addresses []*Address `json:"addresses"`
}

type Address struct {
	Ip       string `json:"ip"`
	GrpcPort int    `json:"grpcPort"`
	HttpPort int    `json:"httpPort"`
	Type     string `json:"type"`
	Tag      string `json:"tag"`
	Monitor  string `json:"monitor"`
}

func (s *Nodes) Load(filePath string) {
	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		log.Fatalln("initialization server config error: ", err.Error())
	}

	r := bufio.NewReader(file)

	data, err := ioutil.ReadAll(r)

	if err != nil {
		log.Fatalln("initialization server config error: ", err.Error())
	}

	err = json.Unmarshal(data, s)

	if err != nil {
		log.Fatalln("initialization server config error: ", err.Error())
	}
}

func (s *Nodes) String() string {

	b, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		log.Fatalln("server config String() error:", err.Error())
	}

	return string(b)
}
