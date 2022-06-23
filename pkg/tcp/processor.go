package tcp

import (
	"context"
	"github.com/pkg/errors"
)

// Processor 是一个由中间件堆砌起来的消息处理栈。
type Processor struct {
	connID                 ConnectionID
	receivedMessageChannel <-chan ReceivedMessage
	sendingMessageChannel  chan<- SendingMessage
	handler                HandlerFunc
}

func NewProcessor(
	connID ConnectionID,
	receivedMessageChannel <-chan ReceivedMessage,
	sendingMessageChannel chan<- SendingMessage,
	handler HandlerFunc,
) *Processor {
	return &Processor{
		connID:                 connID,
		receivedMessageChannel: receivedMessageChannel,
		sendingMessageChannel:  sendingMessageChannel,
		handler:                handler,
	}
}

func (p *Processor) KeepWorking(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("context is done")
		case m, ok := <-p.receivedMessageChannel:
			if !ok {
				return errors.New("channel is closed")
			}
			c := &handleContext{
				connID:                p.connID,
				received:              m,
				sendingMessageChannel: p.sendingMessageChannel,
			}
			if err := p.handler(c); err != nil {
				// fmt.Println("handle failed", err)
			}
		}
	}
}
