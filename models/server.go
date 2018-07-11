package models

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var ServersConfig = new(Servers)

const ServerFilePath = "ServerFile"

func InitServerConfig() {
	path := beego.AppConfig.String(ServerFilePath)
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		log.Fatalln("init server config error:", err.Error())
	}

	ServersConfig.Load(file)
}

type Servers struct {
	Servers []*Server `json:"servers"`
}

type Server struct {
	Setting   *Setting   `json:"setting"`
	Addresses []*Address `json:"addresses"`
}

type Setting struct {
	IsOpenMonitor string `json:"isOpenMonitor"`
	Tag           string `json:"tag"`
}

type Address struct {
	Ip   string `json:"ip"`
	Port int    `json:"port"`
}

func (s *Servers) Load(reader io.Reader) {
	r := bufio.NewReader(reader)

	data, err := ioutil.ReadAll(r)

	if err != nil {
		log.Fatalln("load servers file error:", err.Error())
	}

	err = json.Unmarshal(data, s)

	if err != nil {
		log.Fatalln("load servers file error:", err.Error())
	}
}

func (s *Servers) FlushToFile(writer io.Writer) {
	b, err := json.MarshalIndent(s, "", "  ")

	if err != nil {
		log.Println("flush to file error:", err.Error())
		return
	}

	w := bufio.NewWriter(writer)
	_, err = w.Write(b)

	if err != nil {
		log.Println("flush to file error:", err.Error())
		return
	}

	defer w.Flush()
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

func (s *Servers) GetAllAddresses() []string {
	res := make([]string, 0)

	addresses := make(map[string]bool)
	for _, server := range s.Servers {
		for _, address := range server.Addresses {
			addresses[fmt.Sprintf("%s:%d", address.Ip, address.Port)] = false
		}
	}

	for k, _ := range addresses {
		res = append(res, k)
	}

	return res
}

func (s *Servers) GetAllMonitorAddresses() []string {
	res := make([]string, 0)

	addresses := make(map[string]bool)
	for _, server := range s.Servers {
		if strings.EqualFold("true", server.Setting.IsOpenMonitor) {
			for _, address := range server.Addresses {
				addresses[fmt.Sprintf("%s:%d", address.Ip,
					address.Port)] = false
			}
		}
	}

	for k, _ := range addresses {
		res = append(res, k)
	}

	return res
}
