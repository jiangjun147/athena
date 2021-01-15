package eth

import (
	"context"
	"encoding/hex"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rickone/athena/chain"
)

var (
	transferEventHash = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
)

func (cli *EthClient) CreateKey() (*chain.Key, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	privateKey := hex.EncodeToString(crypto.FromECDSA(key))
	address := crypto.PubkeyToAddress(key.PublicKey).Hex()

	return &chain.Key{
		Address:    address,
		PrivateKey: privateKey,
	}, nil
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

func (cli *EthClient) TokenTransfer(ctx context.Context, token *chain.Token, fromPrivKey string, toAddress string, amount *big.Int) (*chain.Transaction, error) {
	key, err := crypto.HexToECDSA(fromPrivKey)
	if err != nil {
		return nil, err
	}

	usdt, err := NewUSDT(common.HexToAddress(token.Address), cli)
	if err != nil {
		return nil, err
	}

	gasPrice, err := cli.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	option := bind.NewKeyedTransactor(key)
	option.GasPrice = gasPrice
	option.GasLimit = cli.gasLimit

	tx, err := usdt.Transfer(option, common.HexToAddress(toAddress), amount)
	if err != nil {
		return nil, err
	}

	return cli.GetTransaction(ctx, tx.Hash().Hex())
}

func (cli *EthClient) GetTransaction(ctx context.Context, hash string) (*chain.Transaction, error) {
	txHash := common.HexToHash(hash)

	var tx rpcTransaction
	err := cli.raw.CallContext(ctx, &tx, "eth_getTransactionByHash", txHash)
	if err != nil {
		return nil, err
	}
	if tx.Hash == "" {
		return nil, nil
	}

	trx := &chain.Transaction{
		Hash:  tx.Hash,
		From:  tx.From.Hex(),
		To:    tx.To.Hex(),
		Value: big.NewInt(0).SetBytes(common.FromHex(tx.Value)),
		Data:  getData(common.FromHex(tx.Input)),
	}

	if tx.BlockNumber != nil {
		trx.Block = big.NewInt(0).SetBytes(common.FromHex(*tx.BlockNumber)).Int64()

		receipt, err := cli.TransactionReceipt(ctx, txHash)
		if err != nil {
			return nil, err
		}

		gasPrice := big.NewInt(0).SetBytes(common.FromHex(tx.GasPrice))
		gasUsed := big.NewInt(int64(receipt.GasUsed))
		trx.Fee = gasUsed.Mul(gasUsed, gasPrice)

		if receipt.Status != 0 {
			trx.Status = true
		}
	}
	return trx, nil
}

func (cli *EthClient) FindTokenTransaction(ctx context.Context, token *chain.Token, toAddress string, block int64) ([]*chain.Transaction, error) {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(token.Address)},
		Topics: [][]common.Hash{
			{transferEventHash},
			{},
			{common.HexToHash(toAddress)},
		},
		FromBlock: new(big.Int).SetInt64(block + 1),
	}

	logs, err := cli.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(logs) == 0 {
		return nil, nil
	}

	var result []*chain.Transaction
	for _, log := range logs {
		tx, err := cli.GetTransaction(ctx, log.TxHash.Hex())
		if err != nil {
			return nil, err
		}

		result = append(result, tx)
	}
	return result, nil
}

func getData(bytes []byte) *chain.Data {
	if len(bytes) != 4+32+32 {
		return nil
	}

	return &chain.Data{
		Address: common.BytesToAddress(bytes[4:36]).Hex(),
		Value:   big.NewInt(0).SetBytes(bytes[36:]),
	}
}
