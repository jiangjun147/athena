package chain

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rickone/athena/chain/eth"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/eth"
	"github.com/shopspring/decimal"
)

const (
	EtherDecimals = 18 // 1 Ether = 10^18 wei
)

type Ethereum struct {
	conf *config.BlockChainConf
}

var (
	transferEventHash = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
)

func (e *Ethereum) NewAccount(ctx context.Context) (string, string, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return "", "", err
	}

	privateKey := hex.EncodeToString(crypto.FromECDSA(key))
	address := crypto.PubkeyToAddress(key.PublicKey).Hex()

	return address, privateKey, nil
}

func (e *Ethereum) BalanceOf(ctx context.Context, who string) (decimal.Decimal, error) {
	balance, err := eth.Client().BalanceAt(ctx, common.HexToAddress(who), nil)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromBigInt(balance, -EtherDecimals), nil
}

func (e *Ethereum) Transfer(ctx context.Context, key string, who string, amount decimal.Decimal) (txId string, err error) {
	cli := eth.Client()

	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return "", err
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := cli.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return "", err
	}

	value := amount.Mul(decimal.New(1, 18)).BigInt() // in wei
	gasPrice, err := cli.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(nonce, common.HexToAddress(who), value, e.conf.FeeLimit, gasPrice, nil)

	chainId, err := cli.NetworkID(ctx)
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	if err != nil {
		return "", err
	}

	err = cli.SendTransaction(ctx, signedTx)
	return signedTx.Hash().Hex(), err
}

func (e *Ethereum) GetTransactionReceipt(ctx context.Context, txId string) (*TransactionReceipt, error) {
	cli := eth.Client()

	txHash := common.HexToHash(txId)

	tx, isPending, err := cli.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, err
	}
	if isPending {
		return nil, nil
	}

	receipt, err := eth.Client().TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, err
	}

	gasPrice := decimal.NewFromBigInt(tx.GasPrice(), -EtherDecimals)

	return &TransactionReceipt{
		Result: receipt.Status != 0,
		Fee:    gasPrice.Mul(decimal.NewFromInt(int64(receipt.GasUsed))),
	}, nil
}

func (e *Ethereum) TokenBalanceOf(ctx context.Context, who string) (decimal.Decimal, error) {
	cli := eth.Client()

	token, err := eth.NewUSDT(common.HexToAddress(e.conf.Contract), cli)
	if err != nil {
		return decimal.Zero, err
	}

	balance, err := token.BalanceOf(&bind.CallOpts{
		Pending: true,
		Context: ctx,
	}, common.HexToAddress(who))
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromBigInt(balance, -e.conf.Decimals), nil
}

func (e *Ethereum) TokenTransfer(ctx context.Context, key string, who string, amount decimal.Decimal) (txId string, err error) {
	cli := eth.Client()

	token, err := eth.NewUSDT(common.HexToAddress(e.conf.Contract), cli)
	if err != nil {
		return "", err
	}

	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return "", err
	}

	tx, err := token.Transfer(bind.NewKeyedTransactor(privateKey), common.HexToAddress(who), amount.Mul(decimal.New(1, 6)).BigInt())
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

func newTransferEvent(ctx context.Context, log types.Log, conf *config.BlockChainConf) (*TransferEvent, error) {
	block, err := eth.Client().BlockByHash(ctx, log.BlockHash)
	if err != nil {
		return nil, err
	}

	return &TransferEvent{
		Id:          log.TxHash.String(),
		BlockNumber: block.Number(),
		Timestamp:   int64(block.Time() * 1000),
		From:        "0x" + common.Bytes2Hex(log.Topics[1].Bytes()[(common.HashLength-common.AddressLength):]),
		To:          "0x" + common.Bytes2Hex(log.Topics[2].Bytes()[(common.HashLength-common.AddressLength):]),
		Amount:      decimal.NewFromBigInt(new(big.Int).SetBytes(log.Data), -conf.Decimals),
	}, nil
}

func (e *Ethereum) SubscribeTokenTransferEvent(ctx context.Context, from, to string, lastPos int64, f func(event *TransferEvent) error) (TransferEventSubscription, error) {
	cli := eth.Client()

	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(e.conf.Contract)},
		Topics: [][]common.Hash{
			{transferEventHash},
			{},
			{},
		},
	}

	if from != "" {
		query.Topics[1] = append(query.Topics[1], common.HexToHash(from))
	}

	if to != "" {
		query.Topics[2] = append(query.Topics[2], common.HexToHash(to))
	}

	if lastPos > 0 {
		query.FromBlock = new(big.Int).SetInt64(lastPos)
	}

	logs, err := cli.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	for _, log := range logs {
		event, err := newTransferEvent(ctx, log, e.conf)
		if err != nil {
			return nil, err
		}
		if err := f(event); err != nil {
			return nil, err
		}
	}

	ch := make(chan types.Log)
	sub, err := cli.SubscribeFilterLogs(ctx, query, ch)
	if err != nil {
		return nil, err
	}

	return &USDTTransferEventSubscription{
		conf:    e.conf,
		ch:      ch,
		sub:     sub,
		handler: f,
	}, nil
}

type USDTTransferEventSubscription struct {
	conf    *config.BlockChainConf
	ch      chan types.Log
	sub     ethereum.Subscription
	handler func(log *TransferEvent) error
}

func (s *USDTTransferEventSubscription) Subscribe(ctx context.Context) error {
	for {
		var log types.Log
		select {
		case err := <-s.sub.Err():
			return err
		case log = <-s.ch:
		}

		event, err := newTransferEvent(ctx, log, s.conf)
		if err != nil {
			return err
		}
		if err := s.handler(event); err != nil {
			return err
		}
	}
}

func (s *USDTTransferEventSubscription) Close() {
	s.sub.Unsubscribe()
}
