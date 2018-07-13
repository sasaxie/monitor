package controllers

import "github.com/sasaxie/monitor/models"

type ProgramController struct {
	BaseController
}

// @Title Get program info
// @Description get program info
// @router / [get,post]
func (p *ProgramController) Program() {
	p.Data["json"] = models.Program.Runtime.Unix()

	p.ServeJSON()
}
