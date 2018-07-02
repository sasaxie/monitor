package controllers

import (
	"bytes"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

var mutexPing sync.Mutex
var swgp sync.WaitGroup
var ipAddresses []string

func Timer() {
	urlString := beego.AppConfig.String("dingdingURl")
	settings := models.ServersConfig.GetSettings()
	for _, value := range settings {
		if value.IsOpenMonitor {
			//ipAddresses = models.ServersConfig.GetAddressStringByTag(value.Tag)
			ipAddresses = append(ipAddresses, models.ServersConfig.GetAddressStringByTag(value.Tag)...)

		}
	}

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
		TestPost(pingNeedMessage.String(), urlString)
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

func TestPost(content string, url string) {
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
