package tcp

// 包含了 Message 接口，以及一个最基本的实现： Packet

type Message interface {
	Bytes() []byte
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
	return &Packet{data: data}
}
