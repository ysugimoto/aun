package aun

type MessageFrame struct {
	MessageReadable
	Frame *Frame
}

func NewMessageFrame(buffer []byte) (*MessageFrame, error) {
	f := &MessageFrame{
		Frame: NewFrame(),
	}

	if err := f.Frame.parse(buffer); err != nil {
		return nil, err
	}

	return f, nil
}

func (m *MessageFrame) getData() []byte {
	return []byte("Tempoary")
}
