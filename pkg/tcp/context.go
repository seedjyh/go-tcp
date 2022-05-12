package tcp

// Context 是一个TCP消息的上下文。包含收到的消息、解析、分类、处理函数。
// 参考 https://github.com/labstack/echo
type Context interface {
	Received() *Packet
	Send(p *Packet)
}

type handleContext struct {
	received              *Packet
	sendingMessageChannel chan<- *Packet
}

func (c *handleContext) Received() *Packet {
	return c.received
}

func (c *handleContext) Send(p *Packet) {
	c.sendingMessageChannel <- p
}
