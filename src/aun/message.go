package aun

type Message struct {
	Data string
}

func NewMessage(data []byte) *Message {
	return &Message{
		Data: string(data),
	}
}
