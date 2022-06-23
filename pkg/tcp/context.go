package tcp

type ReceivedMessage interface{}
type SendingMessage Serializable

// Context 是一个TCP消息的上下文。包含收到的消息、解析、分类、处理函数。
// 参考 https://github.com/labstack/echo
type Context interface {
	ConnectionID() ConnectionID
	Received() ReceivedMessage
	SetReceived(m ReceivedMessage)
	Send(m SendingMessage)
}

type handleContext struct {
	connID                ConnectionID
	received              ReceivedMessage
	sendingMessageChannel chan<- SendingMessage
}

func (c *handleContext) ConnectionID() ConnectionID {
	return c.connID
}

func (c *handleContext) Received() ReceivedMessage {
	return c.received
}

func (c *handleContext) SetReceived(m ReceivedMessage) {
	c.received = m
}

func (c *handleContext) Send(m SendingMessage) {
	c.sendingMessageChannel <- m
}
