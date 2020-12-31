package chain

import (
	"context"
	"math/big"
)

type Key struct {
	Address    string
	PrivateKey string
}

type Token struct {
	Address  string
	Decimals int32
}

type Data struct {
	Address string
	Value   *big.Int
}

type Transaction struct {
	Hash   string
	Block  int64
	From   string
	To     string
	Value  *big.Int
	Data   *Data
	Fee    *big.Int
	Status bool
}

type Client interface {
	CreateKey() (*Key, error)
	TokenTransfer(ctx context.Context, token *Token, fromPrivKey string, toAddress string, amount *big.Int) (*Transaction, error)
	GetTransaction(ctx context.Context, hash string) (*Transaction, error)
	FindTokenTransaction(ctx context.Context, token *Token, toAddress string, block int64) ([]*Transaction, error)
}
