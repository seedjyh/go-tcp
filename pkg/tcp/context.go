package tcp

// Context 是一个TCP消息的上下文。包含收到的消息、解析、分类、处理函数。
// 参考 https://github.com/labstack/echo
type Context interface {
	ConnectionID() ConnectionID
	Received() Serializable
	SetReceived(m Serializable)
	Send(m Serializable)
}

type handleContext struct {
	connID                ConnectionID
	received              Serializable
	sendingMessageChannel chan<- Serializable
}

func (c *handleContext) ConnectionID() ConnectionID {
	return c.connID
}

func (c *handleContext) Received() Serializable {
	return c.received
}

func (c *handleContext) SetReceived(m Serializable) {
	c.received = m
}

func (c *handleContext) Send(m Serializable) {
	c.sendingMessageChannel <- m
}
