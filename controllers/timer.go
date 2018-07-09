package controllers

import (
	"bytes"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/common/hexutil"
	"github.com/sasaxie/monitor/core"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

var mutexPing sync.Mutex
var mutexPingMonitor sync.Mutex
var swgp sync.WaitGroup

var missBlockMap = make(map[string]int64)

var PingMonitor map[string][]int64

type AddressMonitor struct {
	Count         int64
	StartPostTime time.Time
}

var urlString string

func init() {
	urlString = beego.AppConfig.String("dingdingURl")
}

var addressMonitorMap map[string]*AddressMonitor

func init() {
	addressMonitorMap = make(map[string]*AddressMonitor)
	PingMonitor = make(map[string][]int64)
}

func Timer() {
	ipAddresses := models.ServersConfig.GetAllMonitorAddresses()

	pingMessage := make(map[string]int64)
	pingTimeoutMessage := make(PingMsg)
	pingRecoverMessage := make(PingMsg)

	for _, s := range ipAddresses {
		swgp.Add(1)
		go GetPingMessage(s, pingMessage, &swgp)
	}
	swgp.Wait()

	for address, ping := range pingMessage {
		if ping <= 0 {
			if v, ok := addressMonitorMap[address]; ok {
				v.Count = v.Count + 1
				mutexPingMonitor.Lock()
				addressMonitorMap[address] = v
				mutexPingMonitor.Unlock()
			} else {
				addressMonitor := new(AddressMonitor)
				addressMonitor.Count = 1
				mutexPingMonitor.Lock()
				addressMonitorMap[address] = addressMonitor
				mutexPingMonitor.Unlock()
			}
		}
	}

	// 记录30次ping值
	for address, ping := range pingMessage {
		if pingSlice, ok := PingMonitor[address]; ok {
			pingSlice = append(pingSlice, ping)

			if len(pingSlice) > 30 {
				pingSlice = pingSlice[len(pingSlice)-30:]
			}

			PingMonitor[address] = pingSlice
		} else {
			newPingSlice := make([]int64, 0)
			newPingSlice = append(newPingSlice, ping)
			PingMonitor[address] = newPingSlice
		}
	}

	// address 没有遍历到的直接移除，如果次数>=3的，还提示恢复信息并从map中移除
	addressMonitorMapTmp := make(map[string]*AddressMonitor)
	for k, v := range addressMonitorMap {
		addressMonitorMapTmp[k] = v
	}

	for k, v := range pingMessage {
		if v <= 0 {
			delete(addressMonitorMapTmp, k)
		}
	}

	for k, v := range addressMonitorMapTmp {
		delete(addressMonitorMap, k)

		if v.Count >= 3 {
			pingRecoverMessage[k] = "gRPC接口已恢复正常"
		}
	}

	// 如果次数>=3，并且时间不足1小时，则发送报警，并重置时间为当前时间
	for k, v := range addressMonitorMap {
		if (v.Count >= 3) && (time.Now().UTC().Unix()-v.StartPostTime.UTC().
			Unix() >= 3600) {
			pingTimeoutMessage[k] = fmt.Sprintf("gRPC接口连续%d次超时(>5000ms)",
				v.Count)
			addressMonitorMap[k].StartPostTime = time.Now()
		}
	}

	if len(pingTimeoutMessage) > 0 {
		bodyContent := fmt.Sprintf(`
		{
			"msgtype": "text",
			"text": {
				"content": "%s"
			}
		}
		`, pingTimeoutMessage.String())

		PostDingding(bodyContent, urlString)
	}

	if len(pingRecoverMessage) > 0 {
		bodyContent := fmt.Sprintf(`
		{
			"msgtype": "text",
			"text": {
				"content": "%s"
			}
		}
		`, pingRecoverMessage.String())

		PostDingding(bodyContent, urlString)
	}

	// 判断超级节点不出块
	witnessMissMessage := make(map[string]*core.Witness)
	for k, v := range pingMessage {
		if v > 0 {
			witnesses := service.GrpcClients[k].ListWitnesses()

			if witnesses != nil {
				for _, witness := range witnesses.Witnesses {
					if witness.IsJobs {
						key := hexutil.Encode(witness.Address)
						if oldMissCount, ok := missBlockMap[key]; ok {
							currentMissCount := witness.TotalMissed

							if currentMissCount > oldMissCount {
								witnessMissMessage[key] = witness
							}

							missBlockMap[key] = witness.TotalMissed
						} else {
							missBlockMap[key] = witness.TotalMissed
						}
					}
				}
			}

			break
		}
	}

	if len(witnessMissMessage) > 0 {
		content := ""
		for _, v := range witnessMissMessage {
			content += fmt.Sprintf("[url：%s，当前的totalMissed"+
				"：%d] ", v.Url, v.TotalMissed)
		}

		bodyContent := fmt.Sprintf(`
		{
			"msgtype": "text",
			"text": {
				"content": "超级节点不出块了，一直警告直到恢复正常：%s"
			}
		}
		`, content)

		PostDingding(bodyContent, urlString)
	}
}

func GetPingMessage(address string, pingMessage map[string]int64,
	wg *sync.WaitGroup) {
	defer wg.Done()

	client := service.NewGrpcClient(address)
	client.Start()
	defer client.Conn.Close()

	mutexPing.Lock()
	pingMessage[address] = client.GetPing()
	mutexPing.Unlock()
}

type PingMsg map[string]string

func (p PingMsg) String() string {
	res := ""

	for k, v := range p {
		res += fmt.Sprintf("address: %s, message: %s\n", k, v)
	}

	return res
}

func PostDingding(content string, url string) {
	postBody := []byte(content)

	header := make(map[string]string)

	header["Content-Type"] = "application/json"

	body, err := Post(postBody, url, header)

	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println(string(body))
}

func Post(postBody []byte, u string, header map[string]string) ([]byte, error) {
	client := &http.Client{}
	request, err := http.NewRequest("POST", u, bytes.NewBuffer(postBody))

	if err != nil {
		log.Println("post dingding error:", err.Error())
		return []byte(""), err
	}

	for key, value := range header {
		request.Header.Set(key, value)
	}

	response, err := client.Do(request)
	if err != nil {
		log.Println("post dingding error:", err.Error())
		return []byte(""), err
	}

	defer response.Body.Close()

	if err != nil {
		return []byte(""), err
	}

	data, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return []byte(""), err
	}

	return data, nil
}

func StartMonitorPing() {
	StartTicker(time.Minute, func(now time.Time) {
		fmt.Println("start monitor:", now)
		Timer()
	})
}

func StartTicker(interval time.Duration, f func(now time.Time)) chan bool {
	done := make(chan bool, 1)

	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case now := <-t.C:
				f(now)
			case <-done:
				return
			}
		}
	}()

	return done
}
