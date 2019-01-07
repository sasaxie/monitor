package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/senders"
	"github.com/sasaxie/monitor/util"
	"log"
	"os"
)

const configFilePath = "conf/monitor.toml"
const nodesFilePath = "conf/nodes.json"

var MonitorConfig Config
var Ding *senders.DingTalk

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
	GetDataInterval     int64
	Dingding            string
	ProposalsMonitorUrl string
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
	initSender()
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
# Interval defalut 60 seconds
getDataInterval = 60
dingding = "your DingDing robot url"
proposalsMonitorUrl = "http://54.236.37.243:8090/wallet/getchainparameters"
`

	util.WriteToFile(filename, d)
}

func createDefaultNodes(filename string) {
	n := `
{
  "addresses": [
    {
      "ip": "127.0.0.1",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node",
      "tagName": "本地",
      "monitor": "NowBlock,BlockMissed"
    },
    {
      "ip": "127.0.0.1",
      "grpcPort": 50052,
      "httpPort": 8092,
      "type": "full_node",
      "tagName": "本地",
      "monitor": "NowBlock,BlockMissed"
    },
    {
      "ip": "127.0.0.1",
      "grpcPort": 50053,
      "httpPort": 8094,
      "type": "full_node",
      "tagName": "本地",
      "monitor": "NowBlock,BlockMissed"
    },
    {
      "ip": "127.0.0.1",
      "grpcPort": 50054,
      "httpPort": 8097,
      "type": "solidity_node",
      "tagName": "本地",
      "monitor": "NowBlock,BlockMissed"
    }
  ]
}
`

	util.WriteToFile(filename, n)
}

func initSender() {
	Ding = senders.NewDingTalk(MonitorConfig.Task.Dingding)
	beego.Info("init ding talk, web hook:", Ding.WebHook)
}
