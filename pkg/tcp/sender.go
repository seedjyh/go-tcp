package tcp

import (
	"context"
	"github.com/pkg/errors"
	"net"
	"time"
)

// Sender 负责从一个 channel 获取 Packet 然后发送出去。
// 当发送出错或 channel 被关闭时结束。
// 不负责关闭 net.Conn 。
// 不负责关闭 channel
type Sender struct {
	conn                  net.Conn
	sendingMessageChannel <-chan *Packet
}

func NewSender(conn net.Conn, sendingMessageChannel <-chan *Packet) *Sender {
	return &Sender{
		conn:                  conn,
		sendingMessageChannel: sendingMessageChannel,
	}
}

// KeepWorking 会阻塞并持续从 messages 获取 Packet 并发送。
// ctx Done 或者 channel被关闭则返回。
func (s *Sender) KeepWorking(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("context is done")
		case m, ok := <-s.sendingMessageChannel:
			if !ok {
				return errors.New("channel is closed")
			}
			if err := s.send(m); err != nil {
				return err
			}
		}
	}
}

// send 会阻塞并试图发送 m 到连接。
// 发送成功、出错都会返回。
// 如果发送出错，可能只发送了半条消息。所以如果返回值不是nil，应该立刻关闭连接，避免后续数据出错。
func (s *Sender) send(m *Packet) error {
	const maxWait = time.Second * 1 // 最多 maxWait 要发完
	buf := m.data
	if err := s.conn.SetWriteDeadline(time.Now().Add(maxWait)); err != nil {
		return err
	}
	for len(buf) > 0 {
		if wc, err := s.conn.Write(buf); err != nil {
			return err
		} else {
			buf = buf[wc:]
		}
	}
	return nil
}
