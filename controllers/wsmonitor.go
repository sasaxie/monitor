package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"
	"github.com/sasaxie/monitor/models"
	"net/http"
	"time"
)

var responseMap map[string]*models.Responses

func InitResponseMap() {
	responseMap = make(map[string]*models.Responses, 0)

	go func() {
		ticker := time.NewTicker(5 * time.Second)

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
					response.Total = new(models.TotalData)

					addresses := models.ServersConfig.GetAddressStringByTag(tag)

					for _, address := range addresses {
						waitGroup.Add(1)
						go getResult(address, response)
					}

					waitGroup.Wait()
					response.Total.TotalServerErrorNum = 0
					response.Total.TotalServerSuccessNum = 0
					for _, tableData := range response.Data {

						if tableData.LastSolidityBlockNum == 0 {
							tableData.Message = "timeout"
							response.Total.TotalServerErrorNum++
						} else {
							tableData.Message = "success"
							response.Total.TotalServerSuccessNum++

						}
					}
					var totalBlockNum int64 = 0
					var totalBlockHash string = "0"
					var totalLastSolidityBlockNum int64 = 0
					var totalTransaction int64 = 0
					for i, _ := range response.Data {
						if response.Data[i].NowBlockNum > totalBlockNum {
							totalBlockNum = response.Data[i].NowBlockNum
							totalBlockHash = response.Data[i].NowBlockHash
						}
					}

					for i, _ := range response.Data {
						if response.Data[i].LastSolidityBlockNum > totalLastSolidityBlockNum {
							totalLastSolidityBlockNum = response.Data[i].LastSolidityBlockNum
						}
					}
					for i, _ := range response.Data {
						if response.Data[i].TotalTransaction > totalTransaction {
							totalTransaction = response.Data[i].TotalTransaction
						}
					}
					response.Total.TotalBlockNum = totalBlockNum
					response.Total.TotalBlockHash = totalBlockHash
					response.Total.TotalSolidityBlockNum = totalLastSolidityBlockNum
					response.Total.TotalMaxTransaction = totalTransaction

					response.Total.TotalServerNum = len(addresses)
					fmt.Println(response.Total.TotalServerNum)
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

			if v, ok := responseMap[tag]; ok {
				v.Reduce()
			}

			tag = string(p)

			if v, ok := responseMap[tag]; ok {
				v.Increase()
			}

			if _, ok := responseMap[tag]; !ok {
				continue
			}

			response := responseMap[tag].Response

			b, err := json.Marshal(response)

			if err != nil {
				continue
			}

			msgChan <- b
		}
	}()

	go func(msgChan chan []byte) {
		for {
			if _, ok := responseMap[tag]; !ok {
				continue
			}

			response := responseMap[tag].Response

			b, err := json.Marshal(response)

			if err != nil {
				continue
			}

			msgChan <- b

			time.Sleep(5 * time.Second)
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

func Leave(conn *websocket.Conn, tag string) {
	log.Info("close ws")
	if conn != nil {
		conn.Close()
		if v, ok := responseMap[tag]; ok {
			v.Reduce()
		}
	}
}
