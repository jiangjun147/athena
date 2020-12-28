package eth

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

const (
	etherDecimals = 18
)

type Receipt struct {
	Status uint64
	Block  uint64
	Fee    decimal.Decimal
}

func (cli *EthClient) GetReceipt(ctx context.Context, hash common.Hash) (*Receipt, error) {
	tx, isPending, err := cli.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if isPending {
		return nil, nil
	}

	receipt, err := cli.TransactionReceipt(ctx, hash)
	if err != nil {
		return nil, err
	}

	gasPrice := decimal.NewFromBigInt(tx.GasPrice(), -etherDecimals)

	result := &Receipt{
		Status: receipt.Status,
		Fee:    gasPrice.Mul(decimal.NewFromInt(int64(receipt.GasUsed))),
	}
	if receipt.BlockNumber != nil {
		result.Block = receipt.BlockNumber.Uint64()
	}
	return result, nil
}
