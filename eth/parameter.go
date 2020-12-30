package eth

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
)

type Parameter []byte

func NewParameter(n int) Parameter {
	return make(Parameter, n*32)
}

func NewFromHexParameter(hp []string) (Parameter, error) {
	p := make(Parameter, len(hp)*32)
	for i, h := range hp {
		if err := p.SetHex(i, h); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func NewFromHexInput(input string) (Parameter, error) {
	if len(input) < 8 {
		return nil, nil
	}

	if input[:2] == "0x" {
		input = input[2:]
	}

	var hp []string
	params := input[8:]
	for i := 0; i < len(params); i += 64 {
		hp = append(hp, params[i:i+64])
	}
	return NewFromHexParameter(hp)
}

func (p Parameter) Set(i int, param []byte) {
	s := i * 32
	copy(p[s:s+32], common.LeftPadBytes(param, 32))
}

func (p Parameter) SetHex(i int, h string) error {
	param, err := hex.DecodeString(h)
	if err != nil {
		return err
	}
	p.Set(i, param)
	return nil
}

func (p Parameter) Get(i int) []byte {
	s := i * 32
	return p[s : s+32]
}
