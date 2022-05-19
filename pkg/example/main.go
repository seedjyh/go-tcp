// 这里是使用go-tcp的一个服务器的例子。
package main

import (
	"context"
	"fmt"
	"github.com/seedjyh/go-tcp/pkg/tcp"
	"golang.org/x/sync/errgroup"
	"sort"
	"strings"
)

// 每5个字节一个包
func mySplitter(buf []byte) (tcp.Message, int, error) {
	if len(buf) < 5 {
		return nil, 0, tcp.NoEnoughData
	}
	return tcp.NewPacket(buf[:5]), 5, nil
}

// middlewareResponseFiveZeroes 如果消息是"00000"则返回"11111"。
func middlewareResponseFiveZeroes(next tcp.HandlerFunc) tcp.HandlerFunc {
	return func(c tcp.Context) error {
		p := c.Received()
		if string(p.Bytes()) == "00000" {
			c.Send(tcp.NewPacket([]byte("11111")))
			return nil
		} else {
			return next(c)
		}
	}
}

type MyMessage struct {
	length int
	word   string
}

func (m *MyMessage) Bytes() []byte {
	return []byte(m.word)
}

func isAllAlpha(m tcp.Message) bool {
	mm := m.(*MyMessage)
	for _, c := range mm.word {
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' {
			continue
		} else {
			return false
		}
	}
	return true
}

func isAllDigit(m tcp.Message) bool {
	mm := m.(*MyMessage)
	for _, c := range mm.word {
		if '0' <= c && c <= '9' {
			continue
		} else {
			return false
		}
	}
	return true
}

// middlewareUnpackToMessage 将消息转换成内容包含「长度」和「string格式的内容」的两个成员的struct。
func middlewareUnpackToMessage(next tcp.HandlerFunc) tcp.HandlerFunc {
	return func(c tcp.Context) error {
		rawMessage := c.Received()
		rawPacket := rawMessage.(*tcp.Packet)
		m := &MyMessage{
			length: len(rawPacket.Bytes()),
			word:   string(rawPacket.Bytes()),
		}
		c.SetReceived(m)
		return next(c)
	}
}

type ByteSlice []byte

func (bs ByteSlice) Len() int {
	return len(bs)
}

func (bs ByteSlice) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

func (bs ByteSlice) Less(i, j int) bool {
	return bs[i] < bs[j]
}

func increaseString(raw string) string {
	bs := []byte(raw)
	sort.Sort(ByteSlice(bs))
	return string(bs)
}

func decreaseString(raw string) string {
	bs := []byte(increaseString(raw))
	bs2 := make([]byte, len(bs))
	for i := 0; i < len(bs); i++ {
		bs2[i] = bs[len(bs)-1-i]
	}
	return string(bs2)
}

func main() {
	port := 11223
	s := tcp.NewServer()
	eg, ctx := errgroup.WithContext(context.Background())

	// 分包规则：每5个字节一个包。
	s.SetSplitter(mySplitter)

	// 过滤规则：如果消息是"00000"则返回"11111"。
	s.Use(middlewareResponseFiveZeroes)

	// 转换规则：将消息转换成内容包含「长度」和「string格式的内容」的两个成员的struct。
	s.Use(middlewareUnpackToMessage)

	// 注册处理规则：
	// 如果全是字母，发送两条响应，依次是全大写的和全小写的。
	// 如果全是数字，发送两条响应，依次是递增和递减的。
	// 其他情况，调用默认处理函数。
	s.Add(isAllAlpha, func(c tcp.Context) error {
		m := c.Received().(*MyMessage)
		c.Send(&MyMessage{
			length: m.length,
			word:   strings.ToUpper(m.word),
		})
		c.Send(&MyMessage{
			length: m.length,
			word:   strings.ToLower(m.word),
		})
		return nil
	})
	s.Add(isAllDigit, func(c tcp.Context) error {
		m := c.Received().(*MyMessage)
		c.Send(&MyMessage{
			length: m.length,
			word:   increaseString(m.word),
		})
		c.Send(&MyMessage{
			length: m.length,
			word:   decreaseString(m.word),
		})
		return nil
	})

	// start
	eg.Go(func() error { return s.Start(fmt.Sprintf("0.0.0.0:%d", port)) })

	// wait
	<-ctx.Done()

	// stop all here
	fmt.Println("end err:", eg.Wait())
}
