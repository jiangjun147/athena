package chain

import (
	"context"
	"math/big"

	"github.com/shopspring/decimal"
)

type TransactionReceipt struct {
	Result bool
	Fee    decimal.Decimal
}

type TransferEvent struct {
	Id          string
	BlockNumber *big.Int
	Timestamp   int64
	From        string
	To          string
	Amount      decimal.Decimal
}

type TransferEventSubscription interface {
	Subscribe(ctx context.Context) error
	Close()
}

type Chain interface {
	// 创建账号
	NewAccount(ctx context.Context) (address string, key string, err error)
	// 查询余额
	BalanceOf(ctx context.Context, who string) (balance decimal.Decimal, err error)
	// 转账
	Transfer(ctx context.Context, key string, who string, amount decimal.Decimal) (txId string, err error)
	// 查询事务收据
	GetTransactionReceipt(ctx context.Context, txId string) (receipt *TransactionReceipt, err error)

	// 查询代币余额
	TokenBalanceOf(ctx context.Context, who string) (balance decimal.Decimal, err error)
	// 代币转账
	TokenTransfer(ctx context.Context, key string, who string, amount decimal.Decimal) (txId string, err error)
	// 查询代币转账日志
	SubscribeTokenTransferEvent(ctx context.Context, from, to string, lastPos int64, f func(event *TransferEvent) error) (sub TransferEventSubscription, err error)
}
