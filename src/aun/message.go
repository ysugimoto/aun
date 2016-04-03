package aun

type Message struct {
	MessageReadable
	Data string
}

func NewMessage(data []byte) *Message {
	return &Message{
		Data: string(data),
	}
}

func (m *Message) getData() []byte {
	return []byte(m.Data)
}
