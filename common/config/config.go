package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"log"
)

const configFilePath = "conf/monitor.toml"

var MonitorConfig Config

type Config struct {
	AppName  string
	RunMode  string
	Node     Node
	InfluxDB InfluxDB
	Log      Log
	Http     Http
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

func init() {
	if _, err := toml.DecodeFile(
		configFilePath,
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
