package tcp

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net"
)

// Daemon 负责管理一个 net.Conn 的全生命周期。
// 包括 sender 和 receiver 的协程生命周期，以及 net.Conn 的关闭。
// 这里的设计理念是，不对外界直接提供消息发送接口。外界只有收到消息并被回调处理，才能得到发送接口。
type Daemon struct {
	conn        net.Conn
	splitter    SplitterFunc
	middleware  []MiddlewareFunc
	onConnected OnConnectedFunc
}

func NewDaemon(conn net.Conn, splitter SplitterFunc, middleware []MiddlewareFunc, onConnected OnConnectedFunc) *Daemon {
	return &Daemon{
		conn:        conn,
		splitter:    splitter,
		middleware:  middleware,
		onConnected: onConnected,
	}
}

// KeepWorking 持续工作，直到出错时退出。
// 不会关闭任何外部传入的资源（如 net.Conn, inSiteMessageBuf, outSiteMessageBus 就不会关闭)
func (d *Daemon) KeepWorking(ctx context.Context) error {
	// 1. 创建两个channel
	receivedMessageChannel := make(chan *Packet)
	defer close(receivedMessageChannel)
	sendingMessageChannel := make(chan *Packet)
	defer close(sendingMessageChannel)
	inSiteMessageBus := make(chan *Packet)
	outSiteMessageBus := d.onConnected(inSiteMessageBus)
	// 2. 创建handler
	h := func(c Context) error {
		m := c.Received()
		fmt.Println("daemon.Handler", m.data[0])
		fmt.Println("pushing inSiteMessageBus", inSiteMessageBus)
		inSiteMessageBus <- m
		fmt.Println("pushed inSiteMessageBus")
		return nil
	}
	if d.middleware != nil {
		for i := len(d.middleware) - 1; i >= 0; i-- {
			h = d.middleware[i](h)
		}
	}
	// 3. 创建4个goroutine
	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return NewReceiver(d.conn, d.splitter, receivedMessageChannel).KeepWorking(ctx) })
	eg.Go(func() error { return NewSender(d.conn, sendingMessageChannel).KeepWorking(ctx) })
	eg.Go(func() error { return NewForwarder(outSiteMessageBus, sendingMessageChannel).KeepWorking(ctx) })
	eg.Go(func() error {
		return NewProcessor(receivedMessageChannel, sendingMessageChannel, h).KeepWorking(ctx)
	})
	<-ctx.Done()
	cancel()
	return eg.Wait()
}
