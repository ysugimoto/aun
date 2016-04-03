package aun

import (
	"fmt"
	"strconv"
)

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

type Frame struct {
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

func NewFrame() *Frame {
	return &Frame{}
}

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

	fmt.Println(bits, f.PayloadLength)
	index := 2
	switch {
	case f.PayloadLength == 126:
		bin := ""
		for i := 0; i < 2; i++ {
			bin += fmt.Sprintf("%b", buffer[index+i])
		}
		length, err := strconv.ParseInt(bin, 2, 0)
		if err != nil {
			return err
		}
		f.PayloadLength = int(length)
		index += 2
	case f.PayloadLength == 127:
		bin := ""
		for i := 0; i < 8; i++ {
			bin += fmt.Sprintf("%b", buffer[index+i])
		}
		length, err := strconv.ParseInt(bin, 2, 0)
		if err != nil {
			return err
		}
		f.PayloadLength = int(length)
		index += 8
	}

	if f.Mask > 0 {
		f.MaskingKey = buffer[index:(index + 4)]
		index += 4
	}

	if f.Mask > 0 {
		payload := buffer[index:(index + f.PayloadLength)]
		for i := 0; i < len(payload); i++ {
			f.PayloadData = append(
				f.PayloadData,
				byte((int(payload[i]) ^ int(f.MaskingKey[i%4]))),
			)
		}
	} else {
		f.PayloadData = buffer[index:(index + f.PayloadLength)]
	}

	fmt.Println(string(f.PayloadData))

	return nil
}

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
	case f.PayloadLength > 126 && f.PayloadLength < 65535:
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
