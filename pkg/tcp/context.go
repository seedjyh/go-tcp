package tcp

// Context 是一个TCP消息的上下文。包含收到的消息、解析、分类、处理函数。
// 参考 https://github.com/labstack/echo
type Context interface {
	Received() Message
	SetReceived(m Message)
	Send(m Message)
}

type handleContext struct {
	received              Message
	sendingMessageChannel chan<- Message
}

func (c *handleContext) Received() Message {
	return c.received
}

func (c *handleContext) SetReceived(m Message) {
	c.received = m
}

func (c *handleContext) Send(m Message) {
	c.sendingMessageChannel <- m
}
