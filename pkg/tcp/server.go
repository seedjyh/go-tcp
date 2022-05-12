package tcp

import (
	"context"
	"fmt"
	"sync"
)

type (
	Server struct {
		listener    *Listener
		splitter    SplitterFunc
		middleware  []MiddlewareFunc // 处理函数包裹器（中间件）
		handler     HandlerFunc      // 核心处理函数
		onConnected OnConnectedFunc  // 新连接建立时的回调
	}

	// HandlerFunc 是 Packet 处理函数的标准格式。
	HandlerFunc func(c Context) error

	// MiddlewareFunc 是 Packet 处理函数中间件的标准格式
	MiddlewareFunc func(next HandlerFunc) HandlerFunc

	// OnConnectedFunc 是新连接建立时的回调函数。没有被中间件过滤的消息会送到 inSiteMessageBus，要发送的消息写入 outSiteMessageBus
	OnConnectedFunc func(inSiteMessageBus <-chan *Packet) (outSiteMessageBus <-chan *Packet)
)

// NewServer 传入参数除了端口 port ，还有消息解析器 parser 、消息处理器 processor 。
// 在 processor.ProcessMessage 没有返回之前，并不会调用第二个ProcessMessage。
func NewServer() *Server {
	s := &Server{
		listener:    nil,
		splitter:    nil,
		middleware:  nil,
		onConnected: nil,
	}
	s.splitter = DefaultSplitter
	s.onConnected = DefaultOnConnected
	return s
}

// DefaultOnConnected 所有入站消息都被获取并丢弃；没有出站消息。
func DefaultOnConnected(inSiteMessageBus <-chan *Packet) (outSiteMessageBus <-chan *Packet) {
	go func() {
		for _ = range inSiteMessageBus {
		}
	}()
	return nil
}

func (s *Server) SetOnConnected(onConnected OnConnectedFunc) {
	s.onConnected = onConnected
}

// Start 是一个阻塞式的服务。会一直工作到调用 Stop 为止。
// 收到一个连接，就会启动一个协程去处理该连接。
func (s *Server) Start(address string) error {
	var wg sync.WaitGroup
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.listener = NewListener()
	connChan, err := s.listener.Start(address)
	if err != nil {
		return err
	}

	for conn := range connChan {
		conn := conn
		fmt.Println(conn)
		wg.Add(1)
		go func() {
			wg.Done()
			if err := NewDaemon(conn, s.splitter, s.middleware, s.onConnected).KeepWorking(ctx); err != nil {
				fmt.Println("conn processor exit, error=", err)
			} else {
				fmt.Println("conn processor exit ok")
			}
		}()
	}
	return nil
}

// Stop 仅发送一个停止的信号， Start 需要等关闭所有资源后才返回。
func (s *Server) Stop() error {
	return s.listener.Stop()
}
