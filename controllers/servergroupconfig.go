package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"os"
	"strings"
)

type ServerGroupConfigController struct {
	BaseController
}

// @Title Get server config
// @Description get server config
// @router / [get]
func (s *ServerGroupConfigController) ServerGroupConfig() {
	s.Data["json"] = models.ServersConfig

	s.ServeJSON()
}

// @Title Get tags
// @Description get tags
// @router /tags [get,post]
func (s *ServerGroupConfigController) Tags() {
	s.Data["json"] = models.ServersConfig.GetTags()
	s.ServeJSON()
}

// @Title Get settings
// @Description get settings
// @router /settings [get,post]
func (s *ServerGroupConfigController) Settings() {
	s.Data["json"] = models.ServersConfig.GetSettings()
	s.ServeJSON()
}

// @Title Get server config
// @Description get server config
// @router / [post]
func (s *ServerGroupConfigController) ServerGroupConfigEdit() {

	server := new(models.Servers)
	err := json.Unmarshal(s.Ctx.Input.RequestBody, server)

	if err != nil {
		s.Data["json"] = err.Error()
	} else {

		if server.Servers == nil {
			s.Data["json"] = "error"
		} else {
			path := beego.AppConfig.String(models.ServerFilePath)
			file, err := os.Create(path)
			defer file.Close()

			if err != nil {
				s.Data["json"] = err.Error()
			} else {
				server.FlushToFile(file)

				oldAddresses := models.ServersConfig.GetAllAddresses()

				models.ServersConfig = server

				newAddresses := models.ServersConfig.GetAllAddresses()

				// 删掉的话，必须close
				closeOldGrpcClients(oldAddresses, newAddresses)

				// 新增的话，只重新初始化新增的
				connNewGrpcClients(oldAddresses, newAddresses)

				service.InitGrpcClients()
				s.Data["json"] = "success"
			}
		}
	}

	s.ServeJSON()
}

func closeOldGrpcClients(oldAddresses, newAddresses []string) {
	closeAddresses := make([]string, 0)

	for _, o := range oldAddresses {
		count := 0
		for _, n := range newAddresses {
			count++
			if strings.EqualFold(o, n) {
				break
			}
		}

		if count == len(newAddresses) {
			closeAddresses = append(closeAddresses, o)
		}
	}

	for _, c := range closeAddresses {
		if service.GrpcClients[c].Conn != nil {
			service.GrpcClients[c].Conn.Close()
		}
	}
}

func connNewGrpcClients(oldAddresses, newAddresses []string) {
	connAddresses := make([]string, 0)

	for _, n := range newAddresses {
		count := 0
		for _, o := range oldAddresses {
			count++
			if strings.EqualFold(o, n) {
				break
			}
		}

		if count == len(oldAddresses) {
			connAddresses = append(connAddresses, n)
		}
	}

	for _, c := range connAddresses {
		service.GrpcClients[c] = service.NewGrpcClient(c)
		service.GrpcClients[c].Start()
	}
}
