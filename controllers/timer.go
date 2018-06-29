package controllers

import (
	"bytes"
	"fmt"
	"github.com/sasaxie/monitor/service"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

var mutexPing sync.Mutex
var swgp sync.WaitGroup

func Timer() {

	ipAddresses := []string{
		"18.196.99.16:50051",
		"18.195.254.44:50051",
		"18.196.78.56:50051",
		"54.236.37.243:50051",
		"35.169.107.157:50051",
		"34.237.220.206:50051",
		"52.53.189.99:50051",
		"13.57.246.69:50051",
		"13.57.41.129:50051",
		"52.15.93.92:50051",
		"18.217.144.24:50051",
		"13.58.203.73:50051",
		"34.220.77.106:50051",
		"34.220.59.202:50051",
		"34.216.106.30:50051",
		"34.253.187.192:50051",
		"52.16.167.215:50051",
		"34.250.7.238:50051",
		"52.56.56.149:50051",
		"52.56.115.243:50051",
		"18.130.99.124:50051",
		"35.180.51.163:50051",
		"35.180.22.225:50051",
		"52.47.117.230:50051",
		"54.252.224.209:50051",
		"54.252.238.51:50051",
		"13.211.164.189:50051",
		"18.228.15.36:50051",
		"18.231.118.237:50051",
		"54.232.225.66:50051",
		"35.182.37.246:50051",
		"35.183.101.7:50051",
		"13.229.128.108:50051",
		"13.229.135.228:50051",
		"13.124.62.58:50051",
		"13.125.249.129:50051",
		"13.127.47.162:50051",
		"35.154.40.248:50051"}

	pingMessage := make(map[string]int64)
	pingNeedMessage := make(PingMsg)

	for _, s := range ipAddresses {
		swgp.Add(1)
		go GetPingMessage(s, pingMessage, &swgp)
	}
	swgp.Wait()

	for address, _ := range pingMessage {
		if pingMessage[address] <= 0 {
			pingNeedMessage[address] = "timeout(>5000ms)"
		}
	}

	if len(pingNeedMessage) > 0 {
		TestPost(pingNeedMessage.String())
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

func TestPost(content string) {
	bodyContent := fmt.Sprintf(`
		{
			"msgtype": "text",
			"text": {
				"content": "%s"
			}
		}
		`, content)
	postBody := []byte(bodyContent)

	header := make(map[string]string)

	header["Content-Type"] = "application/json"

	body, err := Post(postBody, "https://oapi.dingtalk.com/robot/send?access_token=1a9f984a9ee7b1c59355563e26725c3dd92354160bdc6b7edd788325e04e1416", header)

	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println(string(body))
}

func Post(postBody []byte, u string, header map[string]string) ([]byte, error) {
	client := &http.Client{}
	request, err := http.NewRequest("POST", u, bytes.NewBuffer(postBody))

	for key, value := range header {
		request.Header.Set(key, value)
	}

	response, err := client.Do(request)
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
	StartTicker(10*time.Second, func(now time.Time) {
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
