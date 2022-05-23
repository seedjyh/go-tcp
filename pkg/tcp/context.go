package tcp

// Context 是一个TCP消息的上下文。包含收到的消息、解析、分类、处理函数。
// 参考 https://github.com/labstack/echo
type Context interface {
	Received() *Envelope
	SetReceived(m *Envelope)
	Send(m *Envelope)
}

type handleContext struct {
	received              *Envelope
	sendingMessageChannel chan<- *Envelope
}

func (c *handleContext) Received() *Envelope {
	return c.received
}

func (c *handleContext) SetReceived(m *Envelope) {
	c.received = m
}

func (c *handleContext) Send(m *Envelope) {
	c.sendingMessageChannel <- m
}
