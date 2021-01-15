package tron

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/rickone/athena/chain"
	"github.com/rickone/athena/errcode"
	"google.golang.org/grpc/status"
)

type Transaction struct {
	Id          string          `json:"txID"`
	Visible     bool            `json:"visible"`
	RawData     json.RawMessage `json:"raw_data"`
	RawDataDesc RawDataDesc     `json:"-"`
	RawDataHex  string          `json:"raw_data_hex"`
	Signature   []string        `json:"signature"`
	Ret         []struct {
		ContractRet string `json:"contractRet"`
	} `json:"ret"`
}

type RawDataDesc struct {
	Contract []*struct {
		Parameter struct {
			Value struct {
				OwnerAddress    string   `json:"owner_address"`
				ToAddress       string   `json:"to_address"`
				ContractAddress string   `json:"contract_address"`
				Amount          *big.Int `json:"amount"`
				Data            string   `json:"data"`
			} `json:"value"`
		} `json:"parameter"`
		Type string `json:"type"`
	} `json:"contract"`
	FeeLimit  *big.Int `json:"fee_limit"`
	Timestamp int64    `json:"timestamp"`
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

func (tx *Transaction) GetFrom() string {
	for _, p := range tx.RawDataDesc.Contract {
		return p.Parameter.Value.OwnerAddress
	}
	return ""
}

func (tx *Transaction) GetTo() string {
	for _, p := range tx.RawDataDesc.Contract {
		switch p.Type {
		case "TransferContract":
			return p.Parameter.Value.ToAddress
		case "TriggerSmartContract":
			return p.Parameter.Value.ContractAddress
		}
	}
	return ""
}

func (tx *Transaction) GetValue() *big.Int {
	for _, p := range tx.RawDataDesc.Contract {
		switch p.Type {
		case "TransferContract":
			return p.Parameter.Value.Amount
		}
	}
	return big.NewInt(0)
}

func (tx *Transaction) GetData() *chain.Data {
	for _, p := range tx.RawDataDesc.Contract {
		switch p.Type {
		case "TriggerSmartContract":
			return getData(common.FromHex(p.Parameter.Value.Data))
		}
	}
	return nil
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
		"value":   id,
		"visible": true,
	}, &tx)
	return &tx, err
}

func getData(bytes []byte) *chain.Data {
	if len(bytes) != 4+32+32 {
		return nil
	}

	var addr address.Address
	addr = append(addr, byte(0x41))
	addr = append(addr, bytes[4+12:36]...)

	return &chain.Data{
		Address: addr.String(),
		Value:   big.NewInt(0).SetBytes(bytes[36:]),
	}
}
