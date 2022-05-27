package tcp

import (
	"context"
	"github.com/pkg/errors"
)

type Forwarder struct {
	outSiteMessageBus     <-chan Serializable
	sendingMessageChannel chan<- Serializable
}

func NewForwarder(outSiteMessageBus <-chan Serializable, sendingMessageChannel chan<- Serializable) *Forwarder {
	return &Forwarder{
		outSiteMessageBus:     outSiteMessageBus,
		sendingMessageChannel: sendingMessageChannel,
	}
}

func (f *Forwarder) KeepWorking(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("context is done")
		case m, ok := <-f.outSiteMessageBus:
			if !ok {
				return errors.New("channel is closed")
			}
			f.sendingMessageChannel <- m
		}
	}
}
