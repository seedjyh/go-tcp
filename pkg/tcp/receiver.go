package tcp

import (
	"bytes"
	"context"
	"errors"
	"net"
)

// Receiver 负责从 net.Conn 收取字节流，拆分成 Message 后写入 channel 。
// 异步工作，在网络出错时停止工作并关闭 channel 。
// 不负责关闭 channel
// 不负责关闭 net.Conn 。
type Receiver struct {
	conn                   net.Conn
	splitter               SplitterFunc
	receivedMessageChannel chan<- *Packet
}

func NewReceiver(conn net.Conn, splitter SplitterFunc, receivedMessageChannel chan<- *Packet) *Receiver {
	return &Receiver{
		conn:                   conn,
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
		case resultChan := <-r.receiveOneData():
			if resultChan.err != nil {
				return resultChan.err
			}
			buf.Write(resultChan.data)
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

type receiveResult struct {
	data []byte
	err  error
}

func (r *Receiver) receiveOneData() <-chan *receiveResult {
	ret := make(chan *receiveResult, 1)
	defer close(ret)
	buf := make([]byte, 1024)
	if n, err := r.conn.Read(buf); err != nil {
		ret <- &receiveResult{
			data: nil,
			err:  err,
		}
	} else {
		ret <- &receiveResult{
			data: buf[:n],
			err:  nil,
		}
	}
	return ret
}
