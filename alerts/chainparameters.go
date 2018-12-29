package alerts

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/dingding"
	"io/ioutil"
	"net/http"
	"strings"
)

// Monitor chain parameters change
type ChainParameters struct {
	MonitorUrl      string
	ChainParameters map[string]*ChainParameter
}

type ChainParameter struct {
	OldValue    int64
	NewValue    int64
	HasOldValue bool
}

func (c *ChainParameters) saveChainParameters(k string, v int64) {
	if c.ChainParameters == nil {
		c.ChainParameters = make(map[string]*ChainParameter)
	}

	if _, ok := c.ChainParameters[k]; ok {
		c.ChainParameters[k].OldValue = c.ChainParameters[k].NewValue
		c.ChainParameters[k].NewValue = v
		c.ChainParameters[k].HasOldValue = true
	} else {
		c.ChainParameters[k] = &ChainParameter{
			NewValue:    v,
			HasOldValue: false,
		}
	}
}

func (c *ChainParameters) RequestData() {
	if len(c.MonitorUrl) == 0 || strings.EqualFold(c.MonitorUrl, "") {
		return
	}

	resp, err := http.Get(c.MonitorUrl)
	if err != nil {
		logs.Error("request chain parameters error:", err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("request chain parameters error:", err.Error())
		return
	}

	var origin interface{}

	err = json.Unmarshal(body, &origin)
	if err != nil {
		logs.Error("request chain parameters error:", err.Error())
		return
	}

	v := origin.(map[string]interface{})
	for k, vv := range v {
		logs.Info(k, vv)
		vvv := vv.([]interface{})

		for _, vvvv := range vvv {
			vvvvv := vvvv.(map[string]interface{})

			key := ""
			var value int64 = 0
			for kkkkkk, vvvvvv := range vvvvv {
				if strings.EqualFold(kkkkkk, "key") {
					key = vvvvvv.(string)
				} else if strings.EqualFold(kkkkkk, "value") {
					value = int64(vvvvvv.(float64))
				}
			}

			c.saveChainParameters(key, value)
		}
	}
}

func (c *ChainParameters) Judge() {
	if c.ChainParameters == nil {
		return
	}

	for key, value := range c.ChainParameters {
		if value.HasOldValue {
			if value.OldValue != value.NewValue {
				msg := fmt.Sprintf(
					"提议生效: %s: [%d] -> [%d]",
					key,
					value.OldValue,
					value.NewValue)

				bodyContent := fmt.Sprintf(`
			{
				"msgtype": "text",
				"text": {
					"content": "%s"
				}
			}
			`, msg)

				dingding.DingAlarm.Alarm([]byte(bodyContent))
			}
		}
	}
}
