package tron

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rickone/athena/errcode"
	"google.golang.org/grpc/status"
)

type Transaction struct {
	Id         string          `json:"txID"`
	Visible    bool            `json:"visible"`
	RawData    json.RawMessage `json:"raw_data"`
	RawDataHex string          `json:"raw_data_hex"`
	Signature  []string        `json:"signature"`
	Ret        []struct {
		ContractRet string `json:"contractRet"`
	} `json:"ret"`
}

type Result struct {
	Result  bool   `json:"result"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey) error {
	rawData, err := hex.DecodeString(tx.RawDataHex)
	if err != nil {
		return err
	}

	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)

	sign, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return err
	}

	tx.Signature = append(tx.Signature, hex.EncodeToString(sign))
	return nil
}

func (tx *Transaction) GetRetMessage() string {
	ret := ""
	for _, r := range tx.Ret {
		ret = r.ContractRet
	}
	return ret
}

func (cli *TronClient) CreateTransaction(ownerAddress string, toAddress string, amount *big.Int) (*Transaction, error) {
	tx := Transaction{}
	err := cli.httpPost("/wallet/createtransaction", map[string]interface{}{
		"owner_address": ownerAddress,
		"to_address":    toAddress,
		"amount":        amount,
		"visible":       true,
	}, &tx)
	return &tx, err
}

func (cli *TronClient) Broadcast(tx *Transaction) error {
	result := Result{}
	err := cli.httpPost("/wallet/broadcasttransaction", tx, &result)
	if err != nil {
		return err
	}

	if !result.Result {
		return status.Errorf(errcode.ErrChainFailed, "Broadcast err: code=%s msg=%s", result.Code, result.Message)
	}
	return nil
}

func (cli *TronClient) GetTransactionById(id string) (*Transaction, error) {
	tx := Transaction{}
	err := cli.httpPost("/wallet/gettransactionbyid", map[string]interface{}{
		"value": id,
	}, &tx)
	return &tx, err
}
