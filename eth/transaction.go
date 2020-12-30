package eth

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Transaction struct {
	rpcTransaction
	param  Parameter
	Fee    *big.Int
	Status uint64
}

type rpcTransaction struct {
	Hash        string         `json:"hash"`
	BlockNumber *string        `json:"blockNumber,omitempty"`
	From        common.Address `json:"from,omitempty"`
	To          common.Address `json:"to,omitempty"`
	Value       string         `json:"value"`
	Input       string         `json:"input"`
	Gas         string         `json:"gas"`
	GasPrice    string         `json:"gasPrice"`
}

func (cli *EthClient) GetTransaction(ctx context.Context, hash string) (*Transaction, error) {
	txHash := common.HexToHash(hash)

	var trx Transaction
	err := cli.raw.CallContext(ctx, &trx, "eth_getTransactionByHash", txHash)
	if err != nil {
		return nil, err
	}
	if trx.Hash == "" {
		return nil, nil
	}

	p, err := NewFromHexInput(trx.Input)
	if err != nil {
		return nil, err
	}
	trx.param = p

	if trx.BlockNumber != nil {
		receipt, err := cli.TransactionReceipt(ctx, txHash)
		if err != nil {
			return nil, err
		}

		gasPrice := big.NewInt(0).SetBytes(common.FromHex(trx.GasPrice))
		gasUsed := big.NewInt(int64(receipt.GasUsed))

		trx.Fee = gasUsed.Mul(gasUsed, gasPrice)
		trx.Status = receipt.Status
	}
	return &trx, nil
}

func (trx *Transaction) GetBlock() int64 {
	if trx.BlockNumber == nil {
		return 0
	}

	return big.NewInt(0).SetBytes(common.FromHex(*trx.BlockNumber)).Int64()
}

func (trx *Transaction) GetParamToAddress() common.Address {
	return common.BytesToAddress(trx.param.Get(0))
}

func (trx *Transaction) GetParamValue() *big.Int {
	return big.NewInt(0).SetBytes(trx.param.Get(1))
}
