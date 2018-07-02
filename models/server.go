package models

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"io/ioutil"
	"log"
	"strings"
)

var ServersConfig *Servers

func init() {
	ServersConfig = new(Servers)
	ServersConfig.Load()
}

type Servers struct {
	Servers []*Server `json:"servers"`
}

type Server struct {
	Setting   *Setting   `json:"setting"`
	Addresses []*Address `json:"addresses"`
}

type Setting struct {
	IsOpenMonitor bool   `json:"isOpenMonitor"`
	Tag           string `json:"tag"`
}

type Address struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

func (s *Servers) Load() {
	path := beego.AppConfig.String("ServerFile")

	data, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatalln("load servers file error:", err.Error())
	}

	err = json.Unmarshal(data, s)

	if err != nil {
		log.Fatalln("load servers file error:", err.Error())
	}
}

func (s *Servers) String() string {

	b, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		log.Fatalln("servers String() error:", err.Error())
	}
	return string(b)
}

func (s *Servers) GetAddressStringByTag(tag string) []string {
	res := make([]string, 0)

	for _, server := range s.Servers {
		if strings.EqualFold(tag, server.Setting.Tag) {
			for _, address := range server.Addresses {
				addressStr := fmt.Sprintf("%s:%d", address.Ip, address.Port)
				res = append(res, addressStr)
			}
		}
	}

	return res
}

func (s *Servers) GetTags() []string {
	res := make([]string, 0)

	for _, server := range s.Servers {
		res = append(res, server.Setting.Tag)
	}

	return res
}

func (s *Servers) GetSettings() []*Setting {
	res := make([]*Setting, 0)

	for _, server := range s.Servers {
		res = append(res, server.Setting)
	}

	return res
}
