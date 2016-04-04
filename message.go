package aun

// Simple message struct.
type Message struct {
	Readable
	Data string
}

// Create []byte wrapped pointer
func NewMessage(data []byte) *Message {
	return &Message{
		Data: string(data),
	}
}

// Readable interface implement
func (m *Message) getData() []byte {
	return []byte(m.Data)
}
