package tcp

// Serializable 具体消息数据接口
type Serializable interface {
	Bytes() []byte
}

// Envelope 包内通用的消息结构。
type Envelope struct {
	ConnID ConnectionID // 收到该消息的连接，或者即将从这个连接发送该消息。
	Data   Serializable // 具体数据
}

func NewEnvelope(connID ConnectionID, data Serializable) *Envelope {
	return &Envelope{
		ConnID: connID,
		Data:   data,
	}
}

// Packet 是一个tcp包。包含定长的字节。是从tcp流拆分得到的结果。
type Packet struct {
	data []byte
}

func (p *Packet) Bytes() []byte {
	return p.data
}

func NewPacket(buf []byte) *Packet {
	l := len(buf)
	data := make([]byte, l)
	copy(data, buf)
	return &Packet{
		data: data,
	}
}
