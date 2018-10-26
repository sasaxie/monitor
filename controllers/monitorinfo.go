package controllers

import (
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/models"
)

type MonitorInfoController struct {
	beego.Controller
}

// @router / [get]
func (n *MonitorInfoController) Get() {
	n.Data["json"] = models.GetMonitorInfo()
	n.ServeJSON()
}
