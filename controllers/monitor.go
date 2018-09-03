package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"net/http"
	"time"
)

// Operations about monitor
type MonitorController struct {
	BaseController
}

type wsResponse struct {
	GRPCResponse             map[string]*GRPCs
	WitnessMissBlockResponse map[string]*WitnessMissBlock
}

// @Title monitor data push
// @Description monitor data push
// @router /ws/tag [get]
func (m *MonitorController) Ws() {
	tag := m.GetString(":tag")

	if tag == "" && len(tag) == 0 {
		tag = models.ServersConfig.GetTags()[0]
	}

	// Upgrade from http request to WebSocket.
	c, err := upgrader.Upgrade(m.Ctx.ResponseWriter, m.Ctx.Request, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(m.Ctx.ResponseWriter, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		beego.Error("Cannot setup WebSocket connection:", err)
		return
	}

	defer c.Close()

	msgChan := make(chan []byte, 2)

	go func() {
		for {
			if c == nil {
				return
			}
			_, p, err := c.ReadMessage()
			if err != nil {
				return
			}

			tag = string(p)

			response := new(wsResponse)
			response.GRPCResponse = make(map[string]*GRPCs)
			response.WitnessMissBlockResponse = make(map[string]*WitnessMissBlock)

			addresses := models.ServersConfig.GetAddressStringByTag(tag)

			for _, address := range addresses {
				if gRPC, ok := TronMonitor.GRPCMonitor.
					LatestGRPCs[address]; ok {
					response.GRPCResponse[address] = gRPC
				}
			}

			if models.ServersConfig.IsMonitorByTag(tag) {
				for _, address := range addresses {
					if gRPC, ok := TronMonitor.GRPCMonitor.GRPC[address]; ok {
						if gRPC > 0 {
							if client, ok := service.GrpcClients[address]; ok {
								witnesses := client.ListWitnesses()

								if witnesses != nil {
									if witnesses.Witnesses != nil {
										for _, witness := range witnesses.Witnesses {
											if witness.IsJobs {
												if totalMissed, ok := TronMonitor.WitnessMonitor.
													LatestWitnessMissBlock[witness.Url]; ok {
													response.WitnessMissBlockResponse[witness.Url] = totalMissed
												}
											}
										}

										break
									}
								}
							}
						}
					}
				}
			}

			b, err := json.Marshal(response)

			if err != nil {
				continue
			}

			msgChan <- b
		}
	}()

	go func(msgChan chan []byte) {
		for {

			response := new(wsResponse)
			response.GRPCResponse = make(map[string]*GRPCs)
			response.WitnessMissBlockResponse = make(map[string]*WitnessMissBlock)

			addresses := models.ServersConfig.GetAddressStringByTag(tag)

			for _, address := range addresses {
				if gRPC, ok := TronMonitor.GRPCMonitor.
					LatestGRPCs[address]; ok {
					response.GRPCResponse[address] = gRPC
				}
			}

			if models.ServersConfig.IsMonitorByTag(tag) {
				for _, address := range addresses {
					if gRPC, ok := TronMonitor.GRPCMonitor.GRPC[address]; ok {
						if gRPC > 0 {
							if client, ok := service.GrpcClients[address]; ok {
								witnesses := client.ListWitnesses()

								if witnesses != nil {
									if witnesses.Witnesses != nil {
										for _, witness := range witnesses.Witnesses {
											if witness.IsJobs {
												if totalMissed, ok := TronMonitor.WitnessMonitor.
													LatestWitnessMissBlock[witness.Url]; ok {
													response.WitnessMissBlockResponse[witness.Url] = totalMissed
												}
											}
										}

										break
									}
								}
							}
						}
					}
				}
			}

			b, err := json.Marshal(response)

			if err != nil {
				continue
			}

			msgChan <- b

			time.Sleep(30 * time.Second)
		}
	}(msgChan)

	for {
		msg := <-msgChan
		err = c.WriteMessage(websocket.TextMessage, msg)

		if err != nil {
			beego.Error(err.Error())
			break
		}
	}
}
