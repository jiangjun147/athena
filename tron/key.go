package tron

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
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
	addr := address.PubkeyToAddress(key.PublicKey).String()

	return &Key{
		Address:    addr,
		PrivateKey: privateKey,
	}, nil
}
