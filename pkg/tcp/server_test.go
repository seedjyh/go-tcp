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
	s.SetSplitter(func(buf []byte) (Message, int, error) {
		if len(buf) < fixedLength {
			return nil, 0, NoEnoughData
		} else {
			return NewPacket(buf[:fixedLength]), fixedLength, nil
		}
	})
	s.SetOnConnected(func(inSiteMessageBus <-chan Message) (outSiteMessageBus <-chan Message) {
		out := make(chan Message)
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

func TestServer_Use(t *testing.T) {
	s := NewServer()
	s.SetSplitter(func(buf []byte) (Message, int, error) {
		if len(buf) < fixedLength {
			return nil, 0, NoEnoughData
		} else {
			return NewPacket(buf[:fixedLength]), fixedLength, nil
		}
	})
	s.SetOnConnected(func(inSiteMessageBus <-chan Message) (outSiteMessageBus <-chan Message) {
		go func() {
			for _ = range inSiteMessageBus {
				assert.Fail(t, "should not access here")
			}
		}()
		return nil
	})
	s.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			received := c.Received()
			assert.Equal(t, requestText, string(received.Bytes()))
			c.Send(NewPacket([]byte(responseText)))
			// no next(c)
			return nil
		}
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
