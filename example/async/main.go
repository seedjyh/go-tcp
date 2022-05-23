// 这里是使用go-tcp的一个服务器的例子。
package main

import (
	"context"
	"fmt"
	"github.com/seedjyh/go-tcp/pkg/tcp"
	"golang.org/x/sync/errgroup"
	"sync"
)

// 每5个字节一个包
func mySplitter(buf []byte) (tcp.Serializable, int, error) {
	if len(buf) < 5 {
		return nil, 0, tcp.NoEnoughData
	}
	return tcp.NewPacket(buf[:5]), 5, nil
}

func main() {

	// 这里创建了一个服务器。监听 port 端口，接受任何连接（可以用telnet连接）。

	// 指定端口
	port := 11223
	inSiteChannel := make(chan *tcp.Envelope)                          // 收到的消息
	outSiteChannelMap := make(map[tcp.ConnectionID]chan *tcp.Envelope) // 要发送的消息
	outSiteChannelMapMutex := sync.RWMutex{}

	s := tcp.NewServer()

	// 设置分包规则：每5个字节一个包。
	s.SetSplitter(mySplitter)

	s.SetOnConnected(func(connectionID tcp.ConnectionID) (outSiteMessageBus <-chan *tcp.Envelope) {
		outSiteChannelMapMutex.Lock()
		defer outSiteChannelMapMutex.Unlock()
		if ch, ok := outSiteChannelMap[connectionID]; ok {
			return ch
		} else {
			ch := make(chan *tcp.Envelope)
			outSiteChannelMap[connectionID] = ch
			return ch
		}
	})
	s.SetOnDisconnected(func(connectionID tcp.ConnectionID) {
		outSiteChannelMapMutex.Lock()
		defer outSiteChannelMapMutex.Unlock()
		delete(outSiteChannelMap, connectionID)
	})

	s.SetDefaultHandler(func(c tcp.Context) error {
		inSiteChannel <- c.Received()
		return nil
	})

	// start
	eg, ctx := errgroup.WithContext(context.Background())
	eg.Go(func() error { return s.Start(fmt.Sprintf("0.0.0.0:%d", port)) })

	// wait
	<-ctx.Done()

	// stop all here
	fmt.Println("end err:", eg.Wait())
}
