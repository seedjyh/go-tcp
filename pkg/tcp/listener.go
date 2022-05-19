package tcp

import (
	"net"
)

type Listener struct {
	listener net.Listener
	err      error
}

func NewListener() *Listener {
	return &Listener{}
}

// Start 会立刻返回一个新的channel，表示监听过程中收到的连接。
// 如果出错，无论是启动监听出错还是运行过程中出错，都会关闭这个channel。这也意味着监听协程中止。
//
// 曾经想过，Start仅仅返回一个 channel of *Connection，启动监听过程中的错误也异步地存放
// 而监听动作放在协程内部。
// 但这里有个问题，如果启动监听就失败，或者协程还没到net.Listen，外面就调用了Stop，这会导致l.listener发生data race。
// 所以监听动作（以及对l.listener的赋值）必须放在协程外面。限制必须 Start 成功才能调用 Stop。
// 所以返回值需要多一个 err。
//
// 由于 listener 的Close保证一定会返回阻塞的Accept函数，所以不需要ctx控制生命周期了。
func (l *Listener) Start(address string) (<-chan net.Conn, error) {
	if ln, err := net.Listen("tcp", address); err != nil {
		return nil, err
	} else {
		l.listener = ln
	}
	connections := make(chan net.Conn)
	go func() {
		defer close(connections)
		//fmt.Println("tcp.listener.proc start!")
		//defer fmt.Println("tcp.listener.proc exit!")
		for {
			if conn, err := l.listener.Accept(); err != nil {
				l.err = err
				break
			} else {
				connections <- conn
			}
		}
	}()
	return connections, nil
}

// Stop 停止监听。
// Stop 返回了错误，只表示停止监听的操作出错，不表示监听本身出错。
func (l *Listener) Stop() error {
	return l.listener.Close()
}

// Err 在 Start 返回的 channel 关闭后，可以用这个函数查看发生了什么问题。
func (l *Listener) Err() error {
	return l.err
}
