package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/config"
	"os"
	"runtime"
	"time"
)

type MonitorInfo struct {
	Version    string
	Runtime    string
	SystemInfo *SystemInfo
	GoInfo     *GoInfo
}

type SystemInfo struct {
	Hostname string
}

type GoInfo struct {
	Version string
	Root    string
	Path    string
	Arch    string
	Os      string
}

func GetMonitorInfo() *MonitorInfo {
	var err error
	monitorInfo := new(MonitorInfo)

	monitorInfo.Version = config.MonitorVersion
	monitorInfo.Runtime = time.Now().Format("2006-01-02 15:04:06")

	monitorInfo.SystemInfo = new(SystemInfo)
	monitorInfo.SystemInfo.Hostname, err = os.Hostname()
	if err != nil {
		logs.Warn(err)
	}

	monitorInfo.GoInfo = new(GoInfo)
	monitorInfo.GoInfo.Version = runtime.Version()
	monitorInfo.GoInfo.Root = runtime.GOROOT()
	monitorInfo.GoInfo.Path = os.Getenv("GOPATH")
	monitorInfo.GoInfo.Arch = runtime.GOARCH
	monitorInfo.GoInfo.Os = runtime.GOOS

	return monitorInfo
}
