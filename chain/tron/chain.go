package tron

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/rickone/athena/chain"
)

func (cli *TronClient) CreateKey() (*chain.Key, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	privateKey := hex.EncodeToString(crypto.FromECDSA(key))
	addr := address.PubkeyToAddress(key.PublicKey).String()

	return &chain.Key{
		Address:    addr,
		PrivateKey: privateKey,
	}, nil
}

func (cli *TronClient) TokenTransfer(ctx context.Context, token *chain.Token, fromPrivKey string, toAddress string, amount *big.Int) (*chain.Transaction, error) {
	to, err := address.Base58ToAddress(toAddress)
	if err != nil {
		return nil, err
	}

	key, err := crypto.HexToECDSA(fromPrivKey)
	if err != nil {
		return nil, err
	}
	fromAddress := address.PubkeyToAddress(key.PublicKey).String()

	p := chain.NewParameter(2)
	p.Set(0, to.Bytes())
	p.Set(1, amount.Bytes())

	// TODO fee_limit

	tx, err := cli.TriggerSmartContract(fromAddress, token.Address, "transfer(address,uint256)", p, 10000)
	if err != nil {
		return nil, err
	}

	err = tx.Sign(key)
	if err != nil {
		return nil, err
	}

	err = cli.Broadcast(tx)
	if err != nil {
		return nil, err
	}

	return cli.GetTransaction(ctx, tx.Id)
}

func (cli *TronClient) GetTransaction(ctx context.Context, hash string) (*chain.Transaction, error) {
	tx, err := cli.GetTransactionById(hash)
	if err != nil {
		return nil, err
	}

	trx := &chain.Transaction{
		Hash:  hash,
		From:  tx.GetFrom(),
		To:    tx.GetTo(),
		Value: tx.GetValue(),
		Data:  tx.GetData(),
	}

	ret := tx.GetRetMessage()
	if ret == "" {
		return trx, nil
	}

	ti, err := cli.GetTransactionInfoById(hash)
	if err != nil {
		return nil, err
	}

	trx.Block = int64(ti.BlockTimestamp)
	trx.Fee = ti.Fee
	if ret == "SUCCESS" {
		trx.Status = true
	}

	return trx, nil
}

func (cli *TronClient) FindTokenTransaction(ctx context.Context, token *chain.Token, toAddress string, block int64) ([]*chain.Transaction, error) {
	records, err := cli.GetTrc20TransactionRecord(token.Address, "", toAddress, block+1000, "Transfer")
	if err != nil {
		return nil, err
	}

	var result []*chain.Transaction
	for _, rec := range records {
		trx, err := cli.GetTransaction(ctx, rec.TxId)
		if err != nil {
			return nil, err
		}
		result = append(result, trx)
	}
	return result, nil
}
