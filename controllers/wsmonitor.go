package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"net/http"
	"sync"
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

					addresses := models.ServersConfig.GetAddressStringByTag(tag)

					for _, address := range addresses {
						waitGroup.Add(1)
						go getResult(address, response)
					}

					waitGroup.Wait()

					for _, tableData := range response.Data {
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

var waitGroup sync.WaitGroup
var mutex sync.Mutex

func getResult(address string, response *models.Response) {
	defer waitGroup.Done()

	var wg sync.WaitGroup
	tableData := new(models.TableData)
	tableData.Address = address

	mutex.Lock()
	client := service.GrpcClients[address]
	mutex.Unlock()

	if client != nil {
		wg.Add(1)
		go client.GetNowBlock(&tableData.NowBlockNum, &tableData.NowBlockHash, &wg)

		wg.Add(1)
		go client.GetLastSolidityBlockNum(&tableData.LastSolidityBlockNum, &wg)

		wg.Add(1)
		go GetPing(client, &tableData.GRPC, &wg)

		wg.Add(1)
		go client.TotalTransaction(&tableData.TotalTransaction, &wg)

		wg.Wait()
	}

	mutex.Lock()
	response.Data = append(response.Data, tableData)
	mutex.Unlock()
}

func GetPing(client *service.GrpcClient, ping *int64,
	wg *sync.WaitGroup) {
	defer wg.Done()

	*ping = client.GetPing()
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
