package tcp

import (
	"context"
	"github.com/pkg/errors"
	"github.com/seedjyh/go-tcp/pkg/tcp/uuid"
	"sync"
)

type (
	// Server 会监听连接，并对于每个连接单独处理。
	// 每个连接收到的字节流的处理顺序如下：
	// 1. 根据 SplitterFunc 拆分成 Packet。
	// 2. 通过一系列 MiddlewareFunc。用 Use 注册自定义 MiddlewareFunc，先注册的先执行。
	// 3. 通过一系列 RouterPair。每个 RouterPair 由一个 IdentifierFunc 和一个 HandlerFunc 组成。用 Add 注册自定义 RouterPair ，先注册的先检查。一个匹配后，不再执行后续的 RouterPair。
	// 4. 如果所有 RouterPair 的 IdentifierFunc 都不匹配，则会调用默认的 HandlerFunc。用 SetDefaultHandler 覆盖默认值。
	// 此外，每个连接创建的时候，会回调一个 OnConnectedFunc，用于配置往该连接发送消息的消息 channel。用 SetOnConnected 覆盖默认值。
	//
	// 一般的使用顺序如下：
	// 1. s := NewServer()
	// 2. （可选）用 SetSplitter 覆盖默认的分包器。
	// 3. （可选）用 Use 注册中间件，用于解析、重写或处理消息。
	// 4. （可选）用 Add 注册路由规则，用于预先配置特定消息的处理函数。
	// 5. （可选）用 SetDefaultHandler 注册默认消息处理函数。
	// 6. （可选）用 SetOnConnected 注册异步发送消息的队列。队列里的消息会均匀分散到已有的连接。
	// 7. Start(address) 将会阻塞。
	// 8. 在要退出时，调用 Stop() 通知上述阻塞的 Start 函数退出。
	Server struct {
		listener              *Listener
		splitter              SplitterFunc
		middleware            []MiddlewareFunc   // 处理函数包裹器（中间件）
		onConnected           OnConnectedFunc    // 新连接建立时的回调
		onDisconnected        OnDisconnectedFunc // 已有连接中断时的回调
		routers               []RouterPair
		defaultHandler        HandlerFunc // 默认处理函数（没有被任何router访问的）
		connectionIDGenerator Generator   // 连接ID的生成器
	}

	IdentifierFunc func(m Serializable) bool

	// HandlerFunc 是 Packet 处理函数的标准格式。
	HandlerFunc func(c Context) error

	RouterPair struct {
		identifier IdentifierFunc
		handler    HandlerFunc
	}

	// MiddlewareFunc 是 Packet 处理函数中间件的标准格式
	MiddlewareFunc func(next HandlerFunc) HandlerFunc

	// OnConnectedFunc 是新连接建立时的回调函数。要发送的消息写入 outSiteMessageBus。
	OnConnectedFunc func(connectionID ConnectionID) (outSiteMessageBus <-chan Serializable)

	// OnDisconnectedFunc 是连接中断时的回调函数。在该函数返回后，各种资源将会被清除。
	OnDisconnectedFunc func(connectionID ConnectionID)
)

func NewServer() *Server {
	s := &Server{
		listener:              nil,
		splitter:              nil,
		middleware:            nil,
		onConnected:           nil,
		routers:               nil,
		defaultHandler:        nil,
		connectionIDGenerator: nil,
	}
	s.splitter = DefaultSplitter
	s.onConnected = DefaultOnConnected
	s.onDisconnected = DefaultOnDisconnected
	s.defaultHandler = DefaultHandler
	s.connectionIDGenerator = uuid.NewUUID32Generator()
	return s
}

// DefaultOnConnected 没有出站消息。
func DefaultOnConnected(connectionID ConnectionID) (outSiteMessageBus <-chan Serializable) {
	return nil
}

func DefaultOnDisconnected(connectionID ConnectionID) {
	return
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

func (s *Server) SetOnDisconnected(onDisconnected OnDisconnectedFunc) {
	s.onDisconnected = onDisconnected
}

func (s *Server) SetDefaultHandler(handler HandlerFunc) {
	s.defaultHandler = handler
}

func (s *Server) SetDefaultConnectionIDGenerator(generator Generator) {
	s.connectionIDGenerator = generator
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
	s.listener = NewListener(s.connectionIDGenerator)
	connChan, err := s.listener.Start(address)
	if err != nil {
		return err
	}

	for conn := range connChan {
		conn := conn
		//fmt.Println("tcp.server.Start took a connection from listener:", conn)

		h := func(c Context) error {
			m := c.Received()
			for _, r := range s.routers {
				if r.identifier(m) {
					return r.handler(c)
				}
			}
			return s.defaultHandler(c)
		}
		for i := len(s.middleware) - 1; i >= 0; i-- {
			h = s.middleware[i](h)
		}

		daemon := NewDaemon(conn, s.splitter, h, s.onConnected, s.onDisconnected)

		wg.Add(1)
		go func() {
			wg.Done()
			if err := daemon.KeepWorking(ctx); err != nil {
				//fmt.Println("conn processor exit, error=", err)
			} else {
				//fmt.Println("conn processor exit ok")
			}
		}()
	}
	return nil
}

// Stop 仅发送一个停止的信号， Start 需要等关闭所有资源后才返回。
func (s *Server) Stop() error {
	return s.listener.Stop()
}
