package aun

// Readable Interface

// This package's channel messaging type must be implement this interface.
// Messages will get the bytes for messaage through the "getData" method.
type Readable interface {
	getData() []byte
}
