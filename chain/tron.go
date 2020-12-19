package chain

import (
	"context"
	"encoding/hex"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/rickone/athena/chain/tron"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/errcode"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
)

var (
	SunPerTrx = decimal.New(1, 6)
)

type Tron struct {
	conf *config.BlockChainConf
}

func (t *Tron) NewAccount(ctx context.Context) (string, string, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return "", "", err
	}

	privateKey := hex.EncodeToString(crypto.FromECDSA(key))
	addr := address.PubkeyToAddress(key.PublicKey).String()

	return addr, privateKey, nil
}

func (t *Tron) BalanceOf(ctx context.Context, who string) (decimal.Decimal, error) {
	acc, err := tron.GetAccount(t.conf.ApiUrl, who)
	if err != nil {
		return decimal.Zero, err
	}

	return acc.Balance.Div(SunPerTrx), nil // sun to trx
}

func (t *Tron) Transfer(ctx context.Context, key string, who string, amount decimal.Decimal) (txId string, err error) {
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return "", err
	}
	fromAddress := address.PubkeyToAddress(privateKey.PublicKey).String()

	tx, err := tron.CreateTransaction(t.conf.ApiUrl, fromAddress, who, amount.Mul(decimal.New(1, 6)).BigInt())
	if err != nil {
		return "", err
	}

	err = tx.Sign(privateKey)
	if err != nil {
		return "", err
	}

	err = tx.Broadcast(t.conf.ApiUrl)
	if err != nil {
		return "", err
	}

	return tx.Id, nil
}

func (t *Tron) GetTransactionReceipt(ctx context.Context, txId string) (*TransactionReceipt, error) {
	tx, err := tron.GetTransactionById(t.conf.ApiUrl, txId)
	if err != nil {
		return nil, err
	}

	retMessage := tx.GetRetMessage()
	if retMessage == "" {
		return nil, nil
	}

	ti, err := tron.GetTransactionInfoById(t.conf.ApiUrl, txId)
	if err != nil {
		return nil, err
	}

	err = nil
	if retMessage != "SUCCESS" {
		err = status.Errorf(errcode.ErrBlockChain, "tron gettransactionbyid err: %s", retMessage)
	}

	return &TransactionReceipt{
		Result: retMessage == "SUCCESS",
		Fee:    ti.Receipt.EnergyFee.Add(ti.Receipt.NetFee),
	}, err
}

func (t *Tron) TokenBalanceOf(ctx context.Context, who string) (decimal.Decimal, error) {
	whoAddress, err := address.Base58ToAddress(who)
	if err != nil {
		return decimal.Zero, err
	}

	p := tron.NewParameter(1)
	p.Set(0, whoAddress.Bytes())

	result, err := tron.TriggerConstantContract(t.conf.ApiUrl, t.conf.Contract, "balanceOf(address)", p)
	if err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromBigInt(new(big.Int).SetBytes(result.Get(0)), -t.conf.Decimals), nil
}

func (t *Tron) TokenTransfer(ctx context.Context, key string, who string, amount decimal.Decimal) (string, error) {
	whoAddress, err := address.Base58ToAddress(who)
	if err != nil {
		return "", err
	}

	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return "", err
	}
	fromAddress := address.PubkeyToAddress(privateKey.PublicKey).String()

	p := tron.NewParameter(2)
	p.Set(0, whoAddress.Bytes())
	p.Set(1, amount.Mul(decimal.New(1, 6)).BigInt().Bytes())

	tx, err := tron.TriggerSmartContract(t.conf.ApiUrl, fromAddress, t.conf.Contract, "transfer(address,uint256)", p, t.conf.FeeLimit)
	if err != nil {
		return "", err
	}

	err = tx.Sign(privateKey)
	if err != nil {
		return "", err
	}

	err = tx.Broadcast(t.conf.ApiUrl)
	if err != nil {
		return "", err
	}

	return tx.Id, nil
}

type TronSubscription struct {
	conf          *config.BlockChainConf
	from          string
	to            string
	trxType       string
	lastTimestamp int64
	handler       func(*TransferEvent) error
}

func trxRecordToEvent(te *tron.TransactionRecord, conf *config.BlockChainConf) (*TransferEvent, error) {
	return &TransferEvent{
		Id:        te.TxId,
		Timestamp: te.BlockTimestamp,
		From:      te.From,
		To:        te.To,
		Amount:    te.Value.Div(decimal.New(1, conf.Decimals)),
	}, nil
}

func (ts *TronSubscription) filterTrc20Tx() error {
	records, err := tron.GetTrc20TransactionRecord(ts.conf.ApiUrl, ts.conf.Contract, ts.from, ts.to, ts.lastTimestamp, ts.trxType)
	if err != nil {
		return err
	}

	for _, rec := range records {
		event, err := trxRecordToEvent(rec, ts.conf)
		if err != nil {
			return err
		}
		ts.lastTimestamp = rec.BlockTimestamp

		if err := ts.handler(event); err != nil {
			return err
		}
	}
	return nil
}

func (ts *TronSubscription) Subscribe(ctx context.Context) error {
	if err := ts.filterTrc20Tx(); err != nil {
		logrus.WithFields(logrus.Fields{
			"from":           ts.from,
			"to":             ts.to,
			"trx_type":       ts.trxType,
			"last_timestamp": ts.lastTimestamp,
			"err":            err.Error(),
		}).Error("Tron subscribe failed")
	}

	t := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-t.C:
			if err := ts.filterTrc20Tx(); err != nil {
				logrus.WithFields(logrus.Fields{
					"from":           ts.from,
					"to":             ts.to,
					"trx_type":       ts.trxType,
					"last_timestamp": ts.lastTimestamp,
					"err":            err.Error(),
				}).Error("Tron subscribe failed")
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (ts *TronSubscription) Close() {
}

func (t *Tron) SubscribeTokenTransferEvent(ctx context.Context, from, to string, lastPos int64, f func(event *TransferEvent) error) (TransferEventSubscription, error) {
	return &TronSubscription{
		conf:          t.conf,
		from:          from,
		to:            to,
		trxType:       "Transfer",
		lastTimestamp: lastPos,
		handler:       f,
	}, nil
}
