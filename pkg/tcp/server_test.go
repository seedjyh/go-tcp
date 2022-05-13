package tcp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"sync"
	"testing"
)

const fixedLength = 5
const requestText = "hello"
const responseText = "world"
const port = 11240

func TestServer_Start(t *testing.T) {
	s := NewServer()
	s.SetSplitter(func(buf []byte) (*Packet, int, error) {
		if len(buf) < fixedLength {
			return nil, 0, NoEnoughData
		} else {
			return NewPacket(buf[:fixedLength]), fixedLength, nil
		}
	})
	s.SetOnConnected(func(inSiteMessageBus <-chan *Packet) (outSiteMessageBus <-chan *Packet) {
		out := make(chan *Packet)
		go func() {
			for m := range inSiteMessageBus {
				assert.Equal(t, requestText, string(m.Bytes()))
				out <- NewPacket([]byte(responseText))
			}
		}()
		return out
	})
	// start
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = s.Start(fmt.Sprintf("0.0.0.0:%d", port))
	}()

	var conn net.Conn
	fmt.Println("test.connecting...")
	if c, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port)); assert.NoError(t, err) {
		conn = c
	}
	fmt.Println("test.connected.")

	req := []byte(requestText)
	for len(req) > 0 {
		if n, err := conn.Write(req); assert.NoError(t, err) {
			req = req[n:]
		}
	}

	rsp := make([]byte, 10)
	fmt.Println("test.receiving...")
	if n, err := conn.Read(rsp); assert.NoError(t, err) {
		assert.Equal(t, responseText, string(rsp[:n]))
	}
	fmt.Println("test.received.")

	assert.NoError(t, conn.Close())
	assert.NoError(t, s.Stop())
	wg.Wait()
}
