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

func (cli *TronClient) BalanceOf(ctx context.Context, b58Address string) (*big.Int, error) {
	acc, err := cli.GetAccount(b58Address)
	if err != nil {
		return nil, err
	}

	return acc.Balance, nil
}

func (cli *TronClient) Transfer(ctx context.Context, fromPrivKey string, toAddress string, amount *big.Int) (*chain.Transaction, error) {
	privateKey, err := crypto.HexToECDSA(fromPrivKey)
	if err != nil {
		return nil, err
	}
	fromAddress := address.PubkeyToAddress(privateKey.PublicKey).String()

	tx, err := cli.CreateTransaction(fromAddress, toAddress, amount)
	if err != nil {
		return nil, err
	}

	err = tx.Sign(privateKey)
	if err != nil {
		return nil, err
	}

	err = cli.Broadcast(tx)
	if err != nil {
		return nil, err
	}

	return cli.GetTransaction(ctx, tx.Id)
}

func (cli *TronClient) TokenBalanceOf(ctx context.Context, token *chain.Token, b58Address string) (*big.Int, error) {
	addr, err := address.Base58ToAddress(b58Address)
	if err != nil {
		return nil, err
	}

	p := NewParameter(1)
	p.Set(0, addr.Bytes())

	result, err := cli.TriggerConstantContract(token.Address, "balanceOf(address)", p)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetBytes(result.Get(0)), nil
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

	p := NewParameter(2)
	p.Set(0, to.Bytes())
	p.Set(1, amount.Bytes())

	tx, err := cli.TriggerSmartContract(fromAddress, token.Address, "transfer(address,uint256)", p, cli.feeLimit)
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
