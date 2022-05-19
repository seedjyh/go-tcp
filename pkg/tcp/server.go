package tcp

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"sync"
)

type (
	Server struct {
		listener       *Listener
		splitter       SplitterFunc
		middleware     []MiddlewareFunc // 处理函数包裹器（中间件）
		onConnected    OnConnectedFunc  // 新连接建立时的回调
		routers        []RouterPair
		defaultHandler HandlerFunc // 默认处理函数（没有被任何router访问的）
	}

	IdentifierFunc func(m Message) bool

	// HandlerFunc 是 Packet 处理函数的标准格式。
	HandlerFunc func(c Context) error

	RouterPair struct {
		identifier IdentifierFunc
		handler    HandlerFunc
	}

	// MiddlewareFunc 是 Packet 处理函数中间件的标准格式
	MiddlewareFunc func(next HandlerFunc) HandlerFunc

	// OnConnectedFunc 是新连接建立时的回调函数。要发送的消息写入 outSiteMessageBus。
	OnConnectedFunc func() (outSiteMessageBus <-chan Message)
)

// NewServer 传入参数除了端口 port ，还有消息解析器 parser 、消息处理器 processor 。
// 在 processor.ProcessMessage 没有返回之前，并不会调用第二个ProcessMessage。
func NewServer() *Server {
	s := &Server{
		listener:       nil,
		splitter:       nil,
		middleware:     nil,
		onConnected:    nil,
		routers:        nil,
		defaultHandler: nil,
	}
	s.splitter = DefaultSplitter
	s.onConnected = DefaultOnConnected
	s.defaultHandler = DefaultHandler
	return s
}

// DefaultOnConnected 没有出站消息。
func DefaultOnConnected() (outSiteMessageBus <-chan Message) {
	return nil
}

// DefaultHandler 默认处理函数
func DefaultHandler(c Context) error {
	return errors.New("unknown message")
}

func (s *Server) SetSplitter(splitter SplitterFunc) {
	s.splitter = splitter
}

func (s *Server) SetOnConnected(onConnected OnConnectedFunc) {
	s.onConnected = onConnected
}

func (s *Server) SetDefaultHandler(handler HandlerFunc) {
	s.defaultHandler = handler
}

func (s *Server) Use(middlewares ...MiddlewareFunc) {
	for _, m := range middlewares {
		s.middleware = append(s.middleware, m)
	}
}

func (s *Server) Add(identifier IdentifierFunc, handler HandlerFunc) {
	s.routers = append(s.routers, RouterPair{
		identifier: identifier,
		handler:    handler,
	})
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
		fmt.Println("tcp.server.Start took a connection from listener:", conn)

		h := func(c Context) error {
			m := c.Received()
			for _, r := range s.routers {
				if r.identifier(m) {
					return r.handler(c)
				}
			}
			return s.defaultHandler(c)
		}
		for _, m := range s.middleware {
			h = m(h)
		}

		daemon := NewDaemon(conn, s.splitter, h, s.onConnected)

		wg.Add(1)
		go func() {
			wg.Done()
			if err := daemon.KeepWorking(ctx); err != nil {
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
