package main

import (
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/controllers"
	_ "github.com/sasaxie/monitor/routers"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	controllers.StartMonitorPing()

	beego.Run()
}
