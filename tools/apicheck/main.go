package main

import (
	"bytes"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/models"
	"io/ioutil"
	"net/http"
)

func main() {
	result := make(map[string][]string)
	i := 0
	for _, address := range models.NodeList.Addresses {
		addr := fmt.Sprintf("%s:%d", address.Ip, address.HttpPort)

		value := addr

		u := fmt.Sprintf("http://%s/wallet/gettransactionbyid",
			addr)

		fmt.Println(u)

		postBody := []byte("{\"value\":\"36854b9c99f5f1af2360c18b03e7e2863078d550da743d387a3d8eb7ced9d136\"}")
		client := &http.Client{}

		req, err := http.NewRequest(
			"POST",
			u,
			bytes.NewBuffer(postBody))
		if err != nil {
			logs.Warn(err)
			key := "connect failed"
			if _, ok := result[key]; !ok {
				result[key] = make([]string, 0)
			}
			result[key] = append(result[key], value)
			i++
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			logs.Warn(err)
			key := "connect failed"
			if _, ok := result[key]; !ok {
				result[key] = make([]string, 0)
			}
			result[key] = append(result[key], value)
			i++
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logs.Warn(err)
			key := "connect failed"
			if _, ok := result[key]; !ok {
				result[key] = make([]string, 0)
			}
			result[key] = append(result[key], value)
			i++
			continue
		}

		key := string(body)
		fmt.Println(addr, key)
		if _, ok := result[key]; !ok {
			result[key] = make([]string, 0)
		}
		result[key] = append(result[key], value)

		i++
		fmt.Println(i, "/", len(models.NodeList.Addresses))
	}

	for k, v := range result {
		fmt.Println(">>>>>>>>>>>>")
		fmt.Println(k, v)
		fmt.Println(">>>>>>>>>>>>")
	}
}
