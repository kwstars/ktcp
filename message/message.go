package message

// Message is the unpacked message object.
type Message struct {
	ID   uint32 // 协议id
	Flag uint16 // message是否正确 ture:正确 false:错误
	Data []byte // 数据
}
