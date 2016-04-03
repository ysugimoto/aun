package aun

type FrameStack []*Frame

func (f FrameStack) synthesize() []byte {
	var payload []byte

	for _, frame := range f {
		payload = append(payload, frame.PayloadData...)
	}

	return payload
}
