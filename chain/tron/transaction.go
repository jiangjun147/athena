package tron

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rickone/athena/errcode"
	"github.com/rickone/athena/jsonex"
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

func (tx *Transaction) Broadcast(apiUrl string) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	r, err := http.Post(fmt.Sprintf("%s/wallet/broadcasttransaction", apiUrl), "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	result := Result{}
	if err := jsonex.UnmarshalFromReader(r.Body, &result); err != nil {
		return err
	}

	if !result.Result {
		return status.Errorf(errcode.ErrBlockChain, "Broadcast err: code=%s msg=%s", result.Code, result.Message)
	}

	return nil
}

func (tx *Transaction) GetRetMessage() string {
	ret := ""
	for _, r := range tx.Ret {
		ret = r.ContractRet
	}
	return ret
}

func CreateTransaction(apiUrl string, ownerAddress string, toAddress string, amount *big.Int) (*Transaction, error) {
	// https://cn.developers.tron.network/reference#%E5%88%9B%E5%BB%BA%E4%BA%A4%E6%98%93

	data, err := json.Marshal(map[string]interface{}{
		"owner_address": ownerAddress,
		"to_address":    toAddress,
		"amount":        amount,
		"visible":       true,
	})
	if err != nil {
		return nil, err
	}

	r, err := http.Post(fmt.Sprintf("%s/wallet/createtransaction", apiUrl), "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	tx := Transaction{}
	err = jsonex.UnmarshalFromReader(r.Body, &tx)
	return &tx, err
}

func GetTransactionById(apiUrl string, id string) (*Transaction, error) {
	// https://cn.developers.tron.network/reference#%E6%8C%89-id-%E4%BA%A4%E6%98%93

	data, err := json.Marshal(map[string]interface{}{
		"value": id,
	})
	if err != nil {
		return nil, err
	}

	r, err := http.Post(fmt.Sprintf("%s/wallet/gettransactionbyid", apiUrl), "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	tx := Transaction{}
	err = jsonex.UnmarshalFromReader(r.Body, &tx)
	return &tx, err
}
