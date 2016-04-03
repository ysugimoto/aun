package aun

import (
	_ "fmt"
	_ "strconv"
	"testing"
)

var fixture = []byte{129, 130, 195, 133, 210, 16, 140, 206}

func TestFrameParser(t *testing.T) {
	s := fixture[1]
	//a, _ := strconv.ParseInt("01111111", 2, 0)

	pl := s & 127
	//bin := []byte(fmt.Sprintf("%b", fixture[0]))
	//bin = bin[1:]
	t.Error(pl)
	//b := bytes.NewBuffer(fixture)
	//var n uint8
	//binary.Read(b, binary.BigEndian, &n)
	//if n !=  {
	//	t.Errorf("n is not 1000: %d", n)
	//}
}
