package main

import (
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/controllers"
	"github.com/sasaxie/monitor/models"
	_ "github.com/sasaxie/monitor/routers"
	"github.com/sasaxie/monitor/service"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	models.InitServerConfig()
	service.InitGrpcClients()
	controllers.InitResponseMap()

	controllers.StartMonitor()

	controllers.GetNodesTask()

	beego.Run()
}
