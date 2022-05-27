package tcp

import (
	"bytes"
	"context"
	"errors"
)

// Receiver 负责从 net.Conn 收取字节流，拆分成 Envelope 后写入 channel 。
// 异步工作，在网络出错时停止工作并关闭 channel 。
// 不负责关闭 channel
// 不负责关闭 net.Conn 。
type Receiver struct {
	connection             *Connection
	splitter               SplitterFunc
	receivedMessageChannel chan<- Serializable
}

func NewReceiver(connection *Connection, splitter SplitterFunc, receivedMessageChannel chan<- Serializable) *Receiver {
	return &Receiver{
		connection:             connection,
		splitter:               splitter,
		receivedMessageChannel: receivedMessageChannel,
	}
}

func (r *Receiver) KeepWorking(ctx context.Context) error {
	buf := bytes.NewBuffer(nil)
	for {
		select {
		case <-ctx.Done():
			return errors.New("context is done")
		default:
		}
		if data, err := r.receiveOneData(); err != nil {
			return err
		} else {
			buf.Write(data)
			if message, messageByteLength, err := r.splitter(buf.Bytes()); err != nil {
				if errors.Is(err, NoEnoughData) {
					continue
				} else {
					return err
				}
			} else {
				r.receivedMessageChannel <- message
				if n, err := buf.Read(make([]byte, messageByteLength)); err != nil {
					return err
				} else if n != messageByteLength {
					return errors.New("something wrong here")
				}
			}
		}
	}
}

func (r *Receiver) receiveOneData() ([]byte, error) {
	buf := make([]byte, 1024)
	if n, err := r.connection.conn.Read(buf); err != nil {
		return nil, err
	} else {
		return buf[:n], nil
	}
}
