package tcp

import (
	"context"
	"golang.org/x/sync/errgroup"
)

// Daemon 负责管理一个 net.Conn 的全生命周期。
// 包括 sender 和 receiver 的协程生命周期，以及 net.Conn 的关闭。
// 这里的设计理念是，不对外界直接提供消息发送接口。外界只有收到消息并被回调处理，才能得到发送接口。
type Daemon struct {
	connection     *Connection
	splitter       SplitterFunc
	handler        HandlerFunc
	onConnected    OnConnectedFunc
	onDisconnected OnDisconnectedFunc
}

func NewDaemon(connection *Connection, splitter SplitterFunc, handler HandlerFunc, onConnected OnConnectedFunc, onDisconnected OnDisconnectedFunc) *Daemon {
	return &Daemon{
		connection:     connection,
		splitter:       splitter,
		handler:        handler,
		onConnected:    onConnected,
		onDisconnected: onDisconnected,
	}
}

// KeepWorking 持续工作，直到出错时退出。
// 不会关闭任何外部传入的资源（如 net.Conn, inSiteMessageBuf, outSiteMessageBus 就不会关闭)
func (d *Daemon) KeepWorking(ctx context.Context) error {
	// 1. 创建两个channel
	receivedMessageChannel := make(chan *Envelope)
	defer close(receivedMessageChannel)
	sendingMessageChannel := make(chan *Envelope)
	defer close(sendingMessageChannel)
	forwardingMessageChannel := d.onConnected(d.connection.connectionID)
	defer d.onDisconnected(d.connection.connectionID)
	// 2. 创建4个goroutine
	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return NewReceiver(d.connection, d.splitter, receivedMessageChannel).KeepWorking(ctx) })
	eg.Go(func() error { return NewSender(d.connection, sendingMessageChannel).KeepWorking(ctx) })
	eg.Go(func() error { return NewForwarder(forwardingMessageChannel, sendingMessageChannel).KeepWorking(ctx) })
	eg.Go(func() error {
		return NewProcessor(receivedMessageChannel, sendingMessageChannel, d.handler).KeepWorking(ctx)
	})
	<-ctx.Done()
	cancel()
	return eg.Wait()
}
