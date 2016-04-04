package aun

// Define type the *Frame's slice.
type FrameStack []*Frame

// Synthesize the queueing frames.
func (f FrameStack) synthesize() []byte {
	var payload []byte

	for _, frame := range f {
		payload = append(payload, frame.PayloadData...)
	}

	return payload
}
