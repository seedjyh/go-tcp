package tcp

// Packet 是一个tcp包。包含定长的字节。是从tcp流拆分得到的结果。
type Packet struct {
	data []byte
}

func NewPacket(buf []byte) *Packet {
	l := len(buf)
	data := make([]byte, l)
	copy(data, buf)
	return &Packet{data: data}
}