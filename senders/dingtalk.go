package senders

import (
	"fmt"
	"github.com/sasaxie/monitor/common/http"
	"github.com/sasaxie/monitor/senders/message"
)

type DingTalk struct {
	WebHook string
}

func NewDingTalk(webHook string) *DingTalk {
	dingTalk := new(DingTalk)
	dingTalk.WebHook = webHook

	return dingTalk
}

func (d *DingTalk) Send(msgType message.Type, content string) ([]byte, error) {
	header := make(map[string]string)

	header["Content-Type"] = "application/json"

	body, err := http.Post(d.wrapperMsg(content), d.WebHook, header)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func (d *DingTalk) wrapperMsg(msg string) []byte {
	wrapper := fmt.Sprintf(`
			{
				"msgtype": "text",
				"text": {
					"content": "%s"
				}
			}
			`, msg)

	return []byte(wrapper)
}
