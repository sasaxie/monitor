package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"
	"github.com/sasaxie/monitor/models"
	"net/http"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{}

// Operations about wsmonitor
type WsMonitorController struct {
	beego.Controller
}

// @Title web socket
// @Description get web socket connection
// @router /tag/:tag [get]
func (w *WsMonitorController) Ws() {
	tag := w.GetString(":tag")

	if tag == "" && len(tag) == 0 {
		w.Data["json"] = "not found tag"
		w.ServeJSON()
	} else {
		// Upgrade from http request to WebSocket.
		c, err := upgrader.Upgrade(w.Ctx.ResponseWriter, w.Ctx.Request, nil)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w.Ctx.ResponseWriter, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			beego.Error("Cannot setup WebSocket connection:", err)
			return
		}

		defer c.Close()

		addresses := models.ServersConfig.GetAddressStringByTag(tag)

		for {
			response := new(models.Response)
			response.Data = make([]*models.TableData, 0)

			for _, address := range addresses {
				waitGroup.Add(1)
				go getResult(address, response)
			}

			waitGroup.Wait()

			for _, tableData := range response.Data {
				tableData.GRPCMonitor = ""

				if pings, ok := PingMonitor[tableData.Address]; ok {
					for index, ping := range pings {
						tableData.GRPCMonitor += strconv.Itoa(int(ping))

						if index != len(pings)-1 {
							tableData.GRPCMonitor += ","
						}
					}
				}

				if tableData.LastSolidityBlockNum == 0 {
					tableData.Message = "timeout"
				} else {
					tableData.Message = "success"
				}
			}

			b, err := json.Marshal(response)

			if err != nil {
				log.Error(err.Error())
				continue
			}

			err = c.WriteMessage(websocket.TextMessage, b)

			if err != nil {
				log.Error(err.Error())
				break
			}

			time.Sleep(1 * time.Second)
		}
	}
}
