package aun

// Message frame spec:
//
// 0                   1                   2                   3
// 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
// +-+-+-+-+-------+-+-------------+-------------------------------+
// |F|R|R|R| opcode|M| Payload len |    Extended payload length    |
// |I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
// |N|V|V|V|       |S|             |   (if payload len==126/127)   |
// | |1|2|3|       |K|             |                               |
// +-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
// |     Extended payload length continued, if payload len == 127  |
// + - - - - - - - - - - - - - - - +-------------------------------+
// |                               |Masking-key, if MASK set to 1  |
// +-------------------------------+-------------------------------+
// | Masking-key (continued)       |          Payload Data         |
// +-------------------------------- - - - - - - - - - - - - - - - +
// :                     Payload Data continued ...                :
// + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
// |                     Payload Data continued ...                |
// +---------------------------------------------------------------+

import (
	"encoding/binary"
)

type Frame struct {
	// embeded interface
	Readable
	Fin           int
	RSV1          int
	RSV2          int
	RSV3          int
	Opcode        int
	Mask          int
	PayloadLength int
	MaskingKey    []byte
	PayloadData   []byte
}

// Create new frame
func NewFrame() *Frame {
	return &Frame{}
}

// Create "pong" frame
func NewPongFrame() *Frame {
	return &Frame{
		Fin:           1,
		RSV1:          0,
		RSV2:          0,
		RSV3:          0,
		Opcode:        10,
		Mask:          0,
		PayloadLength: 0,
		PayloadData:   []byte{},
	}
}

// Create Message frame for S->C sending
func BuildFrame(message []byte, maxSize int) (FrameStack, error) {
	stack := FrameStack{}

	for len(message) > maxSize {
		frame, err := BuildSingleFrame(message[0:maxSize], 0, 0)
		if err != nil {
			return stack, err
		}
		stack = append(stack, frame)
		message = message[maxSize:]
	}
	if len(message) > 0 {
		frame, err := BuildSingleFrame(message, 1, 1)
		if err != nil {
			return stack, err
		}
		stack = append(stack, frame)
	}
	return stack, nil
}

// Create single message frame.
// If finBit is equal to zero,
// Message was split deriverling.
func BuildSingleFrame(message []byte, finBit int, opcode int) (*Frame, error) {
	return &Frame{
		Fin:           finBit,
		RSV1:          0,
		RSV2:          0,
		RSV3:          0,
		Opcode:        opcode,
		Mask:          0,
		PayloadLength: len(message),
		PayloadData:   message,
	}, nil
}

// Parse them incoming message frame.
func (f *Frame) parse(buffer []byte) error {
	bits := int(buffer[0])
	f.Fin = (bits >> 7) & 1
	f.RSV1 = (bits >> 6) & 1
	f.RSV2 = (bits >> 5) & 1
	f.RSV3 = (bits >> 4) & 1
	f.Opcode = bits & 0xF

	bits = int(buffer[1])
	f.Mask = (bits >> 7) & 1
	f.PayloadLength = bits & 0x7F

	index := 2
	switch {
	// payload length = 126, using length of 2 bytes
	case f.PayloadLength == 126:
		n := binary.BigEndian.Uint16(buffer[index:(index + 2)])
		f.PayloadLength = int(n)
		index += 2
	// payload length = 127, using length of 8 bytes
	case f.PayloadLength == 127:
		n := binary.BigEndian.Uint64(buffer[index:(index + 8)])
		f.PayloadLength = int(n)
		index += 8
	}

	// Masking check.
	// C->S message has always need to be masking.
	if f.Mask > 0 {
		f.MaskingKey = buffer[index:(index + 4)]
		index += 4
		payload := buffer[index:(index + f.PayloadLength)]
		size := len(payload)
		for i := 0; i < size; i++ {
			f.PayloadData = append(
				f.PayloadData,
				// Unmasking payload:
				// payload-i ^ masking-key-j mod 4
				byte((int(payload[i]) ^ int(f.MaskingKey[i%4]))),
			)
		}
	} else {
		f.PayloadData = buffer[index:(index + f.PayloadLength)]
	}

	return nil
}

// Implement Readable interface
func (f *Frame) getData() []byte {
	return f.toFrameBytes()
}

// Binarify the frame to send.
func (f *Frame) toFrameBytes() (data []byte) {
	bin := 0
	bin |= (f.Fin << 7)
	bin |= (f.RSV1 << 6)
	bin |= (f.RSV1 << 5)
	bin |= (f.RSV1 << 4)
	bin |= f.Opcode
	data = append(data, byte(bin))

	bin = 0
	bin |= (f.Mask << 7)
	switch {
	case f.PayloadLength > 65535:
		bin |= 127
		data = append(data, byte(bin))
		// extra payload length of 8 bytes
		data = append(
			data,
			byte((f.PayloadLength>>56)&0xFF),
			byte((f.PayloadLength>>48)&0xFF),
			byte((f.PayloadLength>>40)&0xFF),
			byte((f.PayloadLength>>32)&0xFF),
			byte((f.PayloadLength>>24)&0xFF),
			byte((f.PayloadLength>>16)&0xFF),
			byte((f.PayloadLength>>8)&0xFF),
			byte((f.PayloadLength)&0xFF),
		)
	case f.PayloadLength > 126 && f.PayloadLength <= 65535:
		bin |= 126
		data = append(data, byte(bin))
		// extra payload length of 2 bytes
		data = append(
			data,
			byte((f.PayloadLength>>8)&0xFF),
			byte((f.PayloadLength)&0xFF),
		)
	default:
		bin |= f.PayloadLength
		data = append(data, byte(bin))
	}
	data = append(data, f.PayloadData...)
	return
}
