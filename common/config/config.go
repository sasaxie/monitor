package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/util"
	"log"
	"os"
	"time"
)

const configFilePath = "conf/monitor.toml"
const nodesFilePath = "conf/nodes.json"

var MonitorConfig Config

type Config struct {
	AppName  string
	RunMode  string
	Node     Node
	InfluxDB InfluxDB
	Log      Log
	Http     Http
	Task     Task
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

type Log struct {
	Filename string
	Level    string
}

type Http struct {
	Port int
}

type Task struct {
	GetGRPCDataInterval time.Duration
	GetHTTPDataInterval time.Duration
}

func init() {
	configFilePathAbs := fmt.Sprintf("%s%c%s", util.GetCurrentDirectory(), os.PathSeparator, configFilePath)

	if !util.Exists(configFilePathAbs) {
		logs.Warn("create default config")
		createDefaultConfig(configFilePathAbs)
	}

	nodesFilePathAbs := fmt.Sprintf("%s%c%s", util.GetCurrentDirectory(),
		os.PathSeparator, nodesFilePath)

	if !util.Exists(nodesFilePathAbs) {
		logs.Warn("create default nodes file")
		createDefaultNodes(nodesFilePathAbs)
	}

	if _, err := toml.DecodeFile(
		configFilePathAbs,
		&MonitorConfig); err != nil {
		log.Fatal(err)
	}

	initBeego()
	initLog()
}

func initBeego() {
	beego.BConfig.AppName = MonitorConfig.AppName
	beego.BConfig.RunMode = MonitorConfig.RunMode
	beego.BConfig.Listen.HTTPPort = MonitorConfig.Http.Port
}

func initLog() {
	filename := MonitorConfig.Log.Filename
	level := 7
	switch MonitorConfig.Log.Level {
	case "emergency":
		level = 0
	case "alert":
		level = 1
	case "critical":
		level = 2
	case "error":
		level = 3
	case "warning":
		level = 4
	case "notice":
		level = 5
	case "information":
		level = 6
	case "debug":
		level = 7
	default:
		level = 1
	}

	configJson := fmt.Sprintf(`
{
"filename":"%s",
"level":%d
}`, filename, level)

	logs.SetLogger(
		logs.AdapterFile,
		configJson,
	)
	logs.Async()
}

func createDefaultConfig(filename string) {
	d := `
appName = "monitor"

# App run mode:
# 1. dev(default)
# 2. prod
runMode = "dev"

[node]
dataFile = "nodes.json"

[influxdb]
address = "http://localhost:8086"
username = "root"
password = "root"
database = "tronmonitor"

[log]
filename = "monitor.log"

# Log level:
# 1. emergency
# 2. alert
# 3. critical
# 4. error
# 5. warning
# 6. notice
# 7. information
# 8. debug(default)
level = "debug"

[http]
port = 8080

[task]
# gRPC Interval defalut 40 seconds
gRPCInterval = 40
# HTTP Interval defalut 10 seconds
HTTPInterval = 10
`

	util.WriteToFile(filename, d)
}

func createDefaultNodes(filename string) {
	n := `
{
  "addresses": [
    {
      "ip": "172.16.21.39",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "54.236.37.243",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "52.53.189.99",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "18.196.99.16",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "34.253.187.192",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "52.56.56.149",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "35.180.51.163",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "54.252.224.209",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "18.228.15.36",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "52.15.93.92",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "34.220.77.106",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "13.127.47.162",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "13.124.62.58",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "47.74.149.206",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "35.182.37.246",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "47.90.215.84",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "47.254.77.146",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "47.74.242.55",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "47.75.249.119",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "47.90.201.118",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "34.250.140.143",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "35.176.192.130",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "52.47.197.188",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "52.62.210.100",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "13.231.4.243",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "47.254.27.69",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "35.154.90.144",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "13.125.210.234",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "47.88.174.175",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "47.75.249.4",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node"
    },
    {
      "ip": "47.89.187.247",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "solidity_node"
    },
    {
      "ip": "47.91.18.255",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "solidity_node"
    },
    {
      "ip": "47.75.10.71",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "solidity_node"
    },
    {
      "ip": "47.251.52.228",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "solidity_node"
    },
    {
      "ip": "47.251.48.82",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "solidity_node"
    },
    {
      "ip": "47.74.147.80",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "solidity_node"
    },
    {
      "ip": "34.234.164.105",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "solidity_node"
    },
    {
      "ip": "18.221.34.0",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "solidity_node"
    },
    {
      "ip": "35.178.11.0",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "solidity_node"
    },
    {
      "ip": "35.180.18.107",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "solidity_node"
    }
  ]
}
`

	util.WriteToFile(filename, n)
}
