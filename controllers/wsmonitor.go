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

var responseMap map[string]*models.Responses

func InitResponseMap() {
	responseMap = make(map[string]*models.Responses, 0)

	go func() {
		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-ticker.C:
				tags := models.ServersConfig.GetTags()

				for _, tag := range tags {
					var responses *models.Responses

					if v, ok := responseMap[tag]; ok {
						responses = v
						if !v.Runnable() {
							continue
						}
					} else {
						responses = new(models.Responses)
						responses.Count = 0
					}

					response := new(models.Response)

					response.Data = make([]*models.TableData, 0)

					addresses := models.ServersConfig.GetAddressStringByTag(tag)

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

					responses.Response = response
					responseMap[tag] = responses
				}
			}
		}
	}()
}

var upgrader = websocket.Upgrader{}

// Operations about wsmonitor
type WsMonitorController struct {
	beego.Controller
}

// @Title web socket
// @Description get web socket connection
// @router /tag [get]
func (w *WsMonitorController) Ws() {
	tag := w.GetString(":tag")

	if tag == "" && len(tag) == 0 {
		tag = models.ServersConfig.GetTags()[0]
	}

	// Upgrade from http request to WebSocket.
	c, err := upgrader.Upgrade(w.Ctx.ResponseWriter, w.Ctx.Request, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w.Ctx.ResponseWriter, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		beego.Error("Cannot setup WebSocket connection:", err)
		return
	}

	defer Leave(c, tag)

	if v, ok := responseMap[tag]; ok {
		v.Increase()
	}

	for {
		response := responseMap[tag].Response

		b, err := json.Marshal(response)

		if err != nil {
			continue
		}

		err = c.WriteMessage(websocket.TextMessage, b)

		if err != nil {
			log.Error(err.Error())
			break
		}

		time.Sleep(1 * time.Second)

		go func() {
			for {
				if c == nil {
					return
				}
				_, p, err := c.ReadMessage()
				if err != nil {
					return
				}

				if v, ok := responseMap[tag]; ok {
					v.Reduce()
				}

				tag = string(p)

				if v, ok := responseMap[tag]; ok {
					v.Increase()
				}
			}
		}()
	}
}

func Leave(conn *websocket.Conn, tag string) {
	log.Info("close ws")
	if conn != nil {
		conn.Close()
		if v, ok := responseMap[tag]; ok {
			v.Reduce()
		}
	}
}
