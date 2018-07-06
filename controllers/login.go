package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"strings"
)

type UserInfoController struct {
	beego.Controller
}

var (
	userNameValidate = beego.AppConfig.String("username")
	passwordValidate = beego.AppConfig.String("password")
)

// @router /info/tag/:UserName/:Password [get,post]
func (c *UserInfoController) Login() {

	fmt.Println(2)
	userName := c.GetString(":UserName", "")
	password := c.GetString(":Password", "")

	fmt.Println(1, userName, password)

	if strings.EqualFold(userNameValidate, userName) && strings.EqualFold(passwordValidate, password) {
		c.SetSession("Validate", userName)
		c.Data["json"] = "success"
	} else {
		c.Data["json"] = "error"
	}

	c.ServeJSON()
}

//登录
func (c *UserInfoController) Logout() {
	//获得id
	//设置返回对象。
	if c.CruSession == nil {
		c.StartSession()
	}
	sessionId := c.CruSession.SessionID()
	logs.Info("==sessionId %s ==", sessionId)
	//设置 SessionDomain 名称。
	c.DestroySession()
	//设置返回对象。
	c.Ctx.Redirect(302, "/auth/login")
	return
}
