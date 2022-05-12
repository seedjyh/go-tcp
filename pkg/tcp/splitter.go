package tcp

// SplitterFunc 检查buf，从其头部拆分出一个 Packet 。
// 返回内容包括拆分出的 Packet 、读取的长度（通常和 Packet 字节数相同）。如果 error 不是 nil ，那么前两个返回值没有用。
type SplitterFunc func(buf []byte) (*Packet, int, error)

// DefaultSplitter 是最基本的分包器。将buf里所有字节都放在一个包里。
func DefaultSplitter(buf []byte) (*Packet, int, error) {
	if len(buf) == 0 {
		return nil, 0, NoEnoughData
	}
	return NewPacket(buf), len(buf), nil
}
