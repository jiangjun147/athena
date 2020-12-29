package eth

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
)

type Key struct {
	Address    string
	PrivateKey string
}

func CreateKey() (*Key, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	privateKey := hex.EncodeToString(crypto.FromECDSA(key))
	address := crypto.PubkeyToAddress(key.PublicKey).Hex()

	return &Key{
		Address:    address,
		PrivateKey: privateKey,
	}, nil
}
