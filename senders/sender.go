package senders

import "github.com/sasaxie/monitor/senders/message"

type Sender interface {
	Send(msgType message.Type, content string) ([]byte, error)
}
