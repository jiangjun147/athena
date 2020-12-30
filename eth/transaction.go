package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Transaction struct {
	rpcTransaction
	To     common.Address
	Amount *big.Int
}

type rpcTransaction struct {
	Hash        string         `json:"hash"`
	BlockNumber *string        `json:"blockNumber,omitempty"`
	From        common.Address `json:"from,omitempty"`
	to          common.Address `json:"to,omitempty"`
	value       string         `json:"value"`
	input       string         `json:"input"`
}

func (cli *EthClient) GetTransaction(ctx context.Context, hash string) (*Transaction, error) {
	var trx Transaction
	err := cli.raw.CallContext(ctx, &trx, "eth_getTransactionByHash", common.HexToHash(hash))
	if err != nil {
		return nil, err
	}
	if trx.Hash == "" {
		return nil, nil
	}

	p, err := NewFromHexInput(trx.input)
	if err != nil {
		return nil, err
	}

	trx.To = common.BytesToAddress(p.Get(0))
	trx.Amount = big.NewInt(0).SetBytes(p.Get(1))
	return &trx, nil
}
