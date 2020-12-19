package grpcex

import "fmt"

type RawBuf []byte

func (rb RawBuf) Reset() {
}

func (rb RawBuf) String() string {
	return fmt.Sprintf("%x", rb)
}

func (rb RawBuf) ProtoMessage() {
}

func (rb RawBuf) Unmarshal(data []byte) error {
	rb = data
	return nil
}

func (rb RawBuf) Marshal() ([]byte, error) {
	return rb, nil
}
